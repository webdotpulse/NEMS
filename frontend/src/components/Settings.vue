<template>
  <main class="max-w-7xl mx-auto py-6 sm:px-6 lg:px-8">
    <div class="px-4 py-6 sm:px-0">

      <!-- Tabs -->
      <div class="border-b border-gray-200 dark:border-gray-700 mb-8">
        <nav class="-mb-px flex space-x-8" aria-label="Tabs">
          <button @click="activeTab = 'info'" :class="[activeTab === 'info' ? 'border-indigo-500 text-indigo-600 dark:text-indigo-400' : 'border-transparent text-gray-500 hover:text-gray-700 hover:border-gray-300 dark:text-gray-400 dark:hover:text-gray-300', 'whitespace-nowrap py-4 px-1 border-b-2 font-medium text-sm transition-colors']">
            System Info
          </button>
          <button @click="activeTab = 'strategy'" :class="[activeTab === 'strategy' ? 'border-indigo-500 text-indigo-600 dark:text-indigo-400' : 'border-transparent text-gray-500 hover:text-gray-700 hover:border-gray-300 dark:text-gray-400 dark:hover:text-gray-300', 'whitespace-nowrap py-4 px-1 border-b-2 font-medium text-sm transition-colors']">
            Strategy
          </button>
          <button @click="activeTab = 'devices'" :class="[activeTab === 'devices' ? 'border-indigo-500 text-indigo-600 dark:text-indigo-400' : 'border-transparent text-gray-500 hover:text-gray-700 hover:border-gray-300 dark:text-gray-400 dark:hover:text-gray-300', 'whitespace-nowrap py-4 px-1 border-b-2 font-medium text-sm transition-colors']">
            Devices
          </button>
        </nav>
      </div>

      <!-- TAB: SYSTEM INFO -->
      <div v-if="activeTab === 'info'">
        <!-- System Info Section -->
        <div v-if="sysInfo" class="mb-8">
          <div class="flex items-center justify-between mb-4">
            <h2 class="text-2xl font-bold leading-7 text-gray-900 dark:text-white sm:text-3xl sm:truncate">
              System Info
            </h2>
            <button @click="rebootSystem" class="inline-flex items-center px-4 py-2 border border-transparent text-sm font-medium rounded-md shadow-sm text-white bg-red-600 hover:bg-red-700 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-red-500 transition-colors">
              Reboot System
            </button>
          </div>
          <div class="bg-white dark:bg-gray-800 shadow sm:rounded-lg">
            <div class="px-4 py-5 sm:p-6 grid grid-cols-1 md:grid-cols-3 gap-4">
              <div>
                <dt class="text-sm font-medium text-gray-500 dark:text-gray-400">Hostname</dt>
                <dd class="mt-1 text-sm text-gray-900 dark:text-gray-100">{{ sysInfo.hostname }}</dd>
              </div>
              <div>
                <dt class="text-sm font-medium text-gray-500 dark:text-gray-400">IP Address</dt>
                <dd class="mt-1 text-sm text-gray-900 dark:text-gray-100">{{ sysInfo.ip }}</dd>
              </div>
              <div>
                <dt class="text-sm font-medium text-gray-500 dark:text-gray-400">Netmask</dt>
                <dd class="mt-1 text-sm text-gray-900 dark:text-gray-100">{{ sysInfo.netmask }}</dd>
              </div>
              <div>
                <dt class="text-sm font-medium text-gray-500 dark:text-gray-400">Gateway</dt>
                <dd class="mt-1 text-sm text-gray-900 dark:text-gray-100">{{ sysInfo.gateway }}</dd>
              </div>
              <div>
                <dt class="text-sm font-medium text-gray-500 dark:text-gray-400">CPU</dt>
                <dd class="mt-1 text-sm text-gray-900 dark:text-gray-100">{{ sysInfo.cpu }}</dd>
              </div>
              <div>
                <dt class="text-sm font-medium text-gray-500 dark:text-gray-400">Memory</dt>
                <dd class="mt-1 text-sm text-gray-900 dark:text-gray-100">{{ sysInfo.memory }}</dd>
              </div>
              <div>
                <dt class="text-sm font-medium text-gray-500 dark:text-gray-400">Disk</dt>
                <dd class="mt-1 text-sm text-gray-900 dark:text-gray-100">{{ sysInfo.disk }}</dd>
              </div>
            </div>
          </div>
        </div>
      </div>

      <!-- TAB: STRATEGY -->
      <div v-if="activeTab === 'strategy'">
        <!-- Site Optimization Section -->
        <div class="mb-8">
          <h2 class="text-2xl font-bold leading-7 text-gray-900 dark:text-white sm:text-3xl sm:truncate mb-4">
            Site Optimization
          </h2>
          <form @submit.prevent="saveSiteSettings" class="space-y-6">
            <!-- Grid Connection Card -->
            <div class="bg-white dark:bg-gray-800 shadow sm:rounded-lg mb-6">
              <div class="px-4 py-5 sm:p-6">
                <div class="mb-4 border-b border-gray-200 dark:border-gray-700 pb-4">
                  <h3 class="text-lg font-medium leading-6 text-gray-900 dark:text-gray-100">
                    Grid Connection
                  </h3>
                </div>
                <div class="grid grid-cols-1 gap-y-6 gap-x-4 sm:grid-cols-6">
                  <div class="sm:col-span-3">
                    <label for="grid_nominal_current_a" class="block text-sm font-medium text-gray-700 dark:text-gray-300">Nominal Current (A)</label>
                    <div class="mt-1">
                      <input type="number" step="0.1" id="grid_nominal_current_a" v-model="siteSettings.grid_nominal_current_a" class="shadow-sm focus:ring-indigo-500 focus:border-indigo-500 block w-full sm:text-sm border-gray-300 rounded-md dark:bg-gray-700 dark:border-gray-600 dark:text-white" />
                    </div>
                  </div>

                  <div class="sm:col-span-3">
                    <label for="grid_system" class="block text-sm font-medium text-gray-700 dark:text-gray-300">Grid System</label>
                    <div class="mt-1">
                      <select id="grid_system" v-model="siteSettings.grid_system"
                              class="shadow-sm focus:ring-indigo-500 focus:border-indigo-500 block w-full sm:text-sm border-gray-300 rounded-md dark:bg-gray-700 dark:border-gray-600 dark:text-white">
                        <option value="single_phase_230v">Single Phase 230V</option>
                        <option value="three_phase_400v">Three Phase 400V</option>
                        <option value="three_phase_230v_delta">Three Phase 230V Delta</option>
                      </select>
                    </div>
                  </div>

                  <div class="sm:col-span-3">
                    <label for="allowed_grid_import_kw" class="block text-sm font-medium text-gray-700 dark:text-gray-300">Allowed Grid Import (kW)</label>
                    <div class="mt-1">
                      <input type="number" step="0.1" id="allowed_grid_import_kw" v-model="siteSettings.allowed_grid_import_kw" class="shadow-sm focus:ring-indigo-500 focus:border-indigo-500 block w-full sm:text-sm border-gray-300 rounded-md dark:bg-gray-700 dark:border-gray-600 dark:text-white" />
                    </div>
                  </div>

                  <div class="sm:col-span-3">
                    <label for="allowed_grid_export_kw" class="block text-sm font-medium text-gray-700 dark:text-gray-300">Allowed Grid Export (kW)</label>
                    <div class="mt-1">
                      <input type="number" step="0.1" id="allowed_grid_export_kw" v-model="siteSettings.allowed_grid_export_kw" class="shadow-sm focus:ring-indigo-500 focus:border-indigo-500 block w-full sm:text-sm border-gray-300 rounded-md dark:bg-gray-700 dark:border-gray-600 dark:text-white" />
                    </div>
                  </div>
                </div>
              </div>
            </div>

            <!-- Strategy Settings Card -->
            <div class="bg-white dark:bg-gray-800 shadow sm:rounded-lg mb-6">
              <div class="px-4 py-5 sm:p-6">
                <h3 class="text-lg font-medium leading-6 text-gray-900 dark:text-gray-100 mb-4">Strategy</h3>
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

                    <div v-if="siteSettings.strategy_mode === 'flanders'" class="sm:col-span-3">
                      <label for="peak_shaving_buffer_w" class="block text-sm font-medium text-gray-700 dark:text-gray-300">Peak Shaving Buffer (W)</label>
                      <div class="mt-1">
                        <input type="number" step="1" min="0" id="peak_shaving_buffer_w" v-model="siteSettings.peak_shaving_buffer_w"
                               class="shadow-sm focus:ring-indigo-500 focus:border-indigo-500 block w-full sm:text-sm border-gray-300 rounded-md dark:bg-gray-700 dark:border-gray-600 dark:text-white">
                      </div>
                    </div>

                    <div v-if="siteSettings.strategy_mode === 'flanders'" class="sm:col-span-3">
                      <label for="peak_shaving_rampup_w" class="block text-sm font-medium text-gray-700 dark:text-gray-300">Ramp-up Hysteresis (W)</label>
                      <div class="mt-1">
                        <input type="number" step="1" min="0" id="peak_shaving_rampup_w" v-model="siteSettings.peak_shaving_rampup_w"
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
              </div>
            </div>

            <!-- Appliance Control Card -->
            <div class="bg-white dark:bg-gray-800 shadow sm:rounded-lg mb-6">
              <div class="px-4 py-5 sm:p-6">
                <div class="mb-4 border-b border-gray-200 dark:border-gray-700 pb-4">
                  <h3 class="text-lg font-medium leading-6 text-gray-900 dark:text-gray-100">
                    Appliance Control
                  </h3>
                  <p class="mt-1 text-sm text-gray-500 dark:text-gray-400">
                    Configure when to turn on Smart Plugs or Generic Relays to sink excess solar power. Set to 0 to disable.
                  </p>
                </div>

                <div class="grid grid-cols-1 gap-y-6 gap-x-4 sm:grid-cols-6">
                  <div class="sm:col-span-6">
                    <label for="appliance_turn_on_excess_w" class="block text-sm font-medium text-gray-700 dark:text-gray-300">Turn on Relays above Solar Excess (W)</label>
                    <div class="mt-1">
                      <input type="number" step="1" min="0" id="appliance_turn_on_excess_w" v-model="siteSettings.appliance_turn_on_excess_w" class="shadow-sm focus:ring-indigo-500 focus:border-indigo-500 block w-full sm:text-sm border-gray-300 rounded-md dark:bg-gray-700 dark:border-gray-600 dark:text-white" />
                    </div>
                  </div>
                </div>
              </div>
            </div>

            <!-- Dynamic Tariffs Card -->
            <div class="bg-white dark:bg-gray-800 shadow sm:rounded-lg mb-6">
              <div class="px-4 py-5 sm:p-6">
                <div class="mb-4 border-b border-gray-200 dark:border-gray-700 pb-4">
                  <h3 class="text-lg font-medium leading-6 text-gray-900 dark:text-gray-100">
                    Dynamic Tariffs
                  </h3>
                  <p class="mt-1 text-sm text-gray-500 dark:text-gray-400">
                    Configure rules for dynamic energy pricing.
                  </p>
                </div>

                <div class="grid grid-cols-1 gap-y-6 gap-x-4 sm:grid-cols-6">
                  <div class="sm:col-span-6">
                    <label for="force_charge_below_euro" class="block text-sm font-medium text-gray-700 dark:text-gray-300">Force Charge Battery if price drops below (€/kWh)</label>
                    <div class="mt-1">
                      <input type="number" step="0.01" id="force_charge_below_euro" v-model="siteSettings.force_charge_below_euro" class="shadow-sm focus:ring-indigo-500 focus:border-indigo-500 block w-full sm:text-sm border-gray-300 rounded-md dark:bg-gray-700 dark:border-gray-600 dark:text-white" />
                    </div>
                  </div>

                  <div class="sm:col-span-6">
                    <label for="force_discharge_above_euro" class="block text-sm font-medium text-gray-700 dark:text-gray-300">Force Discharge Battery to Grid if price rises above (€/kWh)</label>
                    <div class="mt-1">
                      <input type="number" step="0.01" id="force_discharge_above_euro" v-model="siteSettings.force_discharge_above_euro" class="shadow-sm focus:ring-indigo-500 focus:border-indigo-500 block w-full sm:text-sm border-gray-300 rounded-md dark:bg-gray-700 dark:border-gray-600 dark:text-white" />
                    </div>
                  </div>

                  <div class="sm:col-span-6">
                    <label for="smart_ev_cheapest_hours" class="block text-sm font-medium text-gray-700 dark:text-gray-300">Smart EV Charging: Charge during the cheapest hours of the day</label>
                    <div class="mt-1">
                      <input type="number" step="1" min="0" max="24" id="smart_ev_cheapest_hours" v-model="siteSettings.smart_ev_cheapest_hours" class="shadow-sm focus:ring-indigo-500 focus:border-indigo-500 block w-full sm:text-sm border-gray-300 rounded-md dark:bg-gray-700 dark:border-gray-600 dark:text-white" />
                    </div>
                  </div>
                </div>

              </div>
            </div>

            <!-- Grid Connection Card -->
            <div class="bg-white dark:bg-gray-800 shadow sm:rounded-lg mb-6">
              <div class="px-4 py-5 sm:p-6">
                <div class="mb-4 border-b border-gray-200 dark:border-gray-700 pb-4">
                  <h3 class="text-lg font-medium leading-6 text-gray-900 dark:text-gray-100">
                    Grid Connection
                  </h3>
                </div>
                <div class="grid grid-cols-1 gap-y-6 gap-x-4 sm:grid-cols-6">
                  <div class="sm:col-span-3">
                    <label for="grid_nominal_current_a" class="block text-sm font-medium text-gray-700 dark:text-gray-300">Nominal Current (A)</label>
                    <div class="mt-1">
                      <input type="number" step="0.1" id="grid_nominal_current_a" v-model="siteSettings.grid_nominal_current_a" class="shadow-sm focus:ring-indigo-500 focus:border-indigo-500 block w-full sm:text-sm border-gray-300 rounded-md dark:bg-gray-700 dark:border-gray-600 dark:text-white" />
                    </div>
                  </div>

                  <div class="sm:col-span-3">
                    <label for="grid_system" class="block text-sm font-medium text-gray-700 dark:text-gray-300">Grid System</label>
                    <div class="mt-1">
                      <select id="grid_system" v-model="siteSettings.grid_system"
                              class="shadow-sm focus:ring-indigo-500 focus:border-indigo-500 block w-full sm:text-sm border-gray-300 rounded-md dark:bg-gray-700 dark:border-gray-600 dark:text-white">
                        <option value="single_phase_230v">Single Phase 230V</option>
                        <option value="three_phase_400v">Three Phase 400V</option>
                        <option value="three_phase_230v_delta">Three Phase 230V Delta</option>
                      </select>
                    </div>
                  </div>

                  <div class="sm:col-span-3">
                    <label for="allowed_grid_import_kw" class="block text-sm font-medium text-gray-700 dark:text-gray-300">Allowed Grid Import (kW)</label>
                    <div class="mt-1">
                      <input type="number" step="0.1" id="allowed_grid_import_kw" v-model="siteSettings.allowed_grid_import_kw" class="shadow-sm focus:ring-indigo-500 focus:border-indigo-500 block w-full sm:text-sm border-gray-300 rounded-md dark:bg-gray-700 dark:border-gray-600 dark:text-white" />
                    </div>
                  </div>

                  <div class="sm:col-span-3">
                    <label for="allowed_grid_export_kw" class="block text-sm font-medium text-gray-700 dark:text-gray-300">Allowed Grid Export (kW)</label>
                    <div class="mt-1">
                      <input type="number" step="0.1" id="allowed_grid_export_kw" v-model="siteSettings.allowed_grid_export_kw" class="shadow-sm focus:ring-indigo-500 focus:border-indigo-500 block w-full sm:text-sm border-gray-300 rounded-md dark:bg-gray-700 dark:border-gray-600 dark:text-white" />
                    </div>
                  </div>
                </div>
              </div>
            </div>

            <!-- Save Actions Card -->
            <div class="bg-white dark:bg-gray-800 shadow sm:rounded-lg">
              <div class="px-4 py-5 sm:p-6 flex items-center justify-between">
                <div v-if="saveSettingsSuccess" class="text-sm font-medium text-green-600 dark:text-green-400 transition-opacity">
                  Strategy saved successfully!
                </div>
                <div v-else></div>
                <button type="submit"
                        class="inline-flex justify-center py-2 px-4 border border-transparent shadow-sm text-sm font-medium rounded-md text-white bg-indigo-600 hover:bg-indigo-700 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-indigo-500">
                  Save Strategy
                </button>
              </div>
            </div>
          </form>
        </div>
      </div>

      <!-- TAB: DEVICES -->
      <div v-if="activeTab === 'devices'">
        <!-- Current Devices Section -->
        <div class="mb-8">
          <h2 class="text-2xl font-bold leading-7 text-gray-900 dark:text-white sm:text-3xl sm:truncate mb-6">
            Configured Devices
          </h2>

          <div v-if="devices.length === 0" class="bg-white dark:bg-gray-800 shadow sm:rounded-lg mb-6">
            <div class="px-4 py-5 sm:p-6 text-center text-gray-500 dark:text-gray-400">
              No devices configured yet. Add one below.
            </div>
          </div>

          <div v-else>
            <!-- Categories Iterate -->
            <div v-for="(catDevices, catName) in categorizedDevices" :key="catName">
              <div v-if="catDevices.length > 0" class="mb-6">
                <h3 class="text-lg font-semibold text-gray-700 dark:text-gray-300 mb-3 border-b border-gray-200 dark:border-gray-700 pb-2">
                  {{ catName }}
                </h3>
                <div class="grid grid-cols-1 gap-4 sm:grid-cols-2 lg:grid-cols-3">
                  <div v-for="device in catDevices" :key="device.id" class="bg-white dark:bg-gray-800 overflow-hidden shadow rounded-lg border border-gray-100 dark:border-gray-700">
                    <div class="px-4 py-5 sm:p-6">
                      <h4 class="text-lg leading-6 font-medium text-gray-900 dark:text-white flex items-center">
                        {{ device.name }}
                        <span v-if="device.status === 'online'" class="ml-2 flex w-3 h-3 bg-green-500 rounded-full" title="Online"></span>
                        <span v-else-if="device.status === 'error'" class="ml-2 flex w-3 h-3 bg-red-500 rounded-full" title="Error"></span>
                        <span v-else class="ml-2 flex w-3 h-3 bg-gray-400 rounded-full" title="Offline"></span>
                      </h4>
                      <div class="mt-2 max-w-xl text-sm text-gray-500 dark:text-gray-400">
                        <p><strong>Template:</strong> {{ getTemplateName(device.template) }}</p>
                        <p v-if="device.host"><strong>Host:</strong> {{ device.host }}:{{ device.port }}</p>
                        <p v-if="device.modbus_id"><strong>Modbus ID:</strong> {{ device.modbus_id }}</p>
                      </div>
                      <div class="mt-3 text-sm flex space-x-4 pt-2 border-t border-gray-100 dark:border-gray-700">
                        <button @click="editDevice(device)" class="font-medium text-indigo-600 hover:text-indigo-500 dark:text-indigo-400 dark:hover:text-indigo-300 transition-colors">
                          Edit
                        </button>
                        <button @click="deleteDevice(device.id)" class="font-medium text-red-600 hover:text-red-500 dark:text-red-400 dark:hover:text-red-300 transition-colors">
                          Delete
                        </button>
                      </div>
                    </div>
                  </div>
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
          <div class="bg-white dark:bg-gray-800 shadow sm:rounded-lg border border-gray-100 dark:border-gray-700">
            <div class="px-4 py-5 sm:p-6">
              <form @submit.prevent="addDevice" class="space-y-6">
                <div class="grid grid-cols-1 gap-y-6 gap-x-4 sm:grid-cols-6">

                  <!-- Category Selection Cards -->
                  <div class="sm:col-span-6" v-if="!selectedCategory">
                    <label class="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-4">Select Device Category</label>
                    <div class="grid grid-cols-2 sm:grid-cols-3 md:grid-cols-5 gap-4">
                      <div v-for="cat in deviceCategories" :key="cat.id"
                           @click="selectCategory(cat.id)"
                           class="cursor-pointer border-2 rounded-lg p-4 flex flex-col items-center justify-center transition-all hover:border-indigo-500 hover:bg-indigo-50 dark:hover:bg-indigo-900/20"
                           :class="selectedCategory === cat.id ? 'border-indigo-600 bg-indigo-50 dark:bg-indigo-900/30' : 'border-gray-200 dark:border-gray-700 bg-white dark:bg-gray-800'">
                        <i :class="cat.icon + ' text-3xl mb-2 ' + cat.color"></i>
                        <span class="text-sm font-medium text-gray-900 dark:text-white">{{ cat.name }}</span>
                      </div>
                    </div>
                  </div>

                  <!-- Reset Category Button & Category Info -->
                  <div class="sm:col-span-6 flex items-center justify-between" v-if="selectedCategory">
                    <div class="flex items-center">
                      <i :class="currentCategoryObj?.icon + ' text-xl mr-2 ' + currentCategoryObj?.color"></i>
                      <span class="text-md font-medium text-gray-900 dark:text-white">Adding: {{ currentCategoryObj?.name }}</span>
                    </div>
                    <button type="button" @click="resetCategory" class="text-sm text-indigo-600 hover:text-indigo-500 dark:text-indigo-400 dark:hover:text-indigo-300">
                      &larr; Change Category
                    </button>
                  </div>

                  <template v-if="selectedCategory">
                    <div class="sm:col-span-3">
                      <label for="name" class="block text-sm font-medium text-gray-700 dark:text-gray-300">Device Name</label>
                      <div class="mt-1">
                        <input type="text" id="name" v-model="form.name" required
                               class="shadow-sm focus:ring-indigo-500 focus:border-indigo-500 block w-full sm:text-sm border-gray-300 rounded-md dark:bg-gray-700 dark:border-gray-600 dark:text-white transition-colors">
                      </div>
                    </div>

                    <div class="sm:col-span-3">
                      <label for="template" class="block text-sm font-medium text-gray-700 dark:text-gray-300">Device Template</label>
                      <div class="mt-1">
                        <select id="template" v-model="form.template" required
                                class="shadow-sm focus:ring-indigo-500 focus:border-indigo-500 block w-full sm:text-sm border-gray-300 rounded-md dark:bg-gray-700 dark:border-gray-600 dark:text-white transition-colors">
                          <option disabled value="">Please select one</option>
                          <option v-for="t in filteredTemplates" :key="t.id" :value="t.id">
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
                  <template v-else-if="templates.find(t => t.id === form.template)?.type === 'serial'">
                    <P1SerialTemplate :device="form" />
                  </template>
                  <template v-else-if="templates.find(t => t.id === form.template)?.type === 'network'">
                    <P1NetworkTemplate :device="form" />
                  </template>

                  <!-- EV Charger Mode Selection -->
                  <div v-if="isCharger(form.template)" class="sm:col-span-3">
                    <label for="charge_mode" class="block text-sm font-medium text-gray-700 dark:text-gray-300">Charge Mode</label>
                    <div class="mt-1">
                      <select id="charge_mode" v-model="form.charge_mode"
                              class="shadow-sm focus:ring-indigo-500 focus:border-indigo-500 block w-full sm:text-sm border-gray-300 rounded-md dark:bg-gray-700 dark:border-gray-600 dark:text-white transition-colors">
                        <option value="eco">Eco / Smart</option>
                        <option value="pv_only">PV Only (Solar)</option>
                        <option value="now">Fast (Max Power)</option>
                        <option value="off">Off</option>
                      </select>
                    </div>
                  </div>

                  <!-- Conditional Fields for Huawei Hybrid Inverter Combo -->
                  <template v-if="form.template === 'huawei_inverter'">
                    <div class="sm:col-span-6 border-t border-gray-200 dark:border-gray-700 pt-4 mt-2">
                      <h4 class="text-sm font-medium text-gray-900 dark:text-gray-100 mb-4">Huawei Inverter Features</h4>

                      <div class="flex items-center mb-4">
                        <input id="has_grid_meter" type="checkbox" v-model="form.has_grid_meter"
                               class="h-4 w-4 text-indigo-600 focus:ring-indigo-500 border-gray-300 rounded dark:bg-gray-700 dark:border-gray-600 transition-colors" />
                        <label for="has_grid_meter" class="ml-2 block text-sm text-gray-900 dark:text-gray-300">
                          Grid Meter Connected?
                        </label>
                      </div>

                      <div class="flex items-center mb-4">
                        <input id="has_battery" type="checkbox" v-model="form.has_battery"
                               class="h-4 w-4 text-indigo-600 focus:ring-indigo-500 border-gray-300 rounded dark:bg-gray-700 dark:border-gray-600 transition-colors" />
                        <label for="has_battery" class="ml-2 block text-sm text-gray-900 dark:text-gray-300">
                          Battery Connected?
                        </label>
                      </div>

                      <div v-if="form.has_battery" class="sm:col-span-3 ml-6">
                        <label for="battery_capacity" class="block text-sm font-medium text-gray-700 dark:text-gray-300">Battery Capacity (kWh)</label>
                        <div class="mt-1">
                          <input type="number" step="0.1" id="battery_capacity" v-model="form.battery_capacity"
                                 class="shadow-sm focus:ring-indigo-500 focus:border-indigo-500 block w-full sm:text-sm border-gray-300 rounded-md dark:bg-gray-700 dark:border-gray-600 dark:text-white transition-colors">
                        </div>
                      </div>
                    </div>
                  </template>

                  </template>
                </div>

                <div class="pt-5" v-if="selectedCategory">
                  <div class="flex justify-end">
                    <button type="submit"
                            class="ml-3 inline-flex justify-center py-2 px-4 border border-transparent shadow-sm text-sm font-medium rounded-md text-white bg-indigo-600 hover:bg-indigo-700 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-indigo-500 transition-colors">
                      Save Device
                    </button>
                  </div>
                </div>
              </form>
            </div>
          </div>
        </div>
      </div>

      <!-- Edit Device Modal -->
      <div v-if="editingDevice" class="fixed inset-0 z-50 flex items-center justify-center bg-black bg-opacity-50 transition-opacity">
        <div class="bg-white dark:bg-gray-800 rounded-lg shadow-xl w-11/12 max-w-2xl p-6 relative">
          <button @click="closeEdit" class="absolute top-4 right-4 text-gray-500 hover:text-gray-700 dark:text-gray-400 dark:hover:text-gray-200 transition-colors">
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
                         class="shadow-sm focus:ring-indigo-500 focus:border-indigo-500 block w-full sm:text-sm border-gray-300 rounded-md dark:bg-gray-700 dark:border-gray-600 dark:text-white transition-colors">
                </div>
              </div>

              <div class="sm:col-span-3">
                <label for="edit_template" class="block text-sm font-medium text-gray-700 dark:text-gray-300">Device Template</label>
                <div class="mt-1">
                  <select id="edit_template" v-model="editForm.template" required
                          class="shadow-sm focus:ring-indigo-500 focus:border-indigo-500 block w-full sm:text-sm border-gray-300 rounded-md dark:bg-gray-700 dark:border-gray-600 dark:text-white transition-colors">
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
              <template v-else-if="templates.find(t => t.id === editForm.template)?.type === 'serial'">
                <P1SerialTemplate :device="editForm" />
              </template>
              <template v-else-if="templates.find(t => t.id === editForm.template)?.type === 'network'">
                <P1NetworkTemplate :device="editForm" />
              </template>

              <!-- EV Charger Mode Selection -->
              <div v-if="isCharger(editForm.template)" class="sm:col-span-3">
                <label for="edit_charge_mode" class="block text-sm font-medium text-gray-700 dark:text-gray-300">Charge Mode</label>
                <div class="mt-1">
                  <select id="edit_charge_mode" v-model="editForm.charge_mode"
                          class="shadow-sm focus:ring-indigo-500 focus:border-indigo-500 block w-full sm:text-sm border-gray-300 rounded-md dark:bg-gray-700 dark:border-gray-600 dark:text-white transition-colors">
                    <option value="eco">Eco / Smart</option>
                    <option value="pv_only">PV Only (Solar)</option>
                    <option value="now">Fast (Max Power)</option>
                    <option value="off">Off</option>
                  </select>
                </div>
              </div>

              <!-- Conditional Fields for Huawei Hybrid Inverter Combo -->
              <template v-if="editForm.template === 'huawei_inverter'">
                <div class="sm:col-span-6 border-t border-gray-200 dark:border-gray-700 pt-4 mt-2">
                  <h4 class="text-sm font-medium text-gray-900 dark:text-gray-100 mb-4">Huawei Inverter Features</h4>

                  <div class="flex items-center mb-4">
                    <input id="edit_has_grid_meter" type="checkbox" v-model="editForm.has_grid_meter"
                           class="h-4 w-4 text-indigo-600 focus:ring-indigo-500 border-gray-300 rounded dark:bg-gray-700 dark:border-gray-600 transition-colors" />
                    <label for="edit_has_grid_meter" class="ml-2 block text-sm text-gray-900 dark:text-gray-300">
                      Grid Meter Connected?
                    </label>
                  </div>

                  <div class="flex items-center mb-4">
                    <input id="edit_has_battery" type="checkbox" v-model="editForm.has_battery"
                           class="h-4 w-4 text-indigo-600 focus:ring-indigo-500 border-gray-300 rounded dark:bg-gray-700 dark:border-gray-600 transition-colors" />
                    <label for="edit_has_battery" class="ml-2 block text-sm text-gray-900 dark:text-gray-300">
                      Battery Connected?
                    </label>
                  </div>

                  <div v-if="editForm.has_battery" class="sm:col-span-3 ml-6">
                    <label for="edit_battery_capacity" class="block text-sm font-medium text-gray-700 dark:text-gray-300">Battery Capacity (kWh)</label>
                    <div class="mt-1">
                      <input type="number" step="0.1" id="edit_battery_capacity" v-model="editForm.battery_capacity"
                             class="shadow-sm focus:ring-indigo-500 focus:border-indigo-500 block w-full sm:text-sm border-gray-300 rounded-md dark:bg-gray-700 dark:border-gray-600 dark:text-white transition-colors">
                    </div>
                  </div>
                </div>
              </template>
            </div>

            <div class="pt-5">
              <div class="flex justify-end">
                <button type="button" @click="closeEdit" class="bg-white py-2 px-4 border border-gray-300 rounded-md shadow-sm text-sm font-medium text-gray-700 hover:bg-gray-50 dark:bg-gray-700 dark:border-gray-600 dark:text-gray-200 dark:hover:bg-gray-600 transition-colors">
                  Cancel
                </button>
                <button type="submit"
                        class="ml-3 inline-flex justify-center py-2 px-4 border border-transparent shadow-sm text-sm font-medium rounded-md text-white bg-indigo-600 hover:bg-indigo-700 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-indigo-500 transition-colors">
                  Save Changes
                </button>
              </div>
            </div>
          </form>
        </div>
      </div>

    </div>
  </main>
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { getApiBase } from '../api'
import type { Device, SiteSettings, Template, SystemInfo } from '../types'

import ModbusTemplate from './templates/ModbusTemplate.vue'
import CloudTemplate from './templates/CloudTemplate.vue'
import RestTemplate from './templates/RestTemplate.vue'
import CloudRestTemplate from './templates/CloudRestTemplate.vue'
import P1SerialTemplate from './templates/P1SerialTemplate.vue'
import P1NetworkTemplate from './templates/P1NetworkTemplate.vue'

const activeTab = ref('info')

const sysInfo = ref<SystemInfo | null>(null)

const selectedCategory = ref<string | null>(null)
const deviceCategories = [
  { id: 'inverter', name: 'Inverter / Solar', icon: 'bx bx-sun', color: 'text-yellow-500' },
  { id: 'charger', name: 'EV Charger', icon: 'bx bx-car', color: 'text-purple-500' },
  { id: 'meter', name: 'Smart Meter', icon: 'bx bx-tachometer', color: 'text-blue-500' },
  { id: 'battery', name: 'Battery', icon: 'bx bx-battery', color: 'text-green-500' },
  { id: 'relay', name: 'Relay / Switch', icon: 'bx bx-toggle-right', color: 'text-orange-500' },
  { id: 'demo', name: 'Demo Device', icon: 'bx bx-code-block', color: 'text-gray-500' }
]
const currentCategoryObj = computed(() => deviceCategories.find(c => c.id === selectedCategory.value) || null)
const filteredTemplates = computed(() => {
  if (!selectedCategory.value) return templates.value
  return templates.value.filter(t => t.category === selectedCategory.value)
})
const selectCategory = (cat: string) => {
  selectedCategory.value = cat
}
const resetCategory = () => {
  selectedCategory.value = null
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
  battery_capacity: 0,
  charge_mode: 'eco'
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
  battery_capacity: 0,
  charge_mode: 'eco'
})

const siteSettings = ref<SiteSettings>({
  strategy_mode: 'eco',
  capacity_peak_limit_kw: 2.5,
  active_inverter_curtailment: false,
  force_charge_below_euro: 0.0,
  force_discharge_above_euro: 999.0,
  smart_ev_cheapest_hours: 0,
  grid_nominal_current_a: 25.0,
  grid_system: 'single_phase_230v',
  allowed_grid_import_kw: 0.0,
  allowed_grid_export_kw: 0.0,
  appliance_turn_on_excess_w: 0.0,
  peak_shaving_buffer_w: 200.0,
  peak_shaving_rampup_w: 500.0
})
const saveSettingsSuccess = ref(false)

const getDeviceCategory = (templateId: string) => {
  if (!templateId) return 'Other'
  const id = templateId.toLowerCase()
  if (id.includes('meter') || id.includes('homewizard') || id.includes('p1')) return 'Meters & Grid'
  if (id.includes('inverter') || id.includes('solar')) return 'Inverters & Solar'
  if (id.includes('battery')) return 'Batteries & Storage'
  if (id.includes('charger') || id.includes('wallbox') || id.includes('easee') || id.includes('alfen') || id.includes('raedian') || id.includes('peblar') || id.includes('phoenix') || id.includes('bender')) return 'EV Chargers'
  if (id.includes('relay') || id.includes('plug') || id.includes('appliance')) return 'Relays & Appliances'
  return 'Other'
}

const categorizedDevices = computed(() => {
  const categories: Record<string, Device[]> = {
    'Meters & Grid': [],
    'Inverters & Solar': [],
    'Batteries & Storage': [],
    'EV Chargers': [],
    'Relays & Appliances': [],
    'Other': []
  }

  for (const device of devices.value) {
    const cat = getDeviceCategory(device.template)
    if (categories[cat]) {
      categories[cat].push(device)
    } else {
      categories['Other'].push(device)
    }
  }

  return categories
})

const fetchSiteSettings = async () => {
  try {
    const res = await fetch(`${getApiBase()}/api/settings`)
    if (res.ok) {
      siteSettings.value = await res.json()
    }
  } catch (e) {
    console.error("Failed to fetch site settings:", e)
  }
}

const saveSiteSettings = async () => {
  try {
    const res = await fetch(`${getApiBase()}/api/settings`, {
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
    const res = await fetch(`${getApiBase()}/api/templates`)
    templates.value = await res.json()
  } catch (e) {
    console.error("Failed to fetch templates:", e)
  }
}

const fetchDevices = async () => {
  try {
    const res = await fetch(`${getApiBase()}/api/devices`)
    devices.value = await res.json() || []
  } catch (e) {
    console.error("Failed to fetch devices:", e)
  }
}

const isCharger = (templateId: string) => {
  return templateId.includes('charger') || templateId.includes('wallbox') || templateId.includes('easee') || templateId.includes('alfen') || templateId.includes('raedian') || templateId.includes('peblar') || templateId.includes('phoenix') || templateId.includes('bender')
}

const getTemplateName = (id: string) => {
  const t = templates.value.find(t => t.id === id)
  return t ? t.name : id
}

const addDevice = async () => {
  try {
    const res = await fetch(`${getApiBase()}/api/devices`, {
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
        battery_capacity: 0,
        charge_mode: 'eco'
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
    battery_capacity: device.battery_capacity || 0,
    charge_mode: device.charge_mode || 'eco'
  }
}

const closeEdit = () => {
  editingDevice.value = null
}

const updateDevice = async () => {
  if (!editingDevice.value) return

  try {
    const res = await fetch(`${getApiBase()}/api/devices/${editingDevice.value.id}`, {
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
    const res = await fetch(`${getApiBase()}/api/devices/${id}`, {
      method: 'DELETE'
    })

    if (res.ok) {
      await fetchDevices()
    }
  } catch (e) {
    console.error("Failed to delete device:", e)
  }
}

const rebootSystem = async () => {
  if (confirm("Are you sure you want to reboot the system? This will interrupt power management.")) {
    try {
      await fetch(`${getApiBase()}/api/system/reboot`, { method: 'POST' })
      alert("System is rebooting. Please wait a moment and refresh the page.")
    } catch (e) {
      console.error("Failed to reboot system:", e)
      alert("Failed to send reboot command.")
    }
  }
}

const fetchSystemInfo = async () => {
  try {
    const res = await fetch(`${getApiBase()}/api/system/info`)
    if (res.ok) {
      sysInfo.value = await res.json()
    }
  } catch (e) {
    console.error("Failed to fetch system info:", e)
  }
}

onMounted(() => {
  fetchSystemInfo()
  fetchSiteSettings()
  fetchTemplates()
  fetchDevices()
})
</script>