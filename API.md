# Telegram Gateway API - Документація для клієнта

## Огляд

Telegram Gateway API надає HTTP інтерфейс для взаємодії з Telegram для Symbian-пристроїв. API дозволяє авторизуватися, отримувати чати, читати та відправляти повідомлення.

**Base URL:** `http://localhost:8080`

---

## Авторизація

### 1. Запит коду авторизації

Відправляє SMS-код на номер телефону через Telegram.

**Endpoint:** `POST /auth/request-code`

**Request Body:**
```json
{
  "phone": "+380XXXXXXXXX"
}
```

**Response (200 OK):**
```json
{
  "status": "code_sent",
  "message": "Код відправлено в Telegram"
}
```

**Приклад:**
```bash
curl -X POST http://localhost:8080/auth/request-code \
  -H "Content-Type: application/json" \
  -d '{"phone": "+380XXXXXXXXX"}'
```

---

### 2. Вхід з кодом

Завершує авторизацію та отримує session token, або повідомляє про необхідність 2FA пароля.

**Endpoint:** `POST /auth/login`

**Request Body:**
```json
{
  "phone": "+380XXXXXXXXX",
  "code": "12345"
}
```

**Response (200 OK) - Успішний вхід:**
```json
{
  "status": "success",
  "phone": "+380XXXXXXXXX",
  "session_data": "base64_encoded_telegram_session_data..."
}
```

⚠️ **ВАЖЛИВО:**
- Збережіть `phone` та `session_data` - вони потрібні для КОЖНОГО наступного запиту!
- `session_data` містить Telegram сесію в base64
- Сервер НЕ зберігає сесії - всі дані потрібно передавати в кожному запиті

**Response (200 OK) - Потрібен пароль 2FA:**
```json
{
  "status": "password_required",
  "message": "Обліковий запис захищено 2FA паролем",
  "needs_password": true
}
```

**Response (401 Unauthorized):**
```json
{
  "error": "Invalid code or no pending auth"
}
```

**Приклад:**
```bash
curl -X POST http://localhost:8080/auth/login \
  -H "Content-Type: application/json" \
  -d '{"phone": "+380XXXXXXXXX", "code": "12345"}'
```


---

### 3. Вхід з паролем 2FA

Використовується коли обліковий запис захищений двофакторною аутентифікацією (2FA).

**Endpoint:** `POST /auth/password`

**Request Body:**
```json
{
  "phone": "+380XXXXXXXXX",
  "password": "your_2fa_password"
}
```

**Response (200 OK):**
```json
{
  "status": "success",
  "phone": "+380XXXXXXXXX",
  "session_data": "base64_encoded_telegram_session_data..."
}
```

**Response (401 Unauthorized):**
```json
{
  "error": "No pending password request"
}
```

**Приклад:**
```bash
curl -X POST http://localhost:8080/auth/password \
  -H "Content-Type: application/json" \
  -d '{"phone": "+380XXXXXXXXX", "password": "my_password"}'
```

---

## API для роботи з чатами та повідомленнями

Всі endpoint'и нижче вимагають два headers:
- `X-Phone: <phone_number>` - номер телефону
- `X-Session-Data: <base64_session_data>` - дані Telegram сесії

### 6. Отримання списку чатів

Отримує список діалогів (чатів) користувача.

**Endpoint:** `GET /api/chats`

**Headers:**
- `X-Phone: +380XXXXXXXXX`
- `X-Session-Data: base64_encoded_data`

**Response (200 OK):**
```json
{
  "chats": [
    {
      "id": 123456789,
      "name": "John Doe",
      "last_message": "See you tomorrow!",
      "unread_count": 3,
      "last_update_time": "2025-10-06T14:30:00Z",
      "type": "user"
    },
    {
      "id": 987654321,
      "name": "Family Group",
      "last_message": "👍",
      "unread_count": 0,
      "last_update_time": "2025-10-06T12:15:00Z",
      "type": "chat"
    }
  ],
  "count": 2
}
```

**Поля чату:**
- `id` (int64) - унікальний ідентифікатор чату
- `name` (string) - назва чату або ім'я користувача
- `last_message` (string) - текст останнього повідомлення
- `unread_count` (int) - кількість непрочитаних повідомлень
- `last_update_time` (string) - час останнього оновлення (ISO 8601)
- `type` (string) - тип: "user", "chat", "channel"

**Приклад:**
```bash
curl -X GET http://localhost:8080/api/chats \
  -H "X-Phone: +380XXXXXXXXX" \
  -H "X-Session-Data: eyJkY19pZCI6Miwic2Vzc2lvbl9rZXkiOi4uLn0="
```

---

### 7. Отримання повідомлень з чату

Отримує історію повідомлень конкретного чату.

**Endpoint:** `GET /api/messages/:chat_id`

**Headers:**
- `X-Phone: +380XXXXXXXXX`
- `X-Session-Data: base64_encoded_data`

**URL Parameters:**
- `chat_id` (required) - ID чату з списку чатів

**Query Parameters:**
- `limit` (optional, default: 50) - кількість повідомлень

**Response (200 OK):**
```json
{
  "messages": [
    {
      "id": 1001,
      "chat_id": "123456789",
      "chat_name": "",
      "text": "Hello from Symbian!",
      "sender": "You",
      "timestamp": "2025-10-06T14:25:00Z",
      "is_read": true,
      "out": true
    },
    {
      "id": 1000,
      "chat_id": "123456789",
      "chat_name": "",
      "text": "Hi! How are you?",
      "sender": "@johndoe",
      "timestamp": "2025-10-06T14:20:00Z",
      "is_read": true,
      "out": false
    }
  ],
  "chat_id": "123456789"
}
```

**Поля повідомлення:**
- `id` (int) - ID повідомлення
- `chat_id` (string) - ID чату
- `text` (string) - текст повідомлення
- `sender` (string) - ім'я відправника
- `timestamp` (string) - час відправки (ISO 8601)
- `is_read` (bool) - чи прочитане повідомлення
- `out` (bool) - чи це вихідне повідомлення (від вас)

**Приклад:**
```bash
curl -X GET "http://localhost:8080/api/messages/123456789?limit=20" \
  -H "X-Phone: +380XXXXXXXXX" \
  -H "X-Session-Data: eyJkY19pZCI6Miwic2Vzc2lvbl9rZXkiOi4uLn0="
```

---

### 8. Відправка повідомлення

Відправляє текстове повідомлення в чат.

**Endpoint:** `POST /api/send`

**Headers:**
- `X-Phone: +380XXXXXXXXX`
- `X-Session-Data: base64_encoded_data`
- `Content-Type: application/json`

**Request Body:**
```json
{
  "chat_id": "123456789",
  "text": "Hello from Nokia!"
}
```

**Response (200 OK):**
```json
{
  "status": "sent",
  "message_id": 1002,
  "timestamp": "2025-10-06T14:30:00Z"
}
```

**Response (400 Bad Request):**
```json
{
  "error": "Invalid chat_id"
}
```

**Приклад:**
```bash
curl -X POST http://localhost:8080/api/send \
  -H "X-Phone: +380XXXXXXXXX" \
  -H "X-Session-Data: eyJkY19pZCI6Miwic2Vzc2lvbl9rZXkiOi4uLn0=" \
  -H "Content-Type: application/json" \
  -d '{"chat_id": "123456789", "text": "Hello from Nokia!"}'
```

---

### 9. Позначити повідомлення як прочитані

Позначає повідомлення в чаті як прочитані.

**Endpoint:** `POST /api/mark-read`

**Headers:**
- `X-Phone: +380XXXXXXXXX`
- `X-Session-Data: base64_encoded_data`
- `Content-Type: application/json`

**Request Body:**
```json
{
  "chat_id": "123456789",
  "message_ids": [1000, 1001, 1002]
}
```

**Response (200 OK):**
```json
{
  "status": "marked_read"
}
```

**Приклад:**
```bash
curl -X POST http://localhost:8080/api/mark-read \
  -H "X-Phone: +380XXXXXXXXX" \
  -H "X-Session-Data: eyJkY19pZCI6Miwic2Vzc2lvbl9rZXkiOi4uLn0=" \
  -H "Content-Type: application/json" \
  -d '{"chat_id": "123456789", "message_ids": [1000, 1001]}'
```

---

## Коди помилок

| Код | Значення | Опис |
|-----|----------|------|
| 200 | OK | Запит виконано успішно |
| 204 | No Content | Long polling таймаут без нових даних |
| 400 | Bad Request | Невірний формат запиту або параметри |
| 401 | Unauthorized | Невірний або відсутній session token |
| 500 | Internal Server Error | Помилка на сервері |

**Приклади помилок:**

```json
{
  "error": "No authorization token"
}
```

```json
{
  "error": "Invalid session"
}
```

```json
{
  "error": "Invalid chat_id"
}
```

---

## Приклад повного flow авторизації та відправки повідомлення

### Варіант A: Авторизація без 2FA

#### Крок 1: Запит коду
```bash
curl -X POST http://localhost:8080/auth/request-code \
  -H "Content-Type: application/json" \
  -d '{"phone": "+380501234567"}'
```

#### Крок 2: Введення коду (отриманого в Telegram) та вхід
```bash
curl -X POST http://localhost:8080/auth/login \
  -H "Content-Type: application/json" \
  -d '{"phone": "+380501234567", "code": "12345"}'
```

Відповідь (успіх):
```json
{
  "status": "success",
  "phone": "+380501234567",
  "session_data": "eyJkY19pZCI6Miwic2Vzc2lvbl9rZXkiOi4uLn0="
}
```

### Варіант B: Авторизація з 2FA

#### Крок 1: Запит коду
```bash
curl -X POST http://localhost:8080/auth/request-code \
  -H "Content-Type: application/json" \
  -d '{"phone": "+380501234567"}'
```

#### Крок 2: Введення коду
```bash
curl -X POST http://localhost:8080/auth/login \
  -H "Content-Type: application/json" \
  -d '{"phone": "+380501234567", "code": "12345"}'
```

Відповідь (потрібен пароль):
```json
{
  "status": "password_required",
  "message": "Обліковий запис захищено 2FA паролем",
  "needs_password": true
}
```

#### Крок 3: Введення 2FA пароля
```bash
curl -X POST http://localhost:8080/auth/password \
  -H "Content-Type: application/json" \
  -d '{"phone": "+380501234567", "password": "my_2fa_password"}'
```

Відповідь (успіх):
```json
{
  "status": "success",
  "phone": "+380501234567",
  "session_data": "eyJkY19pZCI6Miwic2Vzc2lvbl9rZXkiOi4uLn0="
}
```

---

### Крок 3: Отримання чатів
```bash
curl -X GET http://localhost:8080/api/chats \
  -H "X-Phone: +380501234567" \
  -H "X-Session-Data: eyJkY19pZCI6Miwic2Vzc2lvbl9rZXkiOi4uLn0="
```

### Крок 4: Читання повідомлень
```bash
curl -X GET http://localhost:8080/api/messages/123456789 \
  -H "X-Phone: +380501234567" \
  -H "X-Session-Data: eyJkY19pZCI6Miwic2Vzc2lvbl9rZXkiOi4uLn0="
```

### Крок 5: Відправка повідомлення
```bash
curl -X POST http://localhost:8080/api/send \
  -H "X-Phone: +380501234567" \
  -H "X-Session-Data: eyJkY19pZCI6Miwic2Vzc2lvbl9rZXkiOi4uLn0=" \
  -H "Content-Type: application/json" \
  -d '{"chat_id": "123456789", "text": "Привіт з Nokia!"}'
```

---

## Рекомендації для Symbian клієнтів

### 1. Збереження даних сесії
Зберігайте `phone` та `session_data` в RecordStore (Java ME) або файлі:
- `phone` - номер телефону
- `session_data` - дані Telegram сесії в base64

**При запуску клієнта:**
1. Перевірити чи є збережені дані
2. Якщо є - використовувати їх для API запитів (передавати в headers)
3. Якщо немає або 401 помилка - показати екран авторизації

### 2. Headers для кожного запиту
ДЛЯ КОЖНОГО API запиту передавайте:
```
X-Phone: +380XXXXXXXXX
X-Session-Data: base64_encoded_data
```

### 3. Обробка помилок
- При 401 (Missing authentication headers / Invalid session data) → повторна авторизація
- При 500 → повторний запит через 5 секунд
- При мережевих помилках → повторний запит

### 4. Економія трафіку
- Кешуйте список чатів локально
- Завантажуйте повідомлення порціями (limit=20-30)

### 5. Обробка 2FA
- Після `/auth/login` перевіряйте поле `needs_password`
- Якщо `needs_password: true` → покажіть форму введення пароля
- Відправте пароль через `/auth/password`
- Обробляйте невірний пароль (помилка 401)

---

## Обмеження

- **Максимальна довжина повідомлення:** ~4096 символів
- **Session timeout:** Сесії НЕ зберігаються на сервері - передавайте дані в кожному запиті
- **Підтримувані типи повідомлень:** тільки текстові
- **Медіафайли:** не підтримуються в MVP

---

## Налагодження

### Логи сервера
Сервер виводить детальні логи в консоль:
```
Starting Telegram Gateway Server
API ID: 27644715
Server: localhost:8080
Server starting on localhost:8080
Auth code requested for: +380XXXXXXXXX
New message: Hello!
```

### Перевірка з'єднання
```bash
curl http://localhost:8080/
```

Якщо сервер працює, отримаєте відповідь 404 (це нормально, просто означає, що сервер працює).

---

## Версія API

**Поточна версія:** 1.0
**Дата оновлення:** Жовтень 2025

---

## Підтримка

Якщо виникли питання або проблеми:
1. Перевірте логи сервера
2. Переконайтеся, що `.env` файл правильно налаштований
3. Перевірте, що Telegram API credentials коректні
