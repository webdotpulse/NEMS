<template>
  <main class="max-w-7xl mx-auto py-6 sm:px-6 lg:px-8">
    <div class="px-4 py-6 sm:px-0">
      <div class="mb-4 flex items-center justify-between">
        <h2 class="text-2xl font-bold leading-7 text-gray-900 dark:text-white sm:text-3xl sm:truncate">
          System Logs
        </h2>
        <button @click="fetchLogs" class="inline-flex items-center px-4 py-2 border border-transparent shadow-sm text-sm font-medium rounded-md text-white bg-indigo-600 hover:bg-indigo-700 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-indigo-500">
          <svg class="-ml-1 mr-2 h-5 w-5" xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" stroke="currentColor">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M4 4v5h.582m15.356 2A8.001 8.001 0 004.582 9m0 0H9m11 11v-5h-.581m0 0a8.003 8.003 0 01-15.357-2m15.357 2H15" />
          </svg>
          Refresh
        </button>
      </div>

      <div class="bg-gray-900 dark:bg-black rounded-lg shadow overflow-hidden h-[600px] flex flex-col">
        <div class="p-4 flex-1 overflow-y-auto font-mono text-xs sm:text-sm text-gray-300" ref="logContainer">
          <div v-if="logs.length === 0" class="text-gray-500 italic">No logs available...</div>
          <div v-for="(log, index) in logs" :key="index" class="whitespace-pre-wrap break-all mb-1 hover:bg-gray-800 px-1 rounded">{{ log }}</div>
        </div>
      </div>
    </div>
  </main>
</template>

<script setup lang="ts">
import { ref, onMounted, onUnmounted, nextTick } from 'vue'
import { getApiBase } from '../api'

const logs = ref<string[]>([])
const logContainer = ref<HTMLElement | null>(null)
let pollInterval: number | null = null

const fetchLogs = async () => {
  try {
    const res = await fetch(`${getApiBase()}/api/logs`)
    if (res.ok) {
      logs.value = await res.json()
      // Auto scroll to bottom
      await nextTick()
      if (logContainer.value) {
        logContainer.value.scrollTop = logContainer.value.scrollHeight
      }
    }
  } catch (e) {
    console.error("Failed to fetch logs:", e)
  }
}

onMounted(() => {
  fetchLogs()
  pollInterval = window.setInterval(fetchLogs, 5000)
})

onUnmounted(() => {
  if (pollInterval) {
    clearInterval(pollInterval)
  }
})
</script>
