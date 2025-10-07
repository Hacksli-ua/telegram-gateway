<script setup>
import { ref, computed } from 'vue'

const API_BASE = 'http://localhost:8080'

const step = ref('phone') // phone, code, password, success
const phone = ref('')
const code = ref('')
const password = ref('')
const loading = ref(false)
const error = ref('')
const sessionData = ref(null)

const emit = defineEmits(['authenticated'])

// Запит коду
async function requestCode() {
  if (!phone.value || phone.value.length < 10) {
    error.value = 'Введіть номер телефону'
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

    const data = await response.json()

    if (response.ok) {
      step.value = 'code'
    } else {
      error.value = data.error || 'Помилка запиту коду'
    }
  } catch (err) {
    error.value = 'Помилка з\'єднання з сервером'
  } finally {
    loading.value = false
  }
}

// Вхід з кодом
async function login() {
  if (!code.value || code.value.length < 5) {
    error.value = 'Введіть код з Telegram'
    return
  }

  loading.value = true
  error.value = ''

  try {
    const response = await fetch(`${API_BASE}/auth/login`, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({
        phone: phone.value,
        code: code.value
      })
    })

    const data = await response.json()

    if (response.ok) {
      if (data.needs_password) {
        step.value = 'password'
      } else {
        // Успішна авторизація
        sessionData.value = {
          phone: data.phone,
          session_data: data.session_data
        }
        // Зберігаємо в localStorage
        localStorage.setItem('telegram_session', JSON.stringify(sessionData.value))
        step.value = 'success'
        emit('authenticated', sessionData.value)
      }
    } else {
      error.value = data.error || 'Невірний код'
    }
  } catch (err) {
    error.value = 'Помилка з\'єднання з сервером'
  } finally {
    loading.value = false
  }
}

// Вхід з паролем 2FA
async function submitPassword() {
  if (!password.value) {
    error.value = 'Введіть пароль 2FA'
    return
  }

  loading.value = true
  error.value = ''

  try {
    const response = await fetch(`${API_BASE}/auth/password`, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({
        phone: phone.value,
        password: password.value
      })
    })

    const data = await response.json()

    if (response.ok) {
      // Успішна авторизація
      sessionData.value = {
        phone: data.phone,
        session_data: data.session_data
      }
      // Зберігаємо в localStorage
      localStorage.setItem('telegram_session', JSON.stringify(sessionData.value))
      step.value = 'success'
      emit('authenticated', sessionData.value)
    } else {
      error.value = data.error || 'Невірний пароль'
    }
  } catch (err) {
    error.value = 'Помилка з\'єднання з сервером'
  } finally {
    loading.value = false
  }
}

function reset() {
  step.value = 'phone'
  phone.value = ''
  code.value = ''
  password.value = ''
  error.value = ''
  sessionData.value = null
}
</script>

<template>
  <div class="auth-container">
    <h2>Telegram Gateway</h2>

    <!-- Крок 1: Введення номера телефону -->
    <div v-if="step === 'phone'" class="auth-step">
      <h3>Авторизація</h3>
      <p>Введіть номер телефону:</p>
      <input
        v-model="phone"
        type="tel"
        placeholder="+380501234567"
        @keyup.enter="requestCode"
        :disabled="loading"
      />
      <button @click="requestCode" :disabled="loading">
        {{ loading ? 'Відправка...' : 'Отримати код' }}
      </button>
    </div>

    <!-- Крок 2: Введення коду -->
    <div v-if="step === 'code'" class="auth-step">
      <h3>Код підтвердження</h3>
      <p>Введіть код з Telegram:</p>
      <input
        v-model="code"
        type="text"
        placeholder="12345"
        @keyup.enter="login"
        :disabled="loading"
      />
      <button @click="login" :disabled="loading">
        {{ loading ? 'Перевірка...' : 'Увійти' }}
      </button>
      <button @click="reset" class="secondary" :disabled="loading">
        Назад
      </button>
    </div>

    <!-- Крок 3: Введення пароля 2FA -->
    <div v-if="step === 'password'" class="auth-step">
      <h3>Двофакторна автентифікація</h3>
      <p>Введіть пароль 2FA:</p>
      <input
        v-model="password"
        type="password"
        placeholder="Пароль 2FA"
        @keyup.enter="submitPassword"
        :disabled="loading"
      />
      <button @click="submitPassword" :disabled="loading">
        {{ loading ? 'Перевірка...' : 'Увійти' }}
      </button>
      <button @click="reset" class="secondary" :disabled="loading">
        Назад
      </button>
    </div>

    <!-- Крок 4: Успіх -->
    <div v-if="step === 'success'" class="auth-step success">
      <h3>✅ Успішно!</h3>
      <p>Ви авторизовані як {{ sessionData?.phone }}</p>
    </div>

    <!-- Помилки -->
    <div v-if="error" class="error">
      {{ error }}
    </div>
  </div>
</template>

<style scoped>
.auth-container {
  max-width: 400px;
  margin: 50px auto;
  padding: 30px;
  border: 1px solid #ddd;
  border-radius: 8px;
  background: #fff;
}

h2 {
  text-align: center;
  color: #0088cc;
  margin-bottom: 30px;
}

h3 {
  margin-bottom: 10px;
  color: #333;
}

.auth-step {
  margin-top: 20px;
}

.auth-step.success {
  text-align: center;
  color: #28a745;
}

p {
  margin-bottom: 10px;
  color: #666;
}

input {
  width: 100%;
  padding: 12px;
  margin-bottom: 15px;
  border: 1px solid #ddd;
  border-radius: 4px;
  font-size: 16px;
  box-sizing: border-box;
}

input:focus {
  outline: none;
  border-color: #0088cc;
}

button {
  width: 100%;
  padding: 12px;
  background: #0088cc;
  color: white;
  border: none;
  border-radius: 4px;
  font-size: 16px;
  cursor: pointer;
  margin-bottom: 10px;
}

button:hover:not(:disabled) {
  background: #006699;
}

button:disabled {
  background: #ccc;
  cursor: not-allowed;
}

button.secondary {
  background: #6c757d;
}

button.secondary:hover:not(:disabled) {
  background: #5a6268;
}

.error {
  margin-top: 15px;
  padding: 12px;
  background: #f8d7da;
  color: #721c24;
  border: 1px solid #f5c6cb;
  border-radius: 4px;
}
</style>
