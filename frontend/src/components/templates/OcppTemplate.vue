<template>
  <div class="sm:col-span-6 mb-4">
    <div class="rounded-md bg-blue-50 dark:bg-blue-900/30 p-4 border border-blue-200 dark:border-blue-800">
      <div class="flex">
        <div class="flex-shrink-0">
          <svg class="h-5 w-5 text-blue-400" xmlns="http://www.w3.org/2000/svg" viewBox="0 0 20 20" fill="currentColor" aria-hidden="true">
            <path fill-rule="evenodd" d="M18 10a8 8 0 11-16 0 8 8 0 0116 0zm-7-4a1 1 0 11-2 0 1 1 0 012 0zM9 9a1 1 0 000 2v3a1 1 0 001 1h1a1 1 0 100-2v-3a1 1 0 00-1-1H9z" clip-rule="evenodd" />
          </svg>
        </div>
        <div class="ml-3 flex-1 md:flex md:justify-between">
          <p class="text-sm text-blue-700 dark:text-blue-300">
            Configure your EV charger to connect to this CSMS WebSocket URL:
            <br>
            <strong class="block mt-1 bg-white dark:bg-gray-800 p-2 rounded border border-blue-100 dark:border-blue-700 break-all select-all font-mono text-xs">
              {{ wsUrl }}
            </strong>
          </p>
        </div>
      </div>
    </div>
  </div>

  <div class="sm:col-span-3">
    <label :for="prefix + 'host'" class="block text-sm font-medium text-gray-700 dark:text-gray-300">ChargePoint ID</label>
    <div class="mt-1">
      <input type="text" :id="prefix + 'host'" :value="modelValue.host" @input="$emit('update:modelValue', { ...modelValue, host: ($event.target as HTMLInputElement).value })"
             placeholder="e.g. CS-001" required
             class="shadow-sm focus:ring-indigo-500 focus:border-indigo-500 block w-full  border-gray-300 rounded-md dark:bg-gray-700 dark:border-gray-600 dark:text-white transition-all duration-200">
    </div>
    <p class="mt-1 text-xs text-gray-500 dark:text-gray-400">Must exactly match the charger's configured ID.</p>
  </div>

</template>

<script setup lang="ts">
import { computed } from 'vue'
import { getApiBase } from '../../api'

const props = defineProps<{
  modelValue: { host: string; port?: number };
  prefix?: string;
}>()

const emit = defineEmits(['update:modelValue'])

const prefix = props.prefix || ''

const wsUrl = computed(() => {
  const chargePointId = props.modelValue.host || '{CHARGE_POINT_ID}'
  let base = getApiBase()

  if (base) {
    base = base.replace(/^http/, 'ws')
  } else {
    const protocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:'
    base = `${protocol}//${window.location.host}`
  }
  return `${base}/api/ocpp/${chargePointId}`
})

</script>
