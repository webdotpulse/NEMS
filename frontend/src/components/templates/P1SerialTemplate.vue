<template>
  <div class="space-y-4">
    <div>
      <label class="block text-sm font-medium text-gray-700 dark:text-gray-300">USB Serial Port</label>
      <input v-model="device.host" type="text" placeholder="/dev/ttyUSB0" class="mt-1 block w-full rounded-md border-gray-300 shadow-sm focus:border-blue-500 focus:ring-blue-500 sm:text-sm dark:bg-gray-700 dark:border-gray-600 dark:text-white">
      <p class="text-xs text-gray-500 dark:text-gray-400 mt-1">Example: /dev/ttyUSB0 or COM3</p>
    </div>
  </div>
</template>

<script setup lang="ts">
const props = defineProps<{
  device: any
}>()
</script>
