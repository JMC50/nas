package share

import (
	"net/http"
	"sync"
	"time"
)

const (
	maxZipStreams        = 3
	maxStreamsPerLink    = 4
	maxZipStreamsPerLink = 1
	// 로그인한 사용자의 zip은 공개 링크와 자리를 따로 씀
	maxBundleStreams        = 3
	maxBundleStreamsPerUser = 2
	recordInterval          = time.Minute
	retryAfterSeconds       = "30"
)

// zip은 폴더 순회와 압축이 무거워 서버 전체에서 동시에 zipSlots 크기만큼만 보냄.
// 링크나 사용자 하나가 그 자리를 다 차지하지 못하게 키마다 zipPerKey개만 받음
type limiter struct {
	zipSlots  chan struct{}
	zipPerKey int
	mu        sync.Mutex
	perLink   map[int64]int
	recorded  map[int64]time.Time
}

func newLimiter(zipStreams, zipStreamsPerKey int) *limiter {
	return &limiter{
		zipSlots:  make(chan struct{}, zipStreams),
		zipPerKey: zipStreamsPerKey,
		perLink:   map[int64]int{},
		recorded:  map[int64]time.Time{},
	}
}

// zipBusy는 자리를 잡지 않고 지금 zip을 받을 수 없는 이유만 HTTP 상태 코드로 돌려줌. 받을 수 있으면 0
func (l *limiter) zipBusy(key int64) int {
	l.mu.Lock()
	defer l.mu.Unlock()
	if l.perLink[key] >= l.zipPerKey {
		return http.StatusTooManyRequests
	}
	if len(l.zipSlots) == cap(l.zipSlots) {
		return http.StatusServiceUnavailable
	}
	return 0
}

// 자리가 없으면 0이 아닌 HTTP 상태 코드를 돌려줌. 자리를 얻었으면 끝날 때 release를 불러야 함
func (l *limiter) acquire(linkID int64, folder bool) (release func(), status int) {
	limit := maxStreamsPerLink
	if folder {
		limit = l.zipPerKey
	}
	l.mu.Lock()
	if l.perLink[linkID] >= limit {
		l.mu.Unlock()
		return nil, http.StatusTooManyRequests
	}
	l.perLink[linkID]++
	l.mu.Unlock()
	if !folder {
		return func() { l.leave(linkID) }, 0
	}
	select {
	case l.zipSlots <- struct{}{}:
		return func() { <-l.zipSlots; l.leave(linkID) }, 0
	default:
		l.leave(linkID)
		return nil, http.StatusServiceUnavailable
	}
}

func (l *limiter) leave(linkID int64) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.perLink[linkID]--
	if l.perLink[linkID] <= 0 {
		delete(l.perLink, linkID)
	}
}

// 같은 링크의 받은 수와 활동 기록은 recordInterval에 한 번만 씀. 이어받기와 되풀이 요청으로 수가 부풀지 않게 함
func (l *limiter) shouldRecord(linkID int64, now time.Time) bool {
	l.mu.Lock()
	defer l.mu.Unlock()
	if last, ok := l.recorded[linkID]; ok && now.Sub(last) < recordInterval {
		return false
	}
	for id, at := range l.recorded {
		if now.Sub(at) >= recordInterval {
			delete(l.recorded, id)
		}
	}
	l.recorded[linkID] = now
	return true
}
