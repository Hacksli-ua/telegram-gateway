/**
 * Чат (діалог)
 */
public class Chat {
    public String id;
    public String name;
    public String lastMessage;
    public int unreadCount;
    public String type;

    public Chat() {
    }

    public Chat(String id, String name) {
        this.id = id;
        this.name = name;
    }
}
