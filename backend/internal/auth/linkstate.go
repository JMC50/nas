package auth

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"sync"
	"time"
)

const (
	linkTTL    = 10 * time.Minute
	nonceBytes = 24
)

// LinkStore tracks short-lived nonces that mark an OAuth round-trip as a
// "link to current user" rather than a sign-in. Entries auto-expire after
// linkTTL. Lost on server restart — callers retry.
type LinkStore struct {
	mu     sync.Mutex
	states map[string]linkRecord
}

type linkRecord struct {
	userID    int64
	provider  string
	expiresAt time.Time
}

func NewLinkStore() *LinkStore {
	return &LinkStore{states: map[string]linkRecord{}}
}

func (s *LinkStore) Issue(userID int64, provider string) (string, error) {
	nonce, err := makeNonce()
	if err != nil {
		return "", err
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	s.purge()
	s.states[nonce] = linkRecord{userID: userID, provider: provider, expiresAt: time.Now().Add(linkTTL)}
	return nonce, nil
}

func (s *LinkStore) Consume(nonce string) (int64, string, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.purge()
	record, ok := s.states[nonce]
	if !ok {
		return 0, "", false
	}
	delete(s.states, nonce)
	return record.userID, record.provider, true
}

func (s *LinkStore) purge() {
	now := time.Now()
	for nonce, record := range s.states {
		if now.After(record.expiresAt) {
			delete(s.states, nonce)
		}
	}
}

func makeNonce() (string, error) {
	raw := make([]byte, nonceBytes)
	if _, err := rand.Read(raw); err != nil {
		return "", fmt.Errorf("nonce rand: %w", err)
	}
	return hex.EncodeToString(raw), nil
}
