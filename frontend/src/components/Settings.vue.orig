<template>
  <main class="max-w-7xl mx-auto py-6 sm:px-6 lg:px-8">
    <div class="px-4 py-6 sm:px-0">

      <!-- Site Optimization Section -->
      <div class="mb-8">
        <h2 class="text-2xl font-bold leading-7 text-gray-900 dark:text-white sm:text-3xl sm:truncate mb-4">
          Site Optimization
        </h2>
        <div class="bg-white dark:bg-gray-800 shadow sm:rounded-lg">
          <div class="px-4 py-5 sm:p-6">
            <form @submit.prevent="saveSiteSettings" class="space-y-6">
              <div class="grid grid-cols-1 gap-y-6 gap-x-4 sm:grid-cols-6">

                <div class="sm:col-span-3">
                  <label for="strategy_mode" class="block text-sm font-medium text-gray-700 dark:text-gray-300">Optimization Strategy</label>
                  <div class="mt-1">
                    <select id="strategy_mode" v-model="siteSettings.strategy_mode"
                            class="shadow-sm focus:ring-indigo-500 focus:border-indigo-500 block w-full sm:text-sm border-gray-300 rounded-md dark:bg-gray-700 dark:border-gray-600 dark:text-white">
                      <option value="eco">Eco (Max Self-Consumption)</option>
                      <option value="flanders">Flanders Mode (Peak Shaving)</option>
                      <option value="netherlands">Netherlands Mode (Zero-Export)</option>
                    </select>
                  </div>
                </div>

                <div v-if="siteSettings.strategy_mode === 'flanders'" class="sm:col-span-3">
                  <label for="capacity_peak_limit" class="block text-sm font-medium text-gray-700 dark:text-gray-300">Capacity Peak Limit (kW)</label>
                  <div class="mt-1">
                    <input type="number" step="0.1" id="capacity_peak_limit" v-model="siteSettings.capacity_peak_limit_kw"
                           class="shadow-sm focus:ring-indigo-500 focus:border-indigo-500 block w-full sm:text-sm border-gray-300 rounded-md dark:bg-gray-700 dark:border-gray-600 dark:text-white">
                  </div>
                </div>

                <div v-if="siteSettings.strategy_mode === 'netherlands'" class="sm:col-span-3 flex items-center pt-6">
                  <input id="active_inverter_curtailment" type="checkbox" v-model="siteSettings.active_inverter_curtailment"
                         class="h-4 w-4 text-indigo-600 focus:ring-indigo-500 border-gray-300 rounded dark:bg-gray-700 dark:border-gray-600" />
                  <label for="active_inverter_curtailment" class="ml-2 block text-sm text-gray-900 dark:text-gray-300">
                    Active Inverter Curtailment
                  </label>
                </div>

              </div>

              <div class="pt-5 border-t border-gray-200 dark:border-gray-700">
                <div class="flex justify-end">
                  <button type="submit"
                          class="ml-3 inline-flex justify-center py-2 px-4 border border-transparent shadow-sm text-sm font-medium rounded-md text-white bg-indigo-600 hover:bg-indigo-700 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-indigo-500">
                    Save Strategy
                  </button>
                </div>
              </div>
            </form>
            <div v-if="saveSettingsSuccess" class="mt-4 text-sm text-green-600 dark:text-green-400">
              Strategy saved successfully!
            </div>
          </div>
        </div>
      </div>

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
              <h3 class="text-lg leading-6 font-medium text-gray-900 dark:text-white flex items-center">
                {{ device.name }}
                <span v-if="device.status === 'online'" class="ml-2 flex w-3 h-3 bg-green-500 rounded-full" title="Online"></span>
                <span v-else-if="device.status === 'error'" class="ml-2 flex w-3 h-3 bg-red-500 rounded-full" title="Error"></span>
                <span v-else class="ml-2 flex w-3 h-3 bg-gray-400 rounded-full" title="Offline"></span>
              </h3>
              <div class="mt-2 max-w-xl text-sm text-gray-500 dark:text-gray-400">
                <p><strong>Template:</strong> {{ getTemplateName(device.template) }}</p>
                <p><strong>Host:</strong> {{ device.host }}:{{ device.port }}</p>
                <p><strong>Modbus ID:</strong> {{ device.modbus_id }}</p>
              </div>
              <div class="mt-3 text-sm flex space-x-4">
                <button @click="editDevice(device)" class="font-medium text-indigo-600 hover:text-indigo-500 dark:text-indigo-400 dark:hover:text-indigo-300">
                  Edit
                </button>
                <button @click="deleteDevice(device.id)" class="font-medium text-red-600 hover:text-red-500 dark:text-red-400 dark:hover:text-red-300">
                  Delete
                </button>
              </div>
            </div>
          </div>
        </div>
      </div>

      <!-- Edit Device Modal -->
      <div v-if="editingDevice" class="fixed inset-0 z-50 flex items-center justify-center bg-black bg-opacity-50 transition-opacity">
        <div class="bg-white dark:bg-gray-800 rounded-lg shadow-xl w-11/12 max-w-2xl p-6 relative">
          <button @click="closeEdit" class="absolute top-4 right-4 text-gray-500 hover:text-gray-700 dark:text-gray-400 dark:hover:text-gray-200">
            <svg class="h-6 w-6" fill="none" viewBox="0 0 24 24" stroke="currentColor">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12" />
            </svg>
          </button>
          <h3 class="text-xl font-semibold mb-4 text-gray-800 dark:text-gray-100">Edit Device</h3>
          <form @submit.prevent="updateDevice" class="space-y-6">
            <div class="grid grid-cols-1 gap-y-6 gap-x-4 sm:grid-cols-6">
              <div class="sm:col-span-3">
                <label for="edit_name" class="block text-sm font-medium text-gray-700 dark:text-gray-300">Device Name</label>
                <div class="mt-1">
                  <input type="text" id="edit_name" v-model="editForm.name" required
                         class="shadow-sm focus:ring-indigo-500 focus:border-indigo-500 block w-full sm:text-sm border-gray-300 rounded-md dark:bg-gray-700 dark:border-gray-600 dark:text-white">
                </div>
              </div>

              <div class="sm:col-span-3">
                <label for="edit_template" class="block text-sm font-medium text-gray-700 dark:text-gray-300">Device Template</label>
                <div class="mt-1">
                  <select id="edit_template" v-model="editForm.template" required
                          class="shadow-sm focus:ring-indigo-500 focus:border-indigo-500 block w-full sm:text-sm border-gray-300 rounded-md dark:bg-gray-700 dark:border-gray-600 dark:text-white">
                    <option disabled value="">Please select one</option>
                    <option v-for="t in templates" :key="t.id" :value="t.id">
                      {{ t.name }}
                    </option>
                  </select>
                </div>
              </div>

              <template v-if="templates.find(t => t.id === editForm.template)?.type === 'modbus'">
                <ModbusTemplate v-model="editForm" prefix="edit_" />
              </template>
              <template v-else-if="templates.find(t => t.id === editForm.template)?.type === 'cloud'">
                <CloudTemplate v-model="editForm" prefix="edit_" />
              </template>
              <template v-else-if="templates.find(t => t.id === editForm.template)?.type === 'rest'">
                <RestTemplate v-model="editForm" prefix="edit_" />
              </template>
              <template v-else-if="templates.find(t => t.id === editForm.template)?.type === 'cloud_rest'">
                <CloudRestTemplate v-model="editForm" prefix="edit_" />
              </template>

              <!-- Conditional Fields for Huawei Hybrid Inverter Combo -->
              <template v-if="editForm.template === 'huawei_inverter'">
                <div class="sm:col-span-6 border-t border-gray-200 dark:border-gray-700 pt-4 mt-2">
                  <h4 class="text-sm font-medium text-gray-900 dark:text-gray-100 mb-4">Huawei Inverter Features</h4>

                  <div class="flex items-center mb-4">
                    <input id="edit_has_grid_meter" type="checkbox" v-model="editForm.has_grid_meter"
                           class="h-4 w-4 text-indigo-600 focus:ring-indigo-500 border-gray-300 rounded dark:bg-gray-700 dark:border-gray-600" />
                    <label for="edit_has_grid_meter" class="ml-2 block text-sm text-gray-900 dark:text-gray-300">
                      Grid Meter Connected?
                    </label>
                  </div>

                  <div class="flex items-center mb-4">
                    <input id="edit_has_battery" type="checkbox" v-model="editForm.has_battery"
                           class="h-4 w-4 text-indigo-600 focus:ring-indigo-500 border-gray-300 rounded dark:bg-gray-700 dark:border-gray-600" />
                    <label for="edit_has_battery" class="ml-2 block text-sm text-gray-900 dark:text-gray-300">
                      Battery Connected?
                    </label>
                  </div>

                  <div v-if="editForm.has_battery" class="sm:col-span-3 ml-6">
                    <label for="edit_battery_capacity" class="block text-sm font-medium text-gray-700 dark:text-gray-300">Battery Capacity (kWh)</label>
                    <div class="mt-1">
                      <input type="number" step="0.1" id="edit_battery_capacity" v-model="editForm.battery_capacity"
                             class="shadow-sm focus:ring-indigo-500 focus:border-indigo-500 block w-full sm:text-sm border-gray-300 rounded-md dark:bg-gray-700 dark:border-gray-600 dark:text-white">
                    </div>
                  </div>
                </div>
              </template>
            </div>

            <div class="pt-5">
              <div class="flex justify-end">
                <button type="button" @click="closeEdit" class="bg-white py-2 px-4 border border-gray-300 rounded-md shadow-sm text-sm font-medium text-gray-700 hover:bg-gray-50 dark:bg-gray-700 dark:border-gray-600 dark:text-gray-200 dark:hover:bg-gray-600">
                  Cancel
                </button>
                <button type="submit"
                        class="ml-3 inline-flex justify-center py-2 px-4 border border-transparent shadow-sm text-sm font-medium rounded-md text-white bg-indigo-600 hover:bg-indigo-700 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-indigo-500">
                  Save Changes
                </button>
              </div>
            </div>
          </form>
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

                <template v-if="templates.find(t => t.id === form.template)?.type === 'modbus'">
                  <ModbusTemplate v-model="form" />
                </template>
                <template v-else-if="templates.find(t => t.id === form.template)?.type === 'cloud'">
                  <CloudTemplate v-model="form" />
                </template>
                <template v-else-if="templates.find(t => t.id === form.template)?.type === 'rest'">
                  <RestTemplate v-model="form" />
                </template>
                <template v-else-if="templates.find(t => t.id === form.template)?.type === 'cloud_rest'">
                  <CloudRestTemplate v-model="form" />
                </template>

                <!-- Conditional Fields for Huawei Hybrid Inverter Combo -->
                <template v-if="form.template === 'huawei_inverter'">
                  <div class="sm:col-span-6 border-t border-gray-200 dark:border-gray-700 pt-4 mt-2">
                    <h4 class="text-sm font-medium text-gray-900 dark:text-gray-100 mb-4">Huawei Inverter Features</h4>

                    <div class="flex items-center mb-4">
                      <input id="has_grid_meter" type="checkbox" v-model="form.has_grid_meter"
                             class="h-4 w-4 text-indigo-600 focus:ring-indigo-500 border-gray-300 rounded dark:bg-gray-700 dark:border-gray-600" />
                      <label for="has_grid_meter" class="ml-2 block text-sm text-gray-900 dark:text-gray-300">
                        Grid Meter Connected?
                      </label>
                    </div>

                    <div class="flex items-center mb-4">
                      <input id="has_battery" type="checkbox" v-model="form.has_battery"
                             class="h-4 w-4 text-indigo-600 focus:ring-indigo-500 border-gray-300 rounded dark:bg-gray-700 dark:border-gray-600" />
                      <label for="has_battery" class="ml-2 block text-sm text-gray-900 dark:text-gray-300">
                        Battery Connected?
                      </label>
                    </div>

                    <div v-if="form.has_battery" class="sm:col-span-3 ml-6">
                      <label for="battery_capacity" class="block text-sm font-medium text-gray-700 dark:text-gray-300">Battery Capacity (kWh)</label>
                      <div class="mt-1">
                        <input type="number" step="0.1" id="battery_capacity" v-model="form.battery_capacity"
                               class="shadow-sm focus:ring-indigo-500 focus:border-indigo-500 block w-full sm:text-sm border-gray-300 rounded-md dark:bg-gray-700 dark:border-gray-600 dark:text-white">
                      </div>
                    </div>
                  </div>
                </template>

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
import ModbusTemplate from './templates/ModbusTemplate.vue'
import CloudTemplate from './templates/CloudTemplate.vue'
import RestTemplate from './templates/RestTemplate.vue'
import CloudRestTemplate from './templates/CloudRestTemplate.vue'

interface Template {
  id: string;
  name: string;
  type: string;
}

interface Device {
  id: number;
  name: string;
  template: string;
  host: string;
  port: number;
  modbus_id: number;
  username?: string;
  password?: string;
  status?: string;
  has_grid_meter?: boolean;
  has_battery?: boolean;
  battery_capacity?: number;
}

const templates = ref<Template[]>([])
const devices = ref<Device[]>([])

const form = ref({
  name: '',
  template: '',
  host: '',
  port: 502,
  modbus_id: 1,
  username: '',
  password: '',
  has_grid_meter: false,
  has_battery: false,
  battery_capacity: 0
})

const editingDevice = ref<Device | null>(null)
const editForm = ref({
  id: 0,
  name: '',
  template: '',
  host: '',
  port: 502,
  modbus_id: 1,
  username: '',
  password: '',
  has_grid_meter: false,
  has_battery: false,
  battery_capacity: 0
})

const siteSettings = ref({
  strategy_mode: 'eco',
  capacity_peak_limit_kw: 2.5,
  active_inverter_curtailment: false
})
const saveSettingsSuccess = ref(false)

const fetchSiteSettings = async () => {
  try {
    const host = window.location.hostname
    const res = await fetch(`http://${host}:8080/api/settings`)
    if (res.ok) {
      siteSettings.value = await res.json()
    }
  } catch (e) {
    console.error("Failed to fetch site settings:", e)
  }
}

const saveSiteSettings = async () => {
  try {
    const host = window.location.hostname
    const res = await fetch(`http://${host}:8080/api/settings`, {
      method: 'PUT',
      headers: {
        'Content-Type': 'application/json'
      },
      body: JSON.stringify(siteSettings.value)
    })

    if (res.ok) {
      saveSettingsSuccess.value = true
      setTimeout(() => {
        saveSettingsSuccess.value = false
      }, 3000)
    }
  } catch (e) {
    console.error("Failed to save site settings:", e)
  }
}

const fetchTemplates = async () => {
  try {
    const host = window.location.hostname
    const res = await fetch(`http://${host}:8080/api/templates`)
    templates.value = await res.json()
  } catch (e) {
    console.error("Failed to fetch templates:", e)
  }
}

const fetchDevices = async () => {
  try {
    const host = window.location.hostname
    const res = await fetch(`http://${host}:8080/api/devices`)
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
    const host = window.location.hostname
    const res = await fetch(`http://${host}:8080/api/devices`, {
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
        modbus_id: 1,
        username: '',
        password: '',
        has_grid_meter: false,
        has_battery: false,
        battery_capacity: 0
      }
      await fetchDevices()
    }
  } catch (e) {
    console.error("Failed to add device:", e)
  }
}

const editDevice = (device: Device) => {
  editingDevice.value = device
  editForm.value = {
    ...device,
    username: device.username || '',
    password: device.password || '',
    has_grid_meter: device.has_grid_meter || false,
    has_battery: device.has_battery || false,
    battery_capacity: device.battery_capacity || 0
  }
}

const closeEdit = () => {
  editingDevice.value = null
}

const updateDevice = async () => {
  if (!editingDevice.value) return

  try {
    const host = window.location.hostname
    const res = await fetch(`http://${host}:8080/api/devices/${editingDevice.value.id}`, {
      method: 'PUT',
      headers: {
        'Content-Type': 'application/json'
      },
      body: JSON.stringify(editForm.value)
    })

    if (res.ok) {
      closeEdit()
      await fetchDevices()
    }
  } catch (e) {
    console.error("Failed to update device:", e)
  }
}

const deleteDevice = async (id: number) => {
  try {
    const host = window.location.hostname
    const res = await fetch(`http://${host}:8080/api/devices/${id}`, {
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
  fetchSiteSettings()
  fetchTemplates()
  fetchDevices()
})
</script>
