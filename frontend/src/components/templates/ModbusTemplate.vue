<template>
  <div class="sm:col-span-2">
    <label :for="prefix + 'host'" class="block text-sm font-medium text-gray-700 dark:text-gray-300">IP Address</label>
    <div class="mt-1">
      <input type="text" :id="prefix + 'host'" :value="modelValue.host" @input="$emit('update:modelValue', { ...modelValue, host: ($event.target as HTMLInputElement).value })"
             class="shadow-sm focus:ring-indigo-500 focus:border-indigo-500 block w-full  border-gray-300 rounded-md dark:bg-gray-700 dark:border-gray-600 dark:text-white">
    </div>
  </div>

  <div class="sm:col-span-2">
    <label :for="prefix + 'port'" class="block text-sm font-medium text-gray-700 dark:text-gray-300">Port</label>
    <div class="mt-1">
      <input type="number" :id="prefix + 'port'" :value="modelValue.port" @input="$emit('update:modelValue', { ...modelValue, port: Number(($event.target as HTMLInputElement).value) })"
             class="shadow-sm focus:ring-indigo-500 focus:border-indigo-500 block w-full  border-gray-300 rounded-md dark:bg-gray-700 dark:border-gray-600 dark:text-white">
    </div>
  </div>

  <div class="sm:col-span-2">
    <label :for="prefix + 'modbus_id'" class="block text-sm font-medium text-gray-700 dark:text-gray-300">Modbus ID</label>
    <div class="mt-1">
      <input type="number" :id="prefix + 'modbus_id'" :value="modelValue.modbus_id" @input="$emit('update:modelValue', { ...modelValue, modbus_id: Number(($event.target as HTMLInputElement).value) })"
             class="shadow-sm focus:ring-indigo-500 focus:border-indigo-500 block w-full  border-gray-300 rounded-md dark:bg-gray-700 dark:border-gray-600 dark:text-white">
    </div>
  </div>
</template>

<script setup lang="ts">


const props = defineProps<{
  modelValue: { host: string; port: number; modbus_id: number };
  prefix?: string;
}>()

const emit = defineEmits(['update:modelValue'])

const prefix = props.prefix || ''
</script>
