package telegram

import (
	"context"
	"log"
	"sync"

	"github.com/gotd/td/telegram/updates"
	"github.com/gotd/td/tg"
)

var (
	updateChannels     = make(map[string]chan *tg.UpdateNewMessage)
	updateChannelMutex sync.RWMutex
)

// UpdatesHandler обробляє оновлення від Telegram
type UpdatesHandler struct {
	client *Client
}

// Handle обробляє оновлення
func (h *UpdatesHandler) Handle(ctx context.Context, u tg.UpdatesClass) error {
	switch upd := u.(type) {
	case *tg.Updates:
		for _, update := range upd.Updates {
			h.handleUpdate(update)
		}
	case *tg.UpdateShort:
		h.handleUpdate(upd.Update)
	}
	return nil
}

// handleUpdate обробляє окреме оновлення
func (h *UpdatesHandler) handleUpdate(u tg.UpdateClass) {
	switch upd := u.(type) {
	case *tg.UpdateNewMessage:
		h.handleNewMessage(upd)
	case *tg.UpdateNewChannelMessage:
		// Конвертуємо в UpdateNewMessage
		h.handleNewMessage(&tg.UpdateNewMessage{
			Message: upd.Message,
		})
	}
}

// handleNewMessage обробляє нові повідомлення
func (h *UpdatesHandler) handleNewMessage(upd *tg.UpdateNewMessage) {
	msg, ok := upd.Message.(*tg.Message)
	if !ok {
		return
	}

	log.Printf("New message: %s", msg.Message)

	// Відправляємо оновлення всім підписаним клієнтам
	updateChannelMutex.RLock()
	defer updateChannelMutex.RUnlock()

	for _, ch := range updateChannels {
		select {
		case ch <- upd:
		default:
			// Канал переповнений, пропускаємо
		}
	}
}

// StartUpdatesListener запускає слухач оновлень
func (c *Client) StartUpdatesListener(ctx context.Context) error {
	return c.Client.Run(ctx, func(ctx context.Context) error {
		// Отримуємо API клієнт всередині Run callback
		api := c.Client.API()

		gaps := updates.New(updates.Config{
			Handler: &UpdatesHandler{client: c},
		})

		// Реєструємо обробник оновлень
		return gaps.Run(ctx, api, 0, updates.AuthOptions{})
	})
}

// SubscribeToUpdates підписується на оновлення
func SubscribeToUpdates(sessionToken string) chan *tg.UpdateNewMessage {
	updateChannelMutex.Lock()
	defer updateChannelMutex.Unlock()

	ch := make(chan *tg.UpdateNewMessage, 100)
	updateChannels[sessionToken] = ch
	return ch
}

// UnsubscribeFromUpdates відписується від оновлень
func UnsubscribeFromUpdates(sessionToken string) {
	updateChannelMutex.Lock()
	defer updateChannelMutex.Unlock()

	if ch, exists := updateChannels[sessionToken]; exists {
		close(ch)
		delete(updateChannels, sessionToken)
	}
}
