package telegram

import (
	"context"
	"fmt"

	"github.com/gotd/td/tg"
)

// GetPhotoData отримує дані фото з повідомлення
func (c *Client) GetPhotoData(ctx context.Context, messageID int, chatID int64) ([]byte, error) {
	var photoData []byte

	err := c.Client.Run(ctx, func(ctx context.Context) error {
		api := c.Client.API()

		// Створюємо InputPeer для чату
		peer, err := c.GetInputPeer(ctx, chatID)
		if err != nil {
			return fmt.Errorf("get input peer error: %w", err)
		}

		// Отримуємо повідомлення
		result, err := api.MessagesGetMessages(ctx, []tg.InputMessageClass{
			&tg.InputMessageID{ID: messageID},
		})
		if err != nil {
			return fmt.Errorf("get message error: %w", err)
		}

		var messages *tg.MessagesMessages
		switch m := result.(type) {
		case *tg.MessagesMessages:
			messages = m
		case *tg.MessagesMessagesSlice:
			messages = &tg.MessagesMessages{
				Messages: m.Messages,
				Chats:    m.Chats,
				Users:    m.Users,
			}
		case *tg.MessagesChannelMessages:
			messages = &tg.MessagesMessages{
				Messages: m.Messages,
				Chats:    m.Chats,
				Users:    m.Users,
			}
		default:
			return fmt.Errorf("unexpected messages type: %T", result)
		}

		if len(messages.Messages) == 0 {
			return fmt.Errorf("message not found")
		}

		msg, ok := messages.Messages[0].(*tg.Message)
		if !ok {
			return fmt.Errorf("invalid message type")
		}

		// Перевіряємо чи є медіа
		if msg.Media == nil {
			return fmt.Errorf("no media in message")
		}

		mediaPhoto, ok := msg.Media.(*tg.MessageMediaPhoto)
		if !ok {
			return fmt.Errorf("media is not a photo")
		}

		photo, ok := mediaPhoto.Photo.(*tg.Photo)
		if !ok {
			return fmt.Errorf("invalid photo type")
		}

		// Знаходимо найбільший доступний розмір
		var bestSize *tg.PhotoSize
		for _, size := range photo.Sizes {
			if ps, ok := size.(*tg.PhotoSize); ok {
				if bestSize == nil || ps.Size > bestSize.Size {
					bestSize = ps
				}
			}
		}

		if bestSize == nil {
			return fmt.Errorf("no suitable photo size found")
		}

		// Завантажуємо файл
		location := &tg.InputPhotoFileLocation{
			ID:            photo.ID,
			AccessHash:    photo.AccessHash,
			FileReference: photo.FileReference,
			ThumbSize:     bestSize.Type,
		}

		// Створюємо буфер для даних
		var buffer []byte
		offset := int64(0)
		limit := 1024 * 512 // 512KB за раз

		for {
			res, err := api.UploadGetFile(ctx, &tg.UploadGetFileRequest{
				Location: location,
				Offset:   offset,
				Limit:    limit,
			})
			if err != nil {
				return fmt.Errorf("failed to download chunk at offset %d: %w", offset, err)
			}

			file, ok := res.(*tg.UploadFile)
			if !ok {
				break
			}

			buffer = append(buffer, file.Bytes...)

			// Якщо отримали менше ніж limit, то це останній chunk
			if len(file.Bytes) < limit {
				break
			}

			offset += int64(len(file.Bytes))
		}

		photoData = buffer
		_ = peer // уникаємо помилки про unused variable

		return nil
	})

	if err != nil {
		return nil, err
	}

	return photoData, nil
}
