package share

import (
	"errors"
	"sync"
	"time"
)

const (
	bundleTicketTTL  = 2 * time.Minute
	maxBundleTickets = 256
)

var errTicketsFull = errors.New("too many pending zip downloads")

// 브라우저가 GET으로 zip을 받을 수 있게, 검사를 마친 경로 목록을 잠깐 들고 있음. 서버를 다시 시작하면 사라짐
type bundleTicket struct {
	userID    int64
	items     []zipItem
	filename  string
	loc       string
	expiresAt time.Time
}

type ticketStore struct {
	mu      sync.Mutex
	tickets map[string]bundleTicket
}

func newTicketStore() *ticketStore {
	return &ticketStore{tickets: map[string]bundleTicket{}}
}

func (s *ticketStore) issue(ticket bundleTicket) (string, error) {
	id, err := newToken()
	if err != nil {
		return "", err
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	s.purge()
	if len(s.tickets) >= maxBundleTickets {
		return "", errTicketsFull
	}
	ticket.expiresAt = time.Now().Add(bundleTicketTTL)
	s.tickets[id] = ticket
	return id, nil
}

// 다운로드 관리자가 같은 주소를 다시 요청해도 받을 수 있게 만료 전에는 지우지 않음
func (s *ticketStore) find(id string) (bundleTicket, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.purge()
	ticket, ok := s.tickets[id]
	return ticket, ok
}

func (s *ticketStore) purge() {
	now := time.Now()
	for id, ticket := range s.tickets {
		if now.After(ticket.expiresAt) {
			delete(s.tickets, id)
		}
	}
}
