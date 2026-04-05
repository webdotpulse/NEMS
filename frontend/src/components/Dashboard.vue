<template>
  <main class="max-w-7xl mx-auto py-6 sm:px-6 lg:px-8">
    <div class="px-4 py-6 sm:px-0">

      <div class="mb-8">
        <PowerFlow :state="state" />
      </div>

      <div v-if="state" class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-6">

        <!-- Grid Power Card -->
        <div v-if="state.grid_power_w !== null" class="bg-white dark:bg-gray-800 overflow-hidden shadow rounded-lg border border-gray-100 dark:border-gray-700">
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
        <div v-if="state.solar_power_w !== null" class="bg-white dark:bg-gray-800 overflow-hidden shadow rounded-lg border border-gray-100 dark:border-gray-700">
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
        <div v-if="state.battery_power_w !== null" class="bg-white dark:bg-gray-800 overflow-hidden shadow rounded-lg border border-gray-100 dark:border-gray-700">
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
        <div v-if="state.total_load_w !== null" class="bg-white dark:bg-gray-800 overflow-hidden shadow rounded-lg border border-gray-100 dark:border-gray-700">
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


      <!-- Daily Aggregates Section (Energy Management) -->
      <div v-if="dailyAggregates" class="mt-8 mb-8 space-y-6">

        <!-- Energy Management Card -->
        <div class="bg-white dark:bg-gray-800 rounded-2xl shadow-sm border border-gray-100 dark:border-gray-700 p-6">

          <!-- Navigation -->
          <div class="flex flex-col sm:flex-row justify-between items-center mb-8 gap-4">
            <div class="bg-gray-100 dark:bg-gray-700 p-1 rounded-full flex w-full sm:w-auto">
    <button @click="setPeriod('day')" :class="['flex-1 sm:flex-none px-6 py-1.5 rounded-full text-sm font-medium', selectedPeriod === 'day' ? 'bg-white dark:bg-gray-600 shadow-sm text-gray-900 dark:text-white' : 'text-gray-500 hover:text-gray-700 dark:text-gray-300']">Day</button>
    <button @click="setPeriod('month')" :class="['flex-1 sm:flex-none px-6 py-1.5 rounded-full text-sm font-medium', selectedPeriod === 'month' ? 'bg-white dark:bg-gray-600 shadow-sm text-gray-900 dark:text-white' : 'text-gray-500 hover:text-gray-700 dark:text-gray-300']">Month</button>
    <button @click="setPeriod('year')" :class="['flex-1 sm:flex-none px-6 py-1.5 rounded-full text-sm font-medium', selectedPeriod === 'year' ? 'bg-white dark:bg-gray-600 shadow-sm text-gray-900 dark:text-white' : 'text-gray-500 hover:text-gray-700 dark:text-gray-300']">Year</button>
    <button @click="setPeriod('lifetime')" :class="['flex-1 sm:flex-none px-6 py-1.5 rounded-full text-sm font-medium', selectedPeriod === 'lifetime' ? 'bg-white dark:bg-gray-600 shadow-sm text-gray-900 dark:text-white' : 'text-gray-500 hover:text-gray-700 dark:text-gray-300']">Lifetime</button>
            </div>

  <div class="flex items-center space-x-4" v-if="selectedPeriod !== 'lifetime'">
              <button @click="changeDate(-1)" class="p-1 rounded-full hover:bg-gray-100 dark:hover:bg-gray-700">
                <svg class="w-5 h-5 text-gray-600 dark:text-gray-300" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M15 19l-7-7 7-7"></path></svg>
              </button>
    <input v-if="selectedPeriod === 'day'" type="date" v-model="selectedDate" @change="fetchEnergyData" class="font-medium text-gray-900 dark:text-white bg-transparent border-none focus:ring-0 cursor-pointer p-0" />
    <input v-else-if="selectedPeriod === 'month'" type="month" v-model="selectedDate" @change="fetchEnergyData" class="font-medium text-gray-900 dark:text-white bg-transparent border-none focus:ring-0 cursor-pointer p-0" />
    <span v-else-if="selectedPeriod === 'year'" class="font-medium text-gray-900 dark:text-white">{{ selectedDate.substring(0,4) }}</span>
              <button @click="changeDate(1)" class="p-1 rounded-full hover:bg-gray-100 dark:hover:bg-gray-700">
                <svg class="w-5 h-5 text-gray-600 dark:text-gray-300" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 5l7 7-7 7"></path></svg>
              </button>
            </div>
          </div>

          <!-- Production Block -->
          <div class="mb-8">
            <div class="flex items-center text-gray-500 dark:text-gray-400 mb-4">
              <span class="mr-1">Production</span>
            </div>

            <div class="bg-gray-50 dark:bg-gray-900 rounded-xl p-4 flex items-center justify-between relative">
              <div class="flex-1 text-center pr-12">
                <div class="text-lg font-semibold text-green-600"><span class="text-xl">{{ prodConsumed.toFixed(2) }}</span> <span class="text-sm font-normal">kWh</span></div>
                <div class="text-xs text-gray-500">Consumed</div>
                <div class="text-xs text-gray-400">({{ prodTotal > 0 ? ((prodConsumed / prodTotal) * 100).toFixed(2) : '0.00' }}%)</div>
              </div>

              <!-- Center Ring -->
              <div class="relative w-28 h-28 flex-shrink-0 z-10 flex flex-col items-center justify-center bg-white dark:bg-gray-800 rounded-full shadow-sm">
                <svg class="absolute inset-0 w-full h-full transform -rotate-90" viewBox="0 0 36 36">
                  <!-- Background circle (light green for Fed to grid) -->
                  <path class="text-green-200 dark:text-green-900" stroke-dasharray="100, 100" d="M18 2.0845 a 15.9155 15.9155 0 0 1 0 31.831 a 15.9155 15.9155 0 0 1 0 -31.831" fill="none" stroke="currentColor" stroke-width="4" stroke-linecap="round"/>
                  <!-- Foreground circle (dark green for Consumed) -->
                  <path class="text-green-600" :stroke-dasharray="`${prodTotal > 0 ? (prodConsumed / prodTotal) * 100 : 0}, 100`" d="M18 2.0845 a 15.9155 15.9155 0 0 1 0 31.831 a 15.9155 15.9155 0 0 1 0 -31.831" fill="none" stroke="currentColor" stroke-width="4" stroke-linecap="round"/>
                </svg>
                <div class="text-xl font-bold text-gray-900 dark:text-white">{{ prodTotal.toFixed(2) }}</div>
                <div class="text-sm text-gray-500">kWh</div>
              </div>

              <div class="flex-1 text-center pl-12">
                <div class="text-lg font-semibold text-green-400"><span class="text-xl">{{ prodFedToGrid.toFixed(2) }}</span> <span class="text-sm font-normal">kWh</span></div>
                <div class="text-xs text-gray-500">Fed to grid</div>
                <div class="text-xs text-gray-400">({{ prodTotal > 0 ? ((prodFedToGrid / prodTotal) * 100).toFixed(2) : '0.00' }}%)</div>
              </div>
            </div>
          </div>

          <!-- Consumption Block -->
          <div>
            <div class="flex items-center text-gray-500 dark:text-gray-400 mb-4">
              <span class="mr-1">Consumption</span>
            </div>

            <div class="bg-gray-50 dark:bg-gray-900 rounded-xl p-4 flex items-center justify-between relative">
              <div class="flex-1 text-center pr-12">
                <div class="text-lg font-semibold text-orange-500"><span class="text-xl">{{ consFromPV.toFixed(2) }}</span> <span class="text-sm font-normal">kWh</span></div>
                <div class="text-xs text-gray-500">From PV</div>
                <div class="text-xs text-gray-400">({{ consTotal > 0 ? ((consFromPV / consTotal) * 100).toFixed(2) : '0.00' }}%)</div>
              </div>

              <!-- Center Ring -->
              <div class="relative w-28 h-28 flex-shrink-0 z-10 flex flex-col items-center justify-center bg-white dark:bg-gray-800 rounded-full shadow-sm">
                <svg class="absolute inset-0 w-full h-full transform -rotate-90" viewBox="0 0 36 36">
                  <!-- Background circle (yellow for From grid) -->
                  <path class="text-yellow-400" stroke-dasharray="100, 100" d="M18 2.0845 a 15.9155 15.9155 0 0 1 0 31.831 a 15.9155 15.9155 0 0 1 0 -31.831" fill="none" stroke="currentColor" stroke-width="4" stroke-linecap="round"/>
                  <!-- Foreground circle (orange for From PV) -->
                  <path class="text-orange-500" :stroke-dasharray="`${consTotal > 0 ? (consFromPV / consTotal) * 100 : 0}, 100`" d="M18 2.0845 a 15.9155 15.9155 0 0 1 0 31.831 a 15.9155 15.9155 0 0 1 0 -31.831" fill="none" stroke="currentColor" stroke-width="4" stroke-linecap="round"/>
                </svg>
                <div class="text-xl font-bold text-gray-900 dark:text-white">{{ consTotal.toFixed(2) }}</div>
                <div class="text-sm text-gray-500">kWh</div>
              </div>

              <div class="flex-1 text-center pl-12">
                <div class="text-lg font-semibold text-yellow-500"><span class="text-xl">{{ consFromGrid.toFixed(2) }}</span> <span class="text-sm font-normal">kWh</span></div>
                <div class="text-xs text-gray-500">From grid</div>
                <div class="text-xs text-gray-400">({{ consTotal > 0 ? ((consFromGrid / consTotal) * 100).toFixed(2) : '0.00' }}%)</div>
              </div>
            </div>
          </div>

        </div>

        <!-- Chart Section -->
        <div v-if="energySeries && energySeries.length > 0" class="mt-8 bg-white dark:bg-gray-800 rounded-2xl shadow-sm border border-gray-100 dark:border-gray-700 p-6 h-96">
           <Bar :data="chartData" :options="chartOptions" />
        </div>

      </div>

      <div v-else class="border-4 border-dashed border-gray-200 dark:border-gray-700 rounded-lg h-96 flex flex-col items-center justify-center mt-6">
        <div class="text-center">
          <svg class="mx-auto h-12 w-12 text-gray-400 animate-pulse" fill="none" viewBox="0 0 24 24" stroke="currentColor" aria-hidden="true">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 19v-6a2 2 0 00-2-2H5a2 2 0 00-2 2v6a2 2 0 002 2h2a2 2 0 002-2zm0 0V9a2 2 0 012-2h2a2 2 0 012 2v10m-6 0a2 2 0 002 2h2a2 2 0 002-2m0 0V5a2 2 0 012-2h2a2 2 0 012 2v14a2 2 0 01-2 2h-2a2 2 0 01-2-2z" />
          </svg>
          <h3 class="mt-2 text-sm font-medium text-gray-900 dark:text-gray-100">Loading Energy Data...</h3>
          <p class="mt-1 text-sm text-gray-500 dark:text-gray-400">
            Fetching historical aggregates...
          </p>
        </div>
      </div>

    </div>
  </main>
</template>

<script setup lang="ts">
import { ref, onMounted, onUnmounted, computed } from 'vue'
import PowerFlow from './PowerFlow.vue'
import { getApiBase } from '../api'
import type { SiteState, DailyAggregates } from '../types'
import {
  Chart as ChartJS,
  CategoryScale,
  LinearScale,
  BarElement,
  Title,
  Tooltip,
  Legend,
  TimeScale
} from 'chart.js'
import { Bar } from 'vue-chartjs'
import 'chartjs-adapter-date-fns'

ChartJS.register(
  CategoryScale,
  LinearScale,
  BarElement,
  Title,
  Tooltip,
  Legend,
  TimeScale
)

const state = ref<SiteState | null>(null)
const dailyAggregates = ref<DailyAggregates | null>(null)
const selectedDate = ref<string>(new Date().toISOString().split('T')[0])
const selectedPeriod = ref<string>('day')
const energySeries = ref<any[]>([])
let eventSource: EventSource | null = null

const setPeriod = (period: string) => {
  selectedPeriod.value = period
  if (period === 'day') {
    selectedDate.value = new Date().toISOString().split('T')[0]
  } else if (period === 'month') {
    selectedDate.value = new Date().toISOString().substring(0, 7)
  } else if (period === 'year') {
    selectedDate.value = new Date().getFullYear().toString() + '-01-01'
  } else {
    selectedDate.value = ''
  }
  fetchEnergyData()
}

const changeDate = (offset: number) => {
  const d = new Date(selectedDate.value || new Date().toISOString())
  if (selectedPeriod.value === 'day') {
    d.setDate(d.getDate() + offset)
    selectedDate.value = d.toISOString().split('T')[0]
  } else if (selectedPeriod.value === 'month') {
    d.setMonth(d.getMonth() + offset)
    selectedDate.value = d.toISOString().substring(0, 7)
  } else if (selectedPeriod.value === 'year') {
    d.setFullYear(d.getFullYear() + offset)
    selectedDate.value = d.getFullYear().toString() + '-01-01'
  }
  fetchEnergyData()
}

const chartData = computed(() => {
  const series = energySeries.value || [];
  return {
    datasets: [
      {
        label: 'Solar Yield',
        data: series.map(d => ({ x: new Date(d.timestamp).getTime(), y: d.solar_yield_kwh || 0 })),
        backgroundColor: '#FBBF24',
        stack: 'Stack 0',
      },
      {
        label: 'Grid Import',
        data: series.map(d => ({ x: new Date(d.timestamp).getTime(), y: d.grid_import_kwh || 0 })),
        backgroundColor: '#3B82F6',
        stack: 'Stack 1',
      },
      {
        label: 'Battery Discharge',
        data: series.map(d => ({ x: new Date(d.timestamp).getTime(), y: d.battery_discharge_kwh || 0 })),
        backgroundColor: '#34D399',
        stack: 'Stack 1',
      },
      {
        label: 'House Consumption',
        data: series.map(d => ({ x: new Date(d.timestamp).getTime(), y: d.house_consumption_kwh || 0 })),
        backgroundColor: '#F97316',
        stack: 'Stack 2',
      }
    ]
  } as any;
});

const chartOptions = computed(() => ({
  responsive: true,
  maintainAspectRatio: false,
  scales: {
    x: {
      type: 'time' as const,
      time: {
        unit: selectedPeriod.value === 'day' ? 'hour' as const : selectedPeriod.value === 'month' ? 'day' as const : 'month' as const,
        tooltipFormat: 'PP HH:mm',
        displayFormats: {
          hour: 'HH:mm'
        }
      },
      stacked: true,
      ticks: { color: '#9CA3AF' },
      grid: { display: false }
    },
    y: {
      stacked: true,
      title: { display: true, text: 'Energy (kWh)', color: '#9CA3AF' },
      ticks: { color: '#9CA3AF' },
      grid: { color: '#374151' }
    }
  },
  plugins: {
    legend: {
      position: 'top' as const,
      labels: { color: '#9CA3AF' }
    }
  }
}));

const prodTotal = computed(() => dailyAggregates.value?.solar_yield_kwh || 0);
const prodFedToGrid = computed(() => dailyAggregates.value?.grid_export_kwh || 0);
const prodConsumed = computed(() => Math.max(0, prodTotal.value - prodFedToGrid.value));

const consTotal = computed(() => dailyAggregates.value?.house_consumption_kwh || 0);
const consFromGrid = computed(() => dailyAggregates.value?.grid_import_kwh || 0);
const consFromPV = computed(() => Math.max(0, consTotal.value - consFromGrid.value));

const fetchEnergyData = async () => {
  try {
    const res = await fetch(`${getApiBase()}/api/energy?period=${selectedPeriod.value}&date=${selectedDate.value}`)
    if (res.ok) {
      const data = await res.json()
      dailyAggregates.value = data.totals
      energySeries.value = data.series
    }
  } catch (e) {
    console.error("Failed to fetch energy data:", e)
  }
}

onMounted(async () => {

  // SSE Connection
  eventSource = new EventSource(`${getApiBase()}/api/live`)

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
  await fetchEnergyData()

})

onUnmounted(() => {
  if (eventSource) {
    eventSource.close()
  }
})
</script>
