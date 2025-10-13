package utils

import (
	"fmt"
	"regexp"

	"os"

	"golang.org/x/crypto/bcrypt"

	"gopkg.in/gomail.v2"
)

func HashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	return string(bytes), err
}

func CheckPasswordHash(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}

func ValidPassword(password string) bool {
	if len(password) < 8 {
		return false
	}

	// Check for at least one uppercase letter
	hasUpper := regexp.MustCompile(`[A-Z]`).MatchString(password)
	// Check for at least one lowercase letter
	hasLower := regexp.MustCompile(`[a-z]`).MatchString(password)
	// Check for at least one digit
	hasDigit := regexp.MustCompile(`\d`).MatchString(password)
	// Check for at least one special character
	hasSpecial := regexp.MustCompile(`[!@#$%^&*()_+\-=\[\]{};':"\\|,.<>/?]`).MatchString(password)

	return hasUpper && hasLower && hasDigit && hasSpecial
}

func SendResetEmail(toEmail string, resetLink string) error {
	m := gomail.NewMessage()
	app_email := os.Getenv("SMTP_USER")
	m.SetHeader("From", "no-reply@lama.com")
	m.SetHeader("To", toEmail)
	m.SetHeader("Subject", "Password Reset Link")
	m.SetBody("text/html", fmt.Sprintf(`
	    <p>You requested to reset your password.</p>
	    <p>Click <a href="%s">here</a> to reset. Link expires in 15 minutes.</p>`, resetLink))

	app_pass := os.Getenv("SMTP_PASS")
	d := gomail.NewDialer("smtp.gmail.com", 587, app_email, app_pass)

	return d.DialAndSend(m)
}
