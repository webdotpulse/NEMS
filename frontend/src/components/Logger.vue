<template>
  <main class="max-w-7xl mx-auto py-6 sm:px-6 lg:px-8">
    <div class="px-4 py-6 sm:px-0">
      <div class="mb-4 flex items-center justify-between">
        <div class="flex items-center gap-4">
          <h2 class="text-2xl font-bold leading-7 text-gray-900 dark:text-white sm:text-3xl sm:truncate">
            System Logs
          </h2>
          <span v-if="isPaused" class="inline-flex items-center rounded-md bg-yellow-50 dark:bg-yellow-900 px-2 py-1 text-xs font-medium text-yellow-800 dark:text-yellow-200 ring-1 ring-inset ring-yellow-600/20">
            Paused
          </span>
        </div>

        <div class="flex space-x-2">
          <button @click="togglePause" :class="[isPaused ? 'bg-green-600 hover:bg-green-700' : 'bg-yellow-600 hover:bg-yellow-700', 'inline-flex items-center px-3 py-2 border border-transparent shadow-sm text-sm font-medium rounded-md text-white focus:outline-none focus:ring-2 focus:ring-offset-2 transition-colors']">
            <svg v-if="isPaused" class="-ml-1 mr-2 h-5 w-5" xmlns="http://www.w3.org/2000/svg" viewBox="0 0 20 20" fill="currentColor">
              <path d="M6.3 2.841A1.5 1.5 0 004 4.11V15.89a1.5 1.5 0 002.3 1.269l9.344-5.89a1.5 1.5 0 000-2.538L6.3 2.84z" />
            </svg>
            <svg v-else class="-ml-1 mr-2 h-5 w-5" xmlns="http://www.w3.org/2000/svg" viewBox="0 0 20 20" fill="currentColor">
              <path d="M5.75 3a.75.75 0 00-.75.75v12.5c0 .414.336.75.75.75h1.5a.75.75 0 00.75-.75V3.75A.75.75 0 007.25 3h-1.5zM12.75 3a.75.75 0 00-.75.75v12.5c0 .414.336.75.75.75h1.5a.75.75 0 00.75-.75V3.75a.75.75 0 00-.75-.75h-1.5z" />
            </svg>
            {{ isPaused ? 'Resume' : 'Pause' }}
          </button>

          <button @click="handleClear" class="inline-flex items-center px-3 py-2 border border-gray-300 dark:border-gray-600 shadow-sm text-sm font-medium rounded-md text-gray-700 dark:text-gray-200 bg-white dark:bg-gray-800 hover:bg-gray-50 dark:hover:bg-gray-700 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-indigo-500 transition-colors">
            <svg class="-ml-1 mr-2 h-5 w-5" xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" stroke="currentColor">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M19 7l-.867 12.142A2 2 0 0116.138 21H7.862a2 2 0 01-1.995-1.858L5 7m5 4v6m4-6v6m1-10V4a1 1 0 00-1-1h-4a1 1 0 00-1 1v3M4 7h16" />
            </svg>
            Clear View
          </button>

          <button @click="fetchLogsManual" class="inline-flex items-center px-3 py-2 border border-transparent shadow-sm text-sm font-medium rounded-md text-white bg-indigo-600 hover:bg-indigo-700 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-indigo-500 transition-colors">
            <svg class="-ml-1 mr-2 h-5 w-5" xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" stroke="currentColor">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M4 4v5h.582m15.356 2A8.001 8.001 0 004.582 9m0 0H9m11 11v-5h-.581m0 0a8.003 8.003 0 01-15.357-2m15.357 2H15" />
            </svg>
            Refresh
          </button>
        </div>
      </div>

      <div class="bg-gray-900 dark:bg-black rounded-lg shadow overflow-hidden h-[600px] flex flex-col">
        <div class="p-4 flex-1 overflow-y-auto font-mono text-xs sm:text-sm text-gray-300" ref="logContainer" @scroll="handleScroll">
          <div v-if="parsedLogs.length === 0" class="text-gray-500 italic">No logs available...</div>
          <table v-else class="min-w-full divide-y divide-gray-800">
            <tbody class="divide-y divide-gray-800/50">
              <tr v-for="(log, index) in parsedLogs" :key="index" class="hover:bg-gray-800/50 transition-colors">
                <td class="py-1 pr-2 whitespace-nowrap text-gray-500 align-top w-1 shrink-0">
                  {{ log.timestamp }}
                </td>
                <td class="py-1 px-2 whitespace-nowrap align-top w-1 shrink-0">
                  <span :class="getLevelClass(log.level)">{{ log.level }}</span>
                </td>
                <td v-if="log.module" class="py-1 px-2 whitespace-nowrap text-indigo-400 align-top w-1 shrink-0">
                  [{{ log.module }}]
                </td>
                <td class="py-1 pl-2 break-all align-top text-gray-300">
                  {{ log.message }}
                </td>
              </tr>
            </tbody>
          </table>
        </div>
      </div>
    </div>
  </main>
</template>

<script setup lang="ts">
import { ref, onMounted, onUnmounted, nextTick, computed } from 'vue'
import { getApiBase } from '../api'

interface ParsedLog {
  timestamp: string;
  level: string;
  module: string;
  message: string;
  original: string;
}

const logs = ref<string[]>([])
const isPaused = ref(false)
const logContainer = ref<HTMLElement | null>(null)
let pollInterval: number | null = null
let userIsScrolling = false
let lastLogMarker = ''

const parseLogLine = (line: string): ParsedLog => {
  // Regex to match typical Go log formats: "2024/02/09 10:15:30 INFO [ModuleName] The log message..."
  // Or "2024-02-09T10:15:30.123Z [INFO] [ModuleName] The log message..."
  const parsed: ParsedLog = { timestamp: '', level: 'INFO', module: '', message: line, original: line }

  // Basic heuristic parser
  const parts = line.split(' ')
  if (parts.length < 2) return parsed

  // Simple check for timestamp (looks like date/time)
  if (parts[0].includes('/') || parts[0].includes('-')) {
    parsed.timestamp = parts[0]
    if (parts[1].includes(':')) {
       parsed.timestamp += ' ' + parts[1]
       parts.splice(0, 2)
    } else {
       parts.splice(0, 1)
    }
  }

  // Next part might be level
  if (parts.length > 0) {
     const p = parts[0].replace(/\[|\]/g, '').toUpperCase()
     if (['INFO', 'WARN', 'ERROR', 'DEBUG', 'FATAL', 'TRACE'].includes(p)) {
         parsed.level = p
         parts.splice(0, 1)
     }
  }

  // Next part might be module in brackets
  if (parts.length > 0 && parts[0].startsWith('[') && parts[0].endsWith(']')) {
      parsed.module = parts[0].substring(1, parts[0].length - 1)
      parts.splice(0, 1)
  }

  parsed.message = parts.join(' ')
  return parsed
}

const parsedLogs = computed(() => {
  return logs.value.map(parseLogLine)
})

const getLevelClass = (level: string) => {
  switch (level) {
    case 'ERROR':
    case 'FATAL':
      return 'text-red-400 font-bold'
    case 'WARN':
      return 'text-yellow-400 font-bold'
    case 'DEBUG':
      return 'text-gray-400'
    case 'INFO':
      return 'text-blue-400 font-bold'
    default:
      return 'text-gray-300'
  }
}

const togglePause = () => {
  isPaused.value = !isPaused.value
}

const handleClear = () => {
  // Mark the last log seen so we can slice out past logs on the next fetch
  if (logs.value.length > 0) {
    lastLogMarker = logs.value[logs.value.length - 1]
  }
  logs.value = []
}

const handleScroll = () => {
  if (!logContainer.value) return

  const { scrollTop, scrollHeight, clientHeight } = logContainer.value
  // If user scrolled up more than 50px from bottom, consider they are scrolling
  userIsScrolling = scrollHeight - scrollTop - clientHeight > 50
}

const fetchLogsData = async () => {
  if (isPaused.value) return

  try {
    const res = await fetch(`${getApiBase()}/api/logs`)
    if (res.ok) {
      let newLogs: string[] = await res.json()

      // If the user clicked "Clear View", we have a marker of the last seen log
      if (lastLogMarker) {
         const idx = newLogs.lastIndexOf(lastLogMarker)
         if (idx !== -1) {
            // Keep only logs that came AFTER the last seen log
            newLogs = newLogs.slice(idx + 1)
         } else {
            // Marker fell out of the backend ring buffer entirely.
            // We just reset the marker and accept the new array to avoid being permanently blank.
            lastLogMarker = ''
         }
      }

      logs.value = newLogs

      if (!userIsScrolling) {
        await nextTick()
        if (logContainer.value) {
          logContainer.value.scrollTop = logContainer.value.scrollHeight
        }
      }
    }
  } catch (e) {
    console.error("Failed to fetch logs:", e)
  }
}

const fetchLogsManual = async () => {
  isPaused.value = false // Resume on manual refresh
  await fetchLogsData()
}

onMounted(() => {
  fetchLogsData()
  pollInterval = window.setInterval(fetchLogsData, 5000)
})

onUnmounted(() => {
  if (pollInterval) {
    clearInterval(pollInterval)
  }
})
</script>