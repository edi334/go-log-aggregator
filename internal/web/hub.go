package web

import (
	"context"
)

type Hub struct {
	register   chan chan []byte
	unregister chan chan []byte
	broadcast  chan []byte
	clients    map[chan []byte]struct{}
}

func NewHub() *Hub {
	return &Hub{
		register:   make(chan chan []byte),
		unregister: make(chan chan []byte),
		broadcast:  make(chan []byte, 256),
		clients:    make(map[chan []byte]struct{}),
	}
}

func (h *Hub) Run(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			for client := range h.clients {
				close(client)
			}
			return
		case client := <-h.register:
			h.clients[client] = struct{}{}
		case client := <-h.unregister:
			if _, ok := h.clients[client]; ok {
				delete(h.clients, client)
				close(client)
			}
		case payload := <-h.broadcast:
			for client := range h.clients {
				select {
				case client <- payload:
				default:
				}
			}
		}
	}
}

func (h *Hub) Register(client chan []byte) {
	h.register <- client
}

func (h *Hub) Unregister(client chan []byte) {
	h.unregister <- client
}

func (h *Hub) Broadcast(payload []byte) {
	if payload == nil {
		return
	}
	select {
	case h.broadcast <- payload:
	default:
	}
}
