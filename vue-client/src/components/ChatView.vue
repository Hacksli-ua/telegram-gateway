<script setup>
import { ref, watch, nextTick, onUnmounted } from 'vue'

const props = defineProps({
  chat: {
    type: Object,
    required: true
  },
  session: {
    type: Object,
    required: true
  }
})

const API_BASE = 'http://localhost:8080'
const messages = ref([])
const loading = ref(false)
const error = ref('')
const newMessage = ref('')
const sending = ref(false)
const messagesContainer = ref(null)
const pollingActive = ref(false)
const pollingAbortController = ref(null)

// Завантаження повідомлень
async function loadMessages() {
  loading.value = true
  error.value = ''

  try {
    const response = await fetch(`${API_BASE}/api/messages/${props.chat.id}?limit=20`, {
      method: 'GET',
      headers: {
        'X-Phone': props.session.phone,
        'X-Session-Data': props.session.session_data
      }
    })

    const data = await response.json()

    if (response.ok) {
      // Реверсуємо масив щоб старіші повідомлення були вгорі
      messages.value = (data.messages || []).reverse()
      // Прокручуємо вниз після завантаження
      await nextTick()
      scrollToBottom()
    } else {
      error.value = data.error || 'Помилка завантаження повідомлень'
    }
  } catch (err) {
    error.value = 'Помилка з\'єднання з сервером: ' + err.message
  } finally {
    loading.value = false
  }
}

// Відправка повідомлення
async function sendMessage() {
  if (!newMessage.value.trim()) return

  const messageText = newMessage.value
  newMessage.value = ''
  sending.value = true
  error.value = ''

  try {
    const response = await fetch(`${API_BASE}/api/send`, {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
        'X-Phone': props.session.phone,
        'X-Session-Data': props.session.session_data
      },
      body: JSON.stringify({
        chat_id: props.chat.id.toString(),
        text: messageText
      })
    })

    const data = await response.json()

    if (response.ok) {
      // Додаємо повідомлення локально
      messages.value.push({
        id: data.message_id,
        chat_id: props.chat.id.toString(),
        text: messageText,
        sender: 'You',
        timestamp: new Date().toISOString(),
        is_read: true,
        out: true,
        has_photo: false
      })

      await nextTick()
      scrollToBottom()
    } else {
      console.error('Send message error:', data)
      error.value = data.error || 'Помилка відправки'
      // Повертаємо текст назад у поле вводу
      newMessage.value = messageText
    }
  } catch (err) {
    error.value = 'Помилка з\'єднання з сервером: ' + err.message
    newMessage.value = messageText
  } finally {
    sending.value = false
  }
}

// Форматування часу
function formatTime(timestamp) {
  if (!timestamp) return ''
  const date = new Date(timestamp)
  return date.toLocaleTimeString('uk-UA', { hour: '2-digit', minute: '2-digit' })
}

// Форматування дати
function formatDate(timestamp) {
  if (!timestamp) return ''
  const date = new Date(timestamp)
  const today = new Date()

  if (date.toDateString() === today.toDateString()) {
    return 'Сьогодні'
  }

  const yesterday = new Date(today)
  yesterday.setDate(yesterday.getDate() - 1)

  if (date.toDateString() === yesterday.toDateString()) {
    return 'Вчора'
  }

  return date.toLocaleDateString('uk-UA', { day: 'numeric', month: 'long', year: 'numeric' })
}

// Групування повідомлень по датам
function groupMessagesByDate() {
  const grouped = []
  let currentDate = null

  messages.value.forEach(msg => {
    const msgDate = new Date(msg.timestamp).toDateString()

    if (msgDate !== currentDate) {
      currentDate = msgDate
      grouped.push({
        type: 'date',
        date: msg.timestamp
      })
    }

    grouped.push({
      type: 'message',
      ...msg
    })
  })

  return grouped
}

// Прокрутка вниз
function scrollToBottom() {
  if (messagesContainer.value) {
    messagesContainer.value.scrollTop = messagesContainer.value.scrollHeight
  }
}

// Отримати URL фото з автентифікацією
function getPhotoUrl(message) {
  // Створюємо base64 token з phone і session_data
  const authToken = btoa(`${props.session.phone}:${props.session.session_data}`)
  return `${API_BASE}/api/photo/${message.chat_id}/${message.id}?token=${authToken}`
}

// Обробка помилок завантаження зображень
function handleImageError(event) {
  event.target.style.display = 'none'
  console.error('Failed to load image')
}

// Long polling для нових повідомлень
async function startPolling() {
  pollingActive.value = true

  while (pollingActive.value) {
    try {
      // Отримуємо ID останнього повідомлення
      const lastMessageId = messages.value.length > 0
        ? Math.max(...messages.value.map(m => m.id))
        : 0

      pollingAbortController.value = new AbortController()

      const response = await fetch(
        `${API_BASE}/api/poll/${props.chat.id}?after_message_id=${lastMessageId}&timeout=30`,
        {
          headers: {
            'X-Phone': props.session.phone,
            'X-Session-Data': props.session.session_data
          },
          signal: pollingAbortController.value.signal
        }
      )

      if (!response.ok) {
        console.error('Polling error:', response.status)
        await new Promise(resolve => setTimeout(resolve, 5000))
        continue
      }

      const data = await response.json()

      if (data.has_new && data.messages && data.messages.length > 0) {
        // Додаємо нові повідомлення
        messages.value.push(...data.messages)
        await nextTick()
        scrollToBottom()
      }
    } catch (err) {
      if (err.name === 'AbortError') {
        // Polling було скасовано
        break
      }
      console.error('Polling error:', err)
      await new Promise(resolve => setTimeout(resolve, 5000))
    }
  }
}

function stopPolling() {
  pollingActive.value = false
  if (pollingAbortController.value) {
    pollingAbortController.value.abort()
  }
}

// Спостереження за зміною чату
watch(() => props.chat.id, () => {
  stopPolling()
  messages.value = []
  loadMessages().then(() => {
    startPolling()
  })
}, { immediate: true })

// Очищення при unmount
onUnmounted(() => {
  stopPolling()
})
</script>

<template>
  <div class="chat-view">
    <!-- Шапка чату -->
    <div class="chat-header">
      <div class="chat-info">
        <div class="avatar-small">
          {{ chat.name.charAt(0).toUpperCase() }}
        </div>
        <div>
          <h3>{{ chat.name }}</h3>
          <span class="chat-type">{{ chat.type === 'user' ? 'Особиста бесіда' : chat.type === 'chat' ? 'Група' : 'Канал' }}</span>
        </div>
      </div>
      <button @click="loadMessages" :disabled="loading" class="refresh-btn">
        {{ loading ? '⏳' : '🔄' }}
      </button>
    </div>

    <!-- Повідомлення -->
    <div ref="messagesContainer" class="messages-container">
      <!-- Індикатор завантаження -->
      <div v-if="loading && messages.length === 0" class="loading">
        <p>⏳ Завантаження повідомлень...</p>
        <p class="hint">Це може зайняти 2-5 секунд</p>
      </div>

      <!-- Помилка -->
      <div v-if="error" class="error">
        {{ error }}
      </div>

      <!-- Список повідомлень -->
      <div v-if="!loading || messages.length > 0" class="messages">
        <div v-for="(item, index) in groupMessagesByDate()" :key="index">
          <!-- Роздільник дати -->
          <div v-if="item.type === 'date'" class="date-divider">
            <span>{{ formatDate(item.date) }}</span>
          </div>

          <!-- Повідомлення -->
          <div v-else-if="item.type === 'message'"
               :class="['message', item.out ? 'message-out' : 'message-in']">
            <div class="message-bubble">
              <div v-if="!item.out" class="message-sender">{{ item.sender }}</div>

              <!-- Фото якщо є -->
              <div v-if="item.has_photo" class="message-photo">
                <img
                  :src="getPhotoUrl(item)"
                  :alt="item.text"
                  loading="lazy"
                  @error="handleImageError"
                />
              </div>

              <div class="message-text" v-if="item.text && item.text !== '📷 Фото'">{{ item.text }}</div>
              <div class="message-time">
                {{ formatTime(item.timestamp) }}
                <span v-if="item.out" class="read-status">
                  {{ item.is_read ? '✓✓' : '✓' }}
                </span>
              </div>
            </div>
          </div>
        </div>

        <!-- Якщо повідомлень немає -->
        <div v-if="messages.length === 0 && !loading" class="no-messages">
          <p>💬 Немає повідомлень</p>
          <p class="hint">Почніть розмову нижче</p>
        </div>
      </div>
    </div>

    <!-- Поле вводу -->
    <div class="message-input-container">
      <input
        v-model="newMessage"
        type="text"
        placeholder="Напишіть повідомлення..."
        @keyup.enter="sendMessage"
        :disabled="sending"
        class="message-input"
      />
      <button
        @click="sendMessage"
        :disabled="sending || !newMessage.trim()"
        class="send-btn"
      >
        {{ sending ? '⏳' : '📤' }}
      </button>
    </div>
  </div>
</template>

<style scoped>
.chat-view {
  display: flex;
  flex-direction: column;
  height: 100%;
  background: #fff;
}

.chat-header {
  padding: 15px 20px;
  border-bottom: 1px solid #ddd;
  display: flex;
  justify-content: space-between;
  align-items: center;
  background: #fff;
}

.chat-info {
  display: flex;
  align-items: center;
  gap: 15px;
}

.avatar-small {
  width: 40px;
  height: 40px;
  border-radius: 50%;
  background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
  color: white;
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 16px;
  font-weight: bold;
}

.chat-header h3 {
  margin: 0;
  font-size: 16px;
  color: #333;
}

.chat-type {
  font-size: 12px;
  color: #999;
}

.refresh-btn {
  background: none;
  border: none;
  font-size: 20px;
  cursor: pointer;
  padding: 5px 10px;
  border-radius: 4px;
}

.refresh-btn:hover:not(:disabled) {
  background: #f0f0f0;
}

.refresh-btn:disabled {
  opacity: 0.5;
  cursor: not-allowed;
}

.messages-container {
  flex: 1;
  overflow-y: auto;
  padding: 20px;
  background: #f5f5f5;
}

.loading {
  text-align: center;
  padding: 40px 20px;
  color: #666;
}

.hint {
  font-size: 12px;
  color: #999;
  margin-top: 5px;
}

.error {
  padding: 15px;
  background: #f8d7da;
  color: #721c24;
  margin: 10px;
  border-radius: 4px;
  text-align: center;
}

.messages {
  max-width: 800px;
  margin: 0 auto;
}

.date-divider {
  text-align: center;
  margin: 20px 0;
}

.date-divider span {
  background: #e1e8ed;
  padding: 5px 15px;
  border-radius: 10px;
  font-size: 12px;
  color: #666;
}

.message {
  margin: 10px 0;
  display: flex;
}

.message-in {
  justify-content: flex-start;
}

.message-out {
  justify-content: flex-end;
}

.message-bubble {
  max-width: 70%;
  padding: 10px 15px;
  border-radius: 12px;
  word-wrap: break-word;
}

.message-in .message-bubble {
  background: #fff;
  border: 1px solid #e1e8ed;
  color: #333;
}

.message-out .message-bubble {
  background: #0088cc;
  color: white;
}

.message-sender {
  font-size: 12px;
  font-weight: 600;
  margin-bottom: 5px;
  color: #0088cc;
}

.message-photo {
  margin-bottom: 8px;
  border-radius: 8px;
  overflow: hidden;
  max-width: 300px;
}

.message-photo img {
  width: 100%;
  height: auto;
  display: block;
  border-radius: 8px;
}

.message-text {
  font-size: 14px;
  line-height: 1.4;
  margin-bottom: 5px;
}

.message-time {
  font-size: 11px;
  opacity: 0.7;
  text-align: right;
  display: flex;
  align-items: center;
  justify-content: flex-end;
  gap: 3px;
}

.read-status {
  font-size: 12px;
}

.no-messages {
  text-align: center;
  padding: 60px 20px;
  color: #999;
}

.no-messages p:first-child {
  font-size: 40px;
  margin-bottom: 10px;
}

.message-input-container {
  padding: 15px 20px;
  border-top: 1px solid #ddd;
  background: #fff;
  display: flex;
  gap: 10px;
}

.message-input {
  flex: 1;
  padding: 12px 15px;
  border: 1px solid #ddd;
  border-radius: 20px;
  font-size: 14px;
  outline: none;
}

.message-input:focus {
  border-color: #0088cc;
}

.send-btn {
  width: 45px;
  height: 45px;
  border: none;
  border-radius: 50%;
  background: #0088cc;
  color: white;
  font-size: 20px;
  cursor: pointer;
  display: flex;
  align-items: center;
  justify-content: center;
}

.send-btn:hover:not(:disabled) {
  background: #006699;
}

.send-btn:disabled {
  background: #ccc;
  cursor: not-allowed;
}
</style>
