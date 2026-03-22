package services

import (
	"QUICK-READ-BACKEND/config"
	"QUICK-READ-BACKEND/models"
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

// RegisterDefaultHandlers sets up the default event handling pipeline
func RegisterDefaultHandlers() {
	notifService := &NotificationService{}
	emailSvc := &EmailService{}
	smsSvc := &SMSService{}

	// When an order is created, notify the pharmacy
	Bus.Subscribe(EventOrderCreated, func(e Event) {
		pharmacyID, _ := e.Payload["pharmacy_id"].(uint)
		orderID, _ := e.Payload["order_id"].(uint)
		if pharmacyID > 0 {
			notifService.Create(pharmacyID, "New Order", "New order #"+uintToStr(orderID)+" received", "order_update")
		}
	})

	// When payment completes, update order and notify
	Bus.Subscribe(EventPaymentCompleted, func(e Event) {
		userID, _ := e.Payload["user_id"].(uint)
		orderID, _ := e.Payload["order_id"].(uint)
		
		// In-app notification
		notifService.NotifyOrderUpdate(userID, orderID, "Payment confirmed")

		// Email Receipt
		var user models.User
		var order models.Order
		if err := config.DB.First(&user, userID).Error; err == nil {
			if err := config.DB.First(&order, orderID).Error; err == nil {
				go emailSvc.SendOrderReceipt(user.Email, order.ID, order.TotalAmount, order.DeliveryFee, order.PaymentMethod)
			}
		}
	})

	// When delivery is assigned, notify the customer
	Bus.Subscribe(EventDeliveryAssigned, func(e Event) {
		userID, _ := e.Payload["user_id"].(uint)
		notifService.NotifyDeliveryUpdate(userID, "A driver has been assigned to your order")

		// SMS Notification
		var user models.User
		if err := config.DB.First(&user, userID).Error; err == nil && user.Phone != "" {
			go smsSvc.SendDeliveryNotification(user.Phone, 0, "assigned to a driver") // orderID not easily available in payload here, but we can pass it if needed
		}
	})

	// When delivery completes, update driver stats
	Bus.Subscribe(EventDeliveryCompleted, func(e Event) {
		userID, _ := e.Payload["user_id"].(uint)
		orderID, _ := e.Payload["order_id"].(uint)
		notifService.NotifyOrderUpdate(userID, orderID, "Delivered successfully!")

		// SMS Notification
		var user models.User
		if err := config.DB.First(&user, userID).Error; err == nil && user.Phone != "" {
			go smsSvc.SendDeliveryNotification(user.Phone, orderID, "DELIVERED")
		}
	})

	// When stock is low, alert the pharmacist
	Bus.Subscribe(EventStockLow, func(e Event) {
		pharmacistID, _ := e.Payload["pharmacist_id"].(uint)
		medicineName, _ := e.Payload["medicine_name"].(string)
		stockLevel, _ := e.Payload["stock_level"].(int)
		notifService.NotifyInventoryAlert(pharmacistID, medicineName, stockLevel)
	})

	fmt.Println("✅ Default event handlers registered")
}
