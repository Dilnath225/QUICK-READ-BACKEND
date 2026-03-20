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
