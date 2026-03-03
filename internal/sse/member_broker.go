package sse

import (
	"sync"
)

type MemberEvent struct {
	Matricula string `json:"matricula"`
	Status    string `json:"status"`
}

var (
	memberSubscribers = make(map[string]chan MemberEvent)
	memberMu          sync.RWMutex
)

func SubscribeMember(teamCode string) chan MemberEvent {
	ch := make(chan MemberEvent, 10)
	memberMu.Lock()
	memberSubscribers[teamCode] = ch
	memberMu.Unlock()
	return ch
}

func UnsubscribeMember(teamCode string) {
	memberMu.Lock()
	delete(memberSubscribers, teamCode)
	memberMu.Unlock()
}

func NotifyMember(teamCode string, event MemberEvent) {
	memberMu.RLock()
	ch, ok := memberSubscribers[teamCode]
	memberMu.RUnlock()
	if ok {
		select {
		case ch <- event:
		default:
		}
	}
}

func NotifyMemberOccupied(teamCode, matricula string) {
	NotifyMember(teamCode, MemberEvent{Matricula: matricula, Status: "occupied"})
}

func NotifyMemberFree(teamCode, matricula string) {
	NotifyMember(teamCode, MemberEvent{Matricula: matricula, Status: "free"})
}
