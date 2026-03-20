package services

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
