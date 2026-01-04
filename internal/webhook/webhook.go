package webhook

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/adityaraj/agentflow/internal/config"
)

// Manager handles sending webhook notifications.
type Manager struct {
	hooks   []config.WebhookConfig
	client  *http.Client
	pending sync.WaitGroup
}

// NewManager creates a new webhook manager.
func NewManager(hooks []config.WebhookConfig) *Manager {
	return &Manager{
		hooks: hooks,
		client: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

// Send dispatches an event to all matching webhooks.
// Events are sent asynchronously and don't block execution.
func (m *Manager) Send(event Event) {
	if len(m.hooks) == 0 {
		return
	}

	for _, hook := range m.hooks {
		if hook.MatchesEvent(event.Type) {
			m.pending.Add(1)
			go m.post(hook, event)
		}
	}
}

// SendSync dispatches an event and waits for all requests to complete.
func (m *Manager) SendSync(event Event) error {
	if len(m.hooks) == 0 {
		return nil
	}

	var wg sync.WaitGroup
	errChan := make(chan error, len(m.hooks))

	for _, hook := range m.hooks {
		if hook.MatchesEvent(event.Type) {
			wg.Add(1)
			go func(h config.WebhookConfig) {
				defer wg.Done()
				if err := m.postSync(h, event); err != nil {
					errChan <- err
				}
			}(hook)
		}
	}

	wg.Wait()
	close(errChan)

	// Return first error, if any
	for err := range errChan {
		return err
	}
	return nil
}

// Wait blocks until all pending webhook requests complete.
func (m *Manager) Wait() {
	m.pending.Wait()
}

// post sends an event to a webhook asynchronously.
func (m *Manager) post(hook config.WebhookConfig, event Event) {
	defer m.pending.Done()
	_ = m.postSync(hook, event) // Ignore errors for async posts
}

// postSync sends an event to a webhook and returns any error.
func (m *Manager) postSync(hook config.WebhookConfig, event Event) error {
	payload, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("failed to marshal event: %w", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, "POST", hook.URL, bytes.NewReader(payload))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-Agent", "Cortex/1.0")

	// Add custom headers
	for key, value := range hook.Headers {
		req.Header.Set(key, value)
	}

	resp, err := m.client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send webhook: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		return fmt.Errorf("webhook returned status %d", resp.StatusCode)
	}

	return nil
}

// HasWebhooks returns true if there are any webhooks configured.
func (m *Manager) HasWebhooks() bool {
	return len(m.hooks) > 0
}

// Count returns the number of configured webhooks.
func (m *Manager) Count() int {
	return len(m.hooks)
}
