// Package events poskytuje in-memory pub/sub a SSE handler pro živý stream událostí aukcí.
package events

import (
	"sync"

	"github.com/ol1n/auction-sim/internal/auction"
)

// Bus je in-memory event bus s odběrem per auction ID.
type Bus struct {
	mu   sync.RWMutex
	subs map[string]map[chan auction.Event]struct{}
}

func NewBus() *Bus {
	return &Bus{subs: map[string]map[chan auction.Event]struct{}{}}
}

// Subscribe vrátí kanál událostí pro daný auction ID a odhlašovací funkci.
func (b *Bus) Subscribe(auctionID string) (<-chan auction.Event, func()) {
	ch := make(chan auction.Event, 64)
	b.mu.Lock()
	if b.subs[auctionID] == nil {
		b.subs[auctionID] = map[chan auction.Event]struct{}{}
	}
	b.subs[auctionID][ch] = struct{}{}
	b.mu.Unlock()

	return ch, func() {
		b.mu.Lock()
		if m := b.subs[auctionID]; m != nil {
			delete(m, ch)
			close(ch)
		}
		b.mu.Unlock()
	}
}

// Publish rozešle událost všem odběratelům daného auction ID (neblokující).
func (b *Bus) Publish(e auction.Event) { b.PublishTo(e.AuctionID, e) }

// PublishTo rozešle událost odběratelům libovolného klíče (např. simulation ID).
func (b *Bus) PublishTo(key string, e auction.Event) {
	b.mu.RLock()
	defer b.mu.RUnlock()
	for ch := range b.subs[key] {
		select {
		case ch <- e:
		default: // pomalý odběratel se přeskočí, ať engine neblokuje
		}
	}
}
