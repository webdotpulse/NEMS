<template>
  <main class="max-w-7xl mx-auto py-6 sm:px-6 lg:px-8">
    <div class="px-4 py-6 sm:px-0">

      <div v-if="state" class="mb-8">
        <PowerFlow :state="state" />
      </div>

      <div v-if="state" class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-6">

        <!-- Grid Power Card -->
        <div v-if="state.grid_power_w !== null" class="bg-white dark:bg-gray-800 overflow-hidden shadow rounded-lg">
          <div class="p-5">
            <div class="flex items-center">
              <div class="flex-shrink-0">
                <svg xmlns="http://www.w3.org/2000/svg" class="h-6 w-6 text-gray-400" fill="currentColor" viewBox="0 0 24 24"><path d="M12 2C6.486 2 2 6.486 2 12s4.486 10 10 10 10-4.486 10-10S17.514 2 12 2zm0 18c-4.411 0-8-3.589-8-8s3.589-8 8-8 8 3.589 8 8-3.589 8-8 8z"/><path d="m13 6-6 7h4v5l6-7h-4z"/></svg>
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
        <div v-if="state.solar_power_w !== null" class="bg-white dark:bg-gray-800 overflow-hidden shadow rounded-lg">
          <div class="p-5">
            <div class="flex items-center">
              <div class="flex-shrink-0">
                <svg xmlns="http://www.w3.org/2000/svg" class="h-6 w-6 text-yellow-400" fill="currentColor" viewBox="0 0 24 24"><path d="M6.993 12c0 2.761 2.246 5.007 5.007 5.007s5.007-2.246 5.007-5.007S14.761 6.993 12 6.993 6.993 9.239 6.993 12zM12 8.993c1.658 0 3.007 1.349 3.007 3.007S13.658 15.007 12 15.007 8.993 13.658 8.993 12 10.342 8.993 12 8.993zM10.998 19h2v3h-2zm0-17h2v3h-2zm-9 9h3v2h-3zm17 0h3v2h-3zM4.219 18.363l2.12-2.122 1.415 1.414-2.12 2.122zM16.24 6.344l2.122-2.122 1.414 1.414-2.122 2.122zM6.342 7.759 4.22 5.637l1.415-1.414 2.12 2.122zm13.434 10.605-1.414 1.414-2.122-2.122 1.414-1.414z"/></svg>
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
        <div v-if="state.battery_power_w !== null" class="bg-white dark:bg-gray-800 overflow-hidden shadow rounded-lg">
          <div class="p-5">
            <div class="flex items-center">
              <div class="flex-shrink-0">
                <svg xmlns="http://www.w3.org/2000/svg" class="h-6 w-6 text-green-400" fill="currentColor" viewBox="0 0 24 24"><path d="M4 18h14c1.103 0 2-.897 2-2v-2h2v-4h-2V8c0-1.103-.897-2-2-2H4c-1.103 0-2 .897-2 2v8c0 1.103.897 2 2 2zM4 8h14l.002 8H4V8z"/></svg>
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
        <div v-if="state.total_load_w !== null" class="bg-white dark:bg-gray-800 overflow-hidden shadow rounded-lg">
          <div class="p-5">
            <div class="flex items-center">
              <div class="flex-shrink-0">
                <svg xmlns="http://www.w3.org/2000/svg" class="h-6 w-6 text-purple-400" fill="currentColor" viewBox="0 0 24 24"><path d="M3 13h1v7c0 1.103.897 2 2 2h12c1.103 0 2-.897 2-2v-7h1a1 1 0 0 0 .707-1.707l-9-9a.999.999 0 0 0-1.414 0l-9 9A1 1 0 0 0 3 13zm7 7v-5h4v5h-4zm2-15.586 6 6V15l.001 5H16v-5c0-1.103-.897-2-2-2h-4c-1.103 0-2 .897-2 2v5H6v-9.586l6-6z"/></svg>
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

      <!-- Daily Aggregates Section -->
      <div v-if="dailyAggregates" class="mt-8 bg-white dark:bg-gray-800 overflow-hidden shadow rounded-lg">
        <div class="px-4 py-5 sm:p-6">
          <h3 class="text-lg leading-6 font-medium text-gray-900 dark:text-gray-100 mb-4">Daily Summary</h3>
          <div class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-4">
            <div class="bg-gray-50 dark:bg-gray-700 p-4 rounded-md">
              <dt class="text-sm font-medium text-gray-500 dark:text-gray-400">Grid (Today)</dt>
              <dd class="mt-1 text-sm text-gray-900 dark:text-gray-100">
                <div>Import: {{ dailyAggregates.grid_import_kwh.toFixed(2) }} kWh</div>
                <div>Export: {{ dailyAggregates.grid_export_kwh.toFixed(2) }} kWh</div>
              </dd>
            </div>
            <div class="bg-gray-50 dark:bg-gray-700 p-4 rounded-md">
              <dt class="text-sm font-medium text-gray-500 dark:text-gray-400">Solar Yield (Today)</dt>
              <dd class="mt-1 text-sm text-gray-900 dark:text-gray-100">
                <div>Production: {{ dailyAggregates.solar_yield_kwh.toFixed(2) }} kWh</div>
              </dd>
            </div>
            <div class="bg-gray-50 dark:bg-gray-700 p-4 rounded-md">
              <dt class="text-sm font-medium text-gray-500 dark:text-gray-400">Battery (Today)</dt>
              <dd class="mt-1 text-sm text-gray-900 dark:text-gray-100">
                <div>Charged: {{ dailyAggregates.battery_charge_kwh.toFixed(2) }} kWh</div>
                <div class="ml-2 text-xs text-gray-500 dark:text-gray-400">From Solar: {{ dailyAggregates.battery_charge_solar_kwh.toFixed(2) }} kWh</div>
                <div class="ml-2 text-xs text-gray-500 dark:text-gray-400">From Grid: {{ dailyAggregates.battery_charge_grid_kwh.toFixed(2) }} kWh</div>
                <div>Discharged: {{ dailyAggregates.battery_discharge_kwh.toFixed(2) }} kWh</div>
              </dd>
            </div>
            <div class="bg-gray-50 dark:bg-gray-700 p-4 rounded-md">
              <dt class="text-sm font-medium text-gray-500 dark:text-gray-400">House (Today)</dt>
              <dd class="mt-1 text-sm text-gray-900 dark:text-gray-100">
                <div>Consumption: {{ dailyAggregates.house_consumption_kwh.toFixed(2) }} kWh</div>
              </dd>
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
  grid_power_w: number | null
  solar_power_w: number | null
  battery_power_w: number | null
  battery_soc: number | null
  total_load_w: number | null
  ev_charger_power_w: number | null
  device_health?: Record<number, string>
}

interface DailyAggregates {
  grid_import_kwh: number
  grid_export_kwh: number
  solar_yield_kwh: number
  battery_charge_kwh: number
  battery_charge_solar_kwh: number
  battery_charge_grid_kwh: number
  battery_discharge_kwh: number
  house_consumption_kwh: number
}

const state = ref<SiteState | null>(null)
const dailyAggregates = ref<DailyAggregates | null>(null)
let eventSource: EventSource | null = null

onMounted(async () => {
  const host = window.location.hostname

  // SSE Connection
  eventSource = new EventSource(`http://${host}:8080/api/live`)

  eventSource.onopen = () => {
    console.log("SSE connected")
  }

  eventSource.onmessage = (event) => {
    try {
      const data = JSON.parse(event.data)
      // Check if it's actual state data and not just an empty connection message
      if (data.grid_power_w !== undefined || data.total_load_w !== undefined) {
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

  // Fetch daily aggregates
  try {
    const res = await fetch(`http://${host}:8080/api/daily`)
    if (res.ok) {
      dailyAggregates.value = await res.json()
    }
  } catch (e) {
    console.error("Failed to fetch daily aggregates:", e)
  }

})

onUnmounted(() => {
  if (eventSource) {
    eventSource.close()
  }
})
</script>
