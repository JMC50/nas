package office

import "sync"

type Dedupe struct {
	flights sync.Map // map[string]chan struct{}
}

// Acquire returns (release, isLeader). The leader does the work and calls release() when finished.
// Followers wait on the returned channel; isLeader is false for them and release is nil.
func (d *Dedupe) Acquire(hash string) (release func(), isLeader bool) {
	ch := make(chan struct{})
	existing, loaded := d.flights.LoadOrStore(hash, ch)
	if loaded {
		<-existing.(chan struct{})
		return nil, false
	}
	return func() {
		close(ch)
		d.flights.Delete(hash)
	}, true
}
