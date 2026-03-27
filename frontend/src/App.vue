<script setup lang="ts">
import { ref, onMounted, onUnmounted } from 'vue'
import NavBar from './components/NavBar.vue'
import Dashboard from './components/Dashboard.vue'
import Settings from './components/Settings.vue'
import Logger from './components/Logger.vue'
import { getApiBase } from './api'

const isConnected = ref(false)
const currentView = ref('dashboard') // 'dashboard' | 'settings' | 'logger'
let pollingInterval: number | undefined

const checkConnection = async () => {
  try {
    const response = await fetch(`${getApiBase()}/api/status`)
    if (response.ok) {
      const data = await response.json()
      isConnected.value = data.status === 'ok'
    } else {
      isConnected.value = false
    }
  } catch (error) {
    isConnected.value = false
  }
}

onMounted(() => {
  // Initial check
  checkConnection()
  // Poll every 2 seconds
  pollingInterval = window.setInterval(checkConnection, 2000)
})

onUnmounted(() => {
  if (pollingInterval) {
    clearInterval(pollingInterval)
  }
})

const setView = (view: string) => {
  currentView.value = view
}
</script>

<template>
  <div class="min-h-screen bg-gray-50 dark:bg-gray-900 transition-colors duration-200">
    <NavBar :connected="isConnected" :currentView="currentView" @navigate="setView" />
    <Dashboard v-if="currentView === 'dashboard'" />
    <Settings v-if="currentView === 'settings'" />
    <Logger v-if="currentView === 'logger'" />
  </div>
</template>

<style scoped>
</style>
