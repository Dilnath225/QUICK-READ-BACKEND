package services

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"
)

type SMSService struct{}

// GenerateOTP creates a random 6-digit OTP
func (s *SMSService) GenerateOTP() string {
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	return fmt.Sprintf("%06d", r.Intn(1000000))
}

// SendOTP sends an OTP via Twilio SMS
func (s *SMSService) SendOTP(phoneNumber string, otp string) error {
	accountSID := os.Getenv("TWILIO_ACCOUNT_SID")
	authToken := os.Getenv("TWILIO_AUTH_TOKEN")
	fromNumber := os.Getenv("TWILIO_FROM_NUMBER")

	if accountSID == "" || authToken == "" {
		fmt.Println("⚠️  Twilio not configured (TWILIO_ACCOUNT_SID missing), skipping SMS send")
		return nil
	}

	urlStr := fmt.Sprintf("https://api.twilio.com/2010-04-01/Accounts/%s/Messages.json", accountSID)

	message := fmt.Sprintf("Your Quick-Read verification code is: %s. Valid for 10 minutes.", otp)

	data := url.Values{}
	data.Set("To", phoneNumber)
	data.Set("From", fromNumber)
	data.Set("Body", message)

	req, err := http.NewRequest("POST", urlStr, strings.NewReader(data.Encode()))
	if err != nil {
		return err
	}

	req.SetBasicAuth(accountSID, authToken)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		fmt.Println("❌ Failed to send SMS:", err)
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		var errorResp map[string]interface{}
		json.NewDecoder(resp.Body).Decode(&errorResp)
		fmt.Println("❌ Twilio error:", errorResp)
		return fmt.Errorf("twilio error: %v", errorResp["message"])
	}

	fmt.Println("✅ SMS sent to:", phoneNumber)
	return nil
}

// SendDeliveryNotification sends an SMS about delivery status
func (s *SMSService) SendDeliveryNotification(phoneNumber string, orderID uint, status string) error {
	accountSID := os.Getenv("TWILIO_ACCOUNT_SID")
	authToken := os.Getenv("TWILIO_AUTH_TOKEN")
	fromNumber := os.Getenv("TWILIO_FROM_NUMBER")

	if accountSID == "" {
		fmt.Println("⚠️  Twilio not configured, skipping SMS")
		return nil
	}

	urlStr := fmt.Sprintf("https://api.twilio.com/2010-04-01/Accounts/%s/Messages.json", accountSID)

	message := fmt.Sprintf("Quick-Read: Your order #%d is now %s. Track your order in the app.", orderID, status)

	data := url.Values{}
	data.Set("To", phoneNumber)
	data.Set("From", fromNumber)
	data.Set("Body", message)

	req, _ := http.NewRequest("POST", urlStr, strings.NewReader(data.Encode()))
	req.SetBasicAuth(accountSID, authToken)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	return nil
}
