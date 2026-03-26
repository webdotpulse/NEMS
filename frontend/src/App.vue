<script setup lang="ts">
import { ref, onMounted, onUnmounted } from 'vue'
import NavBar from './components/NavBar.vue'
import Dashboard from './components/Dashboard.vue'

const isConnected = ref(false)
let pollingInterval: number | undefined

const checkConnection = async () => {
  try {
    const response = await fetch('http://localhost:8080/api/status')
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
</script>

<template>
  <div class="min-h-screen bg-gray-50 dark:bg-gray-900 transition-colors duration-200">
    <NavBar :connected="isConnected" />
    <Dashboard />
  </div>
</template>

<style scoped>
</style>
