package telegram

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/gotd/td/tg"
)

type Message struct {
	ID        int       `json:"id"`
	ChatID    string    `json:"chat_id"`
	ChatName  string    `json:"chat_name"`
	Text      string    `json:"text"`
	Sender    string    `json:"sender"`
	Timestamp time.Time `json:"timestamp"`
	IsRead    bool      `json:"is_read"`
	Out       bool      `json:"out"`
	HasPhoto  bool      `json:"has_photo"`
	PhotoID   int64     `json:"photo_id,omitempty"`
}

// GetMessages отримує повідомлення з чату
func (c *Client) GetMessages(ctx context.Context, chatID int64, limit int) ([]Message, error) {
	var messages []Message

	err := c.Client.Run(ctx, func(ctx context.Context) error {
		// Отримуємо API клієнт всередині Run callback
		api := c.Client.API()

		// Створюємо InputPeer для чату
		peer, err := c.GetInputPeer(ctx, chatID)
		if err != nil {
			return fmt.Errorf("get input peer error: %w", err)
		}

		// Отримуємо історію повідомлень
		result, err := api.MessagesGetHistory(ctx, &tg.MessagesGetHistoryRequest{
			Peer:  peer,
			Limit: limit,
		})
		if err != nil {
			return fmt.Errorf("get history error: %w", err)
		}

		var messagesSlice *tg.MessagesMessages
		switch m := result.(type) {
		case *tg.MessagesMessages:
			messagesSlice = m
		case *tg.MessagesMessagesSlice:
			messagesSlice = &tg.MessagesMessages{
				Messages: m.Messages,
				Chats:    m.Chats,
				Users:    m.Users,
			}
		case *tg.MessagesChannelMessages:
			messagesSlice = &tg.MessagesMessages{
				Messages: m.Messages,
				Chats:    m.Chats,
				Users:    m.Users,
			}
		case *tg.MessagesMessagesNotModified:
			return nil
		default:
			return fmt.Errorf("unexpected messages type: %T", result)
		}

		// Створюємо мапу користувачів
		users := make(map[int64]*tg.User)
		for _, u := range messagesSlice.Users {
			if user, ok := u.(*tg.User); ok {
				users[user.ID] = user
			}
		}

		// Обробляємо повідомлення
		for _, m := range messagesSlice.Messages {
			msg, ok := m.(*tg.Message)
			if !ok {
				continue
			}

			senderName := "Unknown"
			if msg.Out {
				senderName = "You"
			} else {
				// Для вхідних повідомлень визначаємо відправника
				if msg.FromID != nil {
					switch fromPeer := msg.FromID.(type) {
					case *tg.PeerUser:
						if user, exists := users[fromPeer.UserID]; exists {
							senderName = GetUserName(user)
						}
					case *tg.PeerChannel:
						senderName = "Channel"
					case *tg.PeerChat:
						senderName = "Chat"
					}
				} else {
					// Якщо FromID == nil, беремо з PeerID (для особистих чатів)
					if peerUser, ok := msg.PeerID.(*tg.PeerUser); ok {
						if user, exists := users[peerUser.UserID]; exists {
							senderName = GetUserName(user)
						}
					}
				}
			}

			// Визначаємо текст повідомлення і тип медіа
			messageText := msg.Message
			hasPhoto := false
			var photoID int64

			if msg.Media != nil {
				switch media := msg.Media.(type) {
				case *tg.MessageMediaPhoto:
					if photo, ok := media.Photo.(*tg.Photo); ok {
						hasPhoto = true
						photoID = photo.ID
						if messageText == "" {
							messageText = "📷 Фото"
						}
					}
				case *tg.MessageMediaDocument:
					if messageText == "" {
						messageText = "📎 Файл"
					}
				case *tg.MessageMediaGeo:
					if messageText == "" {
						messageText = "📍 Локація"
					}
				case *tg.MessageMediaContact:
					if messageText == "" {
						messageText = "👤 Контакт"
					}
				case *tg.MessageMediaVenue:
					if messageText == "" {
						messageText = "📍 Місце"
					}
				case *tg.MessageMediaWebPage:
					if messageText == "" {
						messageText = "🔗 Посилання"
					}
				default:
					if messageText == "" {
						messageText = "💬 Медіа"
					}
				}
			} else if messageText == "" {
				// Пропускаємо порожні повідомлення без медіа
				continue
			}

			messages = append(messages, Message{
				ID:        msg.ID,
				ChatID:    strconv.FormatInt(chatID, 10),
				ChatName:  "",
				Text:      messageText,
				Sender:    senderName,
				Timestamp: time.Unix(int64(msg.Date), 0),
				IsRead:    !msg.Out || msg.Out && msg.ID <= 0,
				Out:       msg.Out,
				HasPhoto:  hasPhoto,
				PhotoID:   photoID,
			})
		}

		return nil
	})

	return messages, err
}

// SendMessage відправляє повідомлення
func (c *Client) SendMessage(ctx context.Context, chatID int64, text string) (int, error) {
	var messageID int

	err := c.Client.Run(ctx, func(ctx context.Context) error {
		// Отримуємо API клієнт всередині Run callback
		api := c.Client.API()

		// Створюємо InputPeer для чату
		peer, err := c.GetInputPeer(ctx, chatID)
		if err != nil {
			return fmt.Errorf("get input peer error: %w", err)
		}

		// Генеруємо випадковий ID для повідомлення
		randomID := time.Now().UnixNano()

		// Відправляємо повідомлення
		updates, err := api.MessagesSendMessage(ctx, &tg.MessagesSendMessageRequest{
			Peer:     peer,
			Message:  text,
			RandomID: randomID,
		})
		if err != nil {
			return fmt.Errorf("send message error: %w", err)
		}

		// Отримуємо ID відправленого повідомлення
		switch u := updates.(type) {
		case *tg.Updates:
			if len(u.Updates) > 0 {
				if msgUpdate, ok := u.Updates[0].(*tg.UpdateMessageID); ok {
					messageID = msgUpdate.ID
				}
			}
		case *tg.UpdateShortSentMessage:
			messageID = u.ID
		}

		return nil
	})

	return messageID, err
}

// MarkAsRead позначає повідомлення як прочитані
func (c *Client) MarkAsRead(ctx context.Context, chatID int64, maxID int) error {
	return c.Client.Run(ctx, func(ctx context.Context) error {
		// Отримуємо API клієнт всередині Run callback
		api := c.Client.API()

		// Створюємо InputPeer для чату
		peer, err := c.GetInputPeer(ctx, chatID)
		if err != nil {
			return fmt.Errorf("get input peer error: %w", err)
		}

		// Перевіряємо тип peer'а
		if _, ok := peer.(*tg.InputPeerChannel); ok {
			// Для каналів використовуємо інший метод
			_, err = api.ChannelsReadHistory(ctx, &tg.ChannelsReadHistoryRequest{
				Channel: &tg.InputChannel{
					ChannelID:  chatID,
					AccessHash: 0, // TODO: отримувати access hash
				},
				MaxID: maxID,
			})
			return err
		}

		// Для звичайних чатів
		_, err = api.MessagesReadHistory(ctx, &tg.MessagesReadHistoryRequest{
			Peer:  peer,
			MaxID: maxID,
		})
		return err
	})
}

// GetInputPeer створює InputPeer для чату
func (c *Client) GetInputPeer(ctx context.Context, chatID int64) (tg.InputPeerClass, error) {
	// Для простоти спочатку пробуємо InputPeerUser
	// В production потрібно зберігати тип і access_hash для кожного чату
	return &tg.InputPeerUser{
		UserID: chatID,
	}, nil
}
