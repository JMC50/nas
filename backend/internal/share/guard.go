package share

import (
	"errors"
	"log/slog"
	"net/http"
	"sync"
	"sync/atomic"
	"time"
)

const (
	// 빨리 받는 연결은 쓰기 사이에서 이만큼 보낼 때마다, 또는 이 시간이 지날 때마다 다시 확인함
	recheckBytes = 4 << 20
	recheckEvery = 250 * time.Millisecond
	// 쓰기 하나가 막혀 있어도 이 간격으로 따로 다시 확인함
	watchEvery = time.Second
	// 받는 쪽이 이 시간 동안 한 번의 쓰기도 받아 가지 않으면 연결을 끊음
	writeIdleLimit = 30 * time.Second
)

var deadlineUnsupported sync.Once

// 응답 쓰기마다 쓰기 마감을 늘리고, 링크가 닫히면 막혀 있는 쓰기까지 바로 끝냄.
// checkedAt과 unchecked는 응답을 쓰는 고루틴만 만지고, 감시 고루틴과는 closed만 나눔
type streamGuard struct {
	http.ResponseWriter
	control   *http.ResponseController
	recheck   func() error
	closed    atomic.Bool
	checkedAt time.Time
	unchecked int64
}

func newStreamGuard(w http.ResponseWriter, recheck func() error) *streamGuard {
	return &streamGuard{ResponseWriter: w, control: http.NewResponseController(w), recheck: recheck, checkedAt: time.Now()}
}

func (g *streamGuard) Write(p []byte) (int, error) {
	if g.unchecked >= recheckBytes || time.Since(g.checkedAt) >= recheckEvery {
		if err := g.check(); err != nil {
			return 0, err
		}
	}
	g.setDeadline(time.Now().Add(writeIdleLimit))
	// interrupt가 마감을 지난 시각으로 바꾼 직후 위에서 다시 늘렸을 수 있으므로 한 번 더 봄
	if g.closed.Load() {
		return 0, errLinkClosed
	}
	n, err := g.ResponseWriter.Write(p)
	g.unchecked += int64(n)
	return n, err
}

// zip 항목마다 부르지만 마지막 확인에서 recheckEvery가 지났을 때만 데이터베이스를 다시 봄
func (g *streamGuard) checkpoint() error {
	if g.closed.Load() {
		return errLinkClosed
	}
	if time.Since(g.checkedAt) < recheckEvery {
		return nil
	}
	return g.check()
}

func (g *streamGuard) check() error {
	if g.closed.Load() {
		return errLinkClosed
	}
	g.checkedAt = time.Now()
	g.unchecked = 0
	return g.recheck()
}

// 느리게 받는 연결은 커널 버퍼가 비워질 때까지 쓰기 하나가 몇 분씩 막힐 수 있어, 쓰기와 따로 시간으로 다시 확인함.
// 돌려준 함수는 감시를 멈추고 감시 고루틴이 끝날 때까지 기다림
func (g *streamGuard) watch() (stop func()) {
	done := make(chan struct{})
	exited := make(chan struct{})
	go func() {
		defer close(exited)
		ticker := time.NewTicker(watchEvery)
		defer ticker.Stop()
		for {
			select {
			case <-done:
				return
			case <-ticker.C:
				if err := g.recheck(); err != nil {
					g.interrupt()
					return
				}
			}
		}
	}()
	return func() {
		close(done)
		<-exited
	}
}

// 쓰기 마감을 지난 시각으로 옮기면 막혀 있던 쓰기도 곧바로 오류로 끝남
func (g *streamGuard) interrupt() {
	g.closed.Store(true)
	g.setDeadline(time.Now())
}

// 정상으로 끝났으면 net/http가 처리기 밖에서 마감 없이 쓰지 않도록 남은 응답을 먼저 내보내고, 연결을 다시 쓸 때 마감이 남지 않게 지움.
// 링크가 닫혀 끝났으면 남은 쓰기도 바로 실패하도록 마감을 지난 시각으로 둠
func (g *streamGuard) finish() {
	if g.closed.Load() {
		g.setDeadline(time.Now())
		return
	}
	_ = g.control.Flush()
	g.setDeadline(time.Time{})
}

// 연결이 끊긴 뒤의 오류는 무시함. 마감을 아예 지원하지 않으면 느린 연결을 끊지 못하므로 한 번 알림
func (g *streamGuard) setDeadline(at time.Time) {
	if err := g.control.SetWriteDeadline(at); errors.Is(err, http.ErrNotSupported) {
		deadlineUnsupported.Do(func() {
			slog.Warn("share download: response writer has no write deadline; slow readers are not disconnected")
		})
	}
}
