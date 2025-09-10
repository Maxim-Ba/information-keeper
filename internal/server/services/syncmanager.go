package services

import (
	"sync"
	"time"

	"github.com/Maxim-Ba/information-keeper/pkg/proto"
)

// SyncManager управляет real-time синхронизацией
// TODO in process
type SyncManager struct {
	mu              sync.RWMutex
	subscribers     map[string]map[string]chan *proto.SyncEvent // userID -> clientID -> channel
	eventHistory    map[string][]*proto.SyncEvent               // userID -> events history
	historyLimit    int
	retentionPeriod time.Duration
}

// Broadcast implements SyncManagerInterface.
func (s *SyncManager) Broadcast(userID string, event proto.SyncEvent) {
	panic("unimplemented")
}

// Subscribe implements SyncManagerInterface.
func (s *SyncManager) Subscribe(userID string, clientID string) chan proto.SyncEvent {
	panic("unimplemented")
}

// Unsubscribe implements SyncManagerInterface.
func (s *SyncManager) Unsubscribe(userID string, clientID string) {
	panic("unimplemented")
}

func NewSyncManager(historyLimit int, retentionPeriod time.Duration) *SyncManager {
	return &SyncManager{
		subscribers:     make(map[string]map[string]chan *proto.SyncEvent),
		eventHistory:    make(map[string][]*proto.SyncEvent),
		historyLimit:    historyLimit,
		retentionPeriod: retentionPeriod,
	}
}
