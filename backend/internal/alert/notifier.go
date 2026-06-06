package alert

import (
	"bytes"
	"encoding/json"
	"log"
	"net/http"
	"time"
)

type Notifier struct {
	webhookURL string
	client     *http.Client
}

func New(webhookURL string) *Notifier {
	return &Notifier{
		webhookURL: webhookURL,
		client:     &http.Client{Timeout: 5 * time.Second},
	}
}

func (n *Notifier) Send(event, message string) {
	log.Printf("ALERT [%s]: %s", event, message)
	if n.webhookURL == "" {
		return
	}
	body, _ := json.Marshal(map[string]string{"event": event, "message": message})
	resp, err := n.client.Post(n.webhookURL, "application/json", bytes.NewReader(body))
	if err != nil {
		log.Printf("webhook error: %v", err)
		return
	}
	_ = resp.Body.Close()
}
