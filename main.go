package main

import (
	"context"
	"encoding/base64"
	"fmt"
	"log"
	"strconv"
	"sync"
	"telegram-gateway/config"
	tgclient "telegram-gateway/telegram"
	"time"

	"github.com/gin-gonic/gin"
)

// Structures
type User struct {
	ID             string
	Phone          string
	TelegramClient *tgclient.Client
	LastActivity   time.Time
}

type AuthRequest struct {
	Phone string `json:"phone"`
}

type AuthCodeRequest struct {
	Phone string `json:"phone"`
	Code  string `json:"code"`
}

type AuthPasswordRequest struct {
	Phone    string `json:"phone"`
	Password string `json:"password"`
}

type SendMessageRequest struct {
	ChatID string `json:"chat_id"`
	Text   string `json:"text"`
}

type MarkReadRequest struct {
	ChatID     string `json:"chat_id"`
	MessageIDs []int  `json:"message_ids"`
}

// Global storage
var (
	appConfig *config.Config
)

func main() {
	// Завантаження конфігурації
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatal("Failed to load config:", err)
	}
	appConfig = cfg

	log.Printf("Starting Telegram Gateway Server")
	log.Printf("API ID: %d", cfg.TelegramAPIID)
	log.Printf("Server: %s:%s", cfg.ServerHost, cfg.ServerPort)

	r := gin.Default()

	// CORS
	r.Use(func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization, X-Phone, X-Session-Data")
		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}
		c.Next()
	})

	// Routes
	auth := r.Group("/auth")
	{
		auth.POST("/request-code", requestAuthCode)
		auth.POST("/login", login)
		auth.POST("/password", submitPassword)
	}

	api := r.Group("/api")
	{
		// Endpoints з authMiddleware
		authenticated := api.Group("")
		authenticated.Use(authMiddleware())
		{
			authenticated.GET("/chats", getChats)
			authenticated.GET("/messages/:chat_id", getMessages)
			authenticated.POST("/send", sendMessage)
			authenticated.POST("/mark-read", markAsRead)
			authenticated.GET("/poll/:chat_id", pollMessages)
		}

		// Photo endpoint без middleware (використовує token з query)
		api.GET("/photo/:chat_id/:message_id", getPhoto)
	}

	addr := fmt.Sprintf("%s:%s", cfg.ServerHost, cfg.ServerPort)
	log.Printf("Server starting on %s", addr)
	r.Run(addr)
}

// AUTH HANDLERS

// Тимчасове сховище для клієнтів, що авторизуються
var (
	pendingClients      = make(map[string]*tgclient.Client)
	pendingClientsMutex sync.RWMutex
)

func requestAuthCode(c *gin.Context) {
	var req AuthRequest
	if err := c.BindJSON(&req); err != nil {
		c.JSON(400, gin.H{"error": "Invalid request"})
		return
	}

	log.Printf("Auth code requested for: %s", req.Phone)

	// Створюємо Telegram клієнт
	client, err := tgclient.NewClient(appConfig)
	if err != nil {
		c.JSON(500, gin.H{"error": "Failed to create client"})
		return
	}

	// Зберігаємо клієнт для подальшого використання
	pendingClientsMutex.Lock()
	pendingClients[req.Phone] = client
	pendingClientsMutex.Unlock()

	// Запускаємо авторизацію в окремій горутині
	go func() {
		ctx := context.Background()
		if err := client.Auth(ctx, req.Phone); err != nil {
			log.Printf("Auth error: %v", err)
			// Видаляємо клієнт при помилці
			pendingClientsMutex.Lock()
			delete(pendingClients, req.Phone)
			pendingClientsMutex.Unlock()
		}
	}()

	c.JSON(200, gin.H{
		"status":  "code_sent",
		"message": "Код відправлено в Telegram",
	})
}

func login(c *gin.Context) {
	var req AuthCodeRequest
	if err := c.BindJSON(&req); err != nil {
		c.JSON(400, gin.H{"error": "Invalid request"})
		return
	}

	log.Printf("Login: Processing for phone: %s", req.Phone)

	// Відправляємо код авторизації
	if err := tgclient.SubmitCode(req.Phone, req.Code); err != nil {
		c.JSON(401, gin.H{"error": "Invalid code or no pending auth"})
		return
	}

	// Даємо час на обробку - Password() може викликатися пізніше
	time.Sleep(5 * time.Second)

	// Перевіряємо, чи потрібен пароль 2FA
	for i := 0; i < 20; i++ {
		if auth, exists := tgclient.GetPendingAuth(req.Phone); exists && auth.NeedsPassword {
			log.Printf("2FA password required for: %s", req.Phone)
			c.JSON(200, gin.H{
				"status":         "password_required",
				"message":        "Обліковий запис захищено 2FA паролем",
				"needs_password": true,
			})
			return
		}
		time.Sleep(500 * time.Millisecond)
	}

	// Якщо пароль не потрібен - завершуємо авторизацію
	log.Printf("No 2FA password required for: %s, completing login", req.Phone)

	// Очікуємо завершення авторизації
	time.Sleep(2 * time.Second)

	// Очищаємо pending auth
	tgclient.ClearPendingAuth(req.Phone)

	// Отримуємо збережений клієнт
	pendingClientsMutex.Lock()
	client, exists := pendingClients[req.Phone]
	if exists {
		delete(pendingClients, req.Phone)
	}
	pendingClientsMutex.Unlock()

	if !exists {
		log.Printf("Login: ERROR - No pending client for phone: %s", req.Phone)
		c.JSON(500, gin.H{"error": "Auth session expired. Please request code again."})
		return
	}

	// Чекаємо поки session файл буде створений і читаємо його
	var sessionData string
	var err error
	log.Printf("Login: Waiting for session file to be created...")
	for i := 0; i < 20; i++ {
		sessionData, err = client.GetSessionData()
		if err == nil && sessionData != "" {
			log.Printf("Login: Session data ready after %d attempts (%.1f sec)", i+1, float64(i)*0.5)
			break
		}
		if i < 19 {
			time.Sleep(500 * time.Millisecond)
		}
	}

	if sessionData == "" {
		log.Printf("Login: WARNING - Failed to get session data after 10 seconds: %v", err)
	}

	log.Printf("Login: SUCCESS for phone: %s", req.Phone)
	c.JSON(200, gin.H{
		"status":       "success",
		"phone":        req.Phone,
		"session_data": sessionData,
	})
}

func submitPassword(c *gin.Context) {
	var req AuthPasswordRequest
	if err := c.BindJSON(&req); err != nil {
		c.JSON(400, gin.H{"error": "Invalid request"})
		return
	}

	log.Printf("SubmitPassword: Processing for phone: %s", req.Phone)

	// Відправляємо пароль
	if err := tgclient.SubmitPassword(req.Phone, req.Password); err != nil {
		c.JSON(401, gin.H{"error": "No pending password request"})
		return
	}

	// Очищаємо pending auth
	tgclient.ClearPendingAuth(req.Phone)

	// Отримуємо збережений клієнт
	pendingClientsMutex.Lock()
	client, exists := pendingClients[req.Phone]
	if exists {
		delete(pendingClients, req.Phone)
	}
	pendingClientsMutex.Unlock()

	if !exists {
		log.Printf("SubmitPassword: ERROR - No pending client for phone: %s", req.Phone)
		c.JSON(500, gin.H{"error": "Auth session expired. Please request code again."})
		return
	}

	// Чекаємо поки session файл буде створений і читаємо його
	var sessionData string
	var err error
	log.Printf("SubmitPassword: Waiting for session file to be created...")
	for i := 0; i < 20; i++ {
		sessionData, err = client.GetSessionData()
		if err == nil && sessionData != "" {
			log.Printf("SubmitPassword: Session data ready after %d attempts (%.1f sec)", i+1, float64(i)*0.5)
			break
		}
		if i < 19 {
			time.Sleep(500 * time.Millisecond)
		}
	}

	if sessionData == "" {
		log.Printf("SubmitPassword: WARNING - Failed to get session data after 10 seconds: %v", err)
	}

	log.Printf("SubmitPassword: SUCCESS for phone: %s", req.Phone)
	c.JSON(200, gin.H{
		"status":       "success",
		"phone":        req.Phone,
		"session_data": sessionData,
	})
}

// API HANDLERS

func getChats(c *gin.Context) {
	log.Printf("getChats: Starting request")

	user := c.MustGet("user").(*User)
	log.Printf("getChats: User ID: %s, Phone: %s", user.ID, user.Phone)

	user.LastActivity = time.Now()

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	log.Printf("getChats: Calling TelegramClient.GetDialogs")
	dialogs, err := user.TelegramClient.GetDialogs(ctx, 50)
	if err != nil {
		log.Printf("getChats: ERROR - Failed to get dialogs: %v", err)
		c.JSON(500, gin.H{"error": fmt.Sprintf("Failed to get dialogs: %v", err)})
		return
	}

	log.Printf("getChats: Successfully got %d dialogs", len(dialogs))
	c.JSON(200, gin.H{
		"chats": dialogs,
		"count": len(dialogs),
	})
}

func getMessages(c *gin.Context) {
	log.Printf("getMessages: Starting request")

	user := c.MustGet("user").(*User)
	log.Printf("getMessages: User ID: %s", user.ID)

	user.LastActivity = time.Now()

	chatIDStr := c.Param("chat_id")
	chatID, err := strconv.ParseInt(chatIDStr, 10, 64)
	if err != nil {
		log.Printf("getMessages: ERROR - Invalid chat_id: %s", chatIDStr)
		c.JSON(400, gin.H{"error": "Invalid chat_id"})
		return
	}

	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "50"))
	log.Printf("getMessages: Chat ID: %d, Limit: %d", chatID, limit)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	log.Printf("getMessages: Calling TelegramClient.GetMessages")
	messages, err := user.TelegramClient.GetMessages(ctx, chatID, limit)
	if err != nil {
		log.Printf("getMessages: ERROR - Failed to get messages: %v", err)
		c.JSON(500, gin.H{"error": fmt.Sprintf("Failed to get messages: %v", err)})
		return
	}

	log.Printf("getMessages: Successfully got %d messages", len(messages))
	c.JSON(200, gin.H{
		"messages": messages,
		"chat_id":  chatIDStr,
	})
}

func sendMessage(c *gin.Context) {
	log.Printf("sendMessage: Starting request")

	user := c.MustGet("user").(*User)
	user.LastActivity = time.Now()

	var req SendMessageRequest
	if err := c.BindJSON(&req); err != nil {
		log.Printf("sendMessage: ERROR - Invalid request: %v", err)
		c.JSON(400, gin.H{"error": "Invalid request"})
		return
	}

	log.Printf("sendMessage: Chat ID: %s, Text: %s", req.ChatID, req.Text)

	chatID, err := strconv.ParseInt(req.ChatID, 10, 64)
	if err != nil {
		log.Printf("sendMessage: ERROR - Invalid chat_id: %s", req.ChatID)
		c.JSON(400, gin.H{"error": "Invalid chat_id"})
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	log.Printf("sendMessage: Calling TelegramClient.SendMessage")
	messageID, err := user.TelegramClient.SendMessage(ctx, chatID, req.Text)
	if err != nil {
		log.Printf("sendMessage: ERROR - Failed to send message: %v", err)
		c.JSON(500, gin.H{"error": fmt.Sprintf("Failed to send message: %v", err)})
		return
	}

	log.Printf("sendMessage: Successfully sent message, ID: %d", messageID)
	c.JSON(200, gin.H{
		"status":     "sent",
		"message_id": messageID,
		"timestamp":  time.Now(),
	})
}

func markAsRead(c *gin.Context) {
	user := c.MustGet("user").(*User)
	user.LastActivity = time.Now()

	var req MarkReadRequest
	if err := c.BindJSON(&req); err != nil {
		c.JSON(400, gin.H{"error": "Invalid request"})
		return
	}

	chatID, err := strconv.ParseInt(req.ChatID, 10, 64)
	if err != nil {
		c.JSON(400, gin.H{"error": "Invalid chat_id"})
		return
	}

	if len(req.MessageIDs) == 0 {
		c.JSON(400, gin.H{"error": "No message IDs provided"})
		return
	}

	maxID := req.MessageIDs[0]
	for _, id := range req.MessageIDs {
		if id > maxID {
			maxID = id
		}
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := user.TelegramClient.MarkAsRead(ctx, chatID, maxID); err != nil {
		c.JSON(500, gin.H{"error": fmt.Sprintf("Failed to mark as read: %v", err)})
		return
	}

	c.JSON(200, gin.H{"status": "marked_read"})
}

func getPhoto(c *gin.Context) {
	// Намагаємося отримати user з middleware (якщо headers передані)
	var user *User
	if u, exists := c.Get("user"); exists {
		user = u.(*User)
	} else {
		// Якщо немає user з middleware, пробуємо token з query параметрів
		token := c.Query("token")
		if token == "" {
			c.JSON(401, gin.H{"error": "Missing authentication"})
			return
		}

		// Декодуємо token (base64 encoded "phone:session_data")
		decoded, err := base64.StdEncoding.DecodeString(token)
		if err != nil {
			c.JSON(401, gin.H{"error": "Invalid token"})
			return
		}

		parts := string(decoded)
		colonIdx := -1
		for i, ch := range parts {
			if ch == ':' {
				colonIdx = i
				break
			}
		}

		if colonIdx == -1 {
			c.JSON(401, gin.H{"error": "Invalid token format"})
			return
		}

		phone := parts[:colonIdx]
		sessionData := parts[colonIdx+1:]

		// Створюємо клієнт з session data
		client, err := tgclient.NewClientWithSession(appConfig, sessionData)
		if err != nil {
			c.JSON(401, gin.H{"error": "Invalid session data"})
			return
		}

		user = &User{
			ID:             phone,
			Phone:          phone,
			TelegramClient: client,
			LastActivity:   time.Now(),
		}
	}

	user.LastActivity = time.Now()

	chatIDStr := c.Param("chat_id")
	messageIDStr := c.Param("message_id")

	chatID, err := strconv.ParseInt(chatIDStr, 10, 64)
	if err != nil {
		c.JSON(400, gin.H{"error": "Invalid chat_id"})
		return
	}

	messageID, err := strconv.Atoi(messageIDStr)
	if err != nil {
		c.JSON(400, gin.H{"error": "Invalid message_id"})
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	photoData, err := user.TelegramClient.GetPhotoData(ctx, messageID, chatID)
	if err != nil {
		log.Printf("getPhoto: ERROR - Failed to get photo: %v", err)
		c.JSON(500, gin.H{"error": fmt.Sprintf("Failed to get photo: %v", err)})
		return
	}

	// Повертаємо зображення
	c.Data(200, "image/jpeg", photoData)
}

func pollMessages(c *gin.Context) {
	log.Printf("pollMessages: Starting request")

	sessionData := c.GetHeader("X-Session-Data")

	chatIDStr := c.Param("chat_id")
	chatID, err := strconv.ParseInt(chatIDStr, 10, 64)
	if err != nil {
		c.JSON(400, gin.H{"error": "Invalid chat_id"})
		return
	}

	// Отримуємо параметр after_message_id (ID останнього повідомлення яке клієнт має)
	afterMessageIDStr := c.DefaultQuery("after_message_id", "0")
	afterMessageID, err := strconv.Atoi(afterMessageIDStr)
	if err != nil {
		c.JSON(400, gin.H{"error": "Invalid after_message_id"})
		return
	}

	// Таймаут для long polling (30 секунд)
	timeout := 30 * time.Second
	pollTimeout, _ := strconv.Atoi(c.DefaultQuery("timeout", "30"))
	if pollTimeout > 0 && pollTimeout <= 60 {
		timeout = time.Duration(pollTimeout) * time.Second
	}

	log.Printf("pollMessages: Chat ID: %d, After Message ID: %d, Timeout: %v", chatID, afterMessageID, timeout)

	// Створюємо контекст з таймаутом
	pollCtx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	// Запускаємо ticker для перевірки нових повідомлень
	ticker := time.NewTicker(3 * time.Second)
	defer ticker.Stop()

	// Перша перевірка одразу
	checkForNewMessages := func() ([]tgclient.Message, error) {
		// Створюємо новий клієнт для кожної перевірки
		client, err := tgclient.NewClientWithSession(appConfig, sessionData)
		if err != nil {
			return nil, err
		}

		checkCtx, checkCancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer checkCancel()

		return client.GetNewMessages(checkCtx, chatID, afterMessageID, 20)
	}

	for {
		select {
		case <-pollCtx.Done():
			// Таймаут - повертаємо порожній результат
			log.Printf("pollMessages: Timeout - no new messages")
			c.JSON(200, gin.H{
				"has_new":  false,
				"messages": []interface{}{},
			})
			return

		case <-ticker.C:
			// Перевіряємо нові повідомлення
			messages, err := checkForNewMessages()
			if err != nil {
				log.Printf("pollMessages: ERROR - Failed to get messages: %v", err)
				continue
			}

			if len(messages) > 0 {
				log.Printf("pollMessages: Found %d new messages", len(messages))
				c.JSON(200, gin.H{
					"has_new":  true,
					"messages": messages,
				})
				return
			}
		}
	}
}

// MIDDLEWARE
func authMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		log.Printf("authMiddleware: Checking authorization for %s %s", c.Request.Method, c.Request.URL.Path)

		phone := c.GetHeader("X-Phone")
		sessionData := c.GetHeader("X-Session-Data")

		if phone == "" || sessionData == "" {
			log.Printf("authMiddleware: ERROR - Missing required headers (X-Phone or X-Session-Data)")
			c.JSON(401, gin.H{"error": "Missing authentication headers"})
			c.Abort()
			return
		}

		log.Printf("authMiddleware: Phone: %s", phone)

		// Створюємо клієнт з session data
		client, err := tgclient.NewClientWithSession(appConfig, sessionData)
		if err != nil {
			log.Printf("authMiddleware: ERROR - Failed to create client: %v", err)
			c.JSON(401, gin.H{"error": "Invalid session data"})
			c.Abort()
			return
		}

		// Створюємо тимчасовий об'єкт користувача для запиту
		user := &User{
			ID:             phone,
			Phone:          phone,
			TelegramClient: client,
			LastActivity:   time.Now(),
		}

		log.Printf("authMiddleware: User authenticated: %s", user.ID)
		c.Set("user", user)
		c.Next()
	}
}

