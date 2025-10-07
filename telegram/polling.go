package telegram

import (
	"context"
)

// GetNewMessages отримує нові повідомлення для чату з певного message_id
// Використовує прямий виклик без Run() для уникнення закриття клієнта
func (c *Client) GetNewMessages(ctx context.Context, chatID int64, afterMessageID int, limit int) ([]Message, error) {
	var newMessages []Message

	err := c.Client.Run(ctx, func(ctx context.Context) error {
		// Отримуємо всі повідомлення через GetMessages
		allMessages, err := c.getMessagesInternal(ctx, chatID, limit)
		if err != nil {
			return err
		}

		// Фільтруємо тільки нові (з ID > afterMessageID)
		for _, msg := range allMessages {
			if msg.ID > afterMessageID {
				newMessages = append(newMessages, msg)
			}
		}

		return nil
	})

	return newMessages, err
}

// getMessagesInternal - внутрішня функція для отримання повідомлень без Run()
// Використовується всередині існуючого Run() контексту
func (c *Client) getMessagesInternal(ctx context.Context, chatID int64, limit int) ([]Message, error) {
	return c.GetMessages(ctx, chatID, limit)
}
