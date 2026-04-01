<template>
  <div class="space-y-4">
    <div>
      <label class="block text-sm font-medium text-gray-700 dark:text-gray-300">TCP Host</label>
      <input v-model="device.host" type="text" placeholder="192.168.1.50" class="mt-1 block w-full rounded-md border-gray-300 shadow-sm focus:border-blue-500 focus:ring-blue-500  dark:bg-gray-700 dark:border-gray-600 dark:text-white">
    </div>
    <div>
      <label class="block text-sm font-medium text-gray-700 dark:text-gray-300">TCP Port</label>
      <input v-model.number="device.port" type="number" placeholder="23" class="mt-1 block w-full rounded-md border-gray-300 shadow-sm focus:border-blue-500 focus:ring-blue-500  dark:bg-gray-700 dark:border-gray-600 dark:text-white">
    </div>
  </div>
</template>

<script setup lang="ts">
const props = defineProps<{
  device: any
}>()
</script>
