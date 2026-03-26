<template>
  <div class="relative w-full max-w-3xl mx-auto h-[400px] bg-white dark:bg-gray-800 rounded-lg shadow overflow-hidden">
    <!-- SVG paths for animated power flow -->
    <svg viewBox="0 0 100 100" preserveAspectRatio="none" class="absolute inset-0 w-full h-full z-0">
      <!-- Grid to Home -->
      <path
        v-if="hasGrid"
        d="M 20 50 L 50 50"
        class="stroke-gray-300 dark:stroke-gray-600 stroke-2 fill-none"
        vector-effect="non-scaling-stroke"
      />
      <path
        v-if="hasGrid && state?.grid_power_w !== 0"
        d="M 20 50 L 50 50"
        class="stroke-blue-500 stroke-2 fill-none flow-path"
        vector-effect="non-scaling-stroke"
        :style="gridFlowStyle"
      />

      <!-- Solar to Home -->
      <path
        v-if="hasSolar"
        d="M 50 20 L 50 50"
        class="stroke-gray-300 dark:stroke-gray-600 stroke-2 fill-none"
        vector-effect="non-scaling-stroke"
      />
      <path
        v-if="hasSolar && state?.solar_power_w > 0"
        d="M 50 20 L 50 50"
        class="stroke-yellow-400 stroke-2 fill-none flow-path"
        vector-effect="non-scaling-stroke"
        :style="solarFlowStyle"
      />

      <!-- Battery to Home -->
      <path
        v-if="hasBattery"
        d="M 50 80 L 50 50"
        class="stroke-gray-300 dark:stroke-gray-600 stroke-2 fill-none"
        vector-effect="non-scaling-stroke"
      />
      <path
        v-if="hasBattery && state?.battery_power_w !== 0"
        d="M 50 80 L 50 50"
        class="stroke-green-400 stroke-2 fill-none flow-path"
        vector-effect="non-scaling-stroke"
        :style="batteryFlowStyle"
      />

      <!-- Home to EV Charger -->
      <path
        v-if="hasEvCharger"
        d="M 50 50 L 80 50"
        class="stroke-gray-300 dark:stroke-gray-600 stroke-2 fill-none"
        vector-effect="non-scaling-stroke"
      />
      <path
        v-if="hasEvCharger && state?.ev_charger_power_w > 0"
        d="M 50 50 L 80 50"
        class="stroke-purple-500 stroke-2 fill-none flow-path"
        vector-effect="non-scaling-stroke"
        :style="evFlowStyle"
      />
    </svg>

    <!-- Node UI Elements -->

    <!-- Grid Node -->
    <div v-if="hasGrid" class="absolute top-[50%] left-[20%] -translate-x-1/2 -translate-y-1/2 z-10 flex flex-col items-center justify-center w-24 h-24 bg-white dark:bg-gray-800 rounded-full border-4 border-blue-500 shadow-md">
      <svg class="h-8 w-8 text-blue-500 mb-1" fill="none" viewBox="0 0 24 24" stroke="currentColor">
        <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M13 10V3L4 14h7v7l9-11h-7z" />
      </svg>
      <div class="text-xs font-semibold text-gray-700 dark:text-gray-200">Grid</div>
      <div class="text-xs text-gray-500 dark:text-gray-400">{{ Math.abs(state?.grid_power_w || 0).toFixed(0) }}W</div>
    </div>

    <!-- Solar Node -->
    <div v-if="hasSolar" class="absolute top-[20%] left-[50%] -translate-x-1/2 -translate-y-1/2 z-10 flex flex-col items-center justify-center w-24 h-24 bg-white dark:bg-gray-800 rounded-full border-4 border-yellow-400 shadow-md">
      <svg class="h-8 w-8 text-yellow-400 mb-1" fill="none" viewBox="0 0 24 24" stroke="currentColor">
        <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 3v1m0 16v1m9-9h-1M4 12H3m15.364 6.364l-.707-.707M6.343 6.343l-.707-.707m12.728 0l-.707.707M6.343 17.657l-.707.707M16 12a4 4 0 11-8 0 4 4 0 018 0z" />
      </svg>
      <div class="text-xs font-semibold text-gray-700 dark:text-gray-200">Solar</div>
      <div class="text-xs text-gray-500 dark:text-gray-400">{{ (state?.solar_power_w || 0).toFixed(0) }}W</div>
    </div>

    <!-- Battery Node -->
    <div v-if="hasBattery" class="absolute top-[80%] left-[50%] -translate-x-1/2 -translate-y-1/2 z-10 flex flex-col items-center justify-center w-24 h-24 bg-white dark:bg-gray-800 rounded-full border-4 border-green-400 shadow-md">
      <svg class="h-8 w-8 text-green-400 mb-1" fill="none" viewBox="0 0 24 24" stroke="currentColor">
        <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M4 7v10c0 2.21 3.582 4 8 4s8-1.79 8-4V7M4 7c0 2.21 3.582 4 8 4s8-1.79 8-4M4 7c0-2.21 3.582-4 8-4s8 1.79 8 4" />
      </svg>
      <div class="text-xs font-semibold text-gray-700 dark:text-gray-200">Battery</div>
      <div class="text-xs text-gray-500 dark:text-gray-400">{{ Math.abs(state?.battery_power_w || 0).toFixed(0) }}W</div>
    </div>

    <!-- Home Load Node (Center) -->
    <div class="absolute top-[50%] left-[50%] -translate-x-1/2 -translate-y-1/2 z-10 flex flex-col items-center justify-center w-28 h-28 bg-white dark:bg-gray-800 rounded-full border-4 border-indigo-500 shadow-md">
      <svg class="h-10 w-10 text-indigo-500 mb-1" fill="none" viewBox="0 0 24 24" stroke="currentColor">
        <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M3 12l2-2m0 0l7-7 7 7M5 10v10a1 1 0 001 1h3m10-11l2 2m-2-2v10a1 1 0 01-1 1h-3m-6 0a1 1 0 001-1v-4a1 1 0 011-1h2a1 1 0 011 1v4a1 1 0 001 1m-6 0h6" />
      </svg>
      <div class="text-sm font-bold text-gray-700 dark:text-gray-200">Home</div>
      <div class="text-sm text-gray-500 dark:text-gray-400">{{ Math.max(0, homeLoad).toFixed(0) }}W</div>
    </div>

    <!-- EV Charger Node -->
    <div v-if="hasEvCharger" class="absolute top-[50%] left-[80%] -translate-x-1/2 -translate-y-1/2 z-10 flex flex-col items-center justify-center w-24 h-24 bg-white dark:bg-gray-800 rounded-full border-4 border-purple-500 shadow-md">
      <svg class="h-8 w-8 text-purple-500 mb-1" fill="none" viewBox="0 0 24 24" stroke="currentColor">
        <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M13 10V3L4 14h7v7l9-11h-7z" />
        <!-- using a generic power icon for now as EV might be missing from tailwind generic set, it's just a demo representation -->
        <circle cx="12" cy="12" r="3" fill="currentColor"/>
      </svg>
      <div class="text-xs font-semibold text-gray-700 dark:text-gray-200">EV</div>
      <div class="text-xs text-gray-500 dark:text-gray-400">{{ (state?.ev_charger_power_w || 0).toFixed(0) }}W</div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'

interface SiteState {
  grid_power_w: number
  solar_power_w: number
  battery_power_w: number
  total_load_w: number
  ev_charger_power_w: number
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

const hasGrid = computed(() => devices.value.some(d => d.template === 'huawei_dongle'))
const hasSolar = computed(() => devices.value.some(d => d.template === 'huawei_inverter'))
const hasBattery = computed(() => devices.value.some(d => d.template === 'huawei_inverter' && d.name.toLowerCase().includes('battery')))
const hasEvCharger = computed(() => devices.value.some(d => d.template === 'raedian_charger'))

const homeLoad = computed(() => {
  if (!props.state) return 0
  return props.state.total_load_w - (props.state.ev_charger_power_w || 0)
})

// Calculate flow animation styles
// Baseline animation duration (e.g. 1000W = 2s, 5000W = 0.4s)
// The animation moves dashed strokes along the path.
// Normal direction moves from start of path to end. Reverse does the opposite.
const calculateFlowStyle = (power: number, baseDuration: number = 2000) => {
  const absPower = Math.abs(power)
  if (absPower < 10) return { display: 'none' } // don't animate if power is very low

  // Cap animation speed between 0.3s and 5s to keep it looking decent
  const duration = Math.min(Math.max(baseDuration / absPower, 0.3), 5)
  return {
    animationDuration: `${duration}s`,
    animationDirection: power > 0 ? 'normal' : 'reverse'
  }
}

const gridFlowStyle = computed(() => calculateFlowStyle(props.state?.grid_power_w || 0))
const solarFlowStyle = computed(() => calculateFlowStyle(props.state?.solar_power_w || 0))
// Discharging = positive (moving to home), Charging = negative (moving from home)
const batteryFlowStyle = computed(() => calculateFlowStyle(props.state?.battery_power_w || 0))
const evFlowStyle = computed(() => calculateFlowStyle(props.state?.ev_charger_power_w || 0))

</script>

<style scoped>
.flow-path {
  stroke-dasharray: 6 12;
  animation: flow linear infinite;
}

@keyframes flow {
  from {
    stroke-dashoffset: 18;
  }
  to {
    stroke-dashoffset: 0;
  }
}
</style>
