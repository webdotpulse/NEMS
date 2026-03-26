<template>
  <main class="max-w-7xl mx-auto py-6 sm:px-6 lg:px-8">
    <div class="px-4 py-6 sm:px-0">

      <div v-if="state" class="mb-8">
        <PowerFlow :state="state" />
      </div>

      <div v-if="state" class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-6">

        <!-- Grid Power Card -->
        <div class="bg-white dark:bg-gray-800 overflow-hidden shadow rounded-lg">
          <div class="p-5">
            <div class="flex items-center">
              <div class="flex-shrink-0">
                <svg class="h-6 w-6 text-gray-400" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                  <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M13 10V3L4 14h7v7l9-11h-7z" />
                </svg>
              </div>
              <div class="ml-5 w-0 flex-1">
                <dl>
                  <dt class="text-sm font-medium text-gray-500 dark:text-gray-400 truncate">Grid Power</dt>
                  <dd>
                    <div class="text-lg font-medium text-gray-900 dark:text-gray-100">
                      {{ Math.abs(state.grid_power_w).toFixed(0) }} W
                    </div>
                    <div class="text-sm" :class="state.grid_power_w > 0 ? 'text-red-500' : 'text-green-500'">
                      {{ state.grid_power_w > 0 ? 'Importing' : (state.grid_power_w < 0 ? 'Exporting' : 'Idle') }}
                    </div>
                  </dd>
                </dl>
              </div>
            </div>
          </div>
        </div>

        <!-- Solar Power Card -->
        <div class="bg-white dark:bg-gray-800 overflow-hidden shadow rounded-lg">
          <div class="p-5">
            <div class="flex items-center">
              <div class="flex-shrink-0">
                <svg class="h-6 w-6 text-yellow-400" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                  <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 3v1m0 16v1m9-9h-1M4 12H3m15.364 6.364l-.707-.707M6.343 6.343l-.707-.707m12.728 0l-.707.707M6.343 17.657l-.707.707M16 12a4 4 0 11-8 0 4 4 0 018 0z" />
                </svg>
              </div>
              <div class="ml-5 w-0 flex-1">
                <dl>
                  <dt class="text-sm font-medium text-gray-500 dark:text-gray-400 truncate">Solar Power</dt>
                  <dd>
                    <div class="text-lg font-medium text-gray-900 dark:text-gray-100">
                      {{ state.solar_power_w.toFixed(0) }} W
                    </div>
                    <div class="text-sm text-yellow-500">Producing</div>
                  </dd>
                </dl>
              </div>
            </div>
          </div>
        </div>

        <!-- Battery Power Card -->
        <div class="bg-white dark:bg-gray-800 overflow-hidden shadow rounded-lg">
          <div class="p-5">
            <div class="flex items-center">
              <div class="flex-shrink-0">
                <svg class="h-6 w-6 text-green-400" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                  <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M4 7v10c0 2.21 3.582 4 8 4s8-1.79 8-4V7M4 7c0 2.21 3.582 4 8 4s8-1.79 8-4M4 7c0-2.21 3.582-4 8-4s8 1.79 8 4" />
                </svg>
              </div>
              <div class="ml-5 w-0 flex-1">
                <dl>
                  <dt class="text-sm font-medium text-gray-500 dark:text-gray-400 truncate">Battery Power</dt>
                  <dd>
                    <div class="text-lg font-medium text-gray-900 dark:text-gray-100">
                      {{ Math.abs(state.battery_power_w).toFixed(0) }} W
                    </div>
                    <div class="text-sm" :class="state.battery_power_w > 0 ? 'text-blue-500' : 'text-green-500'">
                      {{ state.battery_power_w > 0 ? 'Discharging' : (state.battery_power_w < 0 ? 'Charging' : 'Idle') }}
                    </div>
                  </dd>
                </dl>
              </div>
            </div>
          </div>
        </div>

        <!-- Total Load Card -->
        <div class="bg-white dark:bg-gray-800 overflow-hidden shadow rounded-lg">
          <div class="p-5">
            <div class="flex items-center">
              <div class="flex-shrink-0">
                <svg class="h-6 w-6 text-purple-400" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                  <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M3 12l2-2m0 0l7-7 7 7M5 10v10a1 1 0 001 1h3m10-11l2 2m-2-2v10a1 1 0 01-1 1h-3m-6 0a1 1 0 001-1v-4a1 1 0 011-1h2a1 1 0 011 1v4a1 1 0 001 1m-6 0h6" />
                </svg>
              </div>
              <div class="ml-5 w-0 flex-1">
                <dl>
                  <dt class="text-sm font-medium text-gray-500 dark:text-gray-400 truncate">Total Load</dt>
                  <dd>
                    <div class="text-lg font-medium text-gray-900 dark:text-gray-100">
                      {{ state.total_load_w.toFixed(0) }} W
                    </div>
                    <div class="text-sm text-purple-500">Consuming</div>
                  </dd>
                </dl>
              </div>
            </div>
          </div>
        </div>

      </div>

      <div v-else class="border-4 border-dashed border-gray-200 dark:border-gray-700 rounded-lg h-96 flex flex-col items-center justify-center mt-6">
        <div class="text-center">
          <svg class="mx-auto h-12 w-12 text-gray-400 animate-pulse" fill="none" viewBox="0 0 24 24" stroke="currentColor" aria-hidden="true">
            <path vector-effect="non-scaling-stroke" stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M13 10V3L4 14h7v7l9-11h-7z" />
          </svg>
          <h3 class="mt-2 text-sm font-medium text-gray-900 dark:text-gray-100">Connecting to Energy Engine...</h3>
          <p class="mt-1 text-sm text-gray-500 dark:text-gray-400">
            Waiting for live device data...
          </p>
        </div>
      </div>
    </div>
  </main>
</template>

<script setup lang="ts">
import { ref, onMounted, onUnmounted } from 'vue'
import PowerFlow from './PowerFlow.vue'

interface SiteState {
  grid_power_w: number
  solar_power_w: number
  battery_power_w: number
  total_load_w: number
  ev_charger_power_w: number
}

const state = ref<SiteState | null>(null)
let eventSource: EventSource | null = null

onMounted(() => {
  const host = window.location.hostname
  eventSource = new EventSource(`http://${host}:8080/api/live`)

  eventSource.onopen = () => {
    console.log("SSE connected")
  }

  eventSource.onmessage = (event) => {
    try {
      const data = JSON.parse(event.data)
      // Check if it's actual state data and not just an empty connection message
      if (data.grid_power_w !== undefined) {
        state.value = data as SiteState
      }
    } catch (e) {
      console.error('Failed to parse SSE message', e)
    }
  }

  eventSource.onerror = (error) => {
    console.error('SSE Error:', error)
    // Don't null out state on every error, wait to see if it reconnects
  }
})

onUnmounted(() => {
  if (eventSource) {
    eventSource.close()
  }
})
</script>
