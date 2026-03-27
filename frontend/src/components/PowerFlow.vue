<template>
  <div class="relative w-full h-[400px] bg-transparent rounded-lg overflow-visible">
    <!-- SVG paths for animated power flow -->
    <svg viewBox="0 0 100 100" preserveAspectRatio="none" class="absolute inset-0 w-full h-full z-0 overflow-visible">

      <!-- Central Junction Point -->
      <circle v-if="hasGrid || hasSolar || hasBattery" cx="50" cy="50" r="1.5" class="fill-[#3B82F6]" />

      <!-- GRID PATHS -->
      <!-- Grid to Central Junction (Blue) -->
      <path
        v-if="hasGrid"
        id="path-grid-junction"
        :d="hasBattery ? 'M 20 50 L 50 50' : 'M 25 60 C 35 60, 40 40, 50 40'"
        class="stroke-[#3B82F6] stroke-[0.5] fill-none"
        vector-effect="non-scaling-stroke"
      />

      <!-- Grid to Solar (Purple curve) -->
      <path
        v-if="hasGrid && hasSolar"
        id="path-grid-solar"
        :d="hasBattery ? 'M 20 45 C 35 45, 45 40, 45 20' : 'M 25 55 C 35 55, 45 35, 45 25'"
        class="stroke-[#8B5CF6] stroke-[0.5] fill-none"
        vector-effect="non-scaling-stroke"
      />

      <!-- Grid to Battery (Purple curve) -->
      <path
        v-if="hasGrid && hasBattery"
        id="path-grid-battery"
        d="M 20 55 C 35 55, 45 60, 45 80"
        class="stroke-[#8B5CF6] stroke-[0.5] fill-none"
        vector-effect="non-scaling-stroke"
      />

      <!-- SOLAR PATHS -->
      <!-- Solar to Central Junction (Yellow curve) -->
      <path
        v-if="hasSolar"
        id="path-solar-junction"
        :d="hasBattery ? 'M 55 20 C 55 40, 65 45, 80 45' : 'M 55 25 C 55 45, 65 55, 75 55'"
        class="stroke-[#F59E0B] stroke-[0.5] fill-none"
        vector-effect="non-scaling-stroke"
      />

      <!-- BATTERY PATHS -->
      <!-- Central Junction to Battery (Teal curve) -->
      <path
        v-if="hasBattery"
        id="path-junction-battery"
        d="M 50 50 C 55 50, 55 60, 55 80"
        class="stroke-[#14B8A6] stroke-[0.5] fill-none"
        vector-effect="non-scaling-stroke"
      />

      <!-- Central Junction to Home (Teal line) -->
      <path
        v-if="hasBattery || hasSolar || hasGrid"
        id="path-junction-home"
        :d="hasBattery ? 'M 50 50 L 80 50' : 'M 50 40 C 60 40, 65 60, 75 60'"
        class="stroke-[#14B8A6] stroke-[0.5] fill-none"
        vector-effect="non-scaling-stroke"
      />

      <!-- ANIMATED DOTS using dashed strokes to avoid SVG preserveAspectRatio stretching -->
      <!-- We overlap exactly matching paths but style them as dashed flowing lines -->

      <!-- Grid <-> Junction (Purple dots) -->
      <path
        v-if="hasGrid && Math.abs(state?.grid_power_w || 0) > 10"
        :d="hasBattery ? 'M 20 50 C 40 50, 45 50, 50 50' : 'M 25 60 C 40 60, 45 60, 50 60'"
        class="stroke-[#8B5CF6] stroke-[4px] fill-none flow-path"
        style="stroke-linecap: round; stroke-dasharray: 0 40;"
        vector-effect="non-scaling-stroke"
        :style="getStrokeAnimStyle(state?.grid_power_w, true)"
      />

      <!-- Solar -> Junction (Large Orange dots) -->
      <path
        v-if="hasSolar && (state?.solar_power_w || 0) > 10"
        :d="hasBattery ? 'M 50 20 C 50 35, 50 45, 50 50' : 'M 50 25 C 50 40, 50 50, 50 60'"
        class="stroke-[#F59E0B] stroke-[6px] fill-none flow-path"
        style="stroke-linecap: round; stroke-dasharray: 0 50;"
        vector-effect="non-scaling-stroke"
        :style="getStrokeAnimStyle(state?.solar_power_w, true)"
      />

      <!-- Battery <-> Junction (Blue dots) -->
      <path
        v-if="hasBattery && Math.abs(state?.battery_power_w || 0) > 10"
        d="M 50 80 C 50 65, 50 55, 50 50"
        class="stroke-[#3B82F6] stroke-[4px] fill-none flow-path"
        style="stroke-linecap: round; stroke-dasharray: 0 40;"
        vector-effect="non-scaling-stroke"
        :style="getStrokeAnimStyle(state?.battery_power_w, true)"
      />

      <!-- Junction -> Home (Small Purple dots) -->
      <path
        v-if="(state?.total_load_w || 0) > 10"
        :d="hasBattery ? 'M 50 50 C 65 50, 75 50, 80 50' : 'M 50 60 C 60 60, 65 60, 75 60'"
        class="stroke-[#8B5CF6] stroke-[3px] fill-none flow-path"
        style="stroke-linecap: round; stroke-dasharray: 0 30;"
        vector-effect="non-scaling-stroke"
        :style="getStrokeAnimStyle(state?.total_load_w, true)"
      />

    </svg>

    <!-- Node UI Elements -->

    <!-- Grid Node -->
    <div v-if="hasGrid" @click="openChart('grid')" :style="{ top: nodePositions.grid.top, left: nodePositions.grid.left }" class="absolute -translate-x-1/2 -translate-y-1/2 z-10 flex flex-col items-center w-28 h-28 bg-white dark:bg-gray-800 rounded-full border-[3px] border-[#3B82F6] shadow-sm cursor-pointer hover:scale-105 transition-transform">
      <div class="absolute -bottom-6 text-sm text-gray-500 dark:text-gray-400">Grid</div>
      <div class="mt-4 flex items-center justify-center">
        <svg class="h-8 w-8 text-black dark:text-white" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
          <path d="M4 19L8 5H16L20 19" />
          <path d="M4 15H20" />
          <path d="M6 10H18" />
          <path d="M12 5V19" />
        </svg>
      </div>
      <div class="mt-1 flex flex-col items-center justify-center leading-tight">
        <div v-if="state?.grid_power_w !== null && state?.grid_power_w !== undefined" class="text-sm font-medium" :class="state.grid_power_w > 0 ? 'text-[#8B5CF6]' : 'text-[#3B82F6]'">
          {{ state.grid_power_w > 0 ? '<--' : '-->' }} {{ formatPowerGrid(state.grid_power_w) }}
        </div>
        <div v-else class="text-sm font-medium text-gray-500">--</div>
      </div>
    </div>

    <!-- Solar Node -->
    <div v-if="hasSolar" @click="openChart('solar')" :style="{ top: nodePositions.solar.top, left: nodePositions.solar.left }" class="absolute -translate-x-1/2 -translate-y-1/2 z-10 flex flex-col items-center w-28 h-28 bg-white dark:bg-gray-800 rounded-full border-[3px] border-[#F59E0B] shadow-sm cursor-pointer hover:scale-105 transition-transform">
      <div class="absolute -top-6 text-sm text-gray-500 dark:text-gray-400">Solar</div>
      <div class="mt-4 flex items-center justify-center relative">
        <svg class="h-8 w-8 text-black dark:text-white" viewBox="0 0 24 24" fill="currentColor">
          <path d="M12 2L14.4 7.6L20 10L14.4 12.4L12 18L9.6 12.4L4 10L9.6 7.6L12 2Z" />
        </svg>
        <svg class="h-4 w-4 text-black dark:text-white absolute -bottom-1 -right-1" viewBox="0 0 24 24" fill="currentColor">
          <path d="M13 2L3 14H12L11 22L21 10H12L13 2Z" />
        </svg>
      </div>
      <div class="mt-1 text-sm font-medium text-black dark:text-white">
        {{ state?.solar_power_w !== null && state?.solar_power_w !== undefined ? formatPower(state.solar_power_w) : '--' }}
      </div>
    </div>

    <!-- Battery Node -->
    <div v-if="hasBattery" @click="openChart('battery')" :style="{ top: nodePositions.battery.top, left: nodePositions.battery.left }" class="absolute -translate-x-1/2 -translate-y-1/2 z-10 flex flex-col items-center w-28 h-28 bg-white dark:bg-gray-800 rounded-full border-[3px] border-[#EC4899] shadow-sm cursor-pointer hover:scale-105 transition-transform">
      <div class="absolute -bottom-6 text-sm text-gray-500 dark:text-gray-400">Battery</div>
      <div class="mt-4 flex items-center justify-center">
        <svg class="h-8 w-8 text-black dark:text-white" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
          <rect x="6" y="4" width="12" height="16" rx="2" ry="2" />
          <path d="M9 2H15" />
          <path d="M6 9H18" />
          <path d="M6 14H18" />
          <path d="M6 19H18" />
        </svg>
      </div>
      <div class="mt-1 flex flex-col items-center justify-center leading-tight">
        <div v-if="state?.battery_power_w !== null && state?.battery_power_w !== undefined" class="text-sm font-medium" :class="state.battery_power_w < 0 ? 'text-[#EC4899]' : 'text-[#14B8A6]'">
          {{ state.battery_power_w < 0 ? 'v' : '^' }} {{ formatPowerBattery(state.battery_power_w) }}
        </div>
        <div v-else class="text-sm font-medium text-gray-500">--</div>
      </div>
    </div>

    <!-- Home Load Node -->
    <div :style="{ top: nodePositions.home.top, left: nodePositions.home.left }" class="absolute -translate-x-1/2 -translate-y-1/2 z-10 flex flex-col items-center justify-center w-28 h-28 bg-white dark:bg-gray-800 rounded-full border-[3px] border-[#14B8A6] shadow-sm relative overflow-hidden">
      <!-- Colored arc for Home (Teal and Orange as in image) -->
      <svg class="absolute inset-0 w-full h-full -rotate-90" viewBox="0 0 100 100">
        <circle cx="50" cy="50" r="48" fill="none" stroke="#F59E0B" stroke-width="6" stroke-dasharray="30 270" stroke-dashoffset="0" />
      </svg>
      <div class="absolute -bottom-6 text-sm text-gray-500 dark:text-gray-400">Home</div>
      <svg class="h-8 w-8 text-black dark:text-white mb-1 z-10" viewBox="0 0 24 24" fill="currentColor">
        <path d="M12 3L20 10V21H4V10L12 3Z" />
      </svg>
      <div class="text-sm font-medium text-black dark:text-white z-10">{{ state?.total_load_w !== null && state?.total_load_w !== undefined ? formatPower(Math.max(0, homeLoad)) : '--' }}</div>
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

const getStrokeAnimStyle = (power: number | null | undefined, normalIsPositive: boolean = true, baseDuration: number = 2000) => {
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

const formatPower = (powerW: number) => {
  if (Math.abs(powerW) >= 1000) {
    // format as kWh and replace dot with comma for European format, remove trailing zero
    const val = (Math.abs(powerW) / 1000).toFixed(1)
    return `${val.replace('.', ',')} kWh`
  }
  return `${Math.abs(powerW).toFixed(0)} Wh`
}

const formatPowerGrid = (powerW: number) => {
  return formatPower(powerW)
}

const formatPowerBattery = (powerW: number) => {
  return formatPower(powerW)
}

const nodePositions = computed(() => {
  if (hasBattery.value) {
    return {
      grid: { top: '50%', left: '20%' },
      solar: { top: '20%', left: '50%' },
      battery: { top: '80%', left: '50%' },
      home: { top: '50%', left: '80%' }
    }
  } else {
    // Re-center without battery
    return {
      grid: { top: '60%', left: '25%' },
      solar: { top: '25%', left: '50%' },
      battery: { top: '100%', left: '50%' }, // off-screen or unused
      home: { top: '60%', left: '75%' }
    }
  }
})

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
