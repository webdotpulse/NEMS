<template>
  <div class="w-full flex justify-center items-center py-8">
    <div class="relative w-full h-[460px] bg-white dark:bg-gray-800 rounded-lg shadow overflow-hidden">

      <!-- SVG paths for animated power flow lines -->
      <svg class="absolute inset-0 w-full h-full z-0 pointer-events-none" viewBox="0 0 100 100" preserveAspectRatio="none">


        <!-- Static Background Lines -->
        <path v-for="(device, index) in solarDevices" :key="'static-solar-'+device.id" :d="`M ${50 + (index * 15 - (solarDevices.length-1)*7.5)} ${20 - (index % 2 === 0 ? 0 : 5)} L 50 50`" vector-effect="non-scaling-stroke" stroke-linecap="round" fill="none" stroke="#E5E7EB" stroke-width="6" />
        <path v-for="(device, index) in batteryDevices" :key="'static-battery-'+device.id" :d="`M ${50 + (index * 15 - (batteryDevices.length-1)*7.5)} ${80 + (index % 2 === 0 ? 0 : 5)} L 50 50`" vector-effect="non-scaling-stroke" stroke-linecap="round" fill="none" stroke="#E5E7EB" stroke-width="6" />
        <path v-for="(device, index) in gridDevices" :key="'static-grid-'+device.id" :d="`M ${20 - (index * 5)} ${50 + (index * 10 - (gridDevices.length-1)*5)} L 50 50`" vector-effect="non-scaling-stroke" stroke-linecap="round" fill="none" stroke="#E5E7EB" stroke-width="6" />
        <path d="M 80 50 L 50 50" vector-effect="non-scaling-stroke" stroke-linecap="round" fill="none" stroke="#E5E7EB" stroke-width="6" />
        <path v-for="(device, index) in evDevices" :key="'static-ev-'+device.id" :d="`M ${20 + (index * 12)} ${20 + (index * 8)} Q 20 50 50 50`" vector-effect="non-scaling-stroke" stroke-linecap="round" fill="none" stroke="#E5E7EB" stroke-width="6" />
<!-- Flow lines -->
        <template v-for="segment in activeSegments" :key="segment.id">
          <!-- Outline/glow -->
          <path :d="segment.path" stroke-linecap="round"
                fill="none" :stroke="segment.color" stroke-width="8" stroke-opacity="0.2" class="flow-glow" vector-effect="non-scaling-stroke" />
          <!-- Animated flow line -->
          <path :d="segment.path" stroke-linecap="round"
                fill="none" :stroke="segment.color" stroke-width="4" stroke-dasharray="8 8" class="flow-path"
                vector-effect="non-scaling-stroke"
                :style="getFlowStyle(segment.power, segment.normalIsPositive)" />
        </template>

      </svg>

      <!-- Grid Layout for Nodes -->
      <div class="absolute inset-0">


        <!-- Top Left: EV Charger -->
        <template v-for="(device, index) in evDevices" :key="device.id">
          <div class="absolute flex flex-col items-center justify-center transform -translate-x-1/2 -translate-y-1/2" :style="`left: ${20 + (index * 12)}%; top: ${20 + (index * 8)}%;`">
            <div @click="openChart('ev_charger')" class="z-10 flex flex-col items-center justify-center w-28 h-28 bg-white dark:bg-gray-800 rounded-full border-[4px] border-[#A855F7] shadow-sm cursor-pointer hover:scale-105 transition-transform mb-2 relative">
              <span class="text-xs font-medium text-gray-500 mb-1 absolute -top-6">{{ device.name }}</span>
              <svg xmlns="http://www.w3.org/2000/svg" class="h-8 w-8 text-gray-700 dark:text-gray-300 mb-1" fill="currentColor" viewBox="0 0 24 24"><path d="m20.772 10.156-1.368-4.105A2.995 2.995 0 0 0 16.559 4H7.441a2.995 2.995 0 0 0-2.845 2.051l-1.368 4.105A2.003 2.003 0 0 0 2 12v5c0 .753.423 1.402 1.039 1.743-.013.066-.039.126-.039.195V21a1 1 0 0 0 1 1h1a1 1 0 0 0 1-1v-2h12v2a1 1 0 0 0 1 1h1a1 1 0 0 0 1-1v-2.062c0-.069-.026-.13-.039-.195A1.993 1.993 0 0 0 22 17v-5c0-.829-.508-1.541-1.228-1.844zM4 17v-5h16l.002 5H4zM7.441 6h9.117c.431 0 .813.274.949.684L18.613 10H5.387l1.105-3.316A1 1 0 0 1 7.441 6z"/><circle cx="6.5" cy="14.5" r="1.5"/><circle cx="17.5" cy="14.5" r="1.5"/></svg>
              <div class="text-gray-800 dark:text-gray-200 text-sm font-medium">
                <span v-if="state?.ev_charger_power_w !== null && state?.ev_charger_power_w !== undefined">
                  {{ formatPowerSimple(state.ev_charger_power_w / evDevices.length) }}
                </span>
                <span v-else>0.0 kW</span>
              </div>
            </div>
          </div>
        </template>
<!-- Top Row: Solar -->
        <div class="absolute flex flex-col items-center justify-center transform -translate-x-1/2 -translate-y-1/2" style="left: 50%; top: 20%;">
          <div v-if="solarDevices.length > 0" @click="openChart('solar')" class="z-10 flex flex-col items-center justify-center w-28 h-28 bg-white dark:bg-gray-800 rounded-full border-[4px] border-[#FBBF24] shadow-sm cursor-pointer hover:scale-105 transition-transform mb-2">
            <span class="text-xs font-medium text-gray-500 mb-1 absolute -top-6">Solar</span>
            <svg xmlns="http://www.w3.org/2000/svg" class="h-8 w-8 text-gray-700 dark:text-gray-300 mb-1 relative" fill="currentColor" viewBox="0 0 24 24"><path d="M6.993 12c0 2.761 2.246 5.007 5.007 5.007s5.007-2.246 5.007-5.007S14.761 6.993 12 6.993 6.993 9.239 6.993 12zM12 8.993c1.658 0 3.007 1.349 3.007 3.007S13.658 15.007 12 15.007 8.993 13.658 8.993 12 10.342 8.993 12 8.993zM10.998 19h2v3h-2zm0-17h2v3h-2zm-9 9h3v2h-3zm17 0h3v2h-3zM4.219 18.363l2.12-2.122 1.415 1.414-2.12 2.122zM16.24 6.344l2.122-2.122 1.414 1.414-2.122 2.122zM6.342 7.759 4.22 5.637l1.415-1.414 2.12 2.122zm13.434 10.605-1.414 1.414-2.122-2.122 1.414-1.414z"/></svg>
            <div class="text-gray-800 dark:text-gray-200 text-sm font-medium">
              <span v-if="state?.solar_power_w !== null && state?.solar_power_w !== undefined">
                 {{ formatPowerSimple(state.solar_power_w) }}
              </span>
              <span v-else>0.0 kW</span>
            </div>
          </div>
        </div>

        <!-- Middle Row: Grid (Left), Junction (Center), Home (Right) -->
        <div class="absolute flex flex-col items-center justify-center transform -translate-x-1/2 -translate-y-1/2" style="left: 20%; top: 50%;">
          <div v-if="gridDevices.length > 0" @click="openChart('grid')" class="z-10 flex flex-col items-center justify-center w-28 h-28 bg-white dark:bg-gray-800 rounded-full border-[4px] border-[#3B82F6] shadow-sm cursor-pointer hover:scale-105 transition-transform mb-2 relative">
            <span class="text-xs font-medium text-gray-500 mb-1 absolute -bottom-6">Grid</span>
            <svg xmlns="http://www.w3.org/2000/svg" class="h-8 w-8 text-gray-700 dark:text-gray-300 mb-1" fill="currentColor" viewBox="0 0 24 24"><path d="M12 2C6.486 2 2 6.486 2 12s4.486 10 10 10 10-4.486 10-10S17.514 2 12 2zm0 18c-4.411 0-8-3.589-8-8s3.589-8 8-8 8 3.589 8 8-3.589 8-8 8z"/><path d="m13 6-6 7h4v5l6-7h-4z"/></svg>
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

        <div class="absolute flex items-center justify-center transform -translate-x-1/2 -translate-y-1/2" style="left: 50%; top: 50%;">
          <!-- Junction Node -->
          <div class="z-10 flex items-center justify-center w-[2rem] h-[2rem] bg-[#3B82F6] rounded-full shadow-sm">
            <!-- Small dot for junction like in image -->
          </div>
        </div>

        <div class="absolute flex flex-col items-center justify-center transform -translate-x-1/2 -translate-y-1/2" style="left: 80%; top: 50%;">
          <div @click="openChart('home')" class="z-10 flex flex-col items-center justify-center w-28 h-28 bg-white dark:bg-gray-800 rounded-full border-[4px] border-[#10B981] shadow-sm cursor-pointer hover:scale-105 transition-transform mb-2 relative">
             <span class="text-xs font-medium text-gray-500 mb-1 absolute -bottom-6">Home</span>
            <svg xmlns="http://www.w3.org/2000/svg" class="h-8 w-8 text-gray-700 dark:text-gray-300 mb-1" fill="currentColor" viewBox="0 0 24 24"><path d="M3 13h1v7c0 1.103.897 2 2 2h12c1.103 0 2-.897 2-2v-7h1a1 1 0 0 0 .707-1.707l-9-9a.999.999 0 0 0-1.414 0l-9 9A1 1 0 0 0 3 13zm7 7v-5h4v5h-4zm2-15.586 6 6V15l.001 5H16v-5c0-1.103-.897-2-2-2h-4c-1.103 0-2 .897-2 2v5H6v-9.586l6-6z"/></svg>
            <div class="text-gray-800 dark:text-gray-200 text-sm font-medium">
              <span v-if="state?.total_load_w !== null && state?.total_load_w !== undefined">
                 {{ formatPowerSimple(homeLoad) }}
              </span>
              <span v-else>0.0 kW</span>
            </div>
          </div>
        </div>

        <!-- Bottom Row: Battery -->
        <div class="absolute flex flex-col items-center justify-center transform -translate-x-1/2 -translate-y-1/2" style="left: 50%; top: 80%;">
          <div v-if="batteryDevices.length > 0" @click="openChart('battery')" class="z-10 flex flex-col items-center justify-center w-28 h-28 bg-white dark:bg-gray-800 rounded-full border-[4px] border-[#EC4899] shadow-sm cursor-pointer hover:scale-105 transition-transform mb-2 relative">
            <span class="text-xs font-medium text-gray-500 mb-1 absolute -bottom-6">Battery</span>
            <svg xmlns="http://www.w3.org/2000/svg" class="h-8 w-8 text-gray-700 dark:text-gray-300 mb-1" fill="currentColor" viewBox="0 0 24 24"><path d="M4 18h14c1.103 0 2-.897 2-2v-2h2v-4h-2V8c0-1.103-.897-2-2-2H4c-1.103 0-2 .897-2 2v8c0 1.103.897 2 2 2zM4 8h14l.002 8H4V8z"/></svg>
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
  has_battery?: boolean;
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

const gridDevices = computed(() => devices.value.filter(d => d.template === 'huawei_dongle' || d.template === 'demo_dongle' || (d.template === 'huawei_inverter' && (d as any).has_grid_meter)))
const solarDevices = computed(() => devices.value.filter(d => d.template === 'huawei_inverter' || d.template === 'demo_inverter'))
const batteryDevices = computed(() => devices.value.filter(d => d.has_battery))
const evDevices = computed(() => devices.value.filter(d => ['raedian_charger', 'demo_charger', 'alfen_charger', 'easee_charger', 'bender_charger', 'peblar_charger', 'phoenix_charger'].includes(d.template)))
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
  if (solarDevices.value.length > 0 && Math.abs(solar) >= 10) {
    solarDevices.value.forEach((d, index) => {
      segments.push({
        id: 'solar-segment-' + d.id,
        path: `M ${50 + (index * 15 - (solarDevices.value.length-1)*7.5)} ${20 - (index % 2 === 0 ? 0 : 5)} L 50 50`,
        power: solar / solarDevices.value.length,
        color: '#FBBF24',
        normalIsPositive: true
      });
    });
  }

  // 2. Grid (Left) <-> Junction
  if (gridDevices.value.length > 0 && Math.abs(grid) >= 10) {
    gridDevices.value.forEach((d, index) => {
      segments.push({
        id: 'grid-segment-' + d.id,
        path: `M ${20 - (index * 5)} ${50 + (index * 10 - (gridDevices.value.length-1)*5)} L 50 50`,
        power: grid / gridDevices.value.length,
        color: '#3B82F6',
        normalIsPositive: true
      });
    });
  }

  // 3. Home (Right) <- Junction
  if (Math.abs(home) >= 10) {
    segments.push({
      id: 'home-segment',
      path: 'M 50 50 L 80 50',
      power: home,
      color: '#10B981',
      normalIsPositive: true
    });
  }

  // 4. Battery (Bottom) <-> Junction
  if (batteryDevices.value.length > 0 && Math.abs(battery) >= 10) {
    batteryDevices.value.forEach((d, index) => {
      segments.push({
        id: 'battery-segment-' + d.id,
        path: `M ${50 + (index * 15 - (batteryDevices.value.length-1)*7.5)} ${80 + (index % 2 === 0 ? 0 : 5)} L 50 50`,
        power: battery / batteryDevices.value.length,
        color: '#EC4899',
        normalIsPositive: true
      });
    });
  }

  // 5. EV (Top-Left) <- Junction
  const evPower = props.state.ev_charger_power_w || 0;
  if (evDevices.value.length > 0 && Math.abs(evPower) >= 10) {
    evDevices.value.forEach((d, index) => {
      segments.push({
        id: 'ev-segment-' + d.id,
        path: `M ${20 + (index * 12)} ${20 + (index * 8)} Q 20 50 50 50`,
        power: evPower / evDevices.value.length,
        color: '#A855F7',
        normalIsPositive: true
      });
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
