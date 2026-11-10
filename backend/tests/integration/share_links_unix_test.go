//go:build unix

package integration

import (
	"net/http"
	"os"
	"path/filepath"
	"sync"
	"syscall"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

// 내려받기가 정해진 시간 안에 끝나는지 봄. FIFO를 열다 멈추면 시험이 끝없이 기다리지 않고 실패함
func (f *shareFixture) getWithin(t *testing.T, path string, limit time.Duration) (int, []byte) {
	t.Helper()
	type result struct {
		code int
		body []byte
	}
	done := make(chan result, 1)
	go func() {
		w := f.do("GET", path, "", nil, nil)
		done <- result{w.Code, w.Body.Bytes()}
	}()
	select {
	case got := <-done:
		return got.code, got.body
	case <-time.After(limit):
		t.Fatalf("download of %s did not finish within %s", path, limit)
		return 0, nil
	}
}

// 시험 쪽에서만 쓰는 도우미임. stop을 닫을 때까지 두 이름을 rename으로 계속 맞바꿈
func swapUntil(stop <-chan struct{}, wg *sync.WaitGroup, path, other string) {
	wg.Add(1)
	go func() {
		defer wg.Done()
		parked := path + ".parked"
		for {
			select {
			case <-stop:
				return
			default:
			}
			_ = os.Rename(path, parked)
			_ = os.Rename(other, path)
			_ = os.Rename(path, other)
			_ = os.Rename(parked, path)
		}
	}()
}

func TestShareLink_ZipSkipsNamesWithBackslashes(t *testing.T) {
	f := newShareFixture(t)
	f.write(t, "top2/ok.txt", "ok")
	f.write(t, `top2/..\..\..\evil.txt`, "evil")
	f.write(t, `top2/a\b/c.txt`, "c")
	link := f.create(t, "alice", "/", "top2", "1d")

	w := f.do("GET", link.Path, "", nil, nil)
	require.Equal(t, http.StatusOK, w.Code)
	entries := readZip(t, w.Body.Bytes())
	require.Equal(t, "ok", entries["top2/ok.txt"])
	for entry, content := range entries {
		require.NotContains(t, entry, `\`)
		require.NotEqual(t, "evil", content)
		require.NotEqual(t, "c", content)
	}
}

func TestShareLink_FifoNeverBlocksADownload(t *testing.T) {
	f := newShareFixture(t)
	f.write(t, "pipe.txt", "regular at creation")
	fileLink := f.create(t, "alice", "/", "pipe.txt", "1d")
	pipe := filepath.Join(f.dataDir, "pipe.txt")
	require.NoError(t, os.Remove(pipe))
	require.NoError(t, syscall.Mkfifo(pipe, 0o644))

	code, _ := f.getWithin(t, fileLink.Path, 5*time.Second)
	require.Equal(t, http.StatusNotFound, code)

	f.write(t, "with-pipe/a.txt", "a")
	require.NoError(t, syscall.Mkfifo(filepath.Join(f.dataDir, "with-pipe", "pipe"), 0o644))
	folderLink := f.create(t, "alice", "/", "with-pipe", "1d")
	code, body := f.getWithin(t, folderLink.Path, 5*time.Second)
	require.Equal(t, http.StatusOK, code)
	entries := readZip(t, body)
	require.Equal(t, "a", entries["with-pipe/a.txt"])
	require.NotContains(t, entries, "with-pipe/pipe")
}

func TestShareLink_SwapsDuringDownloadNeverLeaveTheDataFolder(t *testing.T) {
	f := newShareFixture(t)
	secret := "outside secret bytes"
	require.NoError(t, os.WriteFile(filepath.Join(f.outside, "secret.mp4"), []byte(secret), 0o644))
	f.write(t, "top/a.txt", "a")
	f.write(t, "top/sub/inside.txt", "inside")
	f.write(t, "top/f.txt", "inside file")
	f.write(t, "top/p.txt", "regular")
	require.NoError(t, os.Symlink(f.outside, filepath.Join(f.dataDir, "top", "sub-escape")))
	require.NoError(t, os.Symlink(filepath.Join(f.outside, "secret.mp4"), filepath.Join(f.dataDir, "top", "f-escape")))
	require.NoError(t, syscall.Mkfifo(filepath.Join(f.dataDir, "top", "p-fifo"), 0o644))
	folderLink := f.create(t, "alice", "/", "top", "1d")
	fileLink := f.create(t, "alice", "/top", "f.txt", "1d")
	pipeLink := f.create(t, "alice", "/top", "p.txt", "1d")

	stop := make(chan struct{})
	var wg sync.WaitGroup
	top := filepath.Join(f.dataDir, "top")
	swapUntil(stop, &wg, filepath.Join(top, "sub"), filepath.Join(top, "sub-escape"))
	swapUntil(stop, &wg, filepath.Join(top, "f.txt"), filepath.Join(top, "f-escape"))
	swapUntil(stop, &wg, filepath.Join(top, "p.txt"), filepath.Join(top, "p-fifo"))
	defer func() {
		close(stop)
		wg.Wait()
	}()

	for range 150 {
		code, body := f.getWithin(t, folderLink.Path, 5*time.Second)
		if code == http.StatusOK {
			for entry, content := range readZip(t, body) {
				require.NotContains(t, entry, "secret")
				require.NotEqual(t, secret, content)
			}
		}
		code, body = f.getWithin(t, fileLink.Path, 5*time.Second)
		if code == http.StatusOK {
			require.Equal(t, "inside file", string(body))
		}
		f.getWithin(t, pipeLink.Path, 5*time.Second)
	}
}
