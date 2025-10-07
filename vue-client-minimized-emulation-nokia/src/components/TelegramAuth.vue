<script setup>
import { ref } from 'vue'

const API_BASE = 'http://localhost:8080'
const step = ref('phone')
const phone = ref('')
const code = ref('')
const password = ref('')
const loading = ref(false)
const error = ref('')

const emit = defineEmits(['authenticated'])

async function requestCode() {
  if (!phone.value || phone.value.length < 10) {
    error.value = 'Enter phone'
    return
  }
  loading.value = true
  error.value = ''
  try {
    const response = await fetch(`${API_BASE}/auth/request-code`, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ phone: phone.value })
    })
    if (response.ok) {
      step.value = 'code'
    } else {
      error.value = 'Error'
    }
  } catch (err) {
    error.value = 'No server'
  } finally {
    loading.value = false
  }
}

async function login() {
  if (!code.value || code.value.length < 5) {
    error.value = 'Enter code'
    return
  }
  loading.value = true
  error.value = ''
  try {
    const response = await fetch(`${API_BASE}/auth/login`, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ phone: phone.value, code: code.value })
    })
    const data = await response.json()
    if (response.ok) {
      if (data.needs_password) {
        step.value = 'password'
      } else {
        const sessionData = { phone: data.phone, session_data: data.session_data }
        localStorage.setItem('telegram_session', JSON.stringify(sessionData))
        emit('authenticated', sessionData)
      }
    } else {
      error.value = 'Wrong code'
    }
  } catch (err) {
    error.value = 'Error'
  } finally {
    loading.value = false
  }
}

async function submitPassword() {
  if (!password.value) {
    error.value = 'Enter pass'
    return
  }
  loading.value = true
  error.value = ''
  try {
    const response = await fetch(`${API_BASE}/auth/password`, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ phone: phone.value, password: password.value })
    })
    const data = await response.json()
    if (response.ok) {
      const sessionData = { phone: data.phone, session_data: data.session_data }
      localStorage.setItem('telegram_session', JSON.stringify(sessionData))
      emit('authenticated', sessionData)
    } else {
      error.value = 'Wrong pass'
    }
  } catch (err) {
    error.value = 'Error'
  } finally {
    loading.value = false
  }
}
</script>

<template>
  <div class="nokia-auth">
    <div class="nokia-header">Login</div>
    <div class="nokia-content">
      <div v-if="step === 'phone'" class="auth-step">
        <div class="label">Phone:</div>
        <input v-model="phone" type="tel" class="nokia-input" placeholder="+380..." @keyup.enter="requestCode" />
        <button @click="requestCode" :disabled="loading" class="nokia-btn">
          {{ loading ? 'Wait..' : 'Get Code' }}
        </button>
      </div>
      <div v-if="step === 'code'" class="auth-step">
        <div class="label">Code:</div>
        <input v-model="code" type="text" class="nokia-input" placeholder="12345" @keyup.enter="login" />
        <button @click="login" :disabled="loading" class="nokia-btn">
          {{ loading ? 'Wait..' : 'Login' }}
        </button>
      </div>
      <div v-if="step === 'password'" class="auth-step">
        <div class="label">2FA:</div>
        <input v-model="password" type="password" class="nokia-input" @keyup.enter="submitPassword" />
        <button @click="submitPassword" :disabled="loading" class="nokia-btn">
          {{ loading ? 'Wait..' : 'Submit' }}
        </button>
      </div>
      <div v-if="error" class="error">{{ error }}</div>
    </div>
  </div>
</template>

<style scoped>
.nokia-auth { width: 100%; height: 100%; display: flex; flex-direction: column; }
.nokia-header { background: #000; color: #d4e5d4; padding: 3px 5px; font-size: 11px; font-weight: bold; border-bottom: 1px solid #000; }
.nokia-content { flex: 1; padding: 8px 5px; overflow-y: auto; }
.auth-step { margin-bottom: 10px; }
.label { font-size: 9px; margin-bottom: 2px; color: #000; }
.nokia-input { width: 100%; padding: 3px; border: 1px solid #000; background: #fff; font-size: 10px; font-family: monospace; margin-bottom: 4px; }
.nokia-btn { width: 100%; padding: 4px; background: #000; color: #d4e5d4; border: 1px solid #000; font-size: 9px; cursor: pointer; font-weight: bold; }
.nokia-btn:active { background: #333; }
.nokia-btn:disabled { background: #666; cursor: not-allowed; }
.error { font-size: 8px; color: #000; margin-top: 5px; padding: 2px; background: #fff; border: 1px solid #000; }
</style>
