<template>
  <main class="max-w-7xl mx-auto py-6 sm:px-6 lg:px-8">
    <div class="px-4 py-6 sm:px-0">

      <!-- Tabs -->
      <div class="border-b border-gray-200 dark:border-gray-700 mb-8">
        <nav class="-mb-px flex space-x-8" aria-label="Tabs">
          <button @click="activeTab = 'strategy'" :class="[activeTab === 'strategy' ? 'border-indigo-500 text-indigo-600 dark:text-indigo-400' : 'border-transparent text-gray-500 hover:text-gray-700 hover:border-gray-300 dark:text-gray-400 dark:hover:text-gray-300', 'whitespace-nowrap py-4 px-1 border-b-2 font-medium text-sm transition-colors']">
            Strategy
          </button>
          <button @click="activeTab = 'devices'" :class="[activeTab === 'devices' ? 'border-indigo-500 text-indigo-600 dark:text-indigo-400' : 'border-transparent text-gray-500 hover:text-gray-700 hover:border-gray-300 dark:text-gray-400 dark:hover:text-gray-300', 'whitespace-nowrap py-4 px-1 border-b-2 font-medium text-sm transition-colors']">
            Devices
          </button>
          <button @click="activeTab = 'info'" :class="[activeTab === 'info' ? 'border-indigo-500 text-indigo-600 dark:text-indigo-400' : 'border-transparent text-gray-500 hover:text-gray-700 hover:border-gray-300 dark:text-gray-400 dark:hover:text-gray-300', 'whitespace-nowrap py-4 px-1 border-b-2 font-medium text-sm transition-colors']">
            System Info
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
            <div class="flex gap-4">
              <button @click="resetDatabase" class="inline-flex items-center px-4 py-2 border border-red-500 text-sm font-medium rounded-md shadow-sm text-red-500 bg-transparent hover:bg-red-50 dark:hover:bg-red-900/20 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-red-500 transition-colors">
                Reset Database
              </button>
              <button @click="rebootSystem" class="inline-flex items-center px-4 py-2 border border-transparent text-sm font-medium rounded-md shadow-sm text-white bg-red-600 hover:bg-red-700 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-red-500 transition-colors">
                Reboot System
              </button>
            </div>
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
                <dt class="text-sm font-medium text-gray-500 dark:text-gray-400">Memory</dt>
                <dd class="mt-1 text-sm text-gray-900 dark:text-gray-100">{{ sysInfo.memory }}</dd>
              </div>
              <div>
                <dt class="text-sm font-medium text-gray-500 dark:text-gray-400">Disk</dt>
                <dd class="mt-1 text-sm text-gray-900 dark:text-gray-100">{{ sysInfo.disk }}</dd>
              </div>
              <div>
                <dt class="text-sm font-medium text-gray-500 dark:text-gray-400">Build Number</dt>
                <dd class="mt-1 text-sm text-gray-900 dark:text-gray-100">{{ sysInfo.build }}</dd>
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
                      <input type="number" step="0.1" id="grid_nominal_current_a" v-model="siteSettings.grid_nominal_current_a" class="shadow-sm focus:ring-indigo-500 focus:border-indigo-500 block w-full  border-gray-300 rounded-md bg-gray-50 hover:bg-white dark:bg-gray-700 dark:hover:bg-gray-600 dark:border-gray-600 dark:text-white transition-all duration-200" />
                    </div>
                  </div>

                  <div class="sm:col-span-3">
                    <label for="grid_system" class="block text-sm font-medium text-gray-700 dark:text-gray-300">Grid System</label>
                    <div class="mt-1">
                      <select id="grid_system" v-model="siteSettings.grid_system"
                              class="shadow-sm focus:ring-indigo-500 focus:border-indigo-500 block w-full  border-gray-300 rounded-md bg-gray-50 hover:bg-white dark:bg-gray-700 dark:hover:bg-gray-600 dark:border-gray-600 dark:text-white transition-all duration-200">
                        <option value="single_phase_230v">Single Phase 230V</option>
                        <option value="three_phase_400v">Three Phase 400V</option>
                        <option value="three_phase_230v_delta">Three Phase 230V Delta</option>
                      </select>
                    </div>
                  </div>

                  <div class="sm:col-span-3">
                    <label for="timezone" class="block text-sm font-medium text-gray-700 dark:text-gray-300">Time Zone</label>
                    <div class="mt-1">
                      <select id="timezone" v-model="siteSettings.timezone"
                              class="shadow-sm focus:ring-indigo-500 focus:border-indigo-500 block w-full  border-gray-300 rounded-md bg-gray-50 hover:bg-white dark:bg-gray-700 dark:hover:bg-gray-600 dark:border-gray-600 dark:text-white transition-all duration-200">
                        <option value="UTC">UTC</option>
                        <option value="Europe/Brussels">Europe/Brussels</option>
                        <option value="Europe/London">Europe/London</option>
                        <option value="America/New_York">America/New_York</option>
                      </select>
                    </div>
                  </div>

                  <div class="sm:col-span-3">
                    <label for="latitude" class="block text-sm font-medium text-gray-700 dark:text-gray-300">Latitude</label>
                    <div class="mt-1">
                      <input type="number" step="0.0001" id="latitude" v-model="siteSettings.latitude" class="shadow-sm focus:ring-indigo-500 focus:border-indigo-500 block w-full  border-gray-300 rounded-md bg-gray-50 hover:bg-white dark:bg-gray-700 dark:hover:bg-gray-600 dark:border-gray-600 dark:text-white transition-all duration-200" />
                    </div>
                  </div>

                  <div class="sm:col-span-3">
                    <label for="longitude" class="block text-sm font-medium text-gray-700 dark:text-gray-300">Longitude</label>
                    <div class="mt-1">
                      <input type="number" step="0.0001" id="longitude" v-model="siteSettings.longitude" class="shadow-sm focus:ring-indigo-500 focus:border-indigo-500 block w-full  border-gray-300 rounded-md bg-gray-50 hover:bg-white dark:bg-gray-700 dark:hover:bg-gray-600 dark:border-gray-600 dark:text-white transition-all duration-200" />
                    </div>
                  </div>

                  <div class="sm:col-span-3">
                    <label for="allowed_grid_import_kw" class="block text-sm font-medium text-gray-700 dark:text-gray-300">Allowed Grid Import (kW)</label>
                    <div class="mt-1">
                      <input type="number" step="0.1" id="allowed_grid_import_kw" v-model="siteSettings.allowed_grid_import_kw" :placeholder="maxGridPowerKw" class="shadow-sm focus:ring-indigo-500 focus:border-indigo-500 block w-full  border-gray-300 rounded-md bg-gray-50 hover:bg-white dark:bg-gray-700 dark:hover:bg-gray-600 dark:border-gray-600 dark:text-white transition-all duration-200" />
                    </div>
                    <p class="mt-1 text-xs text-gray-500">Proposed max: {{ maxGridPowerKw }} kW</p>
                  </div>

                  <div class="sm:col-span-3">
                    <label for="allowed_grid_export_kw" class="block text-sm font-medium text-gray-700 dark:text-gray-300">Allowed Grid Export (kW)</label>
                    <div class="mt-1">
                      <input type="number" step="0.1" id="allowed_grid_export_kw" v-model="siteSettings.allowed_grid_export_kw" :placeholder="maxGridPowerKw" class="shadow-sm focus:ring-indigo-500 focus:border-indigo-500 block w-full  border-gray-300 rounded-md bg-gray-50 hover:bg-white dark:bg-gray-700 dark:hover:bg-gray-600 dark:border-gray-600 dark:text-white transition-all duration-200" />
                    </div>
                    <p class="mt-1 text-xs text-gray-500">Proposed max: {{ maxGridPowerKw }} kW</p>
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
                                class="shadow-sm focus:ring-indigo-500 focus:border-indigo-500 block w-full  border-gray-300 rounded-md bg-gray-50 hover:bg-white dark:bg-gray-700 dark:hover:bg-gray-600 dark:border-gray-600 dark:text-white transition-all duration-200">
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
                               class="shadow-sm focus:ring-indigo-500 focus:border-indigo-500 block w-full  border-gray-300 rounded-md bg-gray-50 hover:bg-white dark:bg-gray-700 dark:hover:bg-gray-600 dark:border-gray-600 dark:text-white transition-all duration-200">
                      </div>
                    </div>

                    <div v-if="siteSettings.strategy_mode === 'flanders'" class="sm:col-span-3">
                      <label for="peak_shaving_buffer_w" class="block text-sm font-medium text-gray-700 dark:text-gray-300">Peak Shaving Buffer (W)</label>
                      <div class="mt-1">
                        <input type="number" step="1" min="0" id="peak_shaving_buffer_w" v-model="siteSettings.peak_shaving_buffer_w"
                               class="shadow-sm focus:ring-indigo-500 focus:border-indigo-500 block w-full  border-gray-300 rounded-md bg-gray-50 hover:bg-white dark:bg-gray-700 dark:hover:bg-gray-600 dark:border-gray-600 dark:text-white transition-all duration-200">
                      </div>
                    </div>

                    <div v-if="siteSettings.strategy_mode === 'flanders'" class="sm:col-span-3">
                      <label for="peak_shaving_rampup_w" class="block text-sm font-medium text-gray-700 dark:text-gray-300">Ramp-up Hysteresis (W)</label>
                      <div class="mt-1">
                        <input type="number" step="1" min="0" id="peak_shaving_rampup_w" v-model="siteSettings.peak_shaving_rampup_w"
                               class="shadow-sm focus:ring-indigo-500 focus:border-indigo-500 block w-full  border-gray-300 rounded-md bg-gray-50 hover:bg-white dark:bg-gray-700 dark:hover:bg-gray-600 dark:border-gray-600 dark:text-white transition-all duration-200">
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
                      <input type="number" step="1" min="0" id="appliance_turn_on_excess_w" v-model="siteSettings.appliance_turn_on_excess_w" class="shadow-sm focus:ring-indigo-500 focus:border-indigo-500 block w-full  border-gray-300 rounded-md bg-gray-50 hover:bg-white dark:bg-gray-700 dark:hover:bg-gray-600 dark:border-gray-600 dark:text-white transition-all duration-200" />
                    </div>
                  </div>
                </div>
              </div>
            </div>

            <!-- Energy Contract Configuration Card -->
            <div class="bg-white dark:bg-gray-800 shadow sm:rounded-lg mb-6">
              <div class="px-4 py-5 sm:p-6">
                <div class="mb-4 border-b border-gray-200 dark:border-gray-700 pb-4">
                  <h3 class="text-lg font-medium leading-6 text-gray-900 dark:text-gray-100">
                    Energy Contract Configuration
                  </h3>
                  <p class="mt-1 text-sm text-gray-500 dark:text-gray-400">
                    Configure your energy provider contract to accurately calculate costs and optimization thresholds.
                  </p>
                </div>

                <div class="grid grid-cols-1 gap-y-6 gap-x-4 sm:grid-cols-6">
                  <div class="sm:col-span-6">
                    <label for="contract_type" class="block text-sm font-medium text-gray-700 dark:text-gray-300">Contract Type</label>
                    <div class="mt-1">
                      <select id="contract_type" v-model="siteSettings.contract_type"
                              class="shadow-sm focus:ring-indigo-500 focus:border-indigo-500 block w-full  border-gray-300 rounded-md bg-gray-50 hover:bg-white dark:bg-gray-700 dark:hover:bg-gray-600 dark:border-gray-600 dark:text-white transition-all duration-200">
                        <option value="fixed">Fixed Price</option>
                        <option value="dynamic">Standard Dynamic</option>
                        <option value="engie_flextime">Engie EMPOWER Flextime</option>
                      </select>
                    </div>
                  </div>

                  <!-- Fixed Price -->
                  <template v-if="siteSettings.contract_type === 'fixed'">
                    <div class="sm:col-span-2">
                      <label for="fixed_price_peak_kwh" class="block text-sm font-medium text-gray-700 dark:text-gray-300">Peak Price (€/kWh)</label>
                      <div class="mt-1">
                        <input type="number" step="0.0001" id="fixed_price_peak_kwh" v-model="siteSettings.fixed_price_peak_kwh" class="shadow-sm focus:ring-indigo-500 focus:border-indigo-500 block w-full  border-gray-300 rounded-md bg-gray-50 hover:bg-white dark:bg-gray-700 dark:hover:bg-gray-600 dark:border-gray-600 dark:text-white transition-all duration-200" />
                      </div>
                    </div>
                    <div class="sm:col-span-2">
                      <label for="fixed_price_off_peak_kwh" class="block text-sm font-medium text-gray-700 dark:text-gray-300">Off-Peak Price (€/kWh)</label>
                      <div class="mt-1">
                        <input type="number" step="0.0001" id="fixed_price_off_peak_kwh" v-model="siteSettings.fixed_price_off_peak_kwh" class="shadow-sm focus:ring-indigo-500 focus:border-indigo-500 block w-full  border-gray-300 rounded-md bg-gray-50 hover:bg-white dark:bg-gray-700 dark:hover:bg-gray-600 dark:border-gray-600 dark:text-white transition-all duration-200" />
                      </div>
                    </div>
                    <div class="sm:col-span-2">
                      <label for="fixed_inject_price_kwh" class="block text-sm font-medium text-gray-700 dark:text-gray-300">Injection Return (€/kWh)</label>
                      <div class="mt-1">
                        <input type="number" step="0.0001" id="fixed_inject_price_kwh" v-model="siteSettings.fixed_inject_price_kwh" class="shadow-sm focus:ring-indigo-500 focus:border-indigo-500 block w-full  border-gray-300 rounded-md bg-gray-50 hover:bg-white dark:bg-gray-700 dark:hover:bg-gray-600 dark:border-gray-600 dark:text-white transition-all duration-200" />
                      </div>
                    </div>
                  </template>

                  <!-- Standard Dynamic -->
                  <template v-if="siteSettings.contract_type === 'dynamic'">
                    <div class="sm:col-span-6">
                      <label for="dynamic_markup_kwh" class="block text-sm font-medium text-gray-700 dark:text-gray-300">Markup / Taxes on top of EPEX (€/kWh)</label>
                      <div class="mt-1">
                        <input type="number" step="0.0001" id="dynamic_markup_kwh" v-model="siteSettings.dynamic_markup_kwh" class="shadow-sm focus:ring-indigo-500 focus:border-indigo-500 block w-full  border-gray-300 rounded-md bg-gray-50 hover:bg-white dark:bg-gray-700 dark:hover:bg-gray-600 dark:border-gray-600 dark:text-white transition-all duration-200" />
                      </div>
                    </div>
                  </template>

                  <!-- Engie EMPOWER Flextime -->
                  <template v-if="siteSettings.contract_type === 'engie_flextime'">
                    <div class="sm:col-span-6 border border-indigo-100 dark:border-indigo-900 rounded-lg p-4 bg-indigo-50/50 dark:bg-indigo-900/10 mb-4">
                      <div class="flex items-center justify-between">
                        <div>
                          <h4 class="text-sm font-medium text-indigo-900 dark:text-indigo-300">Superdal Optimization Mode</h4>
                          <p class="text-xs text-indigo-700 dark:text-indigo-400 mt-1">
                            <strong>Super-dal:</strong> 01:00 - 07:00 | <strong>Piek:</strong> 07:00 - 11:00 & 17:00 - 22:00 | <strong>Dal:</strong> 11:00 - 17:00 & 22:00 - 01:00 (Weekend)
                          </p>
                          <p class="text-xs text-indigo-700 dark:text-indigo-400 mt-1">Automatically charge battery from grid during Superdal hours to maximize savings.</p>
                        </div>
                        <div class="flex items-center h-5">
                          <input id="superdal_optimization_enabled" type="checkbox" v-model="siteSettings.superdal_optimization_enabled" class="focus:ring-indigo-500 h-5 w-5 text-indigo-600 border-gray-300 rounded transition-colors duration-200 cursor-pointer" />
                        </div>
                      </div>

                      <div v-if="siteSettings.superdal_optimization_enabled" class="mt-4 pt-4 border-t border-indigo-100 dark:border-indigo-800">
                        <label for="superdal_target_soc" class="block text-sm font-medium text-gray-700 dark:text-gray-300">Target SoC %</label>
                        <div class="mt-1 flex items-center">
                          <input type="number" step="1" min="0" max="100" id="superdal_target_soc" v-model="siteSettings.superdal_target_soc" class="shadow-sm focus:ring-indigo-500 focus:border-indigo-500 block w-32  border-gray-300 rounded-md bg-white hover:bg-gray-50 dark:bg-gray-800 dark:hover:bg-gray-700 dark:border-gray-600 dark:text-white transition-all duration-200" />
                          <span class="ml-2 text-sm text-gray-500">%</span>
                        </div>
                      </div>
                    </div>
                    <div class="sm:col-span-3">
                      <label for="engie_multiplier" class="block text-sm font-medium text-gray-700 dark:text-gray-300">EPEX Multiplier</label>
                      <div class="mt-1">
                        <input type="number" step="0.0001" id="engie_multiplier" v-model="siteSettings.engie_multiplier" class="shadow-sm focus:ring-indigo-500 focus:border-indigo-500 block w-full  border-gray-300 rounded-md bg-gray-50 hover:bg-white dark:bg-gray-700 dark:hover:bg-gray-600 dark:border-gray-600 dark:text-white transition-all duration-200" />
                      </div>
                    </div>
                    <div class="sm:col-span-3">
                      <label for="engie_base_fee" class="block text-sm font-medium text-gray-700 dark:text-gray-300">Base Fee (€/kWh)</label>
                      <div class="mt-1">
                        <input type="number" step="0.0001" id="engie_base_fee" v-model="siteSettings.engie_base_fee" class="shadow-sm focus:ring-indigo-500 focus:border-indigo-500 block w-full  border-gray-300 rounded-md bg-gray-50 hover:bg-white dark:bg-gray-700 dark:hover:bg-gray-600 dark:border-gray-600 dark:text-white transition-all duration-200" />
                      </div>
                    </div>
                    <div class="sm:col-span-2">
                      <label for="engie_markup_peak" class="block text-sm font-medium text-gray-700 dark:text-gray-300">Markup Piekuren (€/kWh)</label>
                      <div class="mt-1">
                        <input type="number" step="0.0001" id="engie_markup_peak" v-model="siteSettings.engie_markup_peak" class="shadow-sm focus:ring-indigo-500 focus:border-indigo-500 block w-full  border-gray-300 rounded-md bg-gray-50 hover:bg-white dark:bg-gray-700 dark:hover:bg-gray-600 dark:border-gray-600 dark:text-white transition-all duration-200" />
                      </div>
                    </div>
                    <div class="sm:col-span-2">
                      <label for="engie_markup_off_peak" class="block text-sm font-medium text-gray-700 dark:text-gray-300">Markup Daluren (€/kWh)</label>
                      <div class="mt-1">
                        <input type="number" step="0.0001" id="engie_markup_off_peak" v-model="siteSettings.engie_markup_off_peak" class="shadow-sm focus:ring-indigo-500 focus:border-indigo-500 block w-full  border-gray-300 rounded-md bg-gray-50 hover:bg-white dark:bg-gray-700 dark:hover:bg-gray-600 dark:border-gray-600 dark:text-white transition-all duration-200" />
                      </div>
                    </div>
                    <div class="sm:col-span-2">
                      <label for="engie_markup_super_off_peak" class="block text-sm font-medium text-gray-700 dark:text-gray-300">Markup Super-daluren (€/kWh)</label>
                      <div class="mt-1">
                        <input type="number" step="0.0001" id="engie_markup_super_off_peak" v-model="siteSettings.engie_markup_super_off_peak" class="shadow-sm focus:ring-indigo-500 focus:border-indigo-500 block w-full  border-gray-300 rounded-md bg-gray-50 hover:bg-white dark:bg-gray-700 dark:hover:bg-gray-600 dark:border-gray-600 dark:text-white transition-all duration-200" />
                      </div>
                    </div>
                  </template>
                </div>
              </div>
            </div>

            <!-- Dynamic Tariffs Card -->
            <div class="bg-white dark:bg-gray-800 shadow sm:rounded-lg mb-6">
              <div class="px-4 py-5 sm:p-6">
                <div class="mb-4 border-b border-gray-200 dark:border-gray-700 pb-4">
                  <h3 class="text-lg font-medium leading-6 text-gray-900 dark:text-gray-100">
                    Energy Arbitrage & Optimization
                  </h3>
                  <p class="mt-1 text-sm text-gray-500 dark:text-gray-400">
                    Configure rules based on the calculated effective energy price.
                  </p>
                </div>

                <div class="grid grid-cols-1 gap-y-6 gap-x-4 sm:grid-cols-6">
                  <div class="sm:col-span-6">
                    <label for="force_charge_below_euro" class="block text-sm font-medium text-gray-700 dark:text-gray-300">Force Charge Battery if price drops below (€/kWh)</label>
                    <div class="mt-1">
                      <input type="number" step="0.01" id="force_charge_below_euro" v-model="siteSettings.force_charge_below_euro" class="shadow-sm focus:ring-indigo-500 focus:border-indigo-500 block w-full  border-gray-300 rounded-md bg-gray-50 hover:bg-white dark:bg-gray-700 dark:hover:bg-gray-600 dark:border-gray-600 dark:text-white transition-all duration-200" />
                    </div>
                  </div>

                  <div class="sm:col-span-6">
                    <label for="force_discharge_above_euro" class="block text-sm font-medium text-gray-700 dark:text-gray-300">Force Discharge Battery to Grid if price rises above (€/kWh)</label>
                    <div class="mt-1">
                      <input type="number" step="0.01" id="force_discharge_above_euro" v-model="siteSettings.force_discharge_above_euro" class="shadow-sm focus:ring-indigo-500 focus:border-indigo-500 block w-full  border-gray-300 rounded-md bg-gray-50 hover:bg-white dark:bg-gray-700 dark:hover:bg-gray-600 dark:border-gray-600 dark:text-white transition-all duration-200" />
                    </div>
                  </div>

                  <div class="sm:col-span-6">
                    <label for="smart_ev_cheapest_hours" class="block text-sm font-medium text-gray-700 dark:text-gray-300">Smart EV Charging: Charge during the cheapest hours of the day</label>
                    <div class="mt-1">
                      <input type="number" step="1" min="0" max="24" id="smart_ev_cheapest_hours" v-model="siteSettings.smart_ev_cheapest_hours" class="shadow-sm focus:ring-indigo-500 focus:border-indigo-500 block w-full  border-gray-300 rounded-md bg-gray-50 hover:bg-white dark:bg-gray-700 dark:hover:bg-gray-600 dark:border-gray-600 dark:text-white transition-all duration-200" />
                    </div>
                  </div>

                  <!-- Custom Charging Timetable -->
                  <div class="sm:col-span-6 border-t border-gray-200 dark:border-gray-700 pt-6 mt-2">
                    <div class="flex items-center justify-between mb-4">
                      <div>
                        <h4 class="text-sm font-medium text-gray-900 dark:text-gray-100">Custom Charging Timetable</h4>
                        <p class="text-xs text-gray-500 dark:text-gray-400 mt-1">Define forced charge blocks. Battery will charge from grid (or solar) to target SoC during these hours.</p>
                      </div>
                      <button type="button" @click="addCustomChargeSlot" class="inline-flex items-center px-3 py-1.5 border border-transparent text-xs font-medium rounded shadow-sm text-white bg-indigo-600 hover:bg-indigo-700 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-indigo-500 transition-colors">
                        Add Slot
                      </button>
                    </div>

                    <div v-if="parsedChargeSchedule.length === 0" class="text-sm text-center text-gray-500 dark:text-gray-400 py-4 bg-gray-50 dark:bg-gray-800/50 rounded-lg border border-dashed border-gray-300 dark:border-gray-700">
                      No custom charging slots defined.
                    </div>

                    <div v-else class="space-y-3">
                      <div v-for="(slot, index) in parsedChargeSchedule" :key="index" class="flex items-center gap-4 bg-gray-50 dark:bg-gray-800/50 p-3 rounded-lg border border-gray-200 dark:border-gray-700">
                        <div class="flex-1 grid grid-cols-3 gap-4">
                          <div>
                            <label class="block text-xs font-medium text-gray-500 dark:text-gray-400 mb-1">Start Time</label>
                            <input type="time" v-model="slot.start" @change="updateChargeSchedule" class="shadow-sm focus:ring-indigo-500 focus:border-indigo-500 block w-full  border-gray-300 rounded-md bg-white hover:bg-gray-50 dark:bg-gray-700 dark:hover:bg-gray-600 dark:border-gray-600 dark:text-white transition-all duration-200" required />
                          </div>
                          <div>
                            <label class="block text-xs font-medium text-gray-500 dark:text-gray-400 mb-1">End Time</label>
                            <input type="time" v-model="slot.end" @change="updateChargeSchedule" class="shadow-sm focus:ring-indigo-500 focus:border-indigo-500 block w-full  border-gray-300 rounded-md bg-white hover:bg-gray-50 dark:bg-gray-700 dark:hover:bg-gray-600 dark:border-gray-600 dark:text-white transition-all duration-200" required />
                          </div>
                          <div>
                            <label class="block text-xs font-medium text-gray-500 dark:text-gray-400 mb-1">Target SoC %</label>
                            <div class="flex items-center">
                              <input type="number" min="0" max="100" step="1" v-model="slot.target_soc" @change="updateChargeSchedule" class="shadow-sm focus:ring-indigo-500 focus:border-indigo-500 block w-full  border-gray-300 rounded-md bg-white hover:bg-gray-50 dark:bg-gray-700 dark:hover:bg-gray-600 dark:border-gray-600 dark:text-white transition-all duration-200" required />
                            </div>
                          </div>
                        </div>
                        <button type="button" @click="removeCustomChargeSlot(index)" class="mt-5 inline-flex items-center p-1.5 border border-transparent rounded-full shadow-sm text-white bg-red-600 hover:bg-red-700 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-red-500 transition-colors" title="Remove Slot">
                          <svg class="h-4 w-4" xmlns="http://www.w3.org/2000/svg" viewBox="0 0 20 20" fill="currentColor" aria-hidden="true">
                            <path fill-rule="evenodd" d="M4.293 4.293a1 1 0 011.414 0L10 8.586l4.293-4.293a1 1 0 111.414 1.414L11.414 10l4.293 4.293a1 1 0 01-1.414 1.414L10 11.414l-4.293 4.293a1 1 0 01-1.414-1.414L8.586 10 4.293 5.707a1 1 0 010-1.414z" clip-rule="evenodd" />
                          </svg>
                        </button>
                      </div>
                    </div>
                  </div>
                </div>

              </div>
            </div>

            <!-- Logging Settings Card -->
            <div class="bg-white dark:bg-gray-800 shadow sm:rounded-lg mb-6">
              <div class="px-4 py-5 sm:p-6">
                <div class="mb-4 border-b border-gray-200 dark:border-gray-700 pb-4">
                  <h3 class="text-lg font-medium leading-6 text-gray-900 dark:text-gray-100">
                    System Logging
                  </h3>
                  <p class="mt-1 text-sm text-gray-500 dark:text-gray-400">
                    Configure the detail level of system logs.
                  </p>
                </div>
                <div class="grid grid-cols-1 gap-y-6 gap-x-4 sm:grid-cols-6">
                  <div class="sm:col-span-3">
                    <label for="log_level" class="block text-sm font-medium text-gray-700 dark:text-gray-300">Log Level</label>
                    <div class="mt-1">
                      <select id="log_level" v-model="siteSettings.log_level"
                              class="shadow-sm focus:ring-indigo-500 focus:border-indigo-500 block w-full  border-gray-300 rounded-md bg-gray-50 hover:bg-white dark:bg-gray-700 dark:hover:bg-gray-600 dark:border-gray-600 dark:text-white transition-all duration-200">
                        <option value="TRACE">Trace</option>
                        <option value="DEBUG">Debug</option>
                        <option value="INFO">Info</option>
                        <option value="WARN">Warn</option>
                        <option value="ERROR">Error</option>
                        <option value="FATAL">Fatal</option>
                      </select>
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
                        <p><strong>Vendor:</strong> {{ getTemplateVendor(device.template) }}</p>
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
                        <div v-html="cat.svg" class="w-8 h-8 mb-2" :class="cat.color"></div>
                        <span class="text-sm font-medium text-gray-900 dark:text-white">{{ cat.name }}</span>
                      </div>
                    </div>
                  </div>

                  <!-- Reset Category Button & Category Info -->
                  <div class="sm:col-span-6 flex items-center justify-between" v-if="selectedCategory">
                    <div class="flex items-center">
                      <div v-html="currentCategoryObj?.svg" class="w-6 h-6 mr-2" :class="currentCategoryObj?.color"></div>
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
                               class="shadow-sm focus:ring-indigo-500 focus:border-indigo-500 block w-full  border-gray-300 rounded-md bg-gray-50 hover:bg-white dark:bg-gray-700 dark:hover:bg-gray-600 dark:border-gray-600 dark:text-white transition-all duration-200">
                      </div>
                    </div>

                    <div class="sm:col-span-3">
                      <label for="template" class="block text-sm font-medium text-gray-700 dark:text-gray-300">Device Template</label>
                      <div class="mt-1">
                        <select id="template" v-model="form.template" required
                                class="shadow-sm focus:ring-indigo-500 focus:border-indigo-500 block w-full  border-gray-300 rounded-md bg-gray-50 hover:bg-white dark:bg-gray-700 dark:hover:bg-gray-600 dark:border-gray-600 dark:text-white transition-all duration-200">
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
                  <template v-else-if="templates.find(t => t.id === form.template)?.type === 'ocpp'">
                    <OcppTemplate v-model="form" />
                  </template>

                  <!-- Inverter Rated Power Selection -->
                  <div v-if="getDeviceCategory(form.template) === 'Inverters & Solar'" class="sm:col-span-3">
                    <label for="inverter_rated_power_kw" class="block text-sm font-medium text-gray-700 dark:text-gray-300">Rated Power (kW)</label>
                    <div class="mt-1">
                      <input type="number" step="0.1" id="inverter_rated_power_kw" v-model="form.inverter_rated_power_kw"
                             class="shadow-sm focus:ring-indigo-500 focus:border-indigo-500 block w-full  border-gray-300 rounded-md bg-gray-50 hover:bg-white dark:bg-gray-700 dark:hover:bg-gray-600 dark:border-gray-600 dark:text-white transition-all duration-200">
                    </div>
                  </div>

                  <!-- EV Charger Mode Selection -->
                  <div v-if="isCharger(form.template)" class="sm:col-span-3">
                    <label for="charge_mode" class="block text-sm font-medium text-gray-700 dark:text-gray-300">Charge Mode</label>
                    <div class="mt-1">
                      <select id="charge_mode" v-model="form.charge_mode"
                              class="shadow-sm focus:ring-indigo-500 focus:border-indigo-500 block w-full  border-gray-300 rounded-md bg-gray-50 hover:bg-white dark:bg-gray-700 dark:hover:bg-gray-600 dark:border-gray-600 dark:text-white transition-all duration-200">
                        <option value="eco">Eco / Smart</option>
                        <option value="pv_only">PV Only (Solar)</option>
                        <option value="now">Fast (Max Power)</option>
                        <option value="off">Off</option>
                      </select>
                    </div>
                  </div>

                  <!-- Conditional Fields for Huawei Hybrid Inverter Combo -->
                  <template v-if="form.template === 'huawei_inverter' || form.template === 'enerlution_inverter'">
                    <div class="sm:col-span-6 border-t border-gray-200 dark:border-gray-700 pt-4 mt-2">
                      <h4 class="text-sm font-medium text-gray-900 dark:text-gray-100 mb-4">Hybrid Inverter Features</h4>

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
                                 class="shadow-sm focus:ring-indigo-500 focus:border-indigo-500 block w-full  border-gray-300 rounded-md bg-gray-50 hover:bg-white dark:bg-gray-700 dark:hover:bg-gray-600 dark:border-gray-600 dark:text-white transition-all duration-200">
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
                         class="shadow-sm focus:ring-indigo-500 focus:border-indigo-500 block w-full  border-gray-300 rounded-md bg-gray-50 hover:bg-white dark:bg-gray-700 dark:hover:bg-gray-600 dark:border-gray-600 dark:text-white transition-all duration-200">
                </div>
              </div>

              <div class="sm:col-span-3">
                <label for="edit_template" class="block text-sm font-medium text-gray-700 dark:text-gray-300">Device Template</label>
                <div class="mt-1">
                  <select id="edit_template" v-model="editForm.template" required
                          class="shadow-sm focus:ring-indigo-500 focus:border-indigo-500 block w-full  border-gray-300 rounded-md bg-gray-50 hover:bg-white dark:bg-gray-700 dark:hover:bg-gray-600 dark:border-gray-600 dark:text-white transition-all duration-200">
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
              <template v-else-if="templates.find(t => t.id === editForm.template)?.type === 'ocpp'">
                <OcppTemplate v-model="editForm" prefix="edit_" />
              </template>

              <!-- Inverter Rated Power Selection -->
              <div v-if="getDeviceCategory(editForm.template) === 'Inverters & Solar'" class="sm:col-span-3">
                <label for="edit_inverter_rated_power_kw" class="block text-sm font-medium text-gray-700 dark:text-gray-300">Rated Power (kW)</label>
                <div class="mt-1">
                  <input type="number" step="0.1" id="edit_inverter_rated_power_kw" v-model="editForm.inverter_rated_power_kw"
                         class="shadow-sm focus:ring-indigo-500 focus:border-indigo-500 block w-full  border-gray-300 rounded-md bg-gray-50 hover:bg-white dark:bg-gray-700 dark:hover:bg-gray-600 dark:border-gray-600 dark:text-white transition-all duration-200">
                </div>
              </div>

              <!-- EV Charger Mode Selection -->
              <div v-if="isCharger(editForm.template)" class="sm:col-span-3">
                <label for="edit_charge_mode" class="block text-sm font-medium text-gray-700 dark:text-gray-300">Charge Mode</label>
                <div class="mt-1">
                  <select id="edit_charge_mode" v-model="editForm.charge_mode"
                          class="shadow-sm focus:ring-indigo-500 focus:border-indigo-500 block w-full  border-gray-300 rounded-md bg-gray-50 hover:bg-white dark:bg-gray-700 dark:hover:bg-gray-600 dark:border-gray-600 dark:text-white transition-all duration-200">
                    <option value="eco">Eco / Smart</option>
                    <option value="pv_only">PV Only (Solar)</option>
                    <option value="now">Fast (Max Power)</option>
                    <option value="off">Off</option>
                  </select>
                </div>
              </div>

              <!-- Conditional Fields for Huawei Hybrid Inverter Combo -->
              <template v-if="editForm.template === 'huawei_inverter' || editForm.template === 'enerlution_inverter'">
                <div class="sm:col-span-6 border-t border-gray-200 dark:border-gray-700 pt-4 mt-2">
                  <h4 class="text-sm font-medium text-gray-900 dark:text-gray-100 mb-4">Hybrid Inverter Features</h4>

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
                             class="shadow-sm focus:ring-indigo-500 focus:border-indigo-500 block w-full  border-gray-300 rounded-md bg-gray-50 hover:bg-white dark:bg-gray-700 dark:hover:bg-gray-600 dark:border-gray-600 dark:text-white transition-all duration-200">
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
import OcppTemplate from './templates/OcppTemplate.vue'
import P1SerialTemplate from './templates/P1SerialTemplate.vue'
import P1NetworkTemplate from './templates/P1NetworkTemplate.vue'

const activeTab = ref('strategy')

const sysInfo = ref<SystemInfo | null>(null)

const selectedCategory = ref<string | null>(null)
const deviceCategories = [
  { id: 'inverter', name: 'Inverter / Solar', svg: '<svg xmlns="http://www.w3.org/2000/svg" width="24" height="24" viewBox="0 0 24 24" fill="currentColor"><path d="M6.993 12c0 2.761 2.246 5.007 5.007 5.007s5.007-2.246 5.007-5.007S14.761 6.993 12 6.993 6.993 9.239 6.993 12zM12 8.993c1.658 0 3.007 1.349 3.007 3.007S13.658 15.007 12 15.007 8.993 13.658 8.993 12 10.342 8.993 12 8.993zM10.998 19h2v3h-2zm0-17h2v3h-2zm-9 9h3v2h-3zm17 0h3v2h-3zM4.219 18.363l2.12-2.122 1.415 1.414-2.12 2.122zM16.24 6.344l2.122-2.122 1.414 1.414-2.122 2.122zM6.342 7.759 4.22 5.637l1.415-1.414 2.12 2.122zm13.434 10.605-1.414 1.414-2.122-2.122 1.414-1.414z"/></svg>', color: 'text-yellow-500' },
  { id: 'charger', name: 'EV Charger', svg: '<svg xmlns="http://www.w3.org/2000/svg" width="24" height="24" viewBox="0 0 24 24" fill="currentColor"><path d="m20.772 10.156-1.368-4.105A2.995 2.995 0 0 0 16.559 4H7.441a2.995 2.995 0 0 0-2.845 2.051l-1.368 4.105A2.003 2.003 0 0 0 2 12v5c0 .753.423 1.402 1.039 1.743-.013.066-.039.126-.039.195V21a1 1 0 0 0 1 1h1a1 1 0 0 0 1-1v-2h12v2a1 1 0 0 0 1 1h1a1 1 0 0 0 1-1v-2.062c0-.069-.026-.13-.039-.195A1.993 1.993 0 0 0 22 17v-5c0-.829-.508-1.541-1.228-1.844zM4 17v-5h16l.002 5H4zM7.441 6h9.117c.431 0 .813.274.949.684L18.613 10H5.387l1.105-3.316A1 1 0 0 1 7.441 6z"/><circle cx="6.5" cy="14.5" r="1.5"/><circle cx="17.5" cy="14.5" r="1.5"/></svg>', color: 'text-purple-500' },
  { id: 'meter', name: 'Smart Meter', svg: '<svg xmlns="http://www.w3.org/2000/svg" width="24" height="24" viewBox="0 0 24 24" fill="currentColor"><path d="M12 2C6.486 2 2 6.486 2 12s4.486 10 10 10 10-4.486 10-10S17.514 2 12 2zm0 18c-4.411 0-8-3.589-8-8s3.589-8 8-8 8 3.589 8 8-3.589 8-8 8z"/><path d="m13 6-6 7h4v5l6-7h-4z"/></svg>', color: 'text-blue-500' },
  { id: 'battery', name: 'Battery', svg: '<svg xmlns="http://www.w3.org/2000/svg" width="24" height="24" viewBox="0 0 24 24" fill="currentColor"><path d="M16 7h-2V6c0-1.103-.897-2-2-2h-4c-1.103 0-2 .897-2 2v1H4c-1.103 0-2 .897-2 2v10c0 1.103.897 2 2 2h12c1.103 0 2-.897 2-2V9c0-1.103-.897-2-2-2zM8 6h4v1H8V6zm8 13H4V9h12v10z"/><rect width="2" height="6" x="19" y="11"/></svg>', color: 'text-green-500' },
  { id: 'relay', name: 'Relay / Switch', svg: '<svg xmlns="http://www.w3.org/2000/svg" width="24" height="24" viewBox="0 0 24 24" fill="currentColor"><path d="M16 7H8c-2.757 0-5 2.243-5 5s2.243 5 5 5h8c2.757 0 5-2.243 5-5s-2.243-5-5-5zM8 15c-1.654 0-3-1.346-3-3s1.346-3 3-3 3 1.346 3 3-1.346 3-3 3z"/></svg>', color: 'text-orange-500' }
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
  inverter_rated_power_kw: 0,
  charge_mode: 'eco',
  ocpp_proxy_url: ''
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
  inverter_rated_power_kw: 0,
  charge_mode: 'eco',
  ocpp_proxy_url: ''
})

const siteSettings = ref<SiteSettings>({
  strategy_mode: 'eco',
  capacity_peak_limit_kw: 2.5,
  active_inverter_curtailment: false,
  force_charge_below_euro: 0.0,
  force_discharge_above_euro: 999.0,
  smart_ev_cheapest_hours: 0,
      battery_arbitrage_cheapest_hours: 0,
      battery_arbitrage_expensive_hours: 0,
      solar_forecast_enabled: false,
      latitude: 50.8503,
      longitude: 4.3517,
  grid_nominal_current_a: 25.0,
  grid_system: 'single_phase_230v',
  allowed_grid_import_kw: 0.0,
  allowed_grid_export_kw: 0.0,
  appliance_turn_on_excess_w: 0.0,
  peak_shaving_buffer_w: 200.0,
  peak_shaving_rampup_w: 500.0,
  timezone: 'Europe/Brussels',
  log_level: 'INFO',
  contract_type: 'dynamic',
  fixed_price_peak_kwh: 0.35,
  fixed_price_off_peak_kwh: 0.30,
  fixed_inject_price_kwh: 0.05,
  dynamic_markup_kwh: 0.15,
  engie_markup_peak: 0.15,
  engie_markup_off_peak: 0.15,
  engie_markup_super_off_peak: 0.15,
  engie_multiplier: 0.1448,
  engie_base_fee: 0.0,
  custom_charge_schedule: '[]',
  superdal_optimization_enabled: false,
  superdal_target_soc: 100.0
})

const parsedChargeSchedule = ref<Array<{start: string, end: string, target_soc: number}>>([])

const updateChargeSchedule = () => {
  siteSettings.value.custom_charge_schedule = JSON.stringify(parsedChargeSchedule.value)
}

const addCustomChargeSlot = () => {
  parsedChargeSchedule.value.push({
    start: '00:00',
    end: '06:00',
    target_soc: 100
  })
  updateChargeSchedule()
}

const removeCustomChargeSlot = (index: number) => {
  parsedChargeSchedule.value.splice(index, 1)
  updateChargeSchedule()
}
const saveSettingsSuccess = ref(false)

const maxGridPowerKw = computed(() => {
  const current = siteSettings.value.grid_nominal_current_a
  const system = siteSettings.value.grid_system
  let maxKw = 0
  if (system === 'single_phase_230v') {
    maxKw = (current * 230) / 1000
  } else if (system === 'three_phase_400v') {
    maxKw = (current * 230 * 3) / 1000
  } else if (system === 'three_phase_230v_delta') {
    maxKw = (current * 230 * Math.sqrt(3)) / 1000
  }
  return maxKw.toFixed(1)
})

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
      if (siteSettings.value.custom_charge_schedule) {
        try {
          parsedChargeSchedule.value = JSON.parse(siteSettings.value.custom_charge_schedule)
        } catch (e) {
          parsedChargeSchedule.value = []
        }
      }
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

const getTemplateVendor = (id: string) => {
  const t = templates.value.find(t => t.id === id)
  return t ? t.vendor : 'Unknown'
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
        inverter_rated_power_kw: 0,
        charge_mode: 'eco',
        ocpp_proxy_url: ''
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
    inverter_rated_power_kw: device.inverter_rated_power_kw || 0,
    charge_mode: device.charge_mode || 'eco',
    ocpp_proxy_url: device.ocpp_proxy_url || ''
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

const resetDatabase = async () => {
  if (confirm("WARNING: Are you sure you want to reset the database? This is irreversible and will delete all measurements and configured devices!")) {
    try {
      const res = await fetch(`${getApiBase()}/api/system/reset-db`, { method: 'POST' })
      if (res.ok) {
        alert("Database has been reset. The page will now reload.")
        window.location.reload()
      } else {
        alert("Failed to reset database.")
      }
    } catch (e) {
      console.error("Failed to reset database:", e)
      alert("Failed to send reset command.")
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