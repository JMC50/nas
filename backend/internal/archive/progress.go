package archive

import (
	"sync"
	"time"
)

type Progress struct {
	Percent int    `json:"percent"`
	Status  string `json:"status"`
}

// Tracker stores in-memory progress entries keyed by upload/zip ID.
// Entries are auto-pruned after TTL (default 1 hour) to prevent unbounded growth.
type Tracker struct {
	entries sync.Map
	ttl     time.Duration
}

type entry struct {
	progress Progress
	created  time.Time
}

func NewTracker(ttl time.Duration) *Tracker {
	tracker := &Tracker{ttl: ttl}
	go tracker.pruneLoop()
	return tracker
}

func (t *Tracker) Set(id string, progress Progress) {
	t.entries.Store(id, &entry{progress: progress, created: time.Now()})
}

func (t *Tracker) Get(id string) (Progress, bool) {
	value, ok := t.entries.Load(id)
	if !ok {
		return Progress{}, false
	}
	return value.(*entry).progress, true
}

func (t *Tracker) Delete(id string) {
	t.entries.Delete(id)
}

func (t *Tracker) pruneLoop() {
	ticker := time.NewTicker(t.ttl / 4)
	defer ticker.Stop()
	for range ticker.C {
		now := time.Now()
		t.entries.Range(func(key, value any) bool {
			if now.Sub(value.(*entry).created) > t.ttl {
				t.entries.Delete(key)
			}
			return true
		})
	}
}
