<template>
  <div class="w-full flex justify-center items-center py-8">
    <div class="relative w-[340px] h-[460px]">

      <!-- SVG paths for animated power flow lines -->
      <svg class="absolute inset-0 w-full h-full z-0 pointer-events-none" viewBox="0 0 340 460" preserveAspectRatio="none">

        <!-- Grid to Junction (Vertical line) -->
        <!-- Center x = 170. Grid y ends at ~115 (153-24/2 for bottom padding), Junction y is at 230 -->
        <line v-if="hasGrid" x1="170" y1="120" x2="170" y2="230" stroke="#E5E7EB" stroke-width="6" />
        <line v-if="hasGrid && state?.grid_power_w !== null" x1="170" y1="120" x2="170" y2="230" stroke="#9CA3AF" stroke-width="4" stroke-dasharray="8 8" class="flow-path" :style="getFlowStyle(state?.grid_power_w, false)" />

        <!-- Home to Junction (Vertical line) -->
        <line x1="170" y1="230" x2="170" y2="330" stroke="#E5E7EB" stroke-width="6" />
        <line v-if="state?.total_load_w !== null" x1="170" y1="230" x2="170" y2="330" stroke="#A855F7" stroke-width="4" stroke-dasharray="8 8" class="flow-path" :style="getFlowStyle(state?.total_load_w, true)" />

        <!-- Battery to Junction (Horizontal line) -->
        <line v-if="hasBattery" x1="100" y1="230" x2="170" y2="230" stroke="#E5E7EB" stroke-width="6" />
        <line v-if="hasBattery && state?.battery_power_w !== null" x1="100" y1="230" x2="170" y2="230" stroke="#34D399" stroke-width="4" stroke-dasharray="8 8" class="flow-path" :style="getFlowStyle(state?.battery_power_w, false)" />

        <!-- Solar to Junction (Horizontal line) -->
        <line v-if="hasSolar" x1="170" y1="230" x2="240" y2="230" stroke="#E5E7EB" stroke-width="6" />
        <line v-if="hasSolar && state?.solar_power_w !== null" x1="240" y1="230" x2="170" y2="230" stroke="#FBBF24" stroke-width="4" stroke-dasharray="8 8" class="flow-path" :style="getFlowStyle(state?.solar_power_w, true)" />

      </svg>

      <!-- Grid Layout for Nodes -->
      <div class="absolute inset-0 grid grid-cols-3 grid-rows-3 gap-0">
        <!-- Top Row: Grid -->
        <div class="col-start-2 row-start-1 flex flex-col items-center justify-end pb-2">
          <div v-if="hasGrid" @click="openChart('grid')" class="z-10 flex items-center justify-center w-24 h-24 bg-white dark:bg-gray-800 rounded-full border-[4px] border-[#E5E7EB] dark:border-gray-600 shadow-sm cursor-pointer hover:scale-105 transition-transform mb-2">
            <!-- Icon -->
            <svg class="h-10 w-10 text-gray-700 dark:text-gray-300" viewBox="0 0 24 24" fill="currentColor">
              <path d="M16,7V3h-2v4h-4V3H8v4H6v7.5L9.5,18v3h5v-3l3.5-3.5V7H16z M14,15.5l-2,2l-2-2V13h4V15.5z M14,11h-4V9h4V11z"/>
            </svg>
          </div>
          <!-- Value Display -->
          <div v-if="hasGrid" class="z-10 bg-[#F3F4F6] dark:bg-gray-700 text-gray-800 dark:text-gray-200 text-sm font-medium px-4 py-1 rounded-full shadow-sm">
            <span v-if="state?.grid_power_w !== null && state?.grid_power_w !== undefined">
               {{ formatPower(state.grid_power_w, false) }}
            </span>
            <span v-else>0.0 kW</span>
          </div>
        </div>

        <!-- Middle Row: Battery, Junction, Solar -->
        <div class="col-start-1 row-start-2 flex flex-col items-center justify-center">
          <div v-if="hasBattery" @click="openChart('battery')" class="z-10 flex items-center justify-center w-24 h-24 bg-white dark:bg-gray-800 rounded-full border-[4px] border-[#34D399] shadow-sm cursor-pointer hover:scale-105 transition-transform mb-2">
            <!-- Icon -->
            <svg class="h-10 w-10 text-gray-700 dark:text-gray-300" viewBox="0 0 24 24" fill="currentColor">
              <path d="M16,4h-2V2h-4v2H8C6.9,4,6,4.9,6,6v14c0,1.1,0.9,2,2,2h8c1.1,0,2-0.9,2-2V6C18,4.9,17.1,4,16,4z M16,20H8V6h8V20z M10,8h4v2h-4V8z M10,12h4v2h-4V12z M10,16h4v2h-4V16z"/>
            </svg>
          </div>
          <!-- Value Display -->
          <div v-if="hasBattery" class="z-10 flex flex-col items-center space-y-1">
            <div class="bg-[#F3F4F6] dark:bg-gray-700 text-gray-800 dark:text-gray-200 text-sm font-medium px-4 py-1 rounded-full shadow-sm">
              <span v-if="state?.battery_power_w !== null && state?.battery_power_w !== undefined">
                 {{ formatPower(state.battery_power_w, false, true) }}
              </span>
              <span v-else>0.0 kW</span>
            </div>
            <div class="bg-[#F3F4F6] dark:bg-gray-700 text-gray-800 dark:text-gray-200 text-sm font-medium px-4 py-1 rounded-full shadow-sm">
              <span v-if="state?.battery_soc !== null && state?.battery_soc !== undefined">{{ Math.round(state.battery_soc) }}%</span>
              <span v-else>--%</span>
            </div>
          </div>
        </div>

        <div class="col-start-2 row-start-2 flex items-center justify-center">
          <!-- Junction Node -->
          <div class="z-10 flex items-center justify-center w-[4.5rem] h-[4.5rem] bg-white dark:bg-gray-800 rounded-full border-[4px] border-[#5EEAD4] shadow-sm">
            <svg class="h-6 w-6 text-gray-700 dark:text-gray-300" viewBox="0 0 24 24" fill="currentColor">
              <path d="M11 21h-1l1-7H7.5c-.58 0-.57-.32-.38-.66.19-.34.05-.08.16-.28L11.5 2h1l-1 7h3.5c.49 0 .56.33.47.51l-.07.15C12.96 14.55 11 21 11 21z" />
            </svg>
          </div>
        </div>

        <div class="col-start-3 row-start-2 flex flex-col items-center justify-center">
          <div v-if="hasSolar" @click="openChart('solar')" class="z-10 flex items-center justify-center w-24 h-24 bg-white dark:bg-gray-800 rounded-full border-[4px] border-[#FBBF24] shadow-sm cursor-pointer hover:scale-105 transition-transform mb-2">
            <!-- Icon -->
            <svg class="h-10 w-10 text-gray-700 dark:text-gray-300 relative" viewBox="0 0 24 24" fill="currentColor">
              <!-- Outline panel -->
              <path d="M3 4h18v16H3V4zM5 6v12h14V6H5z"/>
              <!-- Panel grid -->
              <path d="M11 6v12h2V6h-2zM7 6v12h2V6H7zM15 6v12h2V6h-2z"/>
              <path d="M5 11h14v2H5v-2zM5 15h14v2H5v-2z"/>
              <!-- Sun ray dot top right -->
              <circle cx="21" cy="3" r="1.5" />
            </svg>
          </div>
          <!-- Value Display -->
          <div v-if="hasSolar" class="z-10 bg-[#F3F4F6] dark:bg-gray-700 text-gray-800 dark:text-gray-200 text-sm font-medium px-4 py-1 rounded-full shadow-sm">
            <span v-if="state?.solar_power_w !== null && state?.solar_power_w !== undefined">
               {{ formatPower(state.solar_power_w, true, true) }}
            </span>
            <span v-else>0.0 kW</span>
          </div>
        </div>

        <!-- Bottom Row: Home -->
        <div class="col-start-2 row-start-3 flex flex-col items-center justify-start pt-2">
          <div class="z-10 flex items-center justify-center w-24 h-24 bg-white dark:bg-gray-800 rounded-full border-[4px] border-[#A855F7] shadow-sm mb-2">
            <!-- Icon -->
            <svg class="h-10 w-10 text-gray-700 dark:text-gray-300" viewBox="0 0 24 24" fill="currentColor">
              <path d="M12 3L4 9v12h16V9l-8-6zm6 16h-3v-4H9v4H6v-9l6-4.5 6 4.5v9z"/>
            </svg>
          </div>
          <!-- Value Display -->
          <div class="z-10 bg-[#F3F4F6] dark:bg-gray-700 text-gray-800 dark:text-gray-200 text-sm font-medium px-4 py-1 rounded-full shadow-sm">
            <span v-if="state?.total_load_w !== null && state?.total_load_w !== undefined">
               {{ formatPower(homeLoad, true) }}
            </span>
            <span v-else>0.0 kW</span>
          </div>
        </div>
      </div>

    </div>
  </div>

  <!-- Chart Modal -->
  <div v-if="isModalOpen" class="fixed inset-0 z-50 flex items-center justify-center bg-black bg-opacity-50 transition-opacity">
    <div class="bg-white dark:bg-gray-800 rounded-lg shadow-xl w-11/12 max-w-4xl p-6 relative">
      <button @click="closeChart" class="absolute top-4 right-4 text-gray-500 hover:text-gray-700 dark:text-gray-400 dark:hover:text-gray-200">
        <svg class="h-6 w-6" fill="none" viewBox="0 0 24 24" stroke="currentColor">
          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12" />
        </svg>
      </button>

      <h2 class="text-xl font-semibold mb-4 text-gray-800 dark:text-gray-100 capitalize">{{ selectedNode?.replace('_', ' ') }} History</h2>

      <div class="flex space-x-2 mb-4">
        <button v-for="range in ranges" :key="range.value"
                @click="setRange(range.value)"
                :class="[
                  'px-3 py-1 rounded text-sm font-medium transition-colors',
                  selectedRange === range.value
                    ? 'bg-blue-600 text-white'
                    : 'bg-gray-200 text-gray-700 hover:bg-gray-300 dark:bg-gray-700 dark:text-gray-200 dark:hover:bg-gray-600'
                ]">
          {{ range.label }}
        </button>
      </div>

      <div class="h-[400px] w-full">
        <Line v-if="chartData" :data="chartData" :options="chartOptions" />
        <div v-else-if="isLoading" class="flex items-center justify-center h-full text-gray-500 dark:text-gray-400">Loading data...</div>
        <div v-else class="flex items-center justify-center h-full text-gray-500 dark:text-gray-400">No data available</div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'

import {
  Chart as ChartJS,
  CategoryScale,
  LinearScale,
  PointElement,
  LineElement,
  Title,
  Tooltip,
  Legend,
  TimeScale,
} from 'chart.js'
import { Line } from 'vue-chartjs'
import 'chartjs-adapter-date-fns'

ChartJS.register(
  CategoryScale,
  LinearScale,
  PointElement,
  LineElement,
  Title,
  Tooltip,
  Legend,
  TimeScale
)

interface SiteState {
  grid_power_w: number | null
  solar_power_w: number | null
  battery_power_w: number | null
  battery_soc: number | null
  total_load_w: number | null
  ev_charger_power_w: number | null
  device_health?: Record<number, string>
}

interface Device {
  id: number;
  name: string;
  template: string;
}

const props = defineProps<{
  state: SiteState | null
}>()

const devices = ref<Device[]>([])

const fetchDevices = async () => {
  try {
    const host = window.location.hostname
    const res = await fetch(`http://${host}:8080/api/devices`)
    if (res.ok) {
      devices.value = await res.json() || []
    }
  } catch (e) {
    console.error("Failed to fetch devices for power flow map:", e)
  }
}

onMounted(() => {
  fetchDevices()
})

const hasGrid = computed(() => devices.value.some(d => d.template === 'huawei_dongle' || d.template === 'demo_dongle'))
const hasSolar = computed(() => devices.value.some(d => d.template === 'huawei_inverter' || d.template === 'demo_inverter'))
const hasBattery = computed(() => devices.value.some(d => (d.template === 'huawei_inverter' || d.template === 'demo_inverter') && d.name.toLowerCase().includes('battery')))
const homeLoad = computed(() => {
  if (!props.state || props.state.total_load_w === null || props.state.total_load_w === undefined) return 0
  return props.state.total_load_w - (props.state.ev_charger_power_w || 0)
})

// Chart state
const isModalOpen = ref(false)
const selectedNode = ref<string | null>(null)
const selectedRange = ref<string>('today')
const isLoading = ref(false)
const chartData = ref<any>(null)

const ranges = [
  { label: 'Today', value: 'today' },
  { label: 'Last 24h', value: '24h' },
  { label: 'Last 7 Days', value: '7d' },
  { label: 'Last 30 Days', value: '30d' },
]

const getFlowStyle = (power: number | null | undefined, normalIsPositive: boolean = true, baseDuration: number = 2000) => {
  if (!power || Math.abs(power) < 10) return { display: 'none' }
  const absPower = Math.abs(power)
  const duration = Math.min(Math.max(baseDuration / absPower, 0.5), 5)

  let direction = 'normal'
  if (normalIsPositive && power < 0) direction = 'reverse'
  else if (!normalIsPositive && power >= 0) direction = 'reverse'

  return {
    animationDuration: `${duration}s`,
    animationDirection: direction
  }
}

const formatPower = (powerW: number, normalIsPositive: boolean = true, isHorizontal: boolean = false) => {
  const absPower = Math.abs(powerW)
  const valKw = (absPower / 1000).toFixed(1)

  // Arrow logic
  let arrow = ''
  if (absPower >= 10) {
    if (isHorizontal) {
       // e.g. Battery/Solar on sides
       if (normalIsPositive) {
         arrow = powerW > 0 ? '← ' : '→ '
       } else {
         arrow = powerW > 0 ? '→ ' : '← '
       }
    } else {
       // Vertical flows
       if (normalIsPositive) {
         arrow = powerW > 0 ? '↓ ' : '↑ '
       } else {
         arrow = powerW > 0 ? '↑ ' : '↓ '
       }
    }
  }

  return `${arrow}${valKw} kW`
}

const chartOptions = computed(() => ({
  responsive: true,
  maintainAspectRatio: false,
  scales: {
    x: {
      type: 'time' as const,
      time: {
        unit: selectedRange.value === 'today' || selectedRange.value === '24h' ? 'hour' as const : 'day' as const,
      },
      ticks: {
        color: '#9CA3AF',
      },
      grid: {
        color: '#374151',
      }
    },
    y: {
      title: {
        display: true,
        text: 'Power (W)',
        color: '#9CA3AF'
      },
      ticks: {
        color: '#9CA3AF',
      },
      grid: {
        color: '#374151',
      }
    }
  },
  plugins: {
    legend: {
      display: false
    },
    tooltip: {
      callbacks: {
        label: (context: any) => `${context.parsed.y.toFixed(0)} W`
      }
    }
  }
}))

const openChart = (node: string) => {
  selectedNode.value = node
  isModalOpen.value = true
  fetchHistory()
}

const closeChart = () => {
  isModalOpen.value = false
  selectedNode.value = null
  chartData.value = null
}

const setRange = (range: string) => {
  selectedRange.value = range
  fetchHistory()
}

const fetchHistory = async () => {
  if (!selectedNode.value) return
  isLoading.value = true
  chartData.value = null
  try {
    const host = window.location.hostname
    const res = await fetch(`http://${host}:8080/api/history?node=${selectedNode.value}&range=${selectedRange.value}`)
    if (res.ok) {
      const data: {timestamp: string, power_w: number}[] = await res.json()

      let lineColor = '#3B82F6' // Grid - Blue
      if (selectedNode.value === 'solar') lineColor = '#FBBF24' // Yellow
      else if (selectedNode.value === 'battery') lineColor = '#34D399' // Green
      else if (selectedNode.value === 'ev_charger') lineColor = '#A855F7' // Purple

      if (data && data.length > 0) {
        chartData.value = {
          datasets: [
            {
              label: selectedNode.value,
              data: data.map(d => ({ x: new Date(d.timestamp), y: d.power_w })),
              borderColor: lineColor,
              backgroundColor: lineColor,
              borderWidth: 2,
              pointRadius: 1,
              pointHoverRadius: 4,
              fill: false,
              tension: 0.2
            }
          ]
        }
      } else {
        chartData.value = null // No data
      }
    }
  } catch (e) {
    console.error("Failed to fetch history:", e)
  } finally {
    isLoading.value = false
  }
}

</script>

<style scoped>
.flow-path {
  animation: flow linear infinite;
}

@keyframes flow {
  from {
    stroke-dashoffset: 100;
  }
  to {
    stroke-dashoffset: 0;
  }
}
</style>
