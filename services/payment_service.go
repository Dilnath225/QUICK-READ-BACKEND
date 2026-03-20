package services

import (
	"QUICK-READ-SYSTEM/config"
	"QUICK-READ-SYSTEM/models"
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
