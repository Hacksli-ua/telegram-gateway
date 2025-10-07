import javax.microedition.rms.*;

/**
 * Менеджер для збереження/завантаження сесії в RMS
 */
public class SessionManager {
    private static final String STORE_NAME = "TelegramSession";

    public SessionManager() {
    }

    /**
     * Зберегти сесію
     */
    public void saveSession(String phone, String sessionData) {
        RecordStore rs = null;
        try {
            // Видаляємо старий store якщо існує
            try {
                RecordStore.deleteRecordStore(STORE_NAME);
            } catch (Exception e) {
                // Ігноруємо якщо не існує
            }

            rs = RecordStore.openRecordStore(STORE_NAME, true);

            String combined = phone + "|" + sessionData;
            byte[] data = combined.getBytes("UTF-8");

            rs.addRecord(data, 0, data.length);

        } catch (Exception e) {
            e.printStackTrace();
        } finally {
            if (rs != null) {
                try {
                    rs.closeRecordStore();
                } catch (Exception e) {
                }
            }
        }
    }

    /**
     * Завантажити сесію
     * @return масив [phone, sessionData] або null
     */
    public String[] loadSession() {
        RecordStore rs = null;
        try {
            rs = RecordStore.openRecordStore(STORE_NAME, false);

            if (rs.getNumRecords() == 0) {
                return null;
            }

            byte[] data = rs.getRecord(1);
            String combined = new String(data, "UTF-8");

            int pipeIndex = combined.indexOf('|');
            if (pipeIndex == -1) {
                return null;
            }

            String[] result = new String[2];
            result[0] = combined.substring(0, pipeIndex);
            result[1] = combined.substring(pipeIndex + 1);

            return result;

        } catch (RecordStoreNotFoundException e) {
            return null;
        } catch (Exception e) {
            e.printStackTrace();
            return null;
        } finally {
            if (rs != null) {
                try {
                    rs.closeRecordStore();
                } catch (Exception e) {
                }
            }
        }
    }

    /**
     * Очистити сесію
     */
    public void clearSession() {
        try {
            RecordStore.deleteRecordStore(STORE_NAME);
        } catch (Exception e) {
            // Ігноруємо
        }
    }
}
