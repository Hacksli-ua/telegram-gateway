# Інструкція по налаштуванню та запуску

## Швидкий старт

### 1. Встановлення Go

Завантажте і встановіть Go 1.20+ з офіційного сайту: https://go.dev/dl/

Перевірте встановлення:
```bash
go version
```

### 2. Клонування/Завантаження проекту

```bash
cd C:\Users\YourName\Desktop\
git clone <repository-url> telegram-gateway
# або розпакуйте архів
```

### 3. Налаштування Telegram API

1. Перейдіть на https://my.telegram.org
2. Увійдіть зі своїм номером телефону
3. Перейдіть в "API development tools"
4. Створіть новий додаток
5. Збережіть `api_id` та `api_hash`

### 4. Налаштування конфігурації

Файл `.env` вже налаштований з вашими credentials:

```env
TELEGRAM_API_ID=27644715
TELEGRAM_API_HASH=d189d36c471fa1008e7f3f6cc2b373f8
TELEGRAM_PUBLIC_KEY_FILE=telegram_public_key.pem
TELEGRAM_DC_ID=2
TELEGRAM_SERVER_HOST=149.154.167.50
TELEGRAM_SERVER_PORT=443

SERVER_PORT=8080
SERVER_HOST=localhost

SESSION_TIMEOUT=30m
CLEANUP_INTERVAL=5m
POLL_TIMEOUT=50s
```

**Не потрібно нічого змінювати!** Все вже налаштовано.

### 5. Встановлення залежностей

```bash
cd telegram-gateway
go mod tidy
```

### 6. Запуск сервера

#### Варіант A: Режим розробки (через go run)
```bash
start.bat
```

#### Варіант B: Production (компіляція)
```bash
build.bat
run.bat
```

### 7. Перевірка

Якщо все працює, ви побачите:
```
Starting Telegram Gateway Server
API ID: 27644715
Server: localhost:8080
[GIN-debug] Listening and serving HTTP on localhost:8080
```

Сервер готовий до роботи! 🎉

---

## Структура проекту

```
telegram-gateway/
├── main.go                   # Головний HTTP сервер
├── config/
│   └── config.go            # Завантаження конфігурації
├── telegram/
│   ├── client.go            # Telegram клієнт
│   ├── auth.go              # Авторизація
│   ├── dialogs.go           # Отримання чатів
│   ├── messages.go          # Повідомлення
│   └── updates.go           # Real-time оновлення
├── .env                     # Конфігурація (ваші credentials)
├── .env.example             # Приклад конфігурації
├── telegram_public_key.pem  # RSA ключ
├── go.mod                   # Go залежності
├── go.sum                   # Checksums залежностей
├── start.bat                # Запуск в режимі розробки
├── build.bat                # Компіляція
├── run.bat                  # Запуск скомпільованого бінарника
├── README.md                # Основна документація
├── API.md                   # Документація API
└── bin/
    └── telegram-gateway.exe # Скомпільований бінарник
```

---

## Тестування API

### Використання curl

#### 1. Запит коду авторизації
```bash
curl -X POST http://localhost:8080/auth/request-code \
  -H "Content-Type: application/json" \
  -d "{\"phone\": \"+380XXXXXXXXX\"}"
```

Ви отримаєте SMS код в Telegram.

#### 2. Вхід з кодом
```bash
curl -X POST http://localhost:8080/auth/login \
  -H "Content-Type: application/json" \
  -d "{\"phone\": \"+380XXXXXXXXX\", \"code\": \"12345\"}"
```

Отримаєте `session_token`.

#### 3. Отримання чатів
```bash
curl -X GET http://localhost:8080/api/chats \
  -H "Authorization: YOUR_SESSION_TOKEN"
```

#### 4. Відправка повідомлення
```bash
curl -X POST http://localhost:8080/api/send \
  -H "Authorization: YOUR_SESSION_TOKEN" \
  -H "Content-Type: application/json" \
  -d "{\"chat_id\": \"123456789\", \"text\": \"Hello!\"}"
```

### Використання Postman

1. Імпортуйте колекцію з файлу `postman_collection.json` (якщо є)
2. Встановіть змінну `base_url` = `http://localhost:8080`
3. Встановіть змінну `session_token` після авторизації
4. Виконуйте запити

---

## Налагодження

### Проблема: Сервер не запускається

**Помилка:** `bind: address already in use`

**Рішення:** Порт 8080 зайнятий. Змініть порт в `.env`:
```env
SERVER_PORT=8081
```

**Помилка:** `TELEGRAM_API_ID must be set`

**Рішення:** Перевірте, що файл `.env` існує і містить правильні значення.

### Проблема: Не приходить SMS код

**Можливі причини:**
1. Невірний номер телефону (має бути з `+` та кодом країни)
2. Невірні `api_id` або `api_hash`
3. Telegram тимчасово заблокував ваш IP (зачекайте 5-10 хвилин)

### Проблема: Long polling не працює

**Рішення:** Збільште HTTP timeout на клієнті до 60 секунд.

### Проблема: "Invalid session"

**Рішення:**
- Session token застарів (30 хвилин неактивності)
- Повторна авторизація через `/auth/login`

---

## Логи

Сервер виводить детальні логи в консоль:

```
[GIN] 2025/10/06 - 15:30:45 | 200 |   1.234567ms |  127.0.0.1 | POST     "/auth/login"
Auth code requested for: +380XXXXXXXXX
New message: Hello from Telegram!
[GIN] 2025/10/06 - 15:31:00 | 200 |   50.123456s |  127.0.0.1 | GET      "/api/poll"
```

---

## Production Deployment

### Використання з HTTPS

1. Отримайте SSL сертифікат (Let's Encrypt)
2. Оновіть `.env`:
```env
ENABLE_HTTPS=true
SSL_CERT_PATH=/path/to/cert.pem
SSL_KEY_PATH=/path/to/key.pem
```

3. Змініте `SERVER_HOST` на зовнішню адресу:
```env
SERVER_HOST=0.0.0.0
```

### Запуск як сервіс (Windows)

Використовуйте NSSM (Non-Sucking Service Manager):

1. Завантажте NSSM: https://nssm.cc/download
2. Встановіть сервіс:
```bash
nssm install TelegramGateway "C:\path\to\telegram-gateway\bin\telegram-gateway.exe"
nssm set TelegramGateway AppDirectory "C:\path\to\telegram-gateway"
nssm start TelegramGateway
```

### Запуск як сервіс (Linux)

Створіть systemd service файл `/etc/systemd/system/telegram-gateway.service`:

```ini
[Unit]
Description=Telegram Gateway Server
After=network.target

[Service]
Type=simple
User=telegram
WorkingDirectory=/opt/telegram-gateway
ExecStart=/opt/telegram-gateway/telegram-gateway
Restart=on-failure

[Install]
WantedBy=multi-user.target
```

Запустіть:
```bash
sudo systemctl enable telegram-gateway
sudo systemctl start telegram-gateway
sudo systemctl status telegram-gateway
```

---

## Оновлення

### Оновлення залежностей
```bash
go get -u ./...
go mod tidy
```

### Оновлення gotd/td
```bash
go get -u github.com/gotd/td@latest
go mod tidy
```

### Перекомпіляція
```bash
build.bat
```

---

## Backup

### Важливі файли для backup:

1. `.env` - ваші credentials
2. `telegram_public_key.pem` - RSA ключ
3. `session.json` - Telegram сесія (створюється автоматично)

### Створення backup:
```bash
mkdir backup
copy .env backup\
copy telegram_public_key.pem backup\
copy session.json backup\
```

---

## Безпека

⚠️ **ВАЖЛИВО:**

1. **Не публікуйте `.env` в Git!** Файл `.gitignore` вже налаштований.
2. **Зберігайте `api_id` і `api_hash` в секреті**
3. **Використовуйте HTTPS в production**
4. **Змінюйте `CORS_ALLOWED_ORIGINS` для production**:
   ```env
   CORS_ALLOWED_ORIGINS=https://yourdomain.com
   ```

---

## Підтримка

Якщо виникли проблеми:

1. Перевірте логи сервера
2. Перегляньте документацію API в `API.md`
3. Перевірте, що всі залежності встановлені: `go mod tidy`
4. Створіть Issue на GitHub

---

## Корисні посилання

- Telegram API документація: https://core.telegram.org/
- gotd/td бібліотека: https://github.com/gotd/td
- Go документація: https://go.dev/doc/
- Gin framework: https://gin-gonic.com/

---

**Версія:** 1.0
**Дата:** Жовтень 2025
