package telegram

import (
	"context"
	"fmt"
	"log"

	"github.com/gotd/td/telegram/auth"
	"github.com/gotd/td/tg"
)

var (
	pendingAuth        = make(map[string]*PendingAuth)
	authCodeChannel    = make(map[string]chan string)
	authPasswordChannel = make(map[string]chan string)
)

type PendingAuth struct {
	Phone           string
	PhoneHash       string
	NeedsPassword   bool
}

type AuthHandler struct {
	PhoneNumber string
}

// Phone повертає номер телефону
func (a AuthHandler) Phone(_ context.Context) (string, error) {
	return a.PhoneNumber, nil
}

// Password повертає пароль (якщо потрібен)
func (a AuthHandler) Password(_ context.Context) (string, error) {
	log.Printf("2FA Password requested for: %s", a.PhoneNumber)

	// Встановлюємо прапорець, що потрібен пароль
	auth, exists := pendingAuth[a.PhoneNumber]
	if !exists {
		auth = &PendingAuth{
			Phone:         a.PhoneNumber,
			NeedsPassword: true,
		}
		pendingAuth[a.PhoneNumber] = auth
	} else {
		auth.NeedsPassword = true
	}

	log.Printf("Set NeedsPassword=true for: %s", a.PhoneNumber)

	// Чекаємо на пароль з API endpoint
	ch, exists := authPasswordChannel[a.PhoneNumber]
	if !exists {
		ch = make(chan string, 1)
		authPasswordChannel[a.PhoneNumber] = ch
	}

	log.Printf("Waiting for password from API for: %s", a.PhoneNumber)
	password := <-ch
	log.Printf("Received password from API for: %s", a.PhoneNumber)

	return password, nil
}

// Code повертає код авторизації
func (a AuthHandler) Code(_ context.Context, _ *tg.AuthSentCode) (string, error) {
	// Чекаємо на код з API endpoint
	ch, exists := authCodeChannel[a.PhoneNumber]
	if !exists {
		ch = make(chan string, 1)
		authCodeChannel[a.PhoneNumber] = ch
	}

	code := <-ch
	return code, nil
}

// AcceptTermsOfService приймає умови використання
func (a AuthHandler) AcceptTermsOfService(_ context.Context, tos tg.HelpTermsOfService) error {
	return nil
}

// SignUp виконує реєстрацію нового користувача
func (a AuthHandler) SignUp(_ context.Context) (auth.UserInfo, error) {
	return auth.UserInfo{
		FirstName: "Symbian",
		LastName:  "User",
	}, nil
}

// StorePendingAuth зберігає інформацію про очікувану авторизацію
func StorePendingAuth(phone, phoneHash string) {
	pendingAuth[phone] = &PendingAuth{
		Phone:         phone,
		PhoneHash:     phoneHash,
		NeedsPassword: false,
	}
}

// SubmitCode відправляє код авторизації
func SubmitCode(phone, code string) error {
	ch, exists := authCodeChannel[phone]
	if !exists {
		return fmt.Errorf("no pending auth for phone %s", phone)
	}

	ch <- code
	return nil
}

// SubmitPassword відправляє пароль 2FA
func SubmitPassword(phone, password string) error {
	ch, exists := authPasswordChannel[phone]
	if !exists {
		return fmt.Errorf("no pending password request for phone %s", phone)
	}

	ch <- password
	return nil
}

// GetPendingAuth отримує інформацію про очікувану авторизацію
func GetPendingAuth(phone string) (*PendingAuth, bool) {
	auth, exists := pendingAuth[phone]
	return auth, exists
}

// ClearPendingAuth очищає інформацію про очікувану авторизацію
func ClearPendingAuth(phone string) {
	delete(pendingAuth, phone)
	delete(authCodeChannel, phone)
	delete(authPasswordChannel, phone)
}
