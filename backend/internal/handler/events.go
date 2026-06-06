package handler

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

func (a *API) streamScores(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	flusher, ok := w.(http.Flusher)
	if !ok {
		writeError(w, http.StatusInternalServerError, "streaming unsupported")
		return
	}

	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-r.Context().Done():
			return
		case <-ticker.C:
			matches, err := a.Matches.List(r.Context(), "", 50, 0)
			if err != nil {
				continue
			}
			payload, _ := json.Marshal(map[string]interface{}{"items": matches, "ts": time.Now().UTC()})
			fmt.Fprintf(w, "data: %s\n\n", payload)
			flusher.Flush()
		}
	}
}
