package api

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"sync"
)

// SSEEvent represents a server-sent event
type SSEEvent struct {
	Type   string      `json:"type"`
	TaskID int         `json:"task_id,omitempty"`
	Task   interface{} `json:"task,omitempty"`
	UserID int         `json:"user_id"` // Used internally to route the event
}

// Client represents an active SSE connection
type Client struct {
	send   chan SSEEvent
	UserID int
	Role   string
}

// SSEHub maintains active clients and broadcasts events
type SSEHub struct {
	clients    map[*Client]bool
	broadcast  chan SSEEvent
	register   chan *Client
	unregister chan *Client
	mu         sync.RWMutex
}

// AppHub is the global instance of the SSE Hub
var AppHub = &SSEHub{
	broadcast:  make(chan SSEEvent),
	register:   make(chan *Client),
	unregister: make(chan *Client),
	clients:    make(map[*Client]bool),
}

// Run starts the Hub loop
func (h *SSEHub) Run() {
	for {
		select {
		case client := <-h.register:
			h.mu.Lock()
			h.clients[client] = true
			h.mu.Unlock()
			log.Printf("SSE Client registered: UserID %d", client.UserID)
		case client := <-h.unregister:
			h.mu.Lock()
			if _, ok := h.clients[client]; ok {
				delete(h.clients, client)
				close(client.send)
				log.Printf("SSE Client unregistered: UserID %d", client.UserID)
			}
			h.mu.Unlock()
		case event := <-h.broadcast:
			h.mu.RLock()
			for client := range h.clients {
				// Only send the event if the client is an admin, or the owner of the task
				if client.Role == "admin" || client.UserID == event.UserID {
					select {
					case client.send <- event:
					default:
						// If the client's send buffer is full, disconnect them
						close(client.send)
						delete(h.clients, client)
					}
				}
			}
			h.mu.RUnlock()
		}
	}
}

// ServeSSE handles incoming EventSource connections
func ServeSSE(w http.ResponseWriter, r *http.Request) {
	// The AuthMiddleware has already validated the token and set the user in context
	user := GetUserFromContext(r.Context())
	if user == nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Make sure the connection supports Flusher
	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "Streaming unsupported", http.StatusInternalServerError)
		return
	}

	// Set required headers for SSE
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	// CORS headers should be handled by the global CORS middleware, but just in case:
	w.Header().Set("Access-Control-Allow-Origin", "*")

	client := &Client{
		send:   make(chan SSEEvent, 256),
		UserID: user.UserID,
		Role:   user.Role,
	}

	AppHub.register <- client

	// Handle client disconnection
	notify := r.Context().Done()
	go func() {
		<-notify
		AppHub.unregister <- client
	}()

	// Send an initial connected event
	fmt.Fprintf(w, "event: connected\ndata: {\"status\":\"connected\"}\n\n")
	flusher.Flush()

	// Listen for events to send to the client
	for event := range client.send {
		data, err := json.Marshal(event)
		if err != nil {
			log.Printf("Error marshaling SSE event: %v", err)
			continue
		}
		fmt.Fprintf(w, "event: message\ndata: %s\n\n", data)
		flusher.Flush()
	}
}
