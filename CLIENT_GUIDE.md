# Гід для розробників Symbian клієнтів

## ⚠️ ВАЖЛИВА ІНФОРМАЦІЯ ПРО АРХІТЕКТУРУ

**Stateless API** - Сервер НЕ зберігає сесії!

Це означає:
- ✅ Після авторизації отримуєте `phone` та `session_data`
- ✅ ДЛЯ КОЖНОГО API запиту передаєте обидва значення в headers:
  - `X-Phone: +380XXXXXXXXX`
  - `X-Session-Data: base64_encoded_data`
- ✅ Сервер можна перезапускати - сесії не втрачаються
- ⚠️ Кожен запит створює новий Telegram клієнт (займає 2-5 секунд)
- ⚠️ `session_data` може бути великим (кілька KB)

---

## Процес авторизації

### Основний флоу

```
┌─────────────────────────────────────────────────────────────────┐
│                   АВТОРИЗАЦІЯ В TELEGRAM                         │
└─────────────────────────────────────────────────────────────────┘

1. Користувач вводить номер телефону
   │
   ├──> POST /auth/request-code {"phone": "+380..."}
   │
   └──> Відповідь: {"status": "code_sent", "message": "..."}

2. Користувач отримує SMS код в Telegram
   │
   ├──> Показати форму введення коду
   │
   └──> POST /auth/login {"phone": "+380...", "code": "12345"}

3. Перевірка відповіді
   │
   ├──┬──> А) {"status": "success", "phone": "...", "session_data": "..."}
   │  │     └──> Зберегти phone + session_data → Перейти до головного екрану
   │  │
   │  └──> Б) {"status": "password_required", "needs_password": true, ...}
   │        └──> Показати форму введення 2FA пароля
   │             │
   │             └──> POST /auth/password {"phone": "+380...", "password": "..."}
   │                  │
   │                  └──> {"status": "success", "phone": "...", "session_data": "..."}
   │                       └──> Зберегти phone + session_data → Головний екран
```

---

## Реалізація на Java ME (приклад)

### 1. Структура класів

```java
public class TelegramClient {
    private String baseUrl = "http://localhost:8080";
    private String phone = null;
    private String sessionData = null;

    // Авторизація - крок 1
    public void requestCode(String phone) { }

    // Авторизація - крок 2
    public LoginResponse login(String phone, String code) { }

    // Авторизація - крок 3 (якщо потрібно)
    public LoginResponse submitPassword(String phone, String password) { }

    // API методи (всі використовують phone + sessionData в headers)
    public Chat[] getChats() { }
    public Message[] getMessages(String chatId) { }
    public void sendMessage(String chatId, String text) { }
}
```

### 2. Авторизація - запит коду

```java
public void requestCode(String phone) {
    try {
        HttpConnection conn = (HttpConnection) Connector.open(baseUrl + "/auth/request-code");
        conn.setRequestMethod(HttpConnection.POST);
        conn.setRequestProperty("Content-Type", "application/json");

        String json = "{\"phone\":\"" + phone + "\"}";

        OutputStream os = conn.openOutputStream();
        os.write(json.getBytes());
        os.close();

        int responseCode = conn.getResponseCode();
        if (responseCode == 200) {
            // Успіх - показати форму введення коду
            showCodeInputForm();
        }
        conn.close();
    } catch (IOException e) {
        showError("Помилка з'єднання");
    }
}
```

### 3. Авторизація - введення коду

```java
public class LoginResponse {
    public String status;
    public String phone;
    public String sessionData;
    public boolean needsPassword;
    public String message;
}

public LoginResponse login(String phone, String code) {
    try {
        HttpConnection conn = (HttpConnection) Connector.open(baseUrl + "/auth/login");
        conn.setRequestMethod(HttpConnection.POST);
        conn.setRequestProperty("Content-Type", "application/json");

        String json = "{\"phone\":\"" + phone + "\",\"code\":\"" + code + "\"}";

        OutputStream os = conn.openOutputStream();
        os.write(json.getBytes());
        os.close();

        int responseCode = conn.getResponseCode();
        if (responseCode == 200) {
            InputStream is = conn.openInputStream();
            String response = readStream(is);
            is.close();

            LoginResponse lr = parseLoginResponse(response);

            // ВАЖЛИВО: Перевірка на 2FA
            if (lr.needsPassword) {
                // Показати форму введення 2FA пароля
                showPasswordInputForm(phone);
                return lr;
            } else if ("success".equals(lr.status)) {
                // Зберегти токен
                saveSessionToken(lr.sessionToken);
                // Перейти до головного екрану
                showMainScreen();
                return lr;
            }
        }
        conn.close();
    } catch (IOException e) {
        showError("Помилка авторизації");
    }
    return null;
}
```

### 4. Авторизація - введення 2FA пароля

```java
public LoginResponse submitPassword(String phone, String password) {
    try {
        HttpConnection conn = (HttpConnection) Connector.open(baseUrl + "/auth/password");
        conn.setRequestMethod(HttpConnection.POST);
        conn.setRequestProperty("Content-Type", "application/json");

        String json = "{\"phone\":\"" + phone + "\",\"password\":\"" + password + "\"}";

        OutputStream os = conn.openOutputStream();
        os.write(json.getBytes());
        os.close();

        int responseCode = conn.getResponseCode();
        if (responseCode == 200) {
            InputStream is = conn.openInputStream();
            String response = readStream(is);
            is.close();

            LoginResponse lr = parseLoginResponse(response);

            if ("success".equals(lr.status)) {
                // Зберегти дані сесії
                saveSessionData(lr.phone, lr.sessionData);
                this.phone = lr.phone;
                this.sessionData = lr.sessionData;
                // Перейти до головного екрану
                showMainScreen();
                return lr;
            }
        } else if (responseCode == 401) {
            showError("Невірний пароль");
        }
        conn.close();
    } catch (IOException e) {
        showError("Помилка");
    }
    return null;
}
```

### 5. Збереження даних сесії

```java
// Збереження даних сесії
private void saveSessionData(String phone, String sessionData) {
    try {
        RecordStore rs = RecordStore.openRecordStore("TelegramAuth", true);

        // Формат: phone|sessionData
        String combined = phone + "|" + sessionData;
        byte[] data = combined.getBytes();

        if (rs.getNumRecords() == 0) {
            rs.addRecord(data, 0, data.length);
        } else {
            rs.setRecord(1, data, 0, data.length);
        }

        rs.closeRecordStore();
    } catch (RecordStoreException e) {
        // Обробка помилки
    }
}

// Завантаження даних сесії
private String[] loadSessionData() {
    try {
        RecordStore rs = RecordStore.openRecordStore("TelegramAuth", false);
        if (rs.getNumRecords() > 0) {
            byte[] data = rs.getRecord(1);
            rs.closeRecordStore();
            String combined = new String(data);

            // Розділяємо: [0]=phone, [1]=sessionData
            int pipeIndex = combined.indexOf('|');

            if (pipeIndex > 0) {
                String[] parts = new String[2];
                parts[0] = combined.substring(0, pipeIndex);
                parts[1] = combined.substring(pipeIndex + 1);
                return parts;
            }
        }
        rs.closeRecordStore();
    } catch (RecordStoreException e) {
        // Дані не знайдено
    }
    return null;
}

// Очищення збережених даних
private void clearSessionData() {
    try {
        RecordStore.deleteRecordStore("TelegramAuth");
        this.phone = null;
        this.sessionData = null;
    } catch (RecordStoreException e) {
        // Помилка видалення
    }
}
```

### 6. Відновлення сесії при запуску

```java
// При запуску програми
public void onStartup() {
    String[] data = loadSessionData();

    if (data != null) {
        // Є збережені дані - використовуємо їх
        this.phone = data[0];
        this.sessionData = data[1];

        // Переходимо до головного екрану
        // Дані будуть передаватись в кожному API запиті
        showMainScreen();
    } else {
        // Немає збережених даних - показуємо екран логіну
        showLoginScreen();
    }
}
```

### 7. Використання даних сесії в запитах

```java
public Chat[] getChats() {
    try {
        HttpConnection conn = (HttpConnection) Connector.open(baseUrl + "/api/chats");
        conn.setRequestMethod(HttpConnection.GET);

        // ВАЖЛИВО: Додати phone та sessionData в headers
        conn.setRequestProperty("X-Phone", phone);
        conn.setRequestProperty("X-Session-Data", sessionData);

        int responseCode = conn.getResponseCode();
        if (responseCode == 200) {
            InputStream is = conn.openInputStream();
            String response = readStream(is);
            is.close();

            return parseChats(response);
        } else if (responseCode == 401) {
            // Дані сесії недійсні - повернутися до авторизації
            clearSessionData();
            showLoginScreen();
        }
        conn.close();
    } catch (IOException e) {
        showError("Помилка завантаження чатів");
    }
    return null;
}

// Приклад відправки повідомлення
public void sendMessage(String chatId, String text) {
    try {
        HttpConnection conn = (HttpConnection) Connector.open(baseUrl + "/api/send");
        conn.setRequestMethod(HttpConnection.POST);

        // ВАЖЛИВО: Додати headers в КОЖНОМУ запиті
        conn.setRequestProperty("X-Phone", phone);
        conn.setRequestProperty("X-Session-Data", sessionData);
        conn.setRequestProperty("Content-Type", "application/json");

        String json = "{\"chat_id\":\"" + chatId + "\",\"text\":\"" + text + "\"}";

        OutputStream os = conn.openOutputStream();
        os.write(json.getBytes());
        os.close();

        int responseCode = conn.getResponseCode();
        if (responseCode == 200) {
            // Успішно відправлено
            showSuccess("Повідомлення відправлено");
        } else if (responseCode == 401) {
            clearSessionData();
            showLoginScreen();
        }
        conn.close();
    } catch (IOException e) {
        showError("Помилка відправки");
    }
}
```

---

## UI/UX рекомендації

### Екрани авторизації

#### 1. Екран введення номера
```
┌──────────────────────────┐
│   Telegram для Symbian   │
├──────────────────────────┤
│                          │
│  Введіть номер телефону: │
│  ┌────────────────────┐  │
│  │ +380              │  │
│  └────────────────────┘  │
│                          │
│  [Далі]                  │
│                          │
└──────────────────────────┘
```

#### 2. Екран введення коду
```
┌──────────────────────────┐
│   Підтвердження          │
├──────────────────────────┤
│                          │
│  Введіть код з Telegram: │
│  ┌────────────────────┐  │
│  │ _____              │  │
│  └────────────────────┘  │
│                          │
│  [Підтвердити]           │
│  [Змінити номер]         │
│                          │
└──────────────────────────┘
```

#### 3. Екран введення 2FA пароля
```
┌──────────────────────────┐
│   2FA Пароль             │
├──────────────────────────┤
│                          │
│  Обліковий запис         │
│  захищено паролем        │
│                          │
│  Введіть пароль:         │
│  ┌────────────────────┐  │
│  │ ********           │  │
│  └────────────────────┘  │
│                          │
│  [Увійти]                │
│  [Скасувати]             │
│                          │
└──────────────────────────┘
```

---

## Обробка помилок

### Таблиця помилок

| Код | Endpoint | Значення | Дія клієнта |
|-----|----------|----------|-------------|
| 400 | Будь-який | Invalid request | Показати "Невірний формат запиту" |
| 401 | `/auth/login` | Invalid code | "Невірний код. Спробуйте ще раз" |
| 401 | `/auth/password` | Invalid password | "Невірний пароль. Спробуйте ще раз" |
| 401 | `/api/*` | Invalid session | Повернутися до авторизації |
| 500 | Будь-який | Server error | "Помилка сервера. Спробуйте пізніше" |

### Приклад обробки

```java
private void handleHttpError(int code, String endpoint) {
    switch (code) {
        case 401:
            if (endpoint.contains("/auth/")) {
                showError("Невірні дані авторизації");
            } else {
                // Токен застарів
                clearSessionToken();
                showLoginScreen();
            }
            break;
        case 400:
            showError("Невірний запит");
            break;
        case 500:
            showError("Помилка сервера");
            break;
        default:
            showError("Невідома помилка");
    }
}
```

---

## Стани UI

### Діаграма станів

```
        [Запуск програми]
                │
                ▼
        ┌──────────────┐
        │ Перевірка    │
        │ токену       │
        └──────┬───────┘
               │
        ┌──────┴──────┐
        │             │
    [Є токен]    [Немає токену]
        │             │
        ▼             ▼
  ┌─────────┐   ┌──────────┐
  │ Головний│   │  Екран   │
  │  екран  │   │  логіну  │
  └─────────┘   └────┬─────┘
        ▲            │
        │            ▼
        │      ┌──────────┐
        │      │ Введення │
        │      │  коду    │
        │      └────┬─────┘
        │           │
        │    ┌──────┴──────┐
        │    │             │
        │ [Успіх]   [Потрібен пароль]
        │    │             │
        │    │             ▼
        │    │       ┌──────────┐
        │    │       │ Введення │
        │    │       │ пароля   │
        │    │       └────┬─────┘
        │    │            │
        │    └────────────┘
        │         [Успіх]
        └────────────┘
```

---

## Checklist для розробників

### Авторизація
- [ ] Форма введення номера телефону
- [ ] Валідація номера (формат +XXXXXXXXXXX)
- [ ] Запит коду через `/auth/request-code`
- [ ] Форма введення SMS коду
- [ ] Відправка коду через `/auth/login`
- [ ] Перевірка поля `needs_password` у відповіді
- [ ] Форма введення 2FA пароля (якщо потрібно)
- [ ] Відправка пароля через `/auth/password`
- [ ] Збереження `phone` та `session_data` в RecordStore
- [ ] Обробка помилок (401, 400, 500)

### Головний екран
- [ ] Завантаження даних з RecordStore
- [ ] Запит чатів через `/api/chats` з headers (X-Phone, X-Session-Data)
- [ ] Відображення списку чатів
- [ ] Обробка 401 (сесія недійсна → логін)
- [ ] Pull-to-refresh або кнопка оновлення

### Екран чату
- [ ] Завантаження повідомлень через `/api/messages/:chat_id` з headers
- [ ] Відображення історії повідомлень
- [ ] Форма введення повідомлення
- [ ] Відправка через `/api/send` з headers
- [ ] Періодичне оновлення повідомлень (polling)

### Налаштування
- [ ] Кнопка "Вийти" → очищення RecordStore
- [ ] Очищення phone та sessionData при виході

---

## Поради з оптимізації

### 1. Управління пам'яттю
- Обмежте кількість повідомлень в пам'яті (100-200 останніх)
- Очищайте старі повідомлення при завантаженні нових
- Використовуйте `System.gc()` після великих операцій

### 2. Економія трафіку
- Не перезавантажуйте чати при кожному оновленні екрану
- Кешуйте список чатів локально
- Використовуйте `limit` параметр (20-30 повідомлень)
- **ВАЖЛИВО**: `session_data` може бути великим (кілька KB) - враховуйте це при частих запитах

### 3. Користувацький досвід
- Показуйте індикатор завантаження під час HTTP запитів
- **ВАЖЛИВО**: Кожен запит займає 2-5 секунд (створення Telegram клієнта)
- Дозволяйте користувачу скасувати довгі операції
- Зберігайте стан екрану (позицію скролу тощо)

---

## Тестування

### Тестові сценарії

1. **Авторизація без 2FA**
   - Введення номера → Отримання коду → Вхід → Успіх

2. **Авторизація з 2FA**
   - Введення номера → Отримання коду → Вхід → Запит пароля → Введення пароля → Успіх

3. **Невірний код**
   - Введення номера → Введення невірного коду → Помилка → Повторна спроба

4. **Невірний пароль**
   - Авторизація до запиту пароля → Невірний пароль → Помилка → Повторна спроба

5. **Застарілі дані сесії**
   - Збережені дані → Запит чатів → 401 → Очищення даних → Повернення до логіну

6. **Втрата з'єднання**
   - Під час будь-якого запиту → IOException → Показати помилку

---

## Підтримка

Якщо виникли питання:
1. Перегляньте API.md для деталей про endpoints
2. Перевірте SETUP.md для налаштування сервера
3. Створіть Issue на GitHub з описом проблеми

---

**Версія:** 1.0
**Дата:** Жовтень 2025
