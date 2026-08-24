package ws

import (
	"encoding/json"
	"net/http"
	"sync"

	"github.com/alkaid/goapiscanner/internal/logger"
	"github.com/alkaid/goapiscanner/internal/model"
	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 4096,
	CheckOrigin:     func(r *http.Request) bool { return true },
}

type client struct {
	taskID string
	ch     chan []byte
	once   sync.Once
}

type Hub struct {
	mu      sync.Mutex
	clients map[*client]struct{}
}

func NewHub() *Hub {
	return &Hub{clients: make(map[*client]struct{})}
}

func (h *Hub) Broadcast(taskID string, ev model.StreamEvent) {
	b, err := json.Marshal(ev)
	if err != nil {
		return
	}
	h.mu.Lock()
	list := make([]*client, 0, len(h.clients))
	for c := range h.clients {
		if c.taskID == "" || c.taskID == taskID {
			list = append(list, c)
		}
	}
	h.mu.Unlock()
	for _, c := range list {
		select {
		case c.ch <- b:
		default:
		}
	}
}

func (h *Hub) ServeWS(w http.ResponseWriter, r *http.Request) {
	taskID := r.URL.Query().Get("task_id")
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		if logger.L() != nil {
			logger.L().Warn("ws upgrade failed", "err", err.Error())
		}
		return
	}
	c := &client{taskID: taskID, ch: make(chan []byte, 64)}
	h.mu.Lock()
	h.clients[c] = struct{}{}
	h.mu.Unlock()

	go func() {
		defer conn.Close()
		for {
			if _, _, err := conn.ReadMessage(); err != nil {
				h.drop(c)
				return
			}
		}
	}()
	go func() {
		defer conn.Close()
		for msg := range c.ch {
			if err := conn.WriteMessage(websocket.TextMessage, msg); err != nil {
				h.drop(c)
				return
			}
		}
	}()
}

func (h *Hub) drop(c *client) {
	c.once.Do(func() {
		h.mu.Lock()
		delete(h.clients, c)
		h.mu.Unlock()
		close(c.ch)
	})
}
