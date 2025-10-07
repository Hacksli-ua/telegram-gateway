<script setup>
import { ref, onMounted } from 'vue'

const props = defineProps({
  session: {
    type: Object,
    required: true
  }
})

const emit = defineEmits(['chatSelected'])

const API_BASE = 'http://localhost:8080'
const chats = ref([])
const loading = ref(false)
const error = ref('')

// Завантаження чатів
async function loadChats() {
  loading.value = true
  error.value = ''

  try {
    const response = await fetch(`${API_BASE}/api/chats`, {
      method: 'GET',
      headers: {
        'X-Phone': props.session.phone,
        'X-Session-Data': props.session.session_data
      }
    })

    const data = await response.json()

    if (response.ok) {
      chats.value = data.chats || []
    } else {
      error.value = data.error || 'Помилка завантаження чатів'
    }
  } catch (err) {
    error.value = 'Помилка з\'єднання з сервером: ' + err.message
  } finally {
    loading.value = false
  }
}

// Форматування часу
function formatTime(timestamp) {
  if (!timestamp) return ''
  const date = new Date(timestamp)
  const now = new Date()
  const diffDays = Math.floor((now - date) / (1000 * 60 * 60 * 24))

  if (diffDays === 0) {
    // Сьогодні - показуємо час
    return date.toLocaleTimeString('uk-UA', { hour: '2-digit', minute: '2-digit' })
  } else if (diffDays === 1) {
    return 'Вчора'
  } else if (diffDays < 7) {
    return date.toLocaleDateString('uk-UA', { weekday: 'short' })
  } else {
    return date.toLocaleDateString('uk-UA', { day: 'numeric', month: 'short' })
  }
}

// Обрізати довге повідомлення
function truncateMessage(msg, maxLength = 50) {
  if (!msg) return ''
  return msg.length > maxLength ? msg.substring(0, maxLength) + '...' : msg
}

// Вибір чату
function selectChat(chat) {
  emit('chatSelected', chat)
}

onMounted(() => {
  loadChats()
})
</script>

<template>
  <div class="chat-list">
    <div class="chat-list-header">
      <h3>Чати</h3>
      <button @click="loadChats" :disabled="loading" class="refresh-btn">
        {{ loading ? '⏳' : '🔄' }}
      </button>
    </div>

    <!-- Індикатор завантаження -->
    <div v-if="loading && chats.length === 0" class="loading">
      <p>⏳ Завантаження чатів...</p>
      <p class="hint">Це може зайняти 2-5 секунд</p>
    </div>

    <!-- Помилка -->
    <div v-if="error" class="error">
      {{ error }}
      <button @click="loadChats" class="retry-btn">Спробувати знову</button>
    </div>

    <!-- Список чатів -->
    <div v-if="!loading || chats.length > 0" class="chats">
      <div
        v-for="chat in chats"
        :key="chat.id"
        class="chat-item"
        @click="selectChat(chat)"
      >
        <div class="chat-avatar">
          <div class="avatar-circle">
            {{ chat.name.charAt(0).toUpperCase() }}
          </div>
        </div>

        <div class="chat-content">
          <div class="chat-header">
            <span class="chat-name">{{ chat.name }}</span>
            <span class="chat-time">{{ formatTime(chat.last_update_time) }}</span>
          </div>
          <div class="chat-footer">
            <span class="chat-message">{{ truncateMessage(chat.last_message) }}</span>
            <span v-if="chat.unread_count > 0" class="unread-badge">
              {{ chat.unread_count }}
            </span>
          </div>
        </div>
      </div>

      <!-- Якщо чатів немає -->
      <div v-if="chats.length === 0 && !loading" class="no-chats">
        <p>📭 Чатів не знайдено</p>
      </div>
    </div>
  </div>
</template>

<style scoped>
.chat-list {
  width: 100%;
  max-width: 400px;
  border-right: 1px solid #ddd;
  background: #fff;
  display: flex;
  flex-direction: column;
  height: 100%;
}

.chat-list-header {
  padding: 20px;
  border-bottom: 1px solid #ddd;
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.chat-list-header h3 {
  margin: 0;
  color: #333;
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

.loading {
  padding: 40px 20px;
  text-align: center;
  color: #666;
}

.loading .hint {
  font-size: 12px;
  color: #999;
  margin-top: 5px;
}

.error {
  padding: 20px;
  background: #f8d7da;
  color: #721c24;
  margin: 10px;
  border-radius: 4px;
}

.retry-btn {
  margin-top: 10px;
  padding: 8px 16px;
  background: #dc3545;
  color: white;
  border: none;
  border-radius: 4px;
  cursor: pointer;
}

.retry-btn:hover {
  background: #c82333;
}

.chats {
  flex: 1;
  overflow-y: auto;
}

.chat-item {
  display: flex;
  padding: 15px 20px;
  border-bottom: 1px solid #f0f0f0;
  cursor: pointer;
  transition: background 0.2s;
}

.chat-item:hover {
  background: #f8f9fa;
}

.chat-avatar {
  margin-right: 15px;
}

.avatar-circle {
  width: 50px;
  height: 50px;
  border-radius: 50%;
  background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
  color: white;
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 20px;
  font-weight: bold;
}

.chat-content {
  flex: 1;
  min-width: 0;
}

.chat-header {
  display: flex;
  justify-content: space-between;
  margin-bottom: 5px;
}

.chat-name {
  font-weight: 600;
  color: #333;
  font-size: 15px;
}

.chat-time {
  font-size: 12px;
  color: #999;
}

.chat-footer {
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.chat-message {
  color: #666;
  font-size: 14px;
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
  flex: 1;
}

.unread-badge {
  background: #0088cc;
  color: white;
  border-radius: 10px;
  padding: 2px 8px;
  font-size: 11px;
  font-weight: bold;
  margin-left: 10px;
}

.no-chats {
  padding: 60px 20px;
  text-align: center;
  color: #999;
}
</style>
