package middlewares

import (
	"fmt"
	"lama-backend/domain/entities"
	"log"
	"net/http"
	"os"
	"time"

	jwtware "github.com/gofiber/contrib/jwt"
	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v5"
)

func SetJWtHeaderHandler() fiber.Handler {
	return jwtware.New(jwtware.Config{
		SigningKey: jwtware.SigningKey{
			//ตัว secret key ดึงมาจาก .env
			Key: []byte(os.Getenv("JWT_SECRET_KEY")),
			//algorithm ที่เลือกใช้
			JWTAlg: jwtware.HS256,
		},
		ErrorHandler: func(c *fiber.Ctx, err error) error {
			return c.Status(fiber.StatusUnauthorized).JSON(entities.ResponseMessage{Message: "Unauthorization Token."})
		},
	})
}

type TokenDetails struct {
	Token     *string `json:"token"`
	UserID    string  `json:"user_id"`
	Role      string  `json:"role"`
	Purpose   string  `json:"purpose"`
	ExpiresIn *int64  `json:"exp"`
}

func DecodeJWTToken(ctx *fiber.Ctx) (*TokenDetails, error) {
	td := &TokenDetails{
		Token: new(string),
	}

	token, status := ctx.Locals("user").(*jwt.Token)
	if !status {
		return nil, ctx.Status(http.StatusUnauthorized).SendString("Unauthorization Token.")
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return nil, ctx.Status(http.StatusUnauthorized).SendString("Unauthorization Token.")
	}

	for key, value := range claims {
		if key == "user_id" || key == "sub" {
			td.UserID = value.(string)
		}
		if key == "role" {
			td.Role = value.(string)
		}
		if key == "purpose" {
			td.Purpose = value.(string)
		}
	}
	*td.Token = token.Raw
	return td, nil
}

func GenerateJWTToken(userID string, role string, purpose string) (*TokenDetails, error) {
	now := time.Now().UTC()

	td := &TokenDetails{
		ExpiresIn: new(int64),
		Token:     new(string),
	}

	// expiresIn is set to 6 hours from now
	*td.ExpiresIn = now.Add(time.Hour * 6).Unix()

	td.UserID = userID
	td.Role = role
	td.Purpose = purpose

	SigningKey := []byte(os.Getenv("JWT_SECRET_KEY"))

	atClaims := make(jwt.MapClaims)
	atClaims["user_id"] = userID
	atClaims["role"] = role
	atClaims["purpose"] = purpose
	atClaims["exp"] = time.Now().Add(time.Hour * 6).Unix()
	atClaims["iat"] = time.Now().Unix()
	atClaims["nbf"] = time.Now().Unix()

	log.Println("New claims: ", atClaims)

	token, err := jwt.NewWithClaims(jwt.SigningMethodHS256, atClaims).SignedString(SigningKey)
	if err != nil {
		return nil, fmt.Errorf("create: sign token: %w", err)
	}

	*td.Token = token
	return td, nil
}

func DecodeResetPasswordJWTToken(tokenString string) (*TokenDetails, error) {
	td := &TokenDetails{
		Token: new(string),
	}

	parsedToken, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		// Validate the signing method if needed
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(os.Getenv("JWT_SECRET_KEY")), nil
	})

	if err != nil || !parsedToken.Valid {
		return nil, fmt.Errorf("unauthorized token: %v", err)
	}

	claims, ok := parsedToken.Claims.(jwt.MapClaims)
	if !ok {
		return nil, fmt.Errorf("invalid token claims")
	}

	for key, value := range claims {
		if key == "user_id" || key == "sub" {
			td.UserID = value.(string)
		}
		if key == "role" {
			td.Role = value.(string)
		}
		if key == "purpose" {
			td.Purpose = value.(string)
		}
	}

	*td.Token = tokenString
	return td, nil
}

func GenerateResetPasswordJWTToken(userID string, role string, purpose string) (*TokenDetails, error) {
	now := time.Now().UTC()

	td := &TokenDetails{
		ExpiresIn: new(int64),
		Token:     new(string),
	}

	// expiresIn is set to 15 minutes from now
	*td.ExpiresIn = now.Add(time.Minute * 15).Unix()

	td.UserID = userID
	td.Role = role
	td.Purpose = purpose

	SigningKey := []byte(os.Getenv("JWT_SECRET_KEY"))

	atClaims := make(jwt.MapClaims)
	atClaims["user_id"] = userID
	atClaims["role"] = role
	atClaims["purpose"] = purpose
	atClaims["exp"] = time.Now().Add(time.Minute * 15).Unix()
	atClaims["iat"] = time.Now().Unix()
	atClaims["nbf"] = time.Now().Unix()

	log.Println("New claims: ", atClaims)

	token, err := jwt.NewWithClaims(jwt.SigningMethodHS256, atClaims).SignedString(SigningKey)
	if err != nil {
		return nil, fmt.Errorf("create: sign token: %w", err)
	}

	*td.Token = token
	return td, nil
}
