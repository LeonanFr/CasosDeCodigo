package sse

import (
	"sync"
)

type Event struct {
	CaseID string `json:"case_id"`
	Status string `json:"status"`
}

var (
	subscribers = make(map[string]chan Event)
	mu          sync.RWMutex
)

func Subscribe(teamCode string) chan Event {
	ch := make(chan Event, 10)
	mu.Lock()
	subscribers[teamCode] = ch
	mu.Unlock()
	return ch
}

func Unsubscribe(teamCode string) {
	mu.Lock()
	delete(subscribers, teamCode)
	mu.Unlock()
}

func Notify(teamCode string, event Event) {
	mu.RLock()
	ch, ok := subscribers[teamCode]
	mu.RUnlock()
	if ok {
		select {
		case ch <- event:
		default:
		}
	}
}

func NotifyOccupied(teamCode, caseID string) {
	Notify(teamCode, Event{CaseID: caseID, Status: "occupied"})
}

func NotifyFree(teamCode, caseID string) {
	Notify(teamCode, Event{CaseID: caseID, Status: "free"})
}
