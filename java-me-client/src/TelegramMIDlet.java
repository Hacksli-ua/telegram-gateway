import javax.microedition.midlet.*;
import javax.microedition.lcdui.*;

/**
 * Головний MIDlet для Telegram Gateway клієнта
 */
public class TelegramMIDlet extends MIDlet implements CommandListener {
    private Display display;
    private TelegramClient client;
    private SessionManager sessionManager;

    // Екрани
    private LoginForm loginForm;
    private ChatListForm chatListForm;
    private ChatForm chatForm;

    // Команди
    private Command exitCommand;
    private Command backCommand;

    public TelegramMIDlet() {
        display = Display.getDisplay(this);
        sessionManager = new SessionManager();
        client = new TelegramClient();
    }

    protected void startApp() {
        // Перевіряємо чи є збережена сесія
        String[] session = sessionManager.loadSession();

        if (session != null && session[0] != null && session[1] != null) {
            // Є збережена сесія - переходимо до списку чатів
            client.setSession(session[0], session[1]);
            showChatList();
        } else {
            // Немає сесії - показуємо форму входу
            showLogin();
        }
    }

    protected void pauseApp() {
        // Зберігаємо стан при паузі
    }

    protected void destroyApp(boolean unconditional) {
        // Очищення ресурсів
        if (chatListForm != null) {
            chatListForm.stopPolling();
        }
        if (chatForm != null) {
            chatForm.stopPolling();
        }
    }

    /**
     * Показати форму входу
     */
    public void showLogin() {
        if (loginForm == null) {
            loginForm = new LoginForm(this, client, sessionManager);
        }
        display.setCurrent(loginForm);
    }

    /**
     * Показати список чатів
     */
    public void showChatList() {
        if (chatListForm == null) {
            chatListForm = new ChatListForm(this, client);
        }
        chatListForm.loadChats();
        display.setCurrent(chatListForm);
    }

    /**
     * Показати чат
     */
    public void showChat(Chat chat) {
        if (chatForm == null) {
            chatForm = new ChatForm(this, client);
        }
        chatForm.setChat(chat);
        chatForm.loadMessages();
        display.setCurrent(chatForm);
    }

    /**
     * Вийти з програми
     */
    public void exit() {
        destroyApp(true);
        notifyDestroyed();
    }

    /**
     * Вийти з акаунту
     */
    public void logout() {
        sessionManager.clearSession();
        client.clearSession();
        chatListForm = null;
        chatForm = null;
        showLogin();
    }

    public void commandAction(Command c, Displayable d) {
        // Обробка команд
    }

    public Display getDisplay() {
        return display;
    }
}
