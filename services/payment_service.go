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
