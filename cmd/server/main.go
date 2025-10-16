package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"
	"github.com/kevinxiao27/eg-walker/eg"
)

type Server struct {
	documents map[string]*eg.OpLog[rune]
	clients   map[string][]*websocket.Conn
	upgrader  websocket.Upgrader
}

type WSMessage struct {
	Type string      `json:"type"`
	Data interface{} `json:"data"`
}

type DocumentRequest struct {
	Agent string `json:"agent"`
	Pos   int    `json:"pos"`
	Text  string `json:"text,omitempty"`
	Len   int    `json:"len,omitempty"`
}

type DocumentResponse struct {
	Content string `json:"content"`
}

func NewServer() *Server {
	return &Server{
		documents: make(map[string]*eg.OpLog[rune]),
		clients:   make(map[string][]*websocket.Conn),
		upgrader: websocket.Upgrader{
			CheckOrigin: func(r *http.Request) bool { return true },
		},
	}
}

func (s *Server) getDocument(id string) *eg.OpLog[rune] {
	if doc, exists := s.documents[id]; exists {
		return doc
	}
	oplog := eg.NewOpLog[rune]()
	s.documents[id] = &oplog
	return &oplog
}

func (s *Server) handleInsert(w http.ResponseWriter, r *http.Request) {
	var req DocumentRequest
	json.NewDecoder(r.Body).Decode(&req)

	docID := r.URL.Query().Get("doc")
	oplog := s.getDocument(docID)

	log.Printf("INSERT: agent=%s pos=%d text=%s doc=%s", req.Agent, req.Pos, req.Text, docID)

	eg.LocalInsert(oplog, req.Agent, req.Pos, []rune(req.Text))

	content := eg.Checkout(*oplog)
	json.NewEncoder(w).Encode(DocumentResponse{Content: string(content)})
}

func (s *Server) handleDelete(w http.ResponseWriter, r *http.Request) {
	var req DocumentRequest
	json.NewDecoder(r.Body).Decode(&req)

	docID := r.URL.Query().Get("doc")
	oplog := s.getDocument(docID)

	log.Printf("DELETE: agent=%s pos=%d len=%d doc=%s", req.Agent, req.Pos, req.Len, docID)

	eg.LocalDelete(oplog, req.Agent, req.Pos, req.Len)

	content := eg.Checkout(*oplog)
	json.NewEncoder(w).Encode(DocumentResponse{Content: string(content)})
}

func (s *Server) handleGet(w http.ResponseWriter, r *http.Request) {
	docID := r.URL.Query().Get("doc")
	oplog := s.getDocument(docID)

	content := eg.Checkout(*oplog)
	json.NewEncoder(w).Encode(DocumentResponse{Content: string(content)})
}

func (s *Server) handleMerge(w http.ResponseWriter, r *http.Request) {
	// For now, just return success - merge will be implemented later
	json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
}

func (s *Server) broadcastToDocument(docID string, msg WSMessage) {
	if clients, exists := s.clients[docID]; exists {
		log.Printf("BROADCAST: sending %s to %d clients", msg.Type, len(clients))
		for _, conn := range clients {
			conn.WriteJSON(msg)
		}
	}
}

func (s *Server) handleWebSocket(w http.ResponseWriter, r *http.Request) {
	conn, _ := s.upgrader.Upgrade(w, r, nil)
	defer conn.Close()

	docID := r.URL.Query().Get("doc")
	s.clients[docID] = append(s.clients[docID], conn)

	log.Printf("CLIENT CONNECTED: doc=%s total=%d", docID, len(s.clients[docID]))

	// Send current document state
	oplog := s.getDocument(docID)
	content := eg.Checkout(*oplog)
	conn.WriteJSON(WSMessage{
		Type: "init",
		Data: DocumentResponse{Content: string(content)},
	})

	for {
		var msg WSMessage
		if err := conn.ReadJSON(&msg); err != nil {
			break
		}

		log.Printf("MESSAGE: type=%s", msg.Type)

		switch msg.Type {
		case "insert":
			var req DocumentRequest
			json.Unmarshal([]byte(msg.Data.(string)), &req)
			oplog := s.getDocument(docID)
			eg.LocalInsert(oplog, req.Agent, req.Pos, []rune(req.Text))
			s.broadcastToDocument(docID, msg)

		case "delete":
			var req DocumentRequest
			json.Unmarshal([]byte(msg.Data.(string)), &req)
			oplog := s.getDocument(docID)
			eg.LocalDelete(oplog, req.Agent, req.Pos, req.Len)
			s.broadcastToDocument(docID, msg)
		}
	}

	// Remove client
	for i, c := range s.clients[docID] {
		if c == conn {
			s.clients[docID] = append(s.clients[docID][:i], s.clients[docID][i+1:]...)
			break
		}
	}
	log.Printf("CLIENT DISCONNECTED: doc=%s remaining=%d", docID, len(s.clients[docID]))
}

func main() {
	server := NewServer()

	r := mux.NewRouter()
	r.HandleFunc("/ws", server.handleWebSocket)

	fmt.Println("API server starting on :8080")
	fmt.Println("WebSocket API: ws://localhost:8080/ws")
	log.Fatal(http.ListenAndServe(":8080", r))
}
