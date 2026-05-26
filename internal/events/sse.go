package events

import (
	"encoding/json"
	"fmt"
	"net/http"
)

// Handler streamuje události daného auction ID jako Server-Sent Events.
func (b *Bus) Handler(w http.ResponseWriter, r *http.Request, auctionID string) {
	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "streaming nepodporováno", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")

	ch, unsub := b.Subscribe(auctionID)
	defer unsub()

	for {
		select {
		case <-r.Context().Done():
			return
		case e, ok := <-ch:
			if !ok {
				return
			}
			data, _ := json.Marshal(e.Payload)
			fmt.Fprintf(w, "event: %s\ndata: %s\n\n", e.Type, data)
			flusher.Flush()
		}
	}
}
