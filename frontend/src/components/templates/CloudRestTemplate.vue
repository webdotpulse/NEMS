<template>
  <div class="sm:col-span-3">
    <label :for="prefix + 'host'" class="block text-sm font-medium text-gray-700 dark:text-gray-300">IP Address</label>
    <div class="mt-1">
      <input type="text" :id="prefix + 'host'" :value="modelValue.host" @input="$emit('update:modelValue', { ...modelValue, host: ($event.target as HTMLInputElement).value })"
             class="shadow-sm focus:ring-indigo-500 focus:border-indigo-500 block w-full  border-gray-300 rounded-md dark:bg-gray-700 dark:border-gray-600 dark:text-white">
    </div>
  </div>

  <div class="sm:col-span-3">
    <label :for="prefix + 'port'" class="block text-sm font-medium text-gray-700 dark:text-gray-300">Port</label>
    <div class="mt-1">
      <input type="number" :id="prefix + 'port'" :value="modelValue.port" @input="$emit('update:modelValue', { ...modelValue, port: Number(($event.target as HTMLInputElement).value) })"
             class="shadow-sm focus:ring-indigo-500 focus:border-indigo-500 block w-full  border-gray-300 rounded-md dark:bg-gray-700 dark:border-gray-600 dark:text-white">
    </div>
  </div>

  <div class="sm:col-span-3">
    <label :for="prefix + 'username'" class="block text-sm font-medium text-gray-700 dark:text-gray-300">Username</label>
    <div class="mt-1">
      <input type="text" :id="prefix + 'username'" :value="modelValue.username" @input="$emit('update:modelValue', { ...modelValue, username: ($event.target as HTMLInputElement).value })"
             class="shadow-sm focus:ring-indigo-500 focus:border-indigo-500 block w-full  border-gray-300 rounded-md dark:bg-gray-700 dark:border-gray-600 dark:text-white">
    </div>
  </div>

  <div class="sm:col-span-3">
    <label :for="prefix + 'password'" class="block text-sm font-medium text-gray-700 dark:text-gray-300">Password</label>
    <div class="mt-1">
      <input type="password" :id="prefix + 'password'" :value="modelValue.password" @input="$emit('update:modelValue', { ...modelValue, password: ($event.target as HTMLInputElement).value })"
             class="shadow-sm focus:ring-indigo-500 focus:border-indigo-500 block w-full  border-gray-300 rounded-md dark:bg-gray-700 dark:border-gray-600 dark:text-white">
    </div>
  </div>
</template>

<script setup lang="ts">


const props = defineProps<{
  modelValue: { host: string; port: number; username?: string; password?: string };
  prefix?: string;
}>()

const emit = defineEmits(['update:modelValue'])

const prefix = props.prefix || ''
</script>
