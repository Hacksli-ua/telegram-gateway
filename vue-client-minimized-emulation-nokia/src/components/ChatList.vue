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
      <h3>Chats</h3>
      <button @click="loadChats" :disabled="loading" class="refresh-btn">
        {{ loading ? '.' : '⟳' }}
      </button>
    </div>

    <!-- Індикатор завантаження -->
    <div v-if="loading && chats.length === 0" class="loading">
      Loading...
    </div>

    <!-- Помилка -->
    <div v-if="error" class="error">
      Err
      <button @click="loadChats" class="retry-btn">Try</button>
    </div>

    <!-- Список чатів -->
    <div v-if="!loading || chats.length > 0" class="chats">
      <div
        v-for="chat in chats"
        :key="chat.id"
        class="chat-item"
        @click="selectChat(chat)"
      >
        <span class="chat-name">{{ chat.name }}</span>
        <span v-if="chat.unread_count > 0" class="unread-badge">({{ chat.unread_count }})</span>
      </div>

      <!-- Якщо чатів немає -->
      <div v-if="chats.length === 0 && !loading" class="no-chats">
        No chats
      </div>
    </div>
  </div>
</template>

<style scoped>
.chat-list {
  width: 100%;
  display: flex;
  flex-direction: column;
  height: 100%;
  background: #fff;
}

.chat-list-header {
  padding: 3px 5px;
  border-bottom: 1px solid #000;
  display: flex;
  justify-content: space-between;
  align-items: center;
  background: #fff;
}

.chat-list-header h3 {
  margin: 0;
  color: #000;
  font-size: 11px;
  font-weight: bold;
}

.refresh-btn {
  background: none;
  border: none;
  font-size: 9px;
  cursor: pointer;
  padding: 0px 3px;
  color: #000;
}

.refresh-btn:disabled {
  opacity: 0.5;
}

.loading {
  padding: 5px;
  text-align: center;
  font-size: 8px;
  color: #000;
}

.error {
  padding: 3px 5px;
  font-size: 8px;
  color: #000;
  border-bottom: 1px solid #000;
}

.retry-btn {
  margin-left: 5px;
  padding: 1px 3px;
  background: #000;
  color: #d4e5d4;
  border: 1px solid #000;
  font-size: 8px;
  cursor: pointer;
}

.chats {
  flex: 1;
  overflow-y: auto;
}

.chat-item {
  display: flex;
  justify-content: space-between;
  padding: 3px 5px;
  border-bottom: 1px solid #e7e7e7;
  cursor: pointer;
  font-size: 9px;
  color: #000;
}

.chat-item:active {
  background: #000;
  color: #d4e5d4;
}

.chat-name {
  flex: 1;
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
  text-align: left;
}

.unread-badge {
  font-size: 9px;
  margin-left: 3px;
}

.no-chats {
  padding: 10px 5px;
  text-align: center;
  font-size: 8px;
  color: #000;
}
</style>
