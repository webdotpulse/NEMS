<template>
  <div class="w-full flex justify-center items-center py-8">
    <div class="relative w-[340px] h-[460px]">

      <!-- SVG paths for animated power flow lines -->
      <svg class="absolute inset-0 w-full h-full z-0 pointer-events-none" viewBox="0 0 340 460" preserveAspectRatio="none">

        <!-- Static Background Lines -->
        <path v-if="hasSolar" d="M 170 120 L 170 230" stroke-linecap="round" fill="none" stroke="#E5E7EB" stroke-width="6" />
        <path v-if="hasBattery" d="M 170 230 L 170 330" stroke-linecap="round" fill="none" stroke="#E5E7EB" stroke-width="6" />
        <path v-if="hasGrid" d="M 100 230 L 170 230" stroke-linecap="round" fill="none" stroke="#E5E7EB" stroke-width="6" />
        <path d="M 170 230 L 240 230" stroke-linecap="round" fill="none" stroke="#E5E7EB" stroke-width="6" />
        <path v-if="hasEV" d="M 100 120 Q 100 230 170 230" stroke-linecap="round" fill="none" stroke="#E5E7EB" stroke-width="6" />

        <!-- Flow lines -->
        <template v-for="segment in activeSegments" :key="segment.id">
          <!-- Outline/glow -->
          <path :d="segment.path" stroke-linecap="round"
                fill="none" :stroke="segment.color" stroke-width="8" stroke-opacity="0.2" class="flow-glow" />
          <!-- Animated flow line -->
          <path :d="segment.path" stroke-linecap="round"
                fill="none" :stroke="segment.color" stroke-width="4" stroke-dasharray="8 8" class="flow-path"
                :style="getFlowStyle(segment.power, segment.normalIsPositive)" />
        </template>

      </svg>

      <!-- Grid Layout for Nodes -->
      <div class="absolute inset-0 grid grid-cols-3 grid-rows-3 gap-0">

        <!-- Top Left: EV Charger -->
        <div class="col-start-1 row-start-1 flex flex-col items-center justify-end pb-2 pr-4">
          <div v-if="hasEV" @click="openChart('ev_charger')" class="z-10 flex flex-col items-center justify-center w-28 h-28 bg-white dark:bg-gray-800 rounded-full border-[4px] border-[#A855F7] shadow-sm cursor-pointer hover:scale-105 transition-transform mb-2 relative">
            <span class="text-xs font-medium text-gray-500 mb-1 absolute -top-6">EV</span>
            <svg class="h-8 w-8 text-gray-700 dark:text-gray-300 mb-1" viewBox="0 0 24 24" fill="currentColor">
              <path d="M19.77 7.23l.01-.01-3.72-3.72L15 4.56l2.11 2.11C16.17 7 15.5 7.93 15.5 9v11h-2V9c0-2.21 1.79-4 4-4h1V3c0-1.1-.9-2-2-2H5c-1.1 0-2 .9-2 2v18h12V3h-1.5v6H5V3h8v15c0 1.66 1.34 3 3 3h3.5c1.38 0 2.5-1.12 2.5-2.5v-8.38c0-.7-.3-1.39-.73-1.89zM11 11H7v-2h4v2z"/>
            </svg>
            <div class="text-gray-800 dark:text-gray-200 text-sm font-medium">
              <span v-if="state?.ev_charger_power_w !== null && state?.ev_charger_power_w !== undefined">
                 {{ formatPowerSimple(state.ev_charger_power_w) }}
              </span>
              <span v-else>0.0 kW</span>
            </div>
          </div>
        </div>

        <!-- Top Row: Solar -->
        <div class="col-start-2 row-start-1 flex flex-col items-center justify-end pb-2">
          <div v-if="hasSolar" @click="openChart('solar')" class="z-10 flex flex-col items-center justify-center w-28 h-28 bg-white dark:bg-gray-800 rounded-full border-[4px] border-[#FBBF24] shadow-sm cursor-pointer hover:scale-105 transition-transform mb-2">
            <span class="text-xs font-medium text-gray-500 mb-1 absolute -top-6">Solar</span>
            <svg class="h-8 w-8 text-gray-700 dark:text-gray-300 mb-1 relative" viewBox="0 0 24 24" fill="currentColor">
              <path d="M3 4h18v16H3V4zM5 6v12h14V6H5z"/>
              <path d="M11 6v12h2V6h-2zM7 6v12h2V6H7zM15 6v12h2V6h-2z"/>
              <path d="M5 11h14v2H5v-2zM5 15h14v2H5v-2z"/>
              <circle cx="21" cy="3" r="1.5" />
            </svg>
            <div class="text-gray-800 dark:text-gray-200 text-sm font-medium">
              <span v-if="state?.solar_power_w !== null && state?.solar_power_w !== undefined">
                 {{ formatPowerSimple(state.solar_power_w) }}
              </span>
              <span v-else>0.0 kW</span>
            </div>
          </div>
        </div>

        <!-- Middle Row: Grid (Left), Junction (Center), Home (Right) -->
        <div class="col-start-1 row-start-2 flex flex-col items-center justify-center">
          <div v-if="hasGrid" @click="openChart('grid')" class="z-10 flex flex-col items-center justify-center w-28 h-28 bg-white dark:bg-gray-800 rounded-full border-[4px] border-[#3B82F6] shadow-sm cursor-pointer hover:scale-105 transition-transform mb-2 relative">
            <span class="text-xs font-medium text-gray-500 mb-1 absolute -bottom-6">Grid</span>
            <svg class="h-8 w-8 text-gray-700 dark:text-gray-300 mb-1" viewBox="0 0 24 24" fill="currentColor">
              <path d="M16,7V3h-2v4h-4V3H8v4H6v7.5L9.5,18v3h5v-3l3.5-3.5V7H16z M14,15.5l-2,2l-2-2V13h4V15.5z M14,11h-4V9h4V11z"/>
            </svg>
            <div class="text-sm font-medium flex flex-col items-center leading-tight">
              <span v-if="state?.grid_power_w !== null && state?.grid_power_w !== undefined" class="text-purple-600 dark:text-purple-400 text-xs">
                &larr; {{ state.grid_power_w < 0 ? (Math.abs(state.grid_power_w) / 1000).toFixed(1) : '0.0' }} kW
              </span>
              <span v-else class="text-purple-600 dark:text-purple-400 text-xs">&larr; 0.0 kW</span>

              <span v-if="state?.grid_power_w !== null && state?.grid_power_w !== undefined" class="text-blue-500 dark:text-blue-400 text-xs">
                &rarr; {{ state.grid_power_w > 0 ? (state.grid_power_w / 1000).toFixed(1) : '0.0' }} kW
              </span>
              <span v-else class="text-blue-500 dark:text-blue-400 text-xs">&rarr; 0.0 kW</span>
            </div>
          </div>
        </div>

        <div class="col-start-2 row-start-2 flex items-center justify-center">
          <!-- Junction Node -->
          <div class="z-10 flex items-center justify-center w-[2rem] h-[2rem] bg-[#3B82F6] rounded-full shadow-sm">
            <!-- Small dot for junction like in image -->
          </div>
        </div>

        <div class="col-start-3 row-start-2 flex flex-col items-center justify-center">
          <div @click="openChart('home')" class="z-10 flex flex-col items-center justify-center w-28 h-28 bg-white dark:bg-gray-800 rounded-full border-[4px] border-[#10B981] shadow-sm cursor-pointer hover:scale-105 transition-transform mb-2 relative">
             <span class="text-xs font-medium text-gray-500 mb-1 absolute -bottom-6">Home</span>
            <svg class="h-8 w-8 text-gray-700 dark:text-gray-300 mb-1" viewBox="0 0 24 24" fill="currentColor">
              <path d="M12 3L4 9v12h16V9l-8-6zm6 16h-3v-4H9v4H6v-9l6-4.5 6 4.5v9z"/>
            </svg>
            <div class="text-gray-800 dark:text-gray-200 text-sm font-medium">
              <span v-if="state?.total_load_w !== null && state?.total_load_w !== undefined">
                 {{ formatPowerSimple(homeLoad) }}
              </span>
              <span v-else>0.0 kW</span>
            </div>
          </div>
        </div>

        <!-- Bottom Row: Battery -->
        <div class="col-start-2 row-start-3 flex flex-col items-center justify-start pt-2">
          <div v-if="hasBattery" @click="openChart('battery')" class="z-10 flex flex-col items-center justify-center w-28 h-28 bg-white dark:bg-gray-800 rounded-full border-[4px] border-[#EC4899] shadow-sm cursor-pointer hover:scale-105 transition-transform mb-2 relative">
            <span class="text-xs font-medium text-gray-500 mb-1 absolute -bottom-6">Battery</span>
            <svg class="h-8 w-8 text-gray-700 dark:text-gray-300 mb-1" viewBox="0 0 24 24" fill="currentColor">
              <path d="M16,4h-2V2h-4v2H8C6.9,4,6,4.9,6,6v14c0,1.1,0.9,2,2,2h8c1.1,0,2-0.9,2-2V6C18,4.9,17.1,4,16,4z M16,20H8V6h8V20z M10,8h4v2h-4V8z M10,12h4v2h-4V12z M10,16h4v2h-4V16z"/>
            </svg>
            <div class="text-sm font-medium flex flex-col items-center leading-tight">
              <span v-if="state?.battery_power_w !== null && state?.battery_power_w !== undefined" class="text-pink-500 dark:text-pink-400 text-xs">
                &darr; {{ state.battery_power_w < 0 ? (Math.abs(state.battery_power_w) / 1000).toFixed(1) : '0.0' }} kW
              </span>
              <span v-else class="text-pink-500 dark:text-pink-400 text-xs">&darr; 0.0 kW</span>

              <span v-if="state?.battery_power_w !== null && state?.battery_power_w !== undefined" class="text-teal-500 dark:text-teal-400 text-xs">
                &uarr; {{ state.battery_power_w > 0 ? (state.battery_power_w / 1000).toFixed(1) : '0.0' }} kW
              </span>
              <span v-else class="text-teal-500 dark:text-teal-400 text-xs">&uarr; 0.0 kW</span>
            </div>
            <!-- Battery SOC Circle Overlay -->
            <div v-if="state?.battery_soc !== null && state?.battery_soc !== undefined" class="absolute inset-0 rounded-full border-[4px] border-[#EC4899] opacity-50" :style="`clip-path: polygon(0 ${100 - state.battery_soc}%, 100% ${100 - state.battery_soc}%, 100% 100%, 0 100%); border-color: #34D399; z-index: 20;`"></div>
          </div>
        </div>
      </div>

    </div>
  </div>

  <!-- Context-Sensitive Full-screen Panel -->
  <div v-if="isModalOpen" class="fixed inset-0 z-50 flex justify-center items-center bg-black bg-opacity-50 transition-opacity" @click.self="closeChart">
    <div class="w-full h-full bg-white dark:bg-gray-800 shadow-xl p-8 relative flex flex-col transform transition-transform duration-300 translate-x-0 overflow-hidden" @click.stop>
      <button @click="closeChart" class="absolute top-6 right-6 text-gray-500 hover:text-gray-700 dark:text-gray-400 dark:hover:text-gray-200 z-50">
        <svg class="h-8 w-8" fill="none" viewBox="0 0 24 24" stroke="currentColor">
          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12" />
        </svg>
      </button>

      <h2 class="text-3xl font-semibold mb-6 text-gray-800 dark:text-gray-100 capitalize">{{ selectedNode?.replace('_', ' ') }} History</h2>

      <div class="flex flex-wrap gap-4 mb-8">
        <button v-for="range in ranges" :key="range.value"
                @click="setRange(range.value)"
                :class="[
                  'px-4 py-2 rounded-lg text-lg font-medium transition-colors',
                  selectedRange === range.value
                    ? 'bg-blue-600 text-white'
                    : 'bg-gray-200 text-gray-700 hover:bg-gray-300 dark:bg-gray-700 dark:text-gray-200 dark:hover:bg-gray-600'
                ]">
          {{ range.label }}
        </button>
      </div>

      <div class="flex-grow w-full h-full min-h-[500px]">
        <Line v-if="chartData" :data="chartData" :options="chartOptions" />
        <div v-else-if="isLoading" class="flex items-center justify-center h-full text-xl text-gray-500 dark:text-gray-400">Loading data...</div>
        <div v-else class="flex items-center justify-center h-full text-xl text-gray-500 dark:text-gray-400">No data available</div>
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
const hasEV = computed(() => devices.value.some(d => ['raedian_charger', 'demo_charger', 'alfen_charger', 'easee_charger', 'bender_charger', 'peblar_charger', 'phoenix_charger'].includes(d.template)))
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

// Use a dynamic maximum to calculate a relative flow percentage
const flowPercentage = computed(() => {
  if (!props.state) return 0;

  // Find the maximum absolute power flowing through any node right now
  const grid = Math.abs(props.state.grid_power_w || 0);
  const solar = Math.abs(props.state.solar_power_w || 0);
  const battery = Math.abs(props.state.battery_power_w || 0);
  const home = Math.abs(homeLoad.value || 0);

  const maxVal = Math.max(grid, solar, battery, home, 100); // Floor at 100W to avoid 0 div
  return maxVal;
});

const getFlowStyle = (power: number | null | undefined, normalIsPositive: boolean = true) => {
  if (!power || Math.abs(power) < 10) return { display: 'none' }
  const absPower = Math.abs(power)

  // Faster animation for higher percentage of max power
  // Slower animation (e.g., 5s) for small flows. Base 2s. Max duration 5s, Min 0.5s.
  const percentage = absPower / (flowPercentage.value || 100);
  const duration = Math.min(Math.max(5 - (percentage * 4.5), 0.5), 5); // 0% = 5s, 100% = 0.5s


  let direction = 'normal'
  if (normalIsPositive && power < 0) direction = 'reverse'
  else if (!normalIsPositive && power >= 0) direction = 'reverse'

  return {
    animationDuration: `${duration}s`,
    animationDirection: direction
  }
}

interface Segment {
  id: string;
  path: string;
  power: number;
  color: string;
  normalIsPositive: boolean; // Determines direction mapping based on sign
}

const activeSegments = computed<Segment[]>(() => {
  if (!props.state) return [];

  const grid = props.state.grid_power_w || 0;
  const solar = props.state.solar_power_w || 0;
  const battery = props.state.battery_power_w || 0;
  const home = homeLoad.value || 0;

  const segments: Segment[] = [];

  // T-Junction logic: Every segment connects a node to the exact center Site Junction (170, 230)
  // For each node, if power is flowing (abs > 10W), draw the segment to the junction with its native color

  // 1. Solar (Top) -> Junction
  if (hasSolar.value && Math.abs(solar) >= 10) {
    segments.push({
      id: 'solar-segment',
      path: 'M 170 120 L 170 230',
      power: solar,
      color: '#FBBF24',
      normalIsPositive: true
    });
  }

  // 2. Grid (Left) <-> Junction
  if (hasGrid.value && Math.abs(grid) >= 10) {
    segments.push({
      id: 'grid-segment',
      path: 'M 100 230 L 170 230',
      power: grid,
      color: '#3B82F6',
      normalIsPositive: true
    });
  }

  // 3. Home (Right) <- Junction
  if (Math.abs(home) >= 10) {
    segments.push({
      id: 'home-segment',
      path: 'M 170 230 L 240 230',
      power: home,
      color: '#10B981',
      normalIsPositive: true
    });
  }

  // 4. Battery (Bottom) <-> Junction
  if (hasBattery.value && Math.abs(battery) >= 10) {
    segments.push({
      id: 'battery-segment',
      path: 'M 170 330 L 170 230',
      power: battery,
      color: '#EC4899',
      normalIsPositive: true
    });
  }

  // 5. EV (Top-Left) <- Junction
  const evPower = props.state.ev_charger_power_w || 0;
  if (hasEV.value && Math.abs(evPower) >= 10) {
    segments.push({
      id: 'ev-segment',
      path: 'M 100 120 Q 100 230 170 230',
      power: evPower,
      color: '#A855F7',
      normalIsPositive: true
    });
  }

  return segments;
});


const formatPowerSimple = (powerW: number) => {
  const absPower = Math.abs(powerW)
  const valKw = (absPower / 1000).toFixed(1)
  return `${valKw} kW`
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
      else if (selectedNode.value === 'home') lineColor = '#A855F7' // Purple

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
