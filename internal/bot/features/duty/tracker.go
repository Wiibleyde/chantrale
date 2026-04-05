package duty

import "sync"

type dutyHistory struct {
	onDuty   map[string]struct{}
	onCall   map[string]struct{}
	offRadio map[string]struct{}
}

var (
	historyMap = make(map[string]*dutyHistory)
	historyMu  sync.Mutex
)

func ensureHistory(guildID string) *dutyHistory {
	if _, ok := historyMap[guildID]; !ok {
		historyMap[guildID] = &dutyHistory{
			onDuty:   make(map[string]struct{}),
			onCall:   make(map[string]struct{}),
			offRadio: make(map[string]struct{}),
		}
	}
	return historyMap[guildID]
}

func trackDuty(guildID, userID string) {
	historyMu.Lock()
	defer historyMu.Unlock()
	ensureHistory(guildID).onDuty[userID] = struct{}{}
}

func trackOnCall(guildID, userID string) {
	historyMu.Lock()
	defer historyMu.Unlock()
	ensureHistory(guildID).onCall[userID] = struct{}{}
}

func trackOffRadio(guildID, userID string) {
	historyMu.Lock()
	defer historyMu.Unlock()
	ensureHistory(guildID).offRadio[userID] = struct{}{}
}

func popHistory(guildID string) (onDuty, onCall, offRadio []string) {
	historyMu.Lock()
	defer historyMu.Unlock()
	h, ok := historyMap[guildID]
	if !ok {
		return
	}
	for id := range h.onDuty {
		onDuty = append(onDuty, id)
	}
	for id := range h.onCall {
		onCall = append(onCall, id)
	}
	for id := range h.offRadio {
		offRadio = append(offRadio, id)
	}
	delete(historyMap, guildID)
	return
}
