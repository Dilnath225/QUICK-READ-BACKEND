package services

import (
	"fmt"
	"math/rand"
	"time"
)

type SMSService struct{}

// GenerateOTP creates a random 6-digit OTP
func (s *SMSService) GenerateOTP() string {
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	return fmt.Sprintf("%06d", r.Intn(1000000))
}
