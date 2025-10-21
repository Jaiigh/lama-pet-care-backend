package utils

import (
	"fmt"
	// "regexp"

	"os"

	"golang.org/x/crypto/bcrypt"

	"github.com/resend/resend-go/v2"
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
	// if len(password) < 8 {
	// 	return false
	// }

	// // Check for at least one uppercase letter
	// hasUpper := regexp.MustCompile(`[A-Z]`).MatchString(password)
	// // Check for at least one lowercase letter
	// hasLower := regexp.MustCompile(`[a-z]`).MatchString(password)
	// // Check for at least one digit
	// hasDigit := regexp.MustCompile(`\d`).MatchString(password)
	// // Check for at least one special character
	// hasSpecial := regexp.MustCompile(`[!@#$%^&*()_+\-=\[\]{};':"\\|,.<>/?]`).MatchString(password)

	// return hasUpper && hasLower && hasDigit && hasSpecial
	return len(password) >= 8
}

// resend api
func SendResetEmail(toEmail string, resetLink string) error {
	client := resend.NewClient(os.Getenv("RESEND_API_KEY"))

	params := &resend.SendEmailRequest{
		From:    "LAMA Support <onboarding@resend.dev>",
		To:      []string{toEmail},
		Subject: "Password Reset Link",
		Html: fmt.Sprintf(`
			<p>You requested to reset your password.</p>
			<p>Click <a href="%s">here</a> to reset. Link expires in 15 minutes.</p>`, resetLink),
	}

	// it returns sent response and error but sent response which is just id is not useful now
	_, err := client.Emails.Send(params)

	return err
}
