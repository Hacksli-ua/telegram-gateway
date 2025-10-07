/**
 * Повідомлення
 */
public class Message {
    public int id;
    public String chatId;
    public String text;
    public String sender;
    public long timestamp;
    public boolean isRead;
    public boolean out;

    public Message() {
    }

    public Message(int id, String text, String sender, boolean out) {
        this.id = id;
        this.text = text;
        this.sender = sender;
        this.out = out;
        this.timestamp = System.currentTimeMillis();
    }
}
