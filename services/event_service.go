package services

import (
	"fmt"
	"sync"
)

// ========================
// Event-Driven Architecture
// ========================
// Channel-based event bus to decouple order creation, payment, and delivery assignment.
// In production, replace this with RabbitMQ, Kafka, or NATS.

// EventType represents the type of event
type EventType string

const (
	EventOrderCreated      EventType = "order.created"
	EventPaymentCompleted  EventType = "payment.completed"
	EventPaymentFailed     EventType = "payment.failed"
	EventDeliveryAssigned  EventType = "delivery.assigned"
	EventDeliveryCompleted EventType = "delivery.completed"
	EventStockLow          EventType = "stock.low"
)

// Event is a message in the event system
type Event struct {
	Type    EventType
	Payload map[string]interface{}
}

// EventHandler is a function that handles an event
type EventHandler func(event Event)

// EventBus manages event subscriptions and publishing
type EventBus struct {
	mu       sync.RWMutex
	handlers map[EventType][]EventHandler
}

var Bus *EventBus

// InitEventBus creates and starts the global event bus
func InitEventBus() {
	Bus = &EventBus{
		handlers: make(map[EventType][]EventHandler),
	}
	fmt.Println("✅ Event bus initialized")
}

// Subscribe registers a handler for an event type
func (eb *EventBus) Subscribe(eventType EventType, handler EventHandler) {
	eb.mu.Lock()
	defer eb.mu.Unlock()
	eb.handlers[eventType] = append(eb.handlers[eventType], handler)
}

// Publish fires an event to all subscribed handlers (async via goroutines)
func (eb *EventBus) Publish(event Event) {
	eb.mu.RLock()
	defer eb.mu.RUnlock()

	handlers, exists := eb.handlers[event.Type]
	if !exists {
		return
	}

	for _, handler := range handlers {
		go handler(event) // Non-blocking: each handler runs in its own goroutine
	}
}
