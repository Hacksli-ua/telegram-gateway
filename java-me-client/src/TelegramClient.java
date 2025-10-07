import javax.microedition.io.*;
import java.io.*;

/**
 * Клієнт для роботи з Telegram Gateway API
 */
public class TelegramClient {
    private static final String API_BASE = "http://localhost:8080";

    private String phone;
    private String sessionData;

    public TelegramClient() {
    }

    public void setSession(String phone, String sessionData) {
        this.phone = phone;
        this.sessionData = sessionData;
    }

    public void clearSession() {
        this.phone = null;
        this.sessionData = null;
    }

    public boolean hasSession() {
        return phone != null && sessionData != null;
    }

    /**
     * Запит коду авторизації
     */
    public String requestCode(String phoneNumber) throws Exception {
        String url = API_BASE + "/auth/request-code";
        String body = "{\"phone\":\"" + phoneNumber + "\"}";

        String response = httpPost(url, body, null, null);
        return response;
    }

    /**
     * Вхід з кодом
     */
    public AuthResult login(String phoneNumber, String code) throws Exception {
        String url = API_BASE + "/auth/login";
        String body = "{\"phone\":\"" + phoneNumber + "\",\"code\":\"" + code + "\"}";

        String response = httpPost(url, body, null, null);
        return parseAuthResult(response);
    }

    /**
     * Вхід з паролем 2FA
     */
    public AuthResult submitPassword(String phoneNumber, String password) throws Exception {
        String url = API_BASE + "/auth/password";
        String body = "{\"phone\":\"" + phoneNumber + "\",\"password\":\"" + password + "\"}";

        String response = httpPost(url, body, null, null);
        return parseAuthResult(response);
    }

    /**
     * Отримати список чатів
     */
    public Chat[] getChats() throws Exception {
        String url = API_BASE + "/api/chats";
        String response = httpGet(url, phone, sessionData);
        return parseChats(response);
    }

    /**
     * Отримати повідомлення чату
     */
    public Message[] getMessages(String chatId, int limit) throws Exception {
        String url = API_BASE + "/api/messages/" + chatId + "?limit=" + limit;
        String response = httpGet(url, phone, sessionData);
        return parseMessages(response);
    }

    /**
     * Відправити повідомлення
     */
    public void sendMessage(String chatId, String text) throws Exception {
        String url = API_BASE + "/api/send";
        String body = "{\"chat_id\":\"" + chatId + "\",\"text\":\"" + escapeJson(text) + "\"}";
        httpPost(url, body, phone, sessionData);
    }

    /**
     * Long polling для нових повідомлень
     */
    public Message[] pollMessages(String chatId, int afterMessageId) throws Exception {
        String url = API_BASE + "/api/poll/" + chatId + "?after_message_id=" + afterMessageId + "&timeout=30";
        String response = httpGet(url, phone, sessionData);
        return parsePollResult(response);
    }

    // HTTP методи

    private String httpGet(String url, String phone, String sessionData) throws Exception {
        HttpConnection conn = null;
        InputStream is = null;

        try {
            conn = (HttpConnection) Connector.open(url);
            conn.setRequestMethod(HttpConnection.GET);

            if (phone != null && sessionData != null) {
                conn.setRequestProperty("X-Phone", phone);
                conn.setRequestProperty("X-Session-Data", sessionData);
            }

            int rc = conn.getResponseCode();
            if (rc != HttpConnection.HTTP_OK) {
                throw new Exception("HTTP Error: " + rc);
            }

            is = conn.openInputStream();
            return readStream(is);

        } finally {
            if (is != null) try { is.close(); } catch (Exception e) {}
            if (conn != null) try { conn.close(); } catch (Exception e) {}
        }
    }

    private String httpPost(String url, String body, String phone, String sessionData) throws Exception {
        HttpConnection conn = null;
        OutputStream os = null;
        InputStream is = null;

        try {
            conn = (HttpConnection) Connector.open(url);
            conn.setRequestMethod(HttpConnection.POST);
            conn.setRequestProperty("Content-Type", "application/json");

            if (phone != null && sessionData != null) {
                conn.setRequestProperty("X-Phone", phone);
                conn.setRequestProperty("X-Session-Data", sessionData);
            }

            byte[] bodyBytes = body.getBytes("UTF-8");
            conn.setRequestProperty("Content-Length", String.valueOf(bodyBytes.length));

            os = conn.openOutputStream();
            os.write(bodyBytes);
            os.flush();

            int rc = conn.getResponseCode();
            if (rc != HttpConnection.HTTP_OK) {
                throw new Exception("HTTP Error: " + rc);
            }

            is = conn.openInputStream();
            return readStream(is);

        } finally {
            if (os != null) try { os.close(); } catch (Exception e) {}
            if (is != null) try { is.close(); } catch (Exception e) {}
            if (conn != null) try { conn.close(); } catch (Exception e) {}
        }
    }

    private String readStream(InputStream is) throws IOException {
        StringBuffer sb = new StringBuffer();
        int ch;
        while ((ch = is.read()) != -1) {
            sb.append((char) ch);
        }
        return sb.toString();
    }

    // Парсинг JSON (простий парсер без бібліотек)

    private AuthResult parseAuthResult(String json) {
        AuthResult result = new AuthResult();

        result.status = getJsonValue(json, "status");
        result.phone = getJsonValue(json, "phone");
        result.sessionData = getJsonValue(json, "session_data");

        String needsPassword = getJsonValue(json, "needs_password");
        result.needsPassword = "true".equals(needsPassword);

        return result;
    }

    private Chat[] parseChats(String json) {
        // Спрощений парсинг - рахуємо кількість чатів
        int count = countJsonArrayItems(json, "chats");
        Chat[] chats = new Chat[count];

        // TODO: Реалізувати повний парсинг
        // Поки повертаємо порожній масив
        return chats;
    }

    private Message[] parseMessages(String json) {
        int count = countJsonArrayItems(json, "messages");
        Message[] messages = new Message[count];
        // TODO: Реалізувати парсинг
        return messages;
    }

    private Message[] parsePollResult(String json) {
        String hasNew = getJsonValue(json, "has_new");
        if (!"true".equals(hasNew)) {
            return new Message[0];
        }
        return parseMessages(json);
    }

    // Утиліти

    private String getJsonValue(String json, String key) {
        String searchKey = "\"" + key + "\":";
        int start = json.indexOf(searchKey);
        if (start == -1) return null;

        start += searchKey.length();

        // Пропускаємо пробіли
        while (start < json.length() && json.charAt(start) == ' ') {
            start++;
        }

        if (start >= json.length()) return null;

        // Перевіряємо тип значення
        char firstChar = json.charAt(start);

        if (firstChar == '"') {
            // Рядок
            start++;
            int end = json.indexOf('"', start);
            if (end == -1) return null;
            return json.substring(start, end);
        } else if (firstChar == 't' || firstChar == 'f') {
            // Boolean
            int end = start;
            while (end < json.length() && Character.isLetter(json.charAt(end))) {
                end++;
            }
            return json.substring(start, end);
        }

        return null;
    }

    private int countJsonArrayItems(String json, String arrayKey) {
        // Дуже спрощене підрахування
        String searchKey = "\"" + arrayKey + "\":[";
        int start = json.indexOf(searchKey);
        if (start == -1) return 0;

        int count = 0;
        int depth = 0;
        boolean inArray = false;

        for (int i = start + searchKey.length(); i < json.length(); i++) {
            char c = json.charAt(i);
            if (c == '[') {
                depth++;
                inArray = true;
            } else if (c == ']') {
                depth--;
                if (depth == 0) break;
            } else if (c == '{' && depth == 1) {
                count++;
            }
        }

        return count;
    }

    private String escapeJson(String str) {
        StringBuffer sb = new StringBuffer();
        for (int i = 0; i < str.length(); i++) {
            char c = str.charAt(i);
            if (c == '"') {
                sb.append("\\\"");
            } else if (c == '\\') {
                sb.append("\\\\");
            } else if (c == '\n') {
                sb.append("\\n");
            } else if (c == '\r') {
                sb.append("\\r");
            } else {
                sb.append(c);
            }
        }
        return sb.toString();
    }
}
