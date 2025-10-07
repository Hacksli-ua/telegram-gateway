import javax.microedition.lcdui.*;

/**
 * Список чатів
 */
public class ChatListForm extends List implements CommandListener, Runnable {
    private TelegramMIDlet midlet;
    private TelegramClient client;

    private Command selectCommand;
    private Command refreshCommand;
    private Command logoutCommand;
    private Command exitCommand;

    private Chat[] chats;
    private Thread pollingThread;
    private boolean polling;

    public ChatListForm(TelegramMIDlet midlet, TelegramClient client) {
        super("Chats", List.IMPLICIT);
        this.midlet = midlet;
        this.client = client;

        selectCommand = new Command("Open", Command.ITEM, 1);
        refreshCommand = new Command("Refresh", Command.SCREEN, 2);
        logoutCommand = new Command("Logout", Command.SCREEN, 3);
        exitCommand = new Command("Exit", Command.EXIT, 4);

        addCommand(selectCommand);
        addCommand(refreshCommand);
        addCommand(logoutCommand);
        addCommand(exitCommand);
        setCommandListener(this);
        setSelectCommand(selectCommand);
    }

    public void loadChats() {
        deleteAll();
        append("Loading...", null);

        new Thread(new Runnable() {
            public void run() {
                try {
                    chats = client.getChats();
                    displayChats();
                } catch (Exception e) {
                    showAlert("Error", "Failed to load chats: " + e.getMessage());
                }
            }
        }).start();
    }

    private void displayChats() {
        deleteAll();

        if (chats == null || chats.length == 0) {
            append("No chats", null);
            return;
        }

        for (int i = 0; i < chats.length; i++) {
            if (chats[i] != null && chats[i].name != null) {
                String text = chats[i].name;
                if (chats[i].unreadCount > 0) {
                    text += " (" + chats[i].unreadCount + ")";
                }
                append(text, null);
            }
        }
    }

    public void commandAction(Command c, Displayable d) {
        if (c == exitCommand) {
            stopPolling();
            midlet.exit();
        } else if (c == logoutCommand) {
            stopPolling();
            midlet.logout();
        } else if (c == refreshCommand) {
            loadChats();
        } else if (c == selectCommand || c == List.SELECT_COMMAND) {
            int index = getSelectedIndex();
            if (index >= 0 && chats != null && index < chats.length) {
                midlet.showChat(chats[index]);
            }
        }
    }

    public void stopPolling() {
        polling = false;
        if (pollingThread != null) {
            pollingThread.interrupt();
        }
    }

    public void run() {
        // Polling logic - not implemented for chat list
    }

    private void showAlert(String title, String message) {
        Alert alert = new Alert(title, message, null, AlertType.ERROR);
        alert.setTimeout(2000);
        midlet.getDisplay().setCurrent(alert, this);
    }
}
