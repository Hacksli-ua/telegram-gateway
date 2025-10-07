package telegram

import (
	"context"
	"fmt"
	"time"

	"github.com/gotd/td/tg"
)

type Dialog struct {
	ID             int64     `json:"id"`
	Name           string    `json:"name"`
	LastMessage    string    `json:"last_message"`
	UnreadCount    int       `json:"unread_count"`
	LastUpdateTime time.Time `json:"last_update_time"`
	Type           string    `json:"type"` // "user", "chat", "channel"
}

// GetDialogs отримує список діалогів (чатів)
func (c *Client) GetDialogs(ctx context.Context, limit int) ([]Dialog, error) {
	var dialogs []Dialog

	err := c.Client.Run(ctx, func(ctx context.Context) error {
		// Отримуємо API клієнт всередині Run callback
		api := c.Client.API()

		// Отримуємо діалоги
		result, err := api.MessagesGetDialogs(ctx, &tg.MessagesGetDialogsRequest{
			OffsetPeer: &tg.InputPeerEmpty{},
			Limit:      limit,
		})
		if err != nil {
			return fmt.Errorf("get dialogs error: %w", err)
		}

		var dialogsSlice *tg.MessagesDialogsSlice
		switch d := result.(type) {
		case *tg.MessagesDialogs:
			// Конвертуємо MessagesDialogs в MessagesDialogsSlice
			dialogsSlice = &tg.MessagesDialogsSlice{
				Dialogs:  d.Dialogs,
				Messages: d.Messages,
				Chats:    d.Chats,
				Users:    d.Users,
			}
		case *tg.MessagesDialogsSlice:
			dialogsSlice = d
		case *tg.MessagesDialogsNotModified:
			return nil
		default:
			return fmt.Errorf("unexpected dialogs type: %T", result)
		}

		// Створюємо мапи для швидкого доступу
		users := make(map[int64]*tg.User)
		chats := make(map[int64]tg.ChatClass)
		messages := make(map[int64]*tg.Message)

		// Заповнюємо мапу користувачів
		for _, u := range dialogsSlice.Users {
			if user, ok := u.(*tg.User); ok {
				users[user.ID] = user
			}
		}

		// Заповнюємо мапу чатів
		for _, ch := range dialogsSlice.Chats {
			switch chat := ch.(type) {
			case *tg.Chat:
				chats[chat.ID] = chat
			case *tg.ChatForbidden:
				chats[chat.ID] = chat
			case *tg.Channel:
				chats[chat.ID] = chat
			case *tg.ChannelForbidden:
				chats[chat.ID] = chat
			}
		}

		// Заповнюємо мапу повідомлень
		for _, m := range dialogsSlice.Messages {
			if msg, ok := m.(*tg.Message); ok {
				peerID := GetPeerID(msg.PeerID)
				messages[peerID] = msg
			}
		}

		// Обробляємо діалоги
		for _, d := range dialogsSlice.Dialogs {
			dialog, ok := d.(*tg.Dialog)
			if !ok {
				continue
			}

			peerID := GetPeerID(dialog.Peer)
			name := ""
			dialogType := ""

			// Визначаємо тип і ім'я діалогу
			switch peer := dialog.Peer.(type) {
			case *tg.PeerUser:
				if user, exists := users[peer.UserID]; exists {
					name = GetUserName(user)
					dialogType = "user"
				}
			case *tg.PeerChat:
				if chat, exists := chats[peer.ChatID]; exists {
					name = GetChatTitle(chat)
					dialogType = "chat"
				}
			case *tg.PeerChannel:
				if channel, exists := chats[peer.ChannelID]; exists {
					name = GetChatTitle(channel)
					dialogType = "channel"
				}
			}

			// Отримуємо останнє повідомлення
			lastMessage := ""
			lastUpdateTime := time.Now()
			if msg, exists := messages[peerID]; exists {
				lastMessage = msg.Message
				lastUpdateTime = time.Unix(int64(msg.Date), 0)
			}

			dialogs = append(dialogs, Dialog{
				ID:             peerID,
				Name:           name,
				LastMessage:    lastMessage,
				UnreadCount:    dialog.UnreadCount,
				LastUpdateTime: lastUpdateTime,
				Type:           dialogType,
			})
		}

		return nil
	})

	return dialogs, err
}

// GetPeerID отримує ID з Peer
func GetPeerID(peer tg.PeerClass) int64 {
	switch p := peer.(type) {
	case *tg.PeerUser:
		return p.UserID
	case *tg.PeerChat:
		return p.ChatID
	case *tg.PeerChannel:
		return p.ChannelID
	}
	return 0
}

// GetUserName отримує ім'я користувача
func GetUserName(user *tg.User) string {
	if user.Username != "" {
		return "@" + user.Username
	}
	name := user.FirstName
	if user.LastName != "" {
		name += " " + user.LastName
	}
	return name
}

// GetChatTitle отримує назву чату
func GetChatTitle(chat tg.ChatClass) string {
	switch c := chat.(type) {
	case *tg.Chat:
		return c.Title
	case *tg.ChatForbidden:
		return c.Title
	case *tg.Channel:
		return c.Title
	case *tg.ChannelForbidden:
		return c.Title
	}
	return "Unknown"
}
