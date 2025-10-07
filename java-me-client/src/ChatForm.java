import javax.microedition.lcdui.*;

/**
 * Екран чату з повідомленнями
 */
public class ChatForm extends Form implements CommandListener, Runnable {
    private TelegramMIDlet midlet;
    private TelegramClient client;

    private Chat chat;
    private Message[] messages;

    private TextField inputField;
    private Command sendCommand;
    private Command backCommand;
    private Command refreshCommand;

    private Thread pollingThread;
    private boolean polling;

    public ChatForm(TelegramMIDlet midlet, TelegramClient client) {
        super("Chat");
        this.midlet = midlet;
        this.client = client;

        inputField = new TextField("Message:", "", 200, TextField.ANY);

        sendCommand = new Command("Send", Command.OK, 1);
        refreshCommand = new Command("Refresh", Command.SCREEN, 2);
        backCommand = new Command("Back", Command.BACK, 3);

        addCommand(sendCommand);
        addCommand(refreshCommand);
        addCommand(backCommand);
        setCommandListener(this);
    }

    public void setChat(Chat chat) {
        this.chat = chat;
        setTitle(chat.name);
    }

    public void loadMessages() {
        deleteAll();
        append("Loading...");

        new Thread(new Runnable() {
            public void run() {
                try {
                    messages = client.getMessages(chat.id, 20);
                    displayMessages();
                    startPolling();
                } catch (Exception e) {
                    showAlert("Error", "Failed to load messages: " + e.getMessage());
                }
            }
        }).start();
    }

    private void displayMessages() {
        deleteAll();

        if (messages == null || messages.length == 0) {
            append("No messages");
        } else {
            for (int i = 0; i < messages.length; i++) {
                if (messages[i] != null) {
                    String msg = (messages[i].out ? "You: " : messages[i].sender + ": ") + messages[i].text;
                    append(msg);
                }
            }
        }

        append(inputField);
    }

    private void sendMessage() {
        final String text = inputField.getString();

        if (text == null || text.length() == 0) {
            return;
        }

        inputField.setString("");

        new Thread(new Runnable() {
            public void run() {
                try {
                    client.sendMessage(chat.id, text);

                    // Додаємо повідомлення локально
                    Message newMsg = new Message(0, text, "You", true);
                    addMessage(newMsg);
                    displayMessages();

                } catch (Exception e) {
                    showAlert("Error", "Failed to send: " + e.getMessage());
                }
            }
        }).start();
    }

    private void addMessage(Message msg) {
        if (messages == null) {
            messages = new Message[1];
            messages[0] = msg;
        } else {
            Message[] newMessages = new Message[messages.length + 1];
            System.arraycopy(messages, 0, newMessages, 0, messages.length);
            newMessages[messages.length] = msg;
            messages = newMessages;
        }
    }

    private void startPolling() {
        polling = true;
        pollingThread = new Thread(this);
        pollingThread.start();
    }

    public void stopPolling() {
        polling = false;
        if (pollingThread != null) {
            pollingThread.interrupt();
        }
    }

    public void run() {
        while (polling) {
            try {
                int lastMessageId = getLastMessageId();
                Message[] newMessages = client.pollMessages(chat.id, lastMessageId);

                if (newMessages != null && newMessages.length > 0) {
                    for (int i = 0; i < newMessages.length; i++) {
                        addMessage(newMessages[i]);
                    }
                    displayMessages();
                }

                Thread.sleep(1000);
            } catch (InterruptedException e) {
                break;
            } catch (Exception e) {
                // Ігноруємо помилки polling
            }
        }
    }

    private int getLastMessageId() {
        if (messages == null || messages.length == 0) {
            return 0;
        }

        int maxId = 0;
        for (int i = 0; i < messages.length; i++) {
            if (messages[i] != null && messages[i].id > maxId) {
                maxId = messages[i].id;
            }
        }
        return maxId;
    }

    public void commandAction(Command c, Displayable d) {
        if (c == sendCommand) {
            sendMessage();
        } else if (c == refreshCommand) {
            loadMessages();
        } else if (c == backCommand) {
            stopPolling();
            midlet.showChatList();
        }
    }

    private void showAlert(String title, String message) {
        Alert alert = new Alert(title, message, null, AlertType.ERROR);
        alert.setTimeout(2000);
        midlet.getDisplay().setCurrent(alert, this);
    }
}
