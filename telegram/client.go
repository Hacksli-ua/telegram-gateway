package telegram

import (
	"context"
	"encoding/base64"
	"fmt"
	"log"
	"os"
	"telegram-gateway/config"
	"time"

	"github.com/gotd/td/telegram"
	"github.com/gotd/td/telegram/auth"
)

type Client struct {
	Client      *telegram.Client
	Config      *config.Config
	SessionPath string
}

// NewClient створює новий Telegram клієнт
func NewClient(cfg *config.Config) (*Client, error) {
	// Унікальний файл сесії для кожного клієнта
	sessionPath := fmt.Sprintf("session_%d.json", time.Now().UnixNano())

	client := telegram.NewClient(cfg.TelegramAPIID, cfg.TelegramAPIHash, telegram.Options{
		SessionStorage: &telegram.FileSessionStorage{
			Path: sessionPath,
		},
	})

	return &Client{
		Client:      client,
		Config:      cfg,
		SessionPath: sessionPath,
	}, nil
}

// NewClientWithSession створює клієнт з існуючою сесією
func NewClientWithSession(cfg *config.Config, sessionData string) (*Client, error) {
	// Створюємо унікальний файл для цієї сесії
	sessionPath := fmt.Sprintf("session_%d.json", time.Now().UnixNano())

	// Декодуємо base64 та зберігаємо в файл
	decoded, err := base64.StdEncoding.DecodeString(sessionData)
	if err != nil {
		return nil, fmt.Errorf("failed to decode session data: %w", err)
	}

	if err := os.WriteFile(sessionPath, decoded, 0600); err != nil {
		return nil, fmt.Errorf("failed to write session file: %w", err)
	}

	client := telegram.NewClient(cfg.TelegramAPIID, cfg.TelegramAPIHash, telegram.Options{
		SessionStorage: &telegram.FileSessionStorage{
			Path: sessionPath,
		},
	})

	return &Client{
		Client:      client,
		Config:      cfg,
		SessionPath: sessionPath,
	}, nil
}

// GetSessionData повертає дані сесії в base64
func (c *Client) GetSessionData() (string, error) {
	data, err := os.ReadFile(c.SessionPath)
	if err != nil {
		return "", fmt.Errorf("failed to read session file: %w", err)
	}

	return base64.StdEncoding.EncodeToString(data), nil
}

// Connect підключається до Telegram
func (c *Client) Connect(ctx context.Context) error {
	return c.Client.Run(ctx, func(ctx context.Context) error {
		status, err := c.Client.Auth().Status(ctx)
		if err != nil {
			return fmt.Errorf("auth status error: %w", err)
		}

		if !status.Authorized {
			log.Println("Not authorized. Need to login.")
			return nil
		}

		user, ok := status.User.AsNotEmpty()
		if !ok {
			return fmt.Errorf("user is empty")
		}

		log.Printf("Authorized as %s %s (@%s)", user.FirstName, user.LastName, user.Username)
		return nil
	})
}

// Auth виконує авторизацію через номер телефону
func (c *Client) Auth(ctx context.Context, phone string) error {
	return c.Client.Run(ctx, func(ctx context.Context) error {
		flow := auth.NewFlow(
			&AuthHandler{PhoneNumber: phone},
			auth.SendCodeOptions{},
		)

		if err := c.Client.Auth().IfNecessary(ctx, flow); err != nil {
			return fmt.Errorf("auth error: %w", err)
		}

		return nil
	})
}
