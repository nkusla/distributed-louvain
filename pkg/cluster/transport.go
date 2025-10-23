package cluster

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/distributed-louvain/pkg/actor"
	"github.com/distributed-louvain/pkg/messages"
)

// MessageEnvelope wraps messages with metadata for HTTP transport
type MessageEnvelope struct {
	Type    string          `json:"type"`
	To      actor.PID       `json:"to"`
	Payload json.RawMessage `json:"payload"`
}

type Transport struct {
	machineID     string
	server        *http.Server
	system        *actor.ActorSystem
	client        *http.Client
	messageTypes  map[string]func() actor.Message
	mu            sync.RWMutex
	ctx           context.Context
	cancel        context.CancelFunc
}

func NewTransport(machineID string, port int) *Transport {
	transport := &Transport{
		machineID:    machineID,
		client:       &http.Client{Timeout: 30 * time.Second},
		messageTypes: make(map[string]func() actor.Message),
	}

	transport.registerMessageTypes()

	mux := http.NewServeMux()
	mux.HandleFunc("/message", transport.handleMessage)
	mux.HandleFunc("/health", transport.handleHealth)

	transport.server = &http.Server{
		Addr:    fmt.Sprintf(":%d", port),
		Handler: mux,
	}

	return transport
}

func (t *Transport) registerMessageTypes() {
	t.messageTypes["InitialPartitionCreation"] = func() actor.Message { return &messages.InitialPartitionCreation{} }
	t.messageTypes["InitialPartitionCreationComplete"] = func() actor.Message { return &messages.InitialPartitionCreationComplete{} }
	t.messageTypes["StartPhase1"] = func() actor.Message { return &messages.StartPhase1{} }
	t.messageTypes["DegreeRequest"] = func() actor.Message { return &messages.DegreeRequest{} }
	t.messageTypes["DegreeResponse"] = func() actor.Message { return &messages.DegreeResponse{} }
	t.messageTypes["LocalOptimizationComplete"] = func() actor.Message { return &messages.LocalOptimizationComplete{} }
	t.messageTypes["StartPhase2"] = func() actor.Message { return &messages.StartPhase2{} }
	t.messageTypes["EdgeAggregate"] = func() actor.Message { return &messages.EdgeAggregate{} }
	t.messageTypes["EdgeAggregateComplete"] = func() actor.Message { return &messages.EdgeAggregateComplete{} }
	t.messageTypes["AggregationResult"] = func() actor.Message { return &messages.AggregationResult{} }
	t.messageTypes["AggregationComplete"] = func() actor.Message { return &messages.AggregationComplete{} }
	t.messageTypes["CompleteAggregation"] = func() actor.Message { return &messages.CompleteAggregation{} }
	t.messageTypes["AlgorithmComplete"] = func() actor.Message { return &messages.AlgorithmComplete{} }
	t.messageTypes["Shutdown"] = func() actor.Message { return &messages.Shutdown{} }
}

func (t *Transport) SetActorSystem(system *actor.ActorSystem) {
	t.system = system
}

func (t *Transport) Start(ctx context.Context) error {
	t.ctx, t.cancel = context.WithCancel(ctx)

	log.Printf("[Transport] Starting HTTP server on %s", t.server.Addr)

	go func() {
		if err := t.server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Printf("[Transport] Server error: %v", err)
		}
	}()

	return nil
}

func (t *Transport) Send(to actor.PID, address string, msg actor.Message) error {
	if to.IsLocal(t.machineID) {
		return fmt.Errorf("transport should only handle remote messages")
	}

	payload, err := json.Marshal(msg)
	if err != nil {
		return fmt.Errorf("failed to marshal message: %w", err)
	}

	envelope := MessageEnvelope{
		Type:    msg.Type(),
		To:      to,
		Payload: payload,
	}

	data, err := json.Marshal(envelope)
	if err != nil {
		return fmt.Errorf("failed to marshal envelope: %w", err)
	}

	url := fmt.Sprintf("http://%s/message", address)
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(data))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := t.client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send message: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("HTTP error %d: %s", resp.StatusCode, string(body))
	}

	log.Printf("[Transport] Sent message %s to %s", msg.Type(), envelope.To)
	return nil
}

func (t *Transport) Stop() error {
	log.Printf("[Transport] Stopping HTTP server")

	if t.cancel != nil {
		t.cancel()
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	return t.server.Shutdown(ctx)
}

func (t *Transport) handleMessage(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Failed to read body", http.StatusBadRequest)
		return
	}

	var envelope MessageEnvelope
	if err := json.Unmarshal(body, &envelope); err != nil {
		http.Error(w, "Failed to parse envelope", http.StatusBadRequest)
		return
	}

	msg, err := t.decodeMessage(envelope.Type, envelope.Payload)
	if err != nil {
		log.Printf("[Transport] Failed to decode message: %v", err)
		http.Error(w, "Failed to decode message", http.StatusBadRequest)
		return
	}

	log.Printf("[Transport] Received message %s", envelope.Type)

	t.system.Send(envelope.To, msg)

	w.WriteHeader(http.StatusOK)
}

func (t *Transport) handleHealth(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{
		"status":    "healthy",
		"machine_id": t.machineID,
	})
}

func (t *Transport) decodeMessage(msgType string, payload json.RawMessage) (actor.Message, error) {
	t.mu.RLock()
	constructor, exists := t.messageTypes[msgType]
	t.mu.RUnlock()

	if !exists {
		return nil, fmt.Errorf("unknown message type: %s", msgType)
	}

	msg := constructor()
	if err := json.Unmarshal(payload, msg); err != nil {
		return nil, fmt.Errorf("failed to unmarshal %s: %w", msgType, err)
	}

	return msg, nil
}
