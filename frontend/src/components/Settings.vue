<template>
  <main class="max-w-7xl mx-auto py-6 sm:px-6 lg:px-8">
    <div class="px-4 py-6 sm:px-0">

      <!-- Current Devices Section -->
      <div class="mb-8">
        <h2 class="text-2xl font-bold leading-7 text-gray-900 dark:text-white sm:text-3xl sm:truncate mb-4">
          Configured Devices
        </h2>

        <div v-if="devices.length === 0" class="bg-white dark:bg-gray-800 shadow sm:rounded-lg mb-6">
          <div class="px-4 py-5 sm:p-6 text-center text-gray-500 dark:text-gray-400">
            No devices configured yet. Add one below.
          </div>
        </div>

        <div v-else class="grid grid-cols-1 gap-4 sm:grid-cols-2 lg:grid-cols-3 mb-6">
          <div v-for="device in devices" :key="device.id" class="bg-white dark:bg-gray-800 overflow-hidden shadow rounded-lg">
            <div class="px-4 py-5 sm:p-6">
              <h3 class="text-lg leading-6 font-medium text-gray-900 dark:text-white">
                {{ device.name }}
              </h3>
              <div class="mt-2 max-w-xl text-sm text-gray-500 dark:text-gray-400">
                <p><strong>Template:</strong> {{ getTemplateName(device.template) }}</p>
                <p><strong>Host:</strong> {{ device.host }}:{{ device.port }}</p>
                <p><strong>Modbus ID:</strong> {{ device.modbus_id }}</p>
              </div>
              <div class="mt-3 text-sm">
                <button @click="deleteDevice(device.id)" class="font-medium text-red-600 hover:text-red-500 dark:text-red-400 dark:hover:text-red-300">
                  Delete
                </button>
              </div>
            </div>
          </div>
        </div>
      </div>

      <!-- Add Device Section -->
      <div>
        <h2 class="text-2xl font-bold leading-7 text-gray-900 dark:text-white sm:text-3xl sm:truncate mb-4">
          Add Device
        </h2>
        <div class="bg-white dark:bg-gray-800 shadow sm:rounded-lg">
          <div class="px-4 py-5 sm:p-6">
            <form @submit.prevent="addDevice" class="space-y-6">
              <div class="grid grid-cols-1 gap-y-6 gap-x-4 sm:grid-cols-6">

                <div class="sm:col-span-3">
                  <label for="name" class="block text-sm font-medium text-gray-700 dark:text-gray-300">Device Name</label>
                  <div class="mt-1">
                    <input type="text" id="name" v-model="form.name" required
                           class="shadow-sm focus:ring-indigo-500 focus:border-indigo-500 block w-full sm:text-sm border-gray-300 rounded-md dark:bg-gray-700 dark:border-gray-600 dark:text-white">
                  </div>
                </div>

                <div class="sm:col-span-3">
                  <label for="template" class="block text-sm font-medium text-gray-700 dark:text-gray-300">Device Template</label>
                  <div class="mt-1">
                    <select id="template" v-model="form.template" required
                            class="shadow-sm focus:ring-indigo-500 focus:border-indigo-500 block w-full sm:text-sm border-gray-300 rounded-md dark:bg-gray-700 dark:border-gray-600 dark:text-white">
                      <option disabled value="">Please select one</option>
                      <option v-for="t in templates" :key="t.id" :value="t.id">
                        {{ t.name }}
                      </option>
                    </select>
                  </div>
                </div>

                <div class="sm:col-span-2">
                  <label for="host" class="block text-sm font-medium text-gray-700 dark:text-gray-300">IP Address</label>
                  <div class="mt-1">
                    <input type="text" id="host" v-model="form.host" required
                           class="shadow-sm focus:ring-indigo-500 focus:border-indigo-500 block w-full sm:text-sm border-gray-300 rounded-md dark:bg-gray-700 dark:border-gray-600 dark:text-white">
                  </div>
                </div>

                <div class="sm:col-span-2">
                  <label for="port" class="block text-sm font-medium text-gray-700 dark:text-gray-300">Port</label>
                  <div class="mt-1">
                    <input type="number" id="port" v-model.number="form.port" required
                           class="shadow-sm focus:ring-indigo-500 focus:border-indigo-500 block w-full sm:text-sm border-gray-300 rounded-md dark:bg-gray-700 dark:border-gray-600 dark:text-white">
                  </div>
                </div>

                <div class="sm:col-span-2">
                  <label for="modbus_id" class="block text-sm font-medium text-gray-700 dark:text-gray-300">Modbus ID</label>
                  <div class="mt-1">
                    <input type="number" id="modbus_id" v-model.number="form.modbus_id" required
                           class="shadow-sm focus:ring-indigo-500 focus:border-indigo-500 block w-full sm:text-sm border-gray-300 rounded-md dark:bg-gray-700 dark:border-gray-600 dark:text-white">
                  </div>
                </div>
              </div>

              <div class="pt-5">
                <div class="flex justify-end">
                  <button type="submit"
                          class="ml-3 inline-flex justify-center py-2 px-4 border border-transparent shadow-sm text-sm font-medium rounded-md text-white bg-indigo-600 hover:bg-indigo-700 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-indigo-500">
                    Save Device
                  </button>
                </div>
              </div>
            </form>
          </div>
        </div>
      </div>

    </div>
  </main>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'

interface Template {
  id: string;
  name: string;
}

interface Device {
  id: number;
  name: string;
  template: string;
  host: string;
  port: number;
  modbus_id: number;
}

const templates = ref<Template[]>([])
const devices = ref<Device[]>([])

const form = ref({
  name: '',
  template: '',
  host: '',
  port: 502,
  modbus_id: 1
})

const fetchTemplates = async () => {
  try {
    const res = await fetch('http://localhost:8080/api/templates')
    templates.value = await res.json()
  } catch (e) {
    console.error("Failed to fetch templates:", e)
  }
}

const fetchDevices = async () => {
  try {
    const res = await fetch('http://localhost:8080/api/devices')
    devices.value = await res.json() || []
  } catch (e) {
    console.error("Failed to fetch devices:", e)
  }
}

const getTemplateName = (id: string) => {
  const t = templates.value.find(t => t.id === id)
  return t ? t.name : id
}

const addDevice = async () => {
  try {
    const res = await fetch('http://localhost:8080/api/devices', {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json'
      },
      body: JSON.stringify(form.value)
    })

    if (res.ok) {
      // Reset form
      form.value = {
        name: '',
        template: '',
        host: '',
        port: 502,
        modbus_id: 1
      }
      await fetchDevices()
    }
  } catch (e) {
    console.error("Failed to add device:", e)
  }
}

const deleteDevice = async (id: number) => {
  try {
    const res = await fetch(`http://localhost:8080/api/devices/${id}`, {
      method: 'DELETE'
    })

    if (res.ok) {
      await fetchDevices()
    }
  } catch (e) {
    console.error("Failed to delete device:", e)
  }
}

onMounted(() => {
  fetchTemplates()
  fetchDevices()
})
</script>
