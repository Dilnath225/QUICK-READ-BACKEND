package services

import (
	"QUICK-READ-BACKEND/config"
	"QUICK-READ-BACKEND/models"
	"errors"
	"fmt"
	"os"

	"github.com/stripe/stripe-go/v76"
	"github.com/stripe/stripe-go/v76/paymentintent"
)

type PaymentService struct{}

func init() {
	key := os.Getenv("STRIPE_SECRET_KEY")
	if key != "" {
		stripe.Key = key
	}
}

func (s *PaymentService) InitializePayment(orderID string, paymentMethod string) (*models.Order, string, error) {
	var order models.Order
	if err := config.DB.First(&order, orderID).Error; err != nil {
		return nil, "", errors.New("order not found")
	}

	order.PaymentMethod = paymentMethod
	order.PaymentStatus = "pending"

	clientSecret := ""

	switch paymentMethod {
	case "card":
		// Create a real Stripe PaymentIntent
		if stripe.Key == "" || stripe.Key == "sk_test_4eC39HqLyjWDarjtT1zdp7dc" {
			// Fallback to mock if no real key configured
			clientSecret = "pi_mock_secret_" + orderID
			order.TransactionID = "txn_mock_" + orderID
			fmt.Println("⚠️  Using mock Stripe (no real key configured)")
		} else {
			amount := int64((order.TotalAmount + order.DeliveryFee) * 100) // Stripe expects cents
			if amount < 50 {
				amount = 50 // Minimum charge
			}

			params := &stripe.PaymentIntentParams{
				Amount:   stripe.Int64(amount),
				Currency: stripe.String("lkr"), // Sri Lankan Rupee
				AutomaticPaymentMethods: &stripe.PaymentIntentAutomaticPaymentMethodsParams{
					Enabled: stripe.Bool(true),
				},
			}
			params.AddMetadata("order_id", orderID)

			pi, err := paymentintent.New(params)
			if err != nil {
				return nil, "", fmt.Errorf("stripe error: %v", err)
			}

			clientSecret = pi.ClientSecret
			order.TransactionID = pi.ID
		}

	case "cod", "cash":
		order.PaymentStatus = "pending" // Payment on delivery

	case "koko":
		// Koko buy-now-pay-later — mark as pending
		order.PaymentStatus = "pending"
		order.TransactionID = "koko_" + orderID

	default:
		return nil, "", errors.New("invalid payment method: card, cod, cash, or koko")
	}

	if err := config.DB.Save(&order).Error; err != nil {
		return nil, "", err
	}

	// Publish payment event
	if Bus != nil {
		Bus.Publish(Event{
			Type: EventOrderCreated,
			Payload: map[string]interface{}{
				"order_id":    order.ID,
				"user_id":     order.UserID,
				"pharmacy_id": order.PharmacyID,
			},
		})
	}

	return &order, clientSecret, nil
}

// ConfirmPayment is called by Stripe webhook or manual confirmation
func (s *PaymentService) ConfirmPayment(orderID string) (*models.Order, error) {
	var order models.Order
	if err := config.DB.First(&order, orderID).Error; err != nil {
		return nil, errors.New("order not found")
	}

	order.PaymentStatus = "paid"
	if err := config.DB.Save(&order).Error; err != nil {
		return nil, err
	}

	// Publish payment completed event
	if Bus != nil {
		Bus.Publish(Event{
			Type: EventPaymentCompleted,
			Payload: map[string]interface{}{
				"order_id": order.ID,
				"user_id":  order.UserID,
			},
		})
	}

	return &order, nil
}
