package services

import (
	"sync"
	"time"

	eventidgen "github.com/Maxim-Ba/information-keeper/pkg/event-id-gen"
	"github.com/Maxim-Ba/information-keeper/pkg/logger"
	"github.com/Maxim-Ba/information-keeper/pkg/proto"
)

// SyncManager управляет real-time синхронизацией.
type SyncManager struct {
	mu              sync.RWMutex
	subscribers     map[string]map[string]chan *proto.SyncEvent // userID -> clientID -> channel
	eventHistory    map[string][]*proto.SyncEvent               // userID -> events history
	historyLimit    int
	retentionPeriod time.Duration
}

// Broadcast отправляет событие всем подписчикам пользователя.
func (s *SyncManager) Broadcast(userID string, event *proto.SyncEvent) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	// Сохраняем в историю
	s.addToHistory(userID, event)
	logger.Info("Broadcasting event to %d subscribers for user %s\n", len(s.subscribers[userID]), userID)
	// Отправляем всем активным подписчикам
	if clients, ok := s.subscribers[userID]; ok {
		for clientID, ch := range clients {
			select {
			case ch <- event:
			default:
				// Канал заполнен, пропускаем клиента
				logger.Info("Channel full for client %s, skipping\n", clientID)
			}
		}
	}
}

// Subscribe добавляет нового подписчика.
func (s *SyncManager) Subscribe(userID string, clientID string) chan *proto.SyncEvent {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Инициализируем map для пользователя если нужно
	if _, exists := s.subscribers[userID]; !exists {
		s.subscribers[userID] = make(map[string]chan *proto.SyncEvent)
	}

	// Создаем канал для клиента
	ch := make(chan *proto.SyncEvent, 100) // буфер на 100 событий
	s.subscribers[userID][clientID] = ch

	// Отправляем событие о подключении нового клиента
	connectEvent := &proto.SyncEvent{
		EventId:   eventidgen.GenerateEventID(),
		Timestamp: time.Now().Unix(),
		EventType: &proto.SyncEvent_ClientConnected{
			ClientConnected: &proto.ClientConnectedEvent{
				ClientId:    clientID,
				ConnectTime: time.Now().Unix(),
			},
		},
	}

	// Рассылаем всем клиентам пользователя (кроме нового)
	for id, clientCh := range s.subscribers[userID] {
		if id != clientID {
			select {
			case clientCh <- connectEvent:
			default:
				// Пропускаем если канал заполнен
			}
		}
	}

	return ch
}

// Unsubscribe удаляет подписчика.
func (s *SyncManager) Unsubscribe(userID string, clientID string) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if clients, exists := s.subscribers[userID]; exists {
		if ch, exists := clients[clientID]; exists {
			close(ch)
			delete(clients, clientID)

			// Отправляем событие об отключении
			disconnectEvent := &proto.SyncEvent{
				EventId:   eventidgen.GenerateEventID(),
				Timestamp: time.Now().Unix(),
				EventType: &proto.SyncEvent_ClientDisconnected{
					ClientDisconnected: &proto.ClientDisconnectedEvent{
						ClientId:       clientID,
						DisconnectTime: time.Now().Unix(),
					},
				},
			}

			// Рассылаем остальным клиентам
			for id, clientCh := range clients {
				if id != clientID {
					select {
					case clientCh <- disconnectEvent:
					default:
						// Пропускаем если канал заполнен
					}
				}
			}

			// Удаляем map пользователя если не осталось клиентов
			if len(clients) == 0 {
				delete(s.subscribers, userID)
			}
		}
	}
}

// GetEventsSince возвращает события с указанного времени.
func (s *SyncManager) GetEventsSince(userID string, since int64) []*proto.SyncEvent {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var recentEvents []*proto.SyncEvent
	if history, exists := s.eventHistory[userID]; exists {
		for _, event := range history {
			if event.Timestamp > since {
				recentEvents = append(recentEvents, event)
			}
		}
	}
	return recentEvents
}

func (s *SyncManager) addToHistory(userID string, event *proto.SyncEvent) {
	if _, exists := s.eventHistory[userID]; !exists {
		s.eventHistory[userID] = make([]*proto.SyncEvent, 0, s.historyLimit)
	}

	history := s.eventHistory[userID]

	// Добавляем событие
	history = append(history, event)

	// Ограничиваем размер истории
	if len(history) > s.historyLimit {
		history = history[len(history)-s.historyLimit:]
	}

	// Очищаем устаревшие события
	var cleanedHistory []*proto.SyncEvent
	for _, e := range history {
		if time.Unix(e.Timestamp, 0).Add(s.retentionPeriod).After(time.Now()) {
			cleanedHistory = append(cleanedHistory, e)
		}
	}

	s.eventHistory[userID] = cleanedHistory
}

func NewSyncManager(historyLimit int, retentionPeriod time.Duration) *SyncManager {
	return &SyncManager{
		subscribers:     make(map[string]map[string]chan *proto.SyncEvent),
		eventHistory:    make(map[string][]*proto.SyncEvent),
		historyLimit:    historyLimit,
		retentionPeriod: retentionPeriod,
	}
}
