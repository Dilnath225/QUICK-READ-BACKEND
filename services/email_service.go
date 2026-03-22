package services

import (
	"fmt"
	"os"

	"github.com/sendgrid/sendgrid-go"
	"github.com/sendgrid/sendgrid-go/helpers/mail"
)

type EmailService struct{}

// SendOTP sends an OTP verification email
func (s *EmailService) SendOTP(toEmail string, otp string) error {
	subject := "Quick-Read: Your OTP Verification Code"
	body := fmt.Sprintf(`
		<html>
		<body style="font-family: Arial, sans-serif; padding: 20px;">
			<h2 style="color: #2563EB;">Quick-Read Verification</h2>
			<p>Your one-time verification code is:</p>
			<h1 style="color: #1E40AF; letter-spacing: 8px; font-size: 36px;">%s</h1>
			<p>This code expires in 10 minutes.</p>
			<p style="color: #6B7280; font-size: 12px;">If you didn't request this code, please ignore this email.</p>
		</body>
		</html>
	`, otp)

	return s.sendEmail(toEmail, subject, body)
}

// SendOrderReceipt sends a digital receipt via email
func (s *EmailService) SendOrderReceipt(toEmail string, orderID uint, totalAmount float64, deliveryFee float64, paymentMethod string) error {
	subject := fmt.Sprintf("Quick-Read: Order #%d Receipt", orderID)
	body := fmt.Sprintf(`
		<html>
		<body style="font-family: Arial, sans-serif; padding: 20px;">
			<h2 style="color: #2563EB;">Quick-Read Digital Receipt</h2>
			<hr>
			<table style="width: 100%%; border-collapse: collapse;">
				<tr><td style="padding: 8px;"><strong>Order ID:</strong></td><td>#%d</td></tr>
				<tr><td style="padding: 8px;"><strong>Subtotal:</strong></td><td>Rs. %.2f</td></tr>
				<tr><td style="padding: 8px;"><strong>Delivery Fee:</strong></td><td>Rs. %.2f</td></tr>
				<tr style="background: #F3F4F6;"><td style="padding: 8px;"><strong>Total:</strong></td><td><strong>Rs. %.2f</strong></td></tr>
				<tr><td style="padding: 8px;"><strong>Payment:</strong></td><td>%s</td></tr>
			</table>
			<hr>
			<p style="color: #6B7280; font-size: 12px;">Thank you for choosing Quick-Read Pharmacy.</p>
		</body>
		</html>
	`, orderID, totalAmount, deliveryFee, totalAmount+deliveryFee, paymentMethod)

	return s.sendEmail(toEmail, subject, body)
}

// sendEmail is the internal SendGrid API sender
func (s *EmailService) sendEmail(to, subject, htmlBody string) error {
	apiKey := os.Getenv("SENDGRID_API_KEY")
	fromAddress := os.Getenv("SENDGRID_FROM_EMAIL")

	if apiKey == "" || fromAddress == "" {
		fmt.Println("⚠️  Email not configured (SENDGRID_API_KEY or SENDGRID_FROM_EMAIL missing), skipping email send")
		return nil
	}

	// 1. Build the SendGrid email object
	from := mail.NewEmail("Quick-Read System", fromAddress)
	recipient := mail.NewEmail("Quick-Read User", to)

	// plainTextContent is required by SendGrid as a fallback for old email clients
	plainTextContent := "Please view this email in an HTML-compatible email client."

	message := mail.NewSingleEmail(from, subject, recipient, plainTextContent, htmlBody)

	// 2. Send the email using the API key
	client := sendgrid.NewSendClient(apiKey)
	response, err := client.Send(message)

	// 3. Handle errors and logging
	if err != nil {
		fmt.Println("❌ Failed to send email via SendGrid API:", err)
		return err
	}

	if response.StatusCode >= 200 && response.StatusCode <= 299 {
		fmt.Println("✅ Email sent to:", to)
		return nil
	}

	errResponse := fmt.Errorf("SendGrid returned status %d: %s", response.StatusCode, response.Body)
	fmt.Println("❌", errResponse)
	return errResponse
}
