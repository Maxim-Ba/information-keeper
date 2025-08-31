package services

import (
	"sync"
	"time"

	"github.com/Maxim-Ba/information-keeper/proto"
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

func NewSyncManager(historyLimit int, retentionPeriod time.Duration) *SyncManager {
	return &SyncManager{
		subscribers:     make(map[string]map[string]chan *proto.SyncEvent),
		eventHistory:    make(map[string][]*proto.SyncEvent),
		historyLimit:    historyLimit,
		retentionPeriod: retentionPeriod,
	}
}
