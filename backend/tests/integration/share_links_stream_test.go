package integration

import (
	"context"
	"crypto/rand"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

// 받는 쪽이 읽지 않으면 서버가 쓰기에서 멈추도록 소켓 버퍼보다 큰 파일을 씀. 1 MiB 무작위 블록을 되풀이함
func (f *shareFixture) writeBig(t *testing.T, rel string, mebibytes int) int64 {
	t.Helper()
	block := make([]byte, 1<<20)
	_, err := rand.Read(block)
	require.NoError(t, err)
	path := filepath.Join(f.dataDir, filepath.FromSlash(rel))
	require.NoError(t, os.MkdirAll(filepath.Dir(path), 0o755))
	file, err := os.Create(path)
	require.NoError(t, err)
	defer file.Close()
	for range mebibytes {
		_, err := file.Write(block)
		require.NoError(t, err)
	}
	return int64(mebibytes) << 20
}

func (f *shareFixture) serve(t *testing.T) *httptest.Server {
	t.Helper()
	srv := httptest.NewServer(f.router)
	t.Cleanup(srv.Close)
	return srv
}

// 요청마다 따로 연결을 써서 본문을 열어 둔 요청이 다음 요청을 막지 않게 함.
// 시험이 중간에 실패해도 서버를 닫기 전에 본문을 닫아 처리기가 쓰기에서 기다리지 않게 함
func openStream(t *testing.T, srv *httptest.Server, path string) *http.Response {
	t.Helper()
	client := &http.Client{Transport: &http.Transport{DisableKeepAlives: true}}
	resp, err := client.Get(srv.URL + path)
	require.NoError(t, err)
	t.Cleanup(func() { resp.Body.Close() })
	return resp
}

func (f *shareFixture) revoke(t *testing.T, user string, id int64) {
	t.Helper()
	require.Equal(t, http.StatusOK, f.do("DELETE", fmt.Sprintf("/server/share-links/%d", id), user, nil, nil).Code)
}

func TestShareLink_RevokeStopsARunningFileDownload(t *testing.T) {
	f := newShareFixture(t)
	size := f.writeBig(t, "big.bin", 48)
	link := f.create(t, "alice", "/", "big.bin", "1d")
	srv := f.serve(t)

	resp := openStream(t, srv, link.Path)
	defer resp.Body.Close()
	require.Equal(t, http.StatusOK, resp.StatusCode)
	_, err := io.CopyN(io.Discard, resp.Body, 1024)
	require.NoError(t, err)

	f.revoke(t, "alice", link.ID)
	rest, err := io.Copy(io.Discard, resp.Body)
	require.Error(t, err, "the response must end early instead of completing")
	require.Less(t, rest+1024, size)
}

func TestShareLink_RevokeStopsARunningFolderDownload(t *testing.T) {
	f := newShareFixture(t)
	var size int64
	for part := range 4 {
		size += f.writeBig(t, fmt.Sprintf("big/part%d.mp4", part), 12)
	}
	link := f.create(t, "alice", "/", "big", "1d")
	srv := f.serve(t)

	resp := openStream(t, srv, link.Path)
	defer resp.Body.Close()
	require.Equal(t, http.StatusOK, resp.StatusCode)
	_, err := io.CopyN(io.Discard, resp.Body, 1024)
	require.NoError(t, err)

	f.revoke(t, "alice", link.ID)
	rest, err := io.Copy(io.Discard, resp.Body)
	require.Error(t, err, "an unfinished zip must not end like a complete response")
	require.Less(t, rest+1024, size)
}

func TestShareLink_DownloadsAreRecordedOncePerMinute(t *testing.T) {
	f := newShareFixture(t)
	f.write(t, "a.txt", "abc")
	link := f.create(t, "alice", "/", "a.txt", "1d")

	for _, rangeHeader := range []string{"", "bytes=0-", "bytes=0-1", ""} {
		header := map[string]string{}
		if rangeHeader != "" {
			header["Range"] = rangeHeader
		}
		w := f.do("GET", link.Path, "", nil, header)
		require.Contains(t, []int{http.StatusOK, http.StatusPartialContent}, w.Code)
	}

	var count, logs int
	require.NoError(t, f.conn.QueryRow(`SELECT download_count FROM share_links WHERE id = ?`, link.ID).Scan(&count))
	require.NoError(t, f.conn.QueryRow(`SELECT COUNT(*) FROM log WHERE activity = 'DOWNLOAD'`).Scan(&logs))
	require.Equal(t, 1, count)
	require.Equal(t, 1, logs)
}

func requireSlotFreed(t *testing.T, srv *httptest.Server, path string) {
	t.Helper()
	require.Eventually(t, func() bool {
		resp := openStream(t, srv, path)
		defer resp.Body.Close()
		return resp.StatusCode == http.StatusOK
	}, 10*time.Second, 100*time.Millisecond)
}

func TestShareLink_OneLinkServesAtMostFourStreams(t *testing.T) {
	f := newShareFixture(t)
	f.writeBig(t, "big.bin", 48)
	link := f.create(t, "alice", "/", "big.bin", "1d")
	srv := f.serve(t)

	var running []*http.Response
	for range 4 {
		resp := openStream(t, srv, link.Path)
		require.Equal(t, http.StatusOK, resp.StatusCode)
		running = append(running, resp)
	}
	extra := openStream(t, srv, link.Path)
	extra.Body.Close()
	require.Equal(t, http.StatusTooManyRequests, extra.StatusCode)
	require.Equal(t, "30", extra.Header.Get("Retry-After"))

	for _, resp := range running {
		resp.Body.Close()
	}
	requireSlotFreed(t, srv, link.Path)
}

func TestShareLink_ZipStreamsShareAServerWideLimit(t *testing.T) {
	f := newShareFixture(t)
	for part := range 4 {
		f.writeBig(t, fmt.Sprintf("big/part%d.mp4", part), 12)
	}
	var links []createdLink
	for range 4 {
		links = append(links, f.create(t, "alice", "/", "big", "1d"))
	}
	srv := f.serve(t)

	var running []*http.Response
	for _, link := range links[:3] {
		resp := openStream(t, srv, link.Path)
		require.Equal(t, http.StatusOK, resp.StatusCode)
		running = append(running, resp)
	}
	extra := openStream(t, srv, links[3].Path)
	extra.Body.Close()
	require.Equal(t, http.StatusServiceUnavailable, extra.StatusCode)

	for _, resp := range running {
		resp.Body.Close()
	}
	requireSlotFreed(t, srv, links[3].Path)
}

func TestShareLink_AFolderLinkSendsOneZipAtATime(t *testing.T) {
	f := newShareFixture(t)
	for part := range 4 {
		f.writeBig(t, fmt.Sprintf("big/part%d.mp4", part), 12)
	}
	link := f.create(t, "alice", "/", "big", "1d")
	srv := f.serve(t)

	first := openStream(t, srv, link.Path)
	require.Equal(t, http.StatusOK, first.StatusCode)
	second := openStream(t, srv, link.Path)
	second.Body.Close()
	require.Equal(t, http.StatusTooManyRequests, second.StatusCode)

	first.Body.Close()
	requireSlotFreed(t, srv, link.Path)
}

// 받는 버퍼를 16 KiB로 줄이고 초당 약 200 KiB로 천천히 읽음. 서버가 보낸 바이트 수로만 다시 확인하면 끊은 링크의 자리를 20초 넘게 붙잡음
func openSlowStream(t *testing.T, srv *httptest.Server, path string) {
	t.Helper()
	dialer := &net.Dialer{}
	transport := &http.Transport{
		DisableKeepAlives: true,
		DialContext: func(ctx context.Context, network, addr string) (net.Conn, error) {
			conn, err := dialer.DialContext(ctx, network, addr)
			if err != nil {
				return nil, err
			}
			if err := conn.(*net.TCPConn).SetReadBuffer(16 << 10); err != nil {
				conn.Close()
				return nil, err
			}
			return conn, nil
		},
	}
	resp, err := (&http.Client{Transport: transport}).Get(srv.URL + path)
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, resp.StatusCode)
	stop := make(chan struct{})
	done := make(chan struct{})
	go func() {
		defer close(done)
		buf := make([]byte, 4<<10)
		for {
			select {
			case <-stop:
				return
			case <-time.After(20 * time.Millisecond):
			}
			if _, err := resp.Body.Read(buf); err != nil {
				return
			}
		}
	}()
	t.Cleanup(func() {
		close(stop)
		resp.Body.Close()
		<-done
	})
}

func TestShareLink_RevokedSlowZipFreesItsSlotQuickly(t *testing.T) {
	f := newShareFixture(t)
	for part := range 4 {
		f.writeBig(t, fmt.Sprintf("big/part%d.mp4", part), 12)
	}
	var links []createdLink
	for range 4 {
		links = append(links, f.create(t, "alice", "/", "big", "1d"))
	}
	srv := f.serve(t)

	for _, link := range links[:3] {
		openSlowStream(t, srv, link.Path)
	}
	busy := openStream(t, srv, links[3].Path)
	busy.Body.Close()
	require.Equal(t, http.StatusServiceUnavailable, busy.StatusCode)

	f.revoke(t, "alice", links[0].ID)
	require.Eventually(t, func() bool {
		resp := openStream(t, srv, links[3].Path)
		defer resp.Body.Close()
		return resp.StatusCode == http.StatusOK
	}, 5*time.Second, 100*time.Millisecond, "the revoked slow stream must give its zip slot back within seconds")
}
