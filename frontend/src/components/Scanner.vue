<template>
  <div class="max-w-7xl mx-auto py-6 sm:px-6 lg:px-8">
    <div class="bg-white dark:bg-gray-800 shadow sm:rounded-lg overflow-hidden border border-gray-200 dark:border-gray-700">
      <div class="px-4 py-5 sm:px-6 flex justify-between items-center border-b border-gray-200 dark:border-gray-700">
        <div>
          <h3 class="text-lg leading-6 font-medium text-gray-900 dark:text-white">
            Network Scanner
          </h3>
          <p class="mt-1 max-w-2xl text-sm text-gray-500 dark:text-gray-400">
            Scan the local subnet for Modbus, Web, and IoT devices.
          </p>
        </div>
        <button
          @click="startScan"
          :disabled="isScanning"
          class="inline-flex items-center px-4 py-2 border border-transparent text-sm font-medium rounded-md shadow-sm text-white bg-indigo-600 hover:bg-indigo-700 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-indigo-500 disabled:opacity-50 disabled:cursor-not-allowed"
        >
          <svg v-if="isScanning" class="animate-spin -ml-1 mr-2 h-4 w-4 text-white" xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24">
            <circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4"></circle>
            <path class="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"></path>
          </svg>
          <svg v-else class="-ml-1 mr-2 h-4 w-4" xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" stroke="currentColor">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M21 21l-6-6m2-5a7 7 0 11-14 0 7 7 0 0114 0z" />
          </svg>
          {{ isScanning ? 'Scanning...' : 'Start Scan' }}
        </button>
      </div>

      <div v-if="error" class="bg-red-50 dark:bg-red-900/20 border-l-4 border-red-400 p-4 m-4">
        <div class="flex">
          <div class="flex-shrink-0">
            <svg class="h-5 w-5 text-red-400" xmlns="http://www.w3.org/2000/svg" viewBox="0 0 20 20" fill="currentColor">
              <path fill-rule="evenodd" d="M10 18a8 8 0 100-16 8 8 0 000 16zM8.707 7.293a1 1 0 00-1.414 1.414L8.586 10l-1.293 1.293a1 1 0 101.414 1.414L10 11.414l1.293 1.293a1 1 0 001.414-1.414L11.414 10l1.293-1.293a1 1 0 00-1.414-1.414L10 8.586 8.707 7.293z" clip-rule="evenodd" />
            </svg>
          </div>
          <div class="ml-3">
            <p class="text-sm text-red-700 dark:text-red-200">
              {{ error }}
            </p>
          </div>
        </div>
      </div>

      <div class="overflow-x-auto">
        <table class="min-w-full divide-y divide-gray-200 dark:divide-gray-700">
          <thead class="bg-gray-50 dark:bg-gray-700/50">
            <tr>
              <th scope="col" class="px-6 py-3 text-left text-xs font-medium text-gray-500 dark:text-gray-400 uppercase tracking-wider">
                IP Address
              </th>
              <th scope="col" class="px-6 py-3 text-left text-xs font-medium text-gray-500 dark:text-gray-400 uppercase tracking-wider">
                Hostname
              </th>
              <th scope="col" class="px-6 py-3 text-left text-xs font-medium text-gray-500 dark:text-gray-400 uppercase tracking-wider">
                MAC Address
              </th>
              <th scope="col" class="px-6 py-3 text-left text-xs font-medium text-gray-500 dark:text-gray-400 uppercase tracking-wider">
                Vendor
              </th>
              <th scope="col" class="px-6 py-3 text-left text-xs font-medium text-gray-500 dark:text-gray-400 uppercase tracking-wider">
                Open Ports
              </th>
            </tr>
          </thead>
          <tbody class="bg-white dark:bg-gray-800 divide-y divide-gray-200 dark:divide-gray-700">
            <tr v-if="devices.length === 0 && !isScanning && !error">
              <td colspan="4" class="px-6 py-10 text-center text-sm text-gray-500 dark:text-gray-400">
                Click "Start Scan" to discover devices on your local network.
              </td>
            </tr>
            <tr v-else-if="devices.length === 0 && isScanning">
              <td colspan="4" class="px-6 py-10 text-center text-sm text-gray-500 dark:text-gray-400">
                Scanning network... This may take up to 10 seconds.
              </td>
            </tr>
            <tr v-for="device in sortedDevices" :key="device.ip" class="hover:bg-gray-50 dark:hover:bg-gray-700/50">
              <td class="px-6 py-4 whitespace-nowrap text-sm font-medium text-gray-900 dark:text-white">
                {{ device.ip }}
              </td>
              <td class="px-6 py-4 whitespace-nowrap text-sm text-gray-500 dark:text-gray-400">
                {{ device.hostname }}
              </td>
              <td class="px-6 py-4 whitespace-nowrap text-sm text-gray-500 dark:text-gray-400 font-mono">
                {{ device.mac }}
              </td>
              <td class="px-6 py-4 whitespace-nowrap text-sm text-gray-500 dark:text-gray-400">
                {{ device.vendor }}
              </td>
              <td class="px-6 py-4 whitespace-nowrap text-sm text-gray-500 dark:text-gray-400">
                <div class="flex flex-wrap gap-1">
                  <span v-for="port in device.open_ports" :key="port" class="inline-flex items-center px-2 py-0.5 rounded text-xs font-medium bg-blue-100 text-blue-800 dark:bg-blue-900 dark:text-blue-200">
                    {{ port }}
                  </span>
                </div>
              </td>
            </tr>
          </tbody>
        </table>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed } from 'vue';
import { getApiBase } from '../api';

interface DiscoveredDevice {
  ip: string;
  hostname: string;
  mac: string;
  vendor: string;
  open_ports: number[];
}

const devices = ref<DiscoveredDevice[]>([]);
const isScanning = ref(false);
const error = ref<string | null>(null);

const sortedDevices = computed(() => {
  return [...devices.value].sort((a, b) => {
    // Basic IP sorting
    const ipA = a.ip.split('.').map(Number);
    const ipB = b.ip.split('.').map(Number);
    for (let i = 0; i < 4; i++) {
      if (ipA[i] !== ipB[i]) return ipA[i] - ipB[i];
    }
    return 0;
  });
});

const startScan = async () => {
  if (isScanning.value) return;

  isScanning.value = true;
  error.value = null;
  devices.value = [];

  try {
    const response = await fetch(`${getApiBase()}/api/network/scan`);
    if (!response.ok) {
      throw new Error(`Server returned ${response.status}: ${response.statusText}`);
    }
    const data = await response.json();
    if (data) {
      devices.value = data;
    }
  } catch (err: any) {
    error.value = `Failed to scan network: ${err.message}`;
    console.error(err);
  } finally {
    isScanning.value = false;
  }
};
</script>
