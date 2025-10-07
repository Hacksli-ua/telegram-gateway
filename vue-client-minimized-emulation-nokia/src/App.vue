<script setup>
import { ref, onMounted } from 'vue'
import TelegramAuth from './components/TelegramAuth.vue'
import ChatList from './components/ChatList.vue'
import ChatView from './components/ChatView.vue'

const isAuthenticated = ref(false)
const session = ref(null)
const selectedChat = ref(null)

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

function backToChats() {
  selectedChat.value = null
}
</script>

<template>
  <div id="nokia-container">
    <div class="nokia-screen">
      <TelegramAuth v-if="!isAuthenticated" @authenticated="onAuthenticated" />
      <div v-else class="main-view">
        <ChatList v-if="!selectedChat" :session="session" @chatSelected="onChatSelected" @logout="logout" />
        <ChatView v-else :chat="selectedChat" :session="session" @back="backToChats" />
      </div>
    </div>
  </div>
</template>

<style>
* {
  margin: 0;
  padding: 0;
  box-sizing: border-box;
}

body {
  font-family: 'Nokia Pure Text', Arial, sans-serif;
  background: #1a1a1a;
  display: flex;
  align-items: center;
  justify-content: center;
  min-height: 100vh;
}

#nokia-container {
  background: #2a2a2a;
  padding: 20px;
  border-radius: 20px;
  box-shadow: 0 10px 50px rgba(0, 0, 0, 0.5);
}

.nokia-screen {
  width: 230px;
  height: 230px;
  background: #d4e5d4;
  border: 3px solid #000;
  border-radius: 3px;
  overflow: hidden;
  position: relative;
  font-size: 10px;
}

.main-view {
  width: 100%;
  height: 100%;
  background: #d4e5d4;
}
</style>
