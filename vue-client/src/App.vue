<script setup>
import { ref, onMounted } from 'vue'
import TelegramAuth from './components/TelegramAuth.vue'
import ChatList from './components/ChatList.vue'
import ChatView from './components/ChatView.vue'

const isAuthenticated = ref(false)
const session = ref(null)
const selectedChat = ref(null)

// Перевіряємо чи є збережена сесія
onMounted(() => {
  const savedSession = localStorage.getItem('telegram_session')
  if (savedSession) {
    try {
      session.value = JSON.parse(savedSession)
      isAuthenticated.value = true
    } catch (e) {
      localStorage.removeItem('telegram_session')
    }
  }
})

function onAuthenticated(sessionData) {
  session.value = sessionData
  isAuthenticated.value = true
}

function logout() {
  localStorage.removeItem('telegram_session')
  session.value = null
  isAuthenticated.value = false
  selectedChat.value = null
}

function onChatSelected(chat) {
  selectedChat.value = chat
}
</script>

<template>
  <div id="app">
    <!-- Якщо авторизовані -->
    <div v-if="isAuthenticated" class="main-container">
      <div class="header">
        <h2>Telegram Gateway</h2>
        <div class="user-info">
          <span>{{ session?.phone }}</span>
          <button @click="logout" class="logout-btn">Вийти</button>
        </div>
      </div>
      <div class="content">
        <ChatList :session="session" @chatSelected="onChatSelected" />

        <div class="chat-view-container">
          <div v-if="!selectedChat" class="no-chat-selected">
            <p>📱</p>
            <p>Оберіть чат зліва</p>
          </div>
          <ChatView v-else :chat="selectedChat" :session="session" />
        </div>
      </div>
    </div>

    <!-- Якщо не авторизовані -->
    <TelegramAuth v-else @authenticated="onAuthenticated" />
  </div>
</template>

<style>
* {
  margin: 0;
  padding: 0;
  box-sizing: border-box;
}

body {
  font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, Oxygen, Ubuntu, Cantarell, sans-serif;
  background: #f5f5f5;
}

#app {
  min-height: 100vh;
}

.main-container {
  max-width: 1200px;
  margin: 0 auto;
  background: white;
  min-height: 100vh;
}

.header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 20px 30px;
  border-bottom: 1px solid #ddd;
  background: #fff;
}

.header h2 {
  color: #0088cc;
}

.user-info {
  display: flex;
  align-items: center;
  gap: 15px;
}

.user-info span {
  color: #666;
  font-size: 14px;
}

.logout-btn {
  padding: 8px 16px;
  background: #dc3545;
  color: white;
  border: none;
  border-radius: 4px;
  cursor: pointer;
  font-size: 14px;
}

.logout-btn:hover {
  background: #c82333;
}

.content {
  display: flex;
  height: calc(100vh - 81px);
}

.chat-view-container {
  flex: 1;
  display: flex;
  align-items: center;
  justify-content: center;
  background: #f5f5f5;
}

.no-chat-selected {
  text-align: center;
  color: #999;
}

.no-chat-selected p:first-child {
  font-size: 60px;
  margin-bottom: 10px;
}
</style>
