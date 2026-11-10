package share

import (
	"net/http"
	"time"
)

const (
	// 끊거나 만료된 링크는 늦어도 이만큼 더 보낸 뒤나 다음 zip 항목에서 멈춤
	recheckBytes = 4 << 20
	recheckEvery = 250 * time.Millisecond
	// 받는 쪽이 이 시간 동안 한 번의 쓰기도 받아 가지 않으면 연결을 끊음
	writeIdleLimit = 30 * time.Second
)

// 응답 쓰기마다 쓰기 마감을 늘리고, 보내는 동안 링크가 아직 열려 있는지 다시 확인함
type streamGuard struct {
	http.ResponseWriter
	control   *http.ResponseController
	recheck   func() error
	checkedAt time.Time
	unchecked int64
}

func newStreamGuard(w http.ResponseWriter, recheck func() error) *streamGuard {
	return &streamGuard{ResponseWriter: w, control: http.NewResponseController(w), recheck: recheck, checkedAt: time.Now()}
}

func (g *streamGuard) Write(p []byte) (int, error) {
	if g.unchecked >= recheckBytes {
		if err := g.check(); err != nil {
			return 0, err
		}
	}
	// 마감을 지원하지 않는 응답 작성기에서는 마감 없이 씀
	_ = g.control.SetWriteDeadline(time.Now().Add(writeIdleLimit))
	n, err := g.ResponseWriter.Write(p)
	g.unchecked += int64(n)
	return n, err
}

// zip 항목마다 부르지만 마지막 확인에서 recheckEvery가 지났을 때만 데이터베이스를 다시 봄
func (g *streamGuard) checkpoint() error {
	if time.Since(g.checkedAt) < recheckEvery {
		return nil
	}
	return g.check()
}

func (g *streamGuard) check() error {
	g.checkedAt = time.Now()
	g.unchecked = 0
	return g.recheck()
}

// 연결이 다음 요청에 다시 쓰일 때 이 응답의 마감이 남지 않게 지움
func (g *streamGuard) clearDeadline() {
	_ = g.control.SetWriteDeadline(time.Time{})
}
