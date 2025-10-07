import javax.microedition.lcdui.*;

/**
 * Форма входу
 */
public class LoginForm extends Form implements CommandListener {
    private TelegramMIDlet midlet;
    private TelegramClient client;
    private SessionManager sessionManager;

    private TextField phoneField;
    private TextField codeField;
    private TextField passwordField;

    private Command requestCodeCommand;
    private Command loginCommand;
    private Command submitPasswordCommand;
    private Command exitCommand;

    private String currentPhone;
    private int step; // 0=phone, 1=code, 2=password

    public LoginForm(TelegramMIDlet midlet, TelegramClient client, SessionManager sessionManager) {
        super("Telegram Login");
        this.midlet = midlet;
        this.client = client;
        this.sessionManager = sessionManager;

        phoneField = new TextField("Phone:", "+380", 20, TextField.PHONENUMBER);
        codeField = new TextField("Code:", "", 10, TextField.NUMERIC);
        passwordField = new TextField("Password:", "", 30, TextField.PASSWORD);

        requestCodeCommand = new Command("Get Code", Command.OK, 1);
        loginCommand = new Command("Login", Command.OK, 1);
        submitPasswordCommand = new Command("Submit", Command.OK, 1);
        exitCommand = new Command("Exit", Command.EXIT, 2);

        showPhoneStep();
    }

    private void showPhoneStep() {
        step = 0;
        deleteAll();
        append(phoneField);
        removeCommand(loginCommand);
        removeCommand(submitPasswordCommand);
        addCommand(requestCodeCommand);
        addCommand(exitCommand);
        setCommandListener(this);
    }

    private void showCodeStep() {
        step = 1;
        deleteAll();
        append("Code sent to " + currentPhone);
        append(codeField);
        removeCommand(requestCodeCommand);
        removeCommand(submitPasswordCommand);
        addCommand(loginCommand);
        addCommand(exitCommand);
    }

    private void showPasswordStep() {
        step = 2;
        deleteAll();
        append("2FA Password required");
        append(passwordField);
        removeCommand(requestCodeCommand);
        removeCommand(loginCommand);
        addCommand(submitPasswordCommand);
        addCommand(exitCommand);
    }

    public void commandAction(Command c, Displayable d) {
        if (c == exitCommand) {
            midlet.exit();
        } else if (c == requestCodeCommand) {
            requestCode();
        } else if (c == loginCommand) {
            login();
        } else if (c == submitPasswordCommand) {
            submitPassword();
        }
    }

    private void requestCode() {
        currentPhone = phoneField.getString();

        if (currentPhone == null || currentPhone.length() < 10) {
            showAlert("Error", "Enter phone number");
            return;
        }

        showAlert("Wait", "Requesting code...");

        new Thread(new Runnable() {
            public void run() {
                try {
                    client.requestCode(currentPhone);
                    showCodeStep();
                } catch (Exception e) {
                    showAlert("Error", "Failed: " + e.getMessage());
                }
            }
        }).start();
    }

    private void login() {
        final String code = codeField.getString();

        if (code == null || code.length() < 5) {
            showAlert("Error", "Enter code");
            return;
        }

        showAlert("Wait", "Logging in...");

        new Thread(new Runnable() {
            public void run() {
                try {
                    AuthResult result = client.login(currentPhone, code);

                    if (result.needsPassword) {
                        showPasswordStep();
                    } else if ("success".equals(result.status)) {
                        sessionManager.saveSession(result.phone, result.sessionData);
                        client.setSession(result.phone, result.sessionData);
                        midlet.showChatList();
                    } else {
                        showAlert("Error", "Login failed");
                    }
                } catch (Exception e) {
                    showAlert("Error", "Failed: " + e.getMessage());
                }
            }
        }).start();
    }

    private void submitPassword() {
        final String password = passwordField.getString();

        if (password == null || password.length() == 0) {
            showAlert("Error", "Enter password");
            return;
        }

        showAlert("Wait", "Submitting password...");

        new Thread(new Runnable() {
            public void run() {
                try {
                    AuthResult result = client.submitPassword(currentPhone, password);

                    if ("success".equals(result.status)) {
                        sessionManager.saveSession(result.phone, result.sessionData);
                        client.setSession(result.phone, result.sessionData);
                        midlet.showChatList();
                    } else {
                        showAlert("Error", "Login failed");
                    }
                } catch (Exception e) {
                    showAlert("Error", "Failed: " + e.getMessage());
                }
            }
        }).start();
    }

    private void showAlert(String title, String message) {
        Alert alert = new Alert(title, message, null, AlertType.INFO);
        alert.setTimeout(2000);
        midlet.getDisplay().setCurrent(alert, this);
    }
}
