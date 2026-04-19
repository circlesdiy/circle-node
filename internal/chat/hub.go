package chat

import (
	"sync"

	"go.uber.org/zap"
)

// Event is a single fan-out unit delivered to an SSE subscriber.
// Name corresponds to the SSE `event:` field (e.g. "message", "inbox"),
// ID corresponds to `id:` (for Last-Event-ID replay) and Data carries the
// pre-rendered HTML fragment written to `data:`.
type Event struct {
	Name   string
	ID     string
	Data   string
	ChatID string
}

// subscriber represents a single SSE connection for a profile.
type subscriber struct {
	ch   chan Event
	done chan struct{}
}

// Hub is an in-process pub/sub keyed by profile ID. One Hub instance is
// sufficient for a single-node deployment; horizontal scaling would plug a
// Redis Pub/Sub (or NATS) fan-out behind this same interface.
type Hub struct {
	mu           sync.RWMutex
	subs         map[string]map[*subscriber]struct{}
	logger       *zap.Logger
	bufferSize   int
	maxPerProfile int
}

// NewHub constructs a Hub with sensible defaults. bufferSize bounds the
// per-subscriber queue so a stalled client cannot block the publisher;
// maxPerProfile caps concurrent connections from one profile (multi-tab).
func NewHub(logger *zap.Logger) *Hub {
	return &Hub{
		subs:          make(map[string]map[*subscriber]struct{}),
		logger:        logger,
		bufferSize:    32,
		maxPerProfile: 8,
	}
}

// Subscribe registers a listener for the given profile ID. It returns a
// receive-only channel and an unsubscribe function that must be invoked when
// the caller is done (typically via defer).
//
// Publishing is non-blocking: if the channel is full, the event is dropped
// and a warning is logged. Slow readers therefore cannot back-pressure the
// producer or starve other subscribers.
func (h *Hub) Subscribe(profileID string) (<-chan Event, func()) {
	sub := &subscriber{
		ch:   make(chan Event, h.bufferSize),
		done: make(chan struct{}),
	}

	h.mu.Lock()
	profileSubs, ok := h.subs[profileID]
	if !ok {
		profileSubs = make(map[*subscriber]struct{})
		h.subs[profileID] = profileSubs
	}

	if len(profileSubs) >= h.maxPerProfile {
		// Evict the oldest-ish subscriber deterministically by iterating
		// once (map order is randomized, so "oldest" is approximate but
		// ensures we bound memory).
		for existing := range profileSubs {
			delete(profileSubs, existing)
			close(existing.done)
			if h.logger != nil {
				h.logger.Warn("chat hub evicted subscriber (cap reached)",
					zap.String("profile_id", profileID),
					zap.Int("cap", h.maxPerProfile),
				)
			}
			break
		}
	}

	profileSubs[sub] = struct{}{}
	h.mu.Unlock()

	unsubscribe := func() {
		h.mu.Lock()
		if profileSubs, ok := h.subs[profileID]; ok {
			if _, present := profileSubs[sub]; present {
				delete(profileSubs, sub)
				close(sub.done)
				if len(profileSubs) == 0 {
					delete(h.subs, profileID)
				}
			}
		}
		h.mu.Unlock()
	}

	return sub.ch, unsubscribe
}

// Publish fans an event out to every live subscriber for the given profile.
// Delivery is best-effort: full channels drop the event.
func (h *Hub) Publish(profileID string, event Event) {
	h.mu.RLock()
	profileSubs, ok := h.subs[profileID]
	if !ok || len(profileSubs) == 0 {
		h.mu.RUnlock()
		return
	}

	// Copy the subscriber set under the read lock to minimize lock scope while
	// delivering. Deliveries happen outside the lock.
	targets := make([]*subscriber, 0, len(profileSubs))
	for s := range profileSubs {
		targets = append(targets, s)
	}
	h.mu.RUnlock()

	for _, s := range targets {
		select {
		case <-s.done:
		case s.ch <- event:
		default:
			if h.logger != nil {
				h.logger.Warn("chat hub dropped event for slow subscriber",
					zap.String("profile_id", profileID),
					zap.String("event", event.Name),
					zap.String("chat_id", event.ChatID),
				)
			}
		}
	}
}

// PublishAll publishes the same event to a set of profile IDs (e.g. every
// participant except the sender).
func (h *Hub) PublishAll(profileIDs []string, event Event) {
	for _, id := range profileIDs {
		h.Publish(id, event)
	}
}

// SubscriberCount returns the number of live subscriptions for a profile.
// Primarily for tests and diagnostics.
func (h *Hub) SubscriberCount(profileID string) int {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return len(h.subs[profileID])
}
