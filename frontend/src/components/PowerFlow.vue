<template>
  <div class="w-full flex justify-center items-center">
    <div class="relative w-full h-[460px] bg-white dark:bg-gray-800 rounded-lg shadow overflow-hidden">
      <!-- Solar Forecast Button -->
      <button @click="openSolarModal" class="absolute top-4 left-4 z-20 cursor-pointer text-gray-500 hover:text-yellow-500 transition-colors" title="Solar Forecast">
        <svg xmlns="http://www.w3.org/2000/svg" class="h-6 w-6" viewBox="0 0 24 24" fill="currentColor">
          <path d="M6.993 12c0 2.761 2.246 5.007 5.007 5.007s5.007-2.246 5.007-5.007S14.761 6.993 12 6.993 6.993 9.239 6.993 12zM12 8.993c1.658 0 3.007 1.349 3.007 3.007S13.658 15.007 12 15.007 8.993 13.658 8.993 12 10.342 8.993 12 8.993zM10.998 19h2v3h-2zm0-17h2v3h-2zm-9 9h3v2h-3zm17 0h3v2h-3zM4.219 18.363l2.12-2.122 1.415 1.414-2.12 2.122zM16.24 6.344l2.122-2.122 1.414 1.414-2.122 2.122zM6.342 7.759 4.22 5.637l1.415-1.414 2.12 2.122zm13.434 10.605-1.414 1.414-2.122-2.122 1.414-1.414z"/>
        </svg>
      </button>

      <!-- Tariff Forecast Button -->
      <button @click="openTariffModal" class="absolute top-4 right-4 z-20 cursor-pointer text-gray-500 hover:text-blue-500 transition-colors" title="Tariff Forecast">
        <svg xmlns="http://www.w3.org/2000/svg" class="h-6 w-6" viewBox="0 0 24 24" fill="none" stroke="currentColor">
          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M7 12l3-3 3 3 4-4M8 21l4-4 4 4M3 4h18M4 4h16v12a1 1 0 01-1 1H5a1 1 0 01-1-1V4z" />
        </svg>
      </button>

      <!-- SVG paths for animated power flow lines -->
      <svg class="absolute inset-0 w-full h-full z-0 pointer-events-none" viewBox="0 0 100 100" preserveAspectRatio="none">


                <!-- Static Background Lines (Direct Routing) -->
        <!-- Center line (Grid to Home) -->
        <path v-for="device in gridDevices" :key="'static-grid-center-'+device.id" :d="`M 15 50 L 85 50`" vector-effect="non-scaling-stroke" stroke-linecap="round" fill="none" stroke="#E5E7EB" stroke-width="1.5" stroke-opacity="0.5" />

        <!-- Solar -> Home -->
        <path v-for="device in solarDevices" :key="'static-solar-center-'+device.id" :d="`M 50 20 C 50 45, 55 45, 85 45`" vector-effect="non-scaling-stroke" stroke-linecap="round" fill="none" stroke="#E5E7EB" stroke-width="1.5" stroke-opacity="0.5" />

        <!-- Solar -> Grid -->
        <path v-for="device in solarDevices" :key="'static-solar-grid-'+device.id" :d="`M 50 20 C 50 48, 45 48, 15 48`" vector-effect="non-scaling-stroke" stroke-linecap="round" fill="none" stroke="#E5E7EB" stroke-width="1.5" stroke-opacity="0.5" />

        <!-- Battery -> Home -->
        <path v-for="device in batteryDevices" :key="'static-battery-center-'+device.id" :d="`M 50 80 C 50 55, 55 55, 85 55`" vector-effect="non-scaling-stroke" stroke-linecap="round" fill="none" stroke="#E5E7EB" stroke-width="1.5" stroke-opacity="0.5" />

        <!-- Grid -> Battery -->
        <path v-for="device in batteryDevices" :key="'static-grid-battery-'+device.id" :d="`M 15 52 C 45 52, 45 52, 50 80`" vector-effect="non-scaling-stroke" stroke-linecap="round" fill="none" stroke="#E5E7EB" stroke-width="1.5" stroke-opacity="0.5" />

        <!-- Home -> EV -->
        <path v-for="device in evDevices" :key="'static-home-ev-'+device.id" :d="`M 85 50 L 85 80`" vector-effect="non-scaling-stroke" stroke-linecap="round" fill="none" stroke="#E5E7EB" stroke-width="1.5" stroke-opacity="0.5" />

        <!-- Home -> Appliance -->
        <path v-for="device in applianceDevices" :key="'static-home-appliance-'+device.id" :d="`M 85 50 L 85 20`" vector-effect="non-scaling-stroke" stroke-linecap="round" fill="none" stroke="#E5E7EB" stroke-width="1.5" stroke-opacity="0.5" />

        <!-- Flow lines -->
        <template v-for="segment in activeSegments" :key="segment.id">
          <!-- Single perfectly round animated dot per flow line -->
          <path :d="segment.path" stroke-linecap="round"
                fill="none" :stroke="segment.color" stroke-width="8" stroke-dasharray="0.1 200" class="flow-dot"
                vector-effect="non-scaling-stroke"
                :style="getFlowStyle(segment.power, segment.normalIsPositive)" />
        </template>

      </svg>

      <!-- Grid Layout for Nodes -->
      <div class="absolute inset-0">


        <!-- Grid (Top) -->
        <template v-for="(device, index) in gridDevices" :key="device.id">
          <div class="absolute flex flex-col items-center justify-center transform -translate-x-1/2 -translate-y-1/2" :style="`left: 15%; top: ${50 - (index * 15 - (gridDevices.length-1)*7.5)}%;`">
            <div @click="openChart('grid')" class="z-10 flex flex-col items-center justify-center w-28 h-28 bg-white dark:bg-gray-800 rounded-full border-[4px] border-[#3B82F6] shadow-sm cursor-pointer hover:scale-105 transition-transform mb-2 relative">
              <span class="text-xs font-medium text-gray-500 mb-1 absolute -top-6">Grid</span>
              <svg xmlns="http://www.w3.org/2000/svg" class="h-8 w-8 text-gray-700 dark:text-gray-300 mb-1" fill="currentColor" viewBox="0 0 24 24"><path d="M12 2C6.486 2 2 6.486 2 12s4.486 10 10 10 10-4.486 10-10S17.514 2 12 2zm0 18c-4.411 0-8-3.589-8-8s3.589-8 8-8 8 3.589 8 8-3.589 8-8 8z"/><path d="m13 6-6 7h4v5l6-7h-4z"/></svg>
              <div class="text-sm font-medium flex flex-col items-center leading-tight">
                <span v-if="state?.grid_power_w !== null && state?.grid_power_w !== undefined" class="text-purple-600 dark:text-purple-400 text-xs">
                  &larr; {{ state.grid_power_w < 0 ? formatPowerSimple(Math.abs(state.grid_power_w / gridDevices.length)) : '0 W' }}
                </span>
                <span v-else class="text-purple-600 dark:text-purple-400 text-xs">&larr; 0 W</span>

                <span v-if="state?.grid_power_w !== null && state?.grid_power_w !== undefined" class="text-blue-500 dark:text-blue-400 text-xs">
                  &rarr; {{ state.grid_power_w > 0 ? formatPowerSimple(state.grid_power_w / gridDevices.length) : '0 W' }}
                </span>
                <span v-else class="text-blue-500 dark:text-blue-400 text-xs">&rarr; 0 W</span>
              </div>
            </div>
          </div>
        </template>

        <!-- Battery (Left) -->
        <template v-if="batteryDevices.length > 0">
          <template v-for="(device, index) in batteryDevices" :key="device.id">
            <div class="absolute flex flex-col items-center justify-center transform -translate-x-1/2 -translate-y-1/2" :style="`left: ${50 - (index * 15 - (batteryDevices.length-1)*7.5)}%; top: 80%;`">
              <div @click="openChart('battery')" class="z-10 flex flex-col items-center justify-center w-28 h-28 bg-white dark:bg-gray-800 rounded-full border-[4px] border-[#EC4899] shadow-sm cursor-pointer hover:scale-105 transition-transform mb-2 relative">
                <span class="text-xs font-medium text-gray-500 mb-1 absolute -bottom-6">Battery</span>
                <svg xmlns="http://www.w3.org/2000/svg" class="h-8 w-8 text-gray-700 dark:text-gray-300 mb-1" fill="currentColor" viewBox="0 0 24 24"><path d="M4 18h14c1.103 0 2-.897 2-2v-2h2v-4h-2V8c0-1.103-.897-2-2-2H4c-1.103 0-2 .897-2 2v8c0 1.103.897 2 2 2zM4 8h14l.002 8H4V8z"/></svg>
                <div class="text-sm font-medium flex flex-col items-center leading-tight">
                  <span v-if="state?.battery_power_w !== null && state?.battery_power_w !== undefined" class="text-pink-500 dark:text-pink-400 text-xs">
                    &darr; {{ state.battery_power_w < 0 ? formatPowerSimple(Math.abs(state.battery_power_w / batteryDevices.length)) : '0 W' }}
                  </span>
                  <span v-else class="text-pink-500 dark:text-pink-400 text-xs">&darr; 0 W</span>

                  <span v-if="state?.battery_power_w !== null && state?.battery_power_w !== undefined" class="text-teal-500 dark:text-teal-400 text-xs">
                    &uarr; {{ state.battery_power_w > 0 ? formatPowerSimple(state.battery_power_w / batteryDevices.length) : '0 W' }}
                  </span>
                  <span v-else class="text-teal-500 dark:text-teal-400 text-xs">&uarr; 0 W</span>
                </div>
                <!-- Battery SOC Circle Overlay -->
                <div v-if="state?.battery_soc !== null && state?.battery_soc !== undefined" class="absolute inset-0 rounded-full border-[4px] border-[#EC4899] opacity-50" :style="`clip-path: polygon(0 ${100 - state.battery_soc}%, 100% ${100 - state.battery_soc}%, 100% 100%, 0 100%); border-color: #34D399; z-index: 20;`"></div>
              </div>
            </div>
          </template>
        </template>

        <!-- Home (Bottom Center) -->
        <div class="absolute flex flex-col items-center justify-center transform -translate-x-1/2 -translate-y-1/2" style="left: 85%; top: 50%;">
          <!-- Remove direct border, add conic gradient background based on percentages -->
          <div @click="openChart('home')" class="z-10 flex flex-col items-center justify-center w-[120px] h-[120px] rounded-full shadow-sm cursor-pointer hover:scale-105 transition-transform mb-2 relative" :style="homeBorderStyle">
            <span class="text-xs font-medium text-gray-500 mb-1 absolute -top-6">Home</span>
            <!-- Inner white circle to create the border effect -->
            <div class="absolute flex flex-col items-center justify-center w-[104px] h-[104px] bg-white dark:bg-gray-800 rounded-full top-1/2 left-1/2 transform -translate-x-1/2 -translate-y-1/2">
              <svg xmlns="http://www.w3.org/2000/svg" class="h-8 w-8 text-gray-700 dark:text-gray-300 mb-1" fill="currentColor" viewBox="0 0 24 24"><path d="M3 13h1v7c0 1.103.897 2 2 2h12c1.103 0 2-.897 2-2v-7h1a1 1 0 0 0 .707-1.707l-9-9a.999.999 0 0 0-1.414 0l-9 9A1 1 0 0 0 3 13zm7 7v-5h4v5h-4zm2-15.586 6 6V15l.001 5H16v-5c0-1.103-.897-2-2-2h-4c-1.103 0-2 .897-2 2v5H6v-9.586l6-6z"/></svg>
              <div class="text-gray-800 dark:text-gray-200 text-sm font-medium">
                <span v-if="state?.total_load_w !== null && state?.total_load_w !== undefined">
                   {{ formatPowerSimple(homeLoad) }}
                </span>
                <span v-else>0 W</span>
              </div>
            </div>
          </div>
        </div>

        <!-- Solar (Right) -->
        <template v-for="(device, index) in solarDevices" :key="device.id">
          <div class="absolute flex flex-col items-center justify-center transform -translate-x-1/2 -translate-y-1/2" :style="`left: ${50 + (index * 15 - (solarDevices.length-1)*7.5)}%; top: 20%;`">
            <div @click="openChart('solar')" class="z-10 flex flex-col items-center justify-center w-28 h-28 bg-white dark:bg-gray-800 rounded-full border-[4px] border-[#FBBF24] shadow-sm cursor-pointer hover:scale-105 transition-transform mb-2 relative">
              <span class="text-xs font-medium text-gray-500 mb-1 absolute -top-6">Solar</span>
              <svg xmlns="http://www.w3.org/2000/svg" class="h-8 w-8 text-gray-700 dark:text-gray-300 mb-1 relative" fill="currentColor" viewBox="0 0 24 24"><path d="M6.993 12c0 2.761 2.246 5.007 5.007 5.007s5.007-2.246 5.007-5.007S14.761 6.993 12 6.993 6.993 9.239 6.993 12zM12 8.993c1.658 0 3.007 1.349 3.007 3.007S13.658 15.007 12 15.007 8.993 13.658 8.993 12 10.342 8.993 12 8.993zM10.998 19h2v3h-2zm0-17h2v3h-2zm-9 9h3v2h-3zm17 0h3v2h-3zM4.219 18.363l2.12-2.122 1.415 1.414-2.12 2.122zM16.24 6.344l2.122-2.122 1.414 1.414-2.122 2.122zM6.342 7.759 4.22 5.637l1.415-1.414 2.12 2.122zm13.434 10.605-1.414 1.414-2.122-2.122 1.414-1.414z"/></svg>
              <div class="text-gray-800 dark:text-gray-200 text-sm font-medium">
                <span v-if="state?.solar_power_w !== null && state?.solar_power_w !== undefined">
                   {{ formatPowerSimple(state.solar_power_w / solarDevices.length) }}
                </span>
                <span v-else>0 W</span>
              </div>
            </div>
          </div>
        </template>

        <!-- Bottom Left: EV Charger -->
        <template v-for="(device, index) in evDevices" :key="device.id">
          <div class="absolute flex flex-col items-center justify-center transform -translate-x-1/2 -translate-y-1/2" :style="`left: 85%; top: ${80 + (index * 15)}%;`">
            <div @click="openChart('ev_charger')" class="z-10 flex flex-col items-center justify-center w-28 h-28 bg-white dark:bg-gray-800 rounded-full border-[4px] border-[#A855F7] shadow-sm cursor-pointer hover:scale-105 transition-transform mb-2 relative">
              <span class="text-xs font-medium text-gray-500 mb-1 absolute -bottom-6">EV Charger</span>
              <svg xmlns="http://www.w3.org/2000/svg" class="h-8 w-8 text-gray-700 dark:text-gray-300 mb-1" fill="currentColor" viewBox="0 0 24 24"><path d="m20.772 10.156-1.368-4.105A2.995 2.995 0 0 0 16.559 4H7.441a2.995 2.995 0 0 0-2.845 2.051l-1.368 4.105A2.003 2.003 0 0 0 2 12v5c0 .753.423 1.402 1.039 1.743-.013.066-.039.126-.039.195V21a1 1 0 0 0 1 1h1a1 1 0 0 0 1-1v-2h12v2a1 1 0 0 0 1 1h1a1 1 0 0 0 1-1v-2.062c0-.069-.026-.13-.039-.195A1.993 1.993 0 0 0 22 17v-5c0-.829-.508-1.541-1.228-1.844zM4 17v-5h16l.002 5H4zM7.441 6h9.117c.431 0 .813.274.949.684L18.613 10H5.387l1.105-3.316A1 1 0 0 1 7.441 6z"/><circle cx="6.5" cy="14.5" r="1.5"/><circle cx="17.5" cy="14.5" r="1.5"/></svg>
              <div class="text-gray-800 dark:text-gray-200 text-sm font-medium flex flex-col items-center">
                <span v-if="state?.ev_charger_power_w !== null && state?.ev_charger_power_w !== undefined">
                  {{ formatPowerSimple(state.ev_charger_power_w / evDevices.length) }}
                </span>
                <span v-else>0 W</span>
                <span class="text-[10px] font-bold text-purple-600 uppercase mt-0.5">{{ device.charge_mode || 'ECO' }}</span>
              </div>
            </div>
          </div>
        </template>
        <!-- Bottom Right: Appliance (Relay) -->
        <template v-for="(device, index) in applianceDevices" :key="device.id">
          <div class="absolute flex flex-col items-center justify-center transform -translate-x-1/2 -translate-y-1/2" :style="`left: 85%; top: ${20 - (index * 15)}%;`">
            <div @click="openChart('appliance')" class="z-10 flex flex-col items-center justify-center w-28 h-28 bg-white dark:bg-gray-800 rounded-full border-[4px] border-[#F97316] shadow-sm cursor-pointer hover:scale-105 transition-transform mb-2 relative">
              <span class="text-xs font-medium text-gray-500 mb-1 absolute -top-6">Appliance</span>
              <svg xmlns="http://www.w3.org/2000/svg" class="h-8 w-8 text-gray-700 dark:text-gray-300 mb-1" fill="none" viewBox="0 0 24 24" stroke="currentColor"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M13 10V3L4 14h7v7l9-11h-7z" /></svg>
              <div class="text-gray-800 dark:text-gray-200 text-sm font-medium">
                <span class="text-orange-500 font-bold uppercase tracking-wide">{{ device.status === 'online' ? 'ON' : 'OFF' }}</span>
              </div>
            </div>
          </div>
        </template>

      </div>

    </div>
  </div>

  <!-- Solar Forecast Modal -->
  <div v-if="showSolarModal" class="fixed inset-0 z-[100] flex justify-center items-center bg-black bg-opacity-50 transition-opacity" @click.self="closeSolarModal">
    <div class="w-full h-full bg-white dark:bg-gray-800 shadow-xl p-8 relative flex flex-col transform transition-transform duration-300 translate-x-0 overflow-y-auto" @click.stop>
      <button @click="closeSolarModal" class="absolute top-6 right-6 text-gray-500 hover:text-gray-700 dark:text-gray-400 dark:hover:text-gray-200 z-[110]">
        <svg class="h-8 w-8" fill="none" viewBox="0 0 24 24" stroke="currentColor">
          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12" />
        </svg>
      </button>

      <h2 class="text-3xl font-semibold mb-6 text-gray-800 dark:text-gray-100">Solar Forecast (Next 24h)</h2>

      <div class="flex-grow w-full h-full min-h-[500px]">
        <Line v-if="solarForecastData" :data="solarForecastData" :options="solarChartOptions" />
        <div v-else-if="isLoadingSolar" class="flex items-center justify-center h-full text-xl text-gray-500 dark:text-gray-400">Loading forecast...</div>
        <div v-else class="flex items-center justify-center h-full text-xl text-gray-500 dark:text-gray-400">No forecast data available (check location in Settings)</div>
      </div>
    </div>
  </div>

  <!-- Tariff Forecast Modal -->
  <div v-if="showTariffModal" class="fixed inset-0 z-[100] flex justify-center items-center bg-black bg-opacity-50 transition-opacity" @click.self="closeTariffModal">
    <div class="w-full h-full bg-white dark:bg-gray-800 shadow-xl p-8 relative flex flex-col transform transition-transform duration-300 translate-x-0 overflow-y-auto" @click.stop>
      <button @click="closeTariffModal" class="absolute top-6 right-6 text-gray-500 hover:text-gray-700 dark:text-gray-400 dark:hover:text-gray-200 z-[110]">
        <svg class="h-8 w-8" fill="none" viewBox="0 0 24 24" stroke="currentColor">
          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12" />
        </svg>
      </button>

      <h2 class="text-3xl font-semibold mb-6 text-gray-800 dark:text-gray-100">Tariff Forecast (Next 24h)</h2>

      <div class="flex-grow w-full h-full min-h-[500px]">
        <Bar v-if="tariffData" :data="tariffData" :options="tariffChartOptions" />
        <div v-else-if="isLoadingTariff" class="flex items-center justify-center h-full text-xl text-gray-500 dark:text-gray-400">Loading forecast...</div>
        <div v-else class="flex items-center justify-center h-full text-xl text-gray-500 dark:text-gray-400">No forecast data available</div>
      </div>
    </div>
  </div>

  <!-- Context-Sensitive Full-screen Panel -->
  <div v-if="isModalOpen" class="fixed inset-0 z-[100] flex justify-center items-center bg-black bg-opacity-50 transition-opacity" @click.self="closeChart">
    <div class="w-full h-full bg-white dark:bg-gray-800 shadow-xl p-8 relative flex flex-col transform transition-transform duration-300 translate-x-0 overflow-y-auto" @click.stop>
      <button @click="closeChart" class="absolute top-6 right-6 text-gray-500 hover:text-gray-700 dark:text-gray-400 dark:hover:text-gray-200 z-[110]">
        <svg class="h-8 w-8" fill="none" viewBox="0 0 24 24" stroke="currentColor">
          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12" />
        </svg>
      </button>

      <h2 class="text-3xl font-semibold mb-6 text-gray-800 dark:text-gray-100 capitalize">{{ selectedNode?.replace('_', ' ') }} History</h2>

      <div v-if="selectedNode === 'ev_charger' && evDevices && evDevices.length > 0" class="mb-6 bg-gray-50 dark:bg-gray-700/50 p-4 rounded-xl shadow-md border border-gray-200 dark:border-gray-700 flex flex-col items-center z-[105] relative">
        <h3 class="text-sm font-bold text-gray-600 dark:text-gray-300 uppercase tracking-wider mb-4 border-b border-gray-300 dark:border-gray-600 pb-2 w-full text-center">EV Charge Mode Control</h3>
        <div class="flex w-full gap-4 max-w-lg">
          <button @click="setEvMode('off')" :class="['flex-1 py-3 px-4 rounded-xl font-bold transition-all shadow-sm pointer-events-auto', evDevices[0]?.charge_mode === 'off' ? 'bg-purple-600 text-white ring-2 ring-purple-400 ring-offset-2 scale-[1.02]' : 'bg-white dark:bg-gray-800 text-gray-700 dark:text-gray-200 border border-gray-300 dark:border-gray-600 hover:bg-gray-100 dark:hover:bg-gray-700']">Off</button>
          <button @click="setEvMode('eco')" :class="['flex-1 py-3 px-4 rounded-xl font-bold transition-all shadow-sm pointer-events-auto', (evDevices[0]?.charge_mode === 'eco' || !evDevices[0]?.charge_mode) ? 'bg-purple-600 text-white ring-2 ring-purple-400 ring-offset-2 scale-[1.02]' : 'bg-white dark:bg-gray-800 text-gray-700 dark:text-gray-200 border border-gray-300 dark:border-gray-600 hover:bg-gray-100 dark:hover:bg-gray-700']">Eco</button>
          <button @click="setEvMode('now')" :class="['flex-1 py-3 px-4 rounded-xl font-bold transition-all shadow-sm pointer-events-auto', evDevices[0]?.charge_mode === 'now' ? 'bg-purple-600 text-white ring-2 ring-purple-400 ring-offset-2 scale-[1.02]' : 'bg-white dark:bg-gray-800 text-gray-700 dark:text-gray-200 border border-gray-300 dark:border-gray-600 hover:bg-gray-100 dark:hover:bg-gray-700']">Now</button>
        </div>
      </div>

      <div v-if="selectedNode === 'battery' && batteryDevices && batteryDevices.length > 0" class="mb-6 bg-gray-50 dark:bg-gray-700/50 p-4 rounded-xl shadow-md border border-gray-200 dark:border-gray-700 flex flex-col items-center z-[105] relative pointer-events-auto">
        <h3 class="text-sm font-bold text-gray-600 dark:text-gray-300 uppercase tracking-wider mb-4 border-b border-gray-300 dark:border-gray-600 pb-2 w-full text-center">Battery Operations Mode</h3>
        <div class="flex w-full gap-4 max-w-lg">
          <button @click="setBatteryMode('auto')" :class="['flex-1 py-3 px-4 rounded-xl font-bold transition-all shadow-sm pointer-events-auto', (batteryDevices[0]?.battery_mode === 'auto' || !batteryDevices[0]?.battery_mode) ? 'bg-emerald-500 text-white ring-2 ring-emerald-400 ring-offset-2 scale-[1.02]' : 'bg-white dark:bg-gray-800 text-gray-700 dark:text-gray-200 border border-gray-300 dark:border-gray-600 hover:bg-gray-100 dark:hover:bg-gray-700']">Auto</button>
          <button @click="setBatteryMode('hold')" :class="['flex-1 py-3 px-4 rounded-xl font-bold transition-all shadow-sm pointer-events-auto', batteryDevices[0]?.battery_mode === 'hold' ? 'bg-amber-500 text-white ring-2 ring-amber-400 ring-offset-2 scale-[1.02]' : 'bg-white dark:bg-gray-800 text-gray-700 dark:text-gray-200 border border-gray-300 dark:border-gray-600 hover:bg-gray-100 dark:hover:bg-gray-700']">Hold</button>
          <button @click="setBatteryMode('force_charge')" :class="['flex-1 py-3 px-4 rounded-xl font-bold transition-all shadow-sm pointer-events-auto', batteryDevices[0]?.battery_mode === 'force_charge' ? 'bg-blue-600 text-white ring-2 ring-blue-400 ring-offset-2 scale-[1.02]' : 'bg-white dark:bg-gray-800 text-gray-700 dark:text-gray-200 border border-gray-300 dark:border-gray-600 hover:bg-gray-100 dark:hover:bg-gray-700']">Force Charge</button>
        </div>
      </div>

      <div class="flex flex-wrap gap-4 mb-8">
        <button v-for="range in ranges" :key="range.value"
                @click="setRange(range.value)"
                :class="[
                  'px-4 py-2 rounded-lg text-lg font-medium transition-colors',
                  selectedRange === range.value
                    ? 'bg-blue-600 text-white'
                    : 'bg-gray-200 text-gray-700 hover:bg-gray-300 dark:bg-gray-700 dark:text-gray-200 dark:hover:bg-gray-600'
                ]">
          {{ range.label }}
        </button>
      </div>

      <div class="flex-grow w-full h-full min-h-[500px]">
        <Line v-if="chartData" :data="chartData" :options="chartOptions" />
        <div v-else-if="isLoading" class="flex items-center justify-center h-full text-xl text-gray-500 dark:text-gray-400">Loading data...</div>
        <div v-else class="flex items-center justify-center h-full text-xl text-gray-500 dark:text-gray-400">No data available</div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { getApiBase } from '../api'

import {
  Chart as ChartJS,
  CategoryScale,
  LinearScale,
  PointElement,
  LineElement,
  BarElement,
  Title,
  Tooltip,
  Legend,
  TimeScale,
} from 'chart.js'
import { Line, Bar } from 'vue-chartjs'
import 'chartjs-adapter-date-fns'

ChartJS.register(
  CategoryScale,
  LinearScale,
  PointElement,
  LineElement,
  BarElement,
  Title,
  Tooltip,
  Legend,
  TimeScale
)

interface SiteState {
  grid_power_w: number | null
  solar_power_w: number | null
  battery_power_w: number | null
  battery_soc: number | null
  total_load_w: number | null
  ev_charger_power_w: number | null
  device_health?: Record<number, string>
}

interface Device {
  id: number;
  name: string;
  template: string;
  has_battery?: boolean;
  charge_mode?: string;
  battery_mode?: string;
  status?: string;
}

const props = defineProps<{
  state: SiteState | null
}>()

const devices = ref<Device[]>([])

const fetchDevices = async () => {
  try {
    const res = await fetch(`${getApiBase()}/api/devices`)
    if (res.ok) {
      devices.value = await res.json() || []
    }
  } catch (e) {
    console.error("Failed to fetch devices for power flow map:", e)
  }
}

onMounted(() => {
  fetchDevices()
})

const gridDevices = computed(() => devices.value.filter(d => (d as any).category === 'meter' || ((d as any).category === 'inverter' && (d as any).has_grid_meter)))
const solarDevices = computed(() => devices.value.filter(d => (d as any).category === 'inverter'))
const batteryDevices = computed(() => devices.value.filter(d => d.has_battery))
const evDevices = computed(() => devices.value.filter(d => (d as any).category === 'charger'))
const applianceDevices = computed(() => devices.value.filter(d => (d as any).category === 'relay'))
const homeLoad = computed(() => {
  if (!props.state || props.state.total_load_w === null || props.state.total_load_w === undefined) return 0
  return props.state.total_load_w - (props.state.ev_charger_power_w || 0)
})

// Chart state
const isModalOpen = ref(false)
const selectedNode = ref<string | null>(null)
const selectedRange = ref<string>('today')
const isLoading = ref(false)
const chartData = ref<any>(null)

// Solar Forecast Modal state
const showSolarModal = ref(false)
const solarForecastData = ref<any>(null)
const isLoadingSolar = ref(false)

const openSolarModal = () => {
  showSolarModal.value = true
  fetchSolarForecast()
}

const closeSolarModal = () => {
  showSolarModal.value = false
  solarForecastData.value = null
}

const fetchSolarForecast = async () => {
  isLoadingSolar.value = true
  solarForecastData.value = null
  try {
    const res = await fetch(`${getApiBase()}/api/solar/forecast`)
    if (res.ok) {
      const data: {timestamp: string, estimated_power_w: number}[] = await res.json()
      if (data && data.length > 0) {
        solarForecastData.value = {
          datasets: [
            {
              label: 'Estimated Power (W)',
              data: data.map(d => ({ x: new Date(d.timestamp), y: d.estimated_power_w })),
              borderColor: '#F59E0B',
              backgroundColor: '#FDE68A',
              borderWidth: 2,
              fill: true,
              tension: 0.4
            }
          ]
        }
      }
    } else {
        console.error("Failed to fetch solar forecast, status: ", res.status)
    }
  } catch (e) {
    console.error("Failed to fetch solar forecast:", e)
  } finally {
    isLoadingSolar.value = false
  }
}

// Tariff Modal state
const showTariffModal = ref(false)
const tariffData = ref<any>(null)
const isLoadingTariff = ref(false)

const openTariffModal = () => {
  showTariffModal.value = true
  fetchTariffForecast()
}

const closeTariffModal = () => {
  showTariffModal.value = false
  tariffData.value = null
}

const fetchTariffForecast = async () => {
  isLoadingTariff.value = true
  tariffData.value = null
  try {
    const res = await fetch(`${getApiBase()}/api/tariffs/forecast`)
    if (res.ok) {
      const data: {timestamp: string, price_per_kwh: number}[] = await res.json()

      if (data && data.length > 0) {
        tariffData.value = {
          datasets: [
            {
              label: 'Price',
              data: data.map(d => ({ x: new Date(d.timestamp), y: d.price_per_kwh })),
              backgroundColor: data.map(d => {
                if (d.price_per_kwh <= 0) return '#10B981' // Green (Cheap/Negative)
                if (d.price_per_kwh >= 0.15) return '#EF4444' // Red (Expensive)
                return '#3B82F6' // Blue (Normal)
              }),
              borderWidth: 1,
              borderColor: data.map(d => {
                if (d.price_per_kwh <= 0) return '#059669'
                if (d.price_per_kwh >= 0.15) return '#DC2626'
                return '#2563EB'
              }),
            }
          ]
        }
      } else {
        tariffData.value = null
      }
    }
  } catch (e) {
    console.error("Failed to fetch tariff forecast:", e)
  } finally {
    isLoadingTariff.value = false
  }
}

const ranges = [
  { label: 'Today', value: 'today' },
  { label: 'Last 24h', value: '24h' },
  { label: 'Last 7 Days', value: '7d' },
  { label: 'Last 30 Days', value: '30d' },
]

// Use a dynamic maximum to calculate a relative flow percentage
const homeBorderStyle = computed(() => {
  if (!props.state || !homeLoad.value) {
    return { background: '#10B981' }; // default green
  }

  // Determine what is supplying the home load based on the same logic in activeSegments
  const homeLoadRemaining = homeLoad.value;
  if (homeLoadRemaining <= 0) return { background: '#10B981' };

  let remainingSolar = props.state.solar_power_w !== null ? Math.max(0, props.state.solar_power_w) : 0;
  let remainingBattery = props.state.battery_power_w !== null && props.state.battery_power_w > 0 ? props.state.battery_power_w : 0;
  let remainingGrid = props.state.grid_power_w !== null && props.state.grid_power_w > 0 ? props.state.grid_power_w : 0;

  const solarToHome = Math.min(remainingSolar, homeLoadRemaining);
  const batteryToHome = Math.min(remainingBattery, homeLoadRemaining - solarToHome);
  const gridToHome = Math.min(remainingGrid, homeLoadRemaining - solarToHome - batteryToHome);

  const totalSupplied = solarToHome + batteryToHome + gridToHome;

  if (totalSupplied <= 0) return { background: '#10B981' };

  const solarPct = (solarToHome / totalSupplied) * 100;
  const gridPct = (gridToHome / totalSupplied) * 100;
  const batteryPct = (batteryToHome / totalSupplied) * 100;

  // Colors: Solar #FBBF24, Grid #3B82F6, Battery #34D399 (or #10B981)
  let gradient = 'conic-gradient(';
  let currentStop = 0;

  if (solarPct > 0) {
    gradient += `#FBBF24 ${currentStop}% ${currentStop + solarPct}%, `;
    currentStop += solarPct;
  }
  if (gridPct > 0) {
    gradient += `#3B82F6 ${currentStop}% ${currentStop + gridPct}%, `;
    currentStop += gridPct;
  }
  if (batteryPct > 0) {
    gradient += `#34D399 ${currentStop}% ${currentStop + batteryPct}%, `;
    currentStop += batteryPct;
  }

  // fallback if less than 100% due to precision
  if (currentStop < 100) {
      gradient += `#10B981 ${currentStop}% 100%`;
  } else {
      gradient = gradient.slice(0, -2); // remove last comma and space
  }

  gradient += ')';

  return { background: gradient };
});

const flowPercentage = computed(() => {
  if (!props.state) return 0;

  // Find the maximum absolute power flowing through any node right now
  const grid = Math.abs(props.state.grid_power_w || 0);
  const solar = Math.abs(props.state.solar_power_w || 0);
  const battery = Math.abs(props.state.battery_power_w || 0);
  const home = Math.abs(homeLoad.value || 0);

  const maxVal = Math.max(grid, solar, battery, home, 100); // Floor at 100W to avoid 0 div
  return maxVal;
});

const getFlowStyle = (power: number | null | undefined, normalIsPositive: boolean = true) => {
  if (!power || Math.abs(power) < 10) return { display: 'none' }
  const absPower = Math.abs(power)

  // Faster animation for higher percentage of max power
  // Slower animation (e.g., 5s) for small flows. Base 2s. Max duration 5s, Min 0.5s.
  const percentage = absPower / (flowPercentage.value || 100);
  const duration = Math.min(Math.max(5 - (percentage * 4.5), 0.5), 5); // 0% = 5s, 100% = 0.5s


  let direction = 'normal'
  if (normalIsPositive && power < 0) direction = 'reverse'
  else if (!normalIsPositive && power >= 0) direction = 'reverse'

  return {
    animationDuration: `${duration}s`,
    animationDirection: direction
  }
}

interface Segment {
  id: string;
  path: string;
  power: number;
  color: string;
  normalIsPositive: boolean; // Determines direction mapping based on sign
}

const activeSegments = computed<Segment[]>(() => {
  if (!props.state) return [];

  const grid = props.state.grid_power_w !== null ? props.state.grid_power_w : 0;
  const solar = props.state.solar_power_w !== null ? Math.max(0, props.state.solar_power_w) : 0;
  const battery = props.state.battery_power_w !== null ? props.state.battery_power_w : 0;
  const home = homeLoad.value || 0;

  const segments: Segment[] = [];
  const evPower = props.state.ev_charger_power_w || 0;

  // Track remaining power to distribute
  let remainingSolar = solar;
  let remainingBatteryDischarge = battery > 0 ? battery : 0;
  let remainingGridImport = grid > 0 ? grid : 0;

  let homeLoadRemaining = home;
  let evLoadRemaining = evPower;
  let batteryChargeRemaining = battery < 0 ? Math.abs(battery) : 0;
  let gridExportRemaining = grid < 0 ? Math.abs(grid) : 0;

  // Direct connection helper
  const addSegment = (sourceType: string, targetType: string, powerFlow: number, color: string) => {
    if (powerFlow < 10) return;

    let path = '';

    // We assume 1 device for now for path logic, but keep loop for ID generation
    if (sourceType === 'solar') {
      solarDevices.value.forEach((d) => {
        if (targetType === 'home') path = `M 50 20 C 50 45, 55 45, 85 45`;
        else if (targetType === 'ev') {
            evDevices.value.forEach((ev) => {
                path = `M 50 20 C 50 45, 55 45, 85 45 L 85 80`;
                segments.push({
                  id: `solar-${targetType}-${d.id}-${ev.id}`,
                  path, power: (powerFlow / solarDevices.value.length) / evDevices.value.length, color, normalIsPositive: true
                });
            });
            return;
        }
        else if (targetType === 'battery') {
             batteryDevices.value.forEach((bat) => {
                 path = `M 50 20 C 40 40, 40 60, 50 80`; // Optional: solar to battery direct curve
                 segments.push({
                   id: `solar-${targetType}-${d.id}-${bat.id}`,
                   path, power: (powerFlow / solarDevices.value.length) / batteryDevices.value.length, color, normalIsPositive: true
                 });
             });
             return;
        }
        else if (targetType === 'grid') {
            gridDevices.value.forEach((gd) => {
                 path = `M 50 20 C 50 48, 45 48, 15 48`;
                 segments.push({
                   id: `solar-${targetType}-${d.id}-${gd.id}`,
                   path, power: (powerFlow / solarDevices.value.length) / gridDevices.value.length, color, normalIsPositive: true
                 });
             });
             return;
        }

        segments.push({
          id: `solar-${targetType}-${d.id}`,
          path, power: powerFlow / solarDevices.value.length, color, normalIsPositive: true
        });
      });
    } else if (sourceType === 'battery') {
       batteryDevices.value.forEach((d) => {
        if (targetType === 'home') path = `M 50 80 C 50 55, 55 55, 85 55`;
        else if (targetType === 'ev') {
            evDevices.value.forEach((ev) => {
                path = `M 50 80 C 50 55, 55 55, 85 55 L 85 80`;
                segments.push({
                  id: `battery-${targetType}-${d.id}-${ev.id}`,
                  path, power: (powerFlow / batteryDevices.value.length) / evDevices.value.length, color, normalIsPositive: true
                });
            });
            return;
        }
        else if (targetType === 'grid') {
             gridDevices.value.forEach((gd) => {
                 path = `M 50 80 C 45 52, 45 52, 15 52`; // Reverse of grid->battery
                 segments.push({
                   id: `battery-${targetType}-${d.id}-${gd.id}`,
                   path, power: (powerFlow / batteryDevices.value.length) / gridDevices.value.length, color, normalIsPositive: true
                 });
             });
             return;
        }

        segments.push({
          id: `battery-${targetType}-${d.id}`,
          path, power: powerFlow / batteryDevices.value.length, color, normalIsPositive: true
        });
      });
    } else if (sourceType === 'grid') {
      gridDevices.value.forEach((d) => {
        if (targetType === 'home') path = `M 15 50 L 85 50`;
        else if (targetType === 'ev') {
             evDevices.value.forEach((ev) => {
                path = `M 15 50 L 85 50 L 85 80`;
                segments.push({
                  id: `grid-${targetType}-${d.id}-${ev.id}`,
                  path, power: (powerFlow / gridDevices.value.length) / evDevices.value.length, color, normalIsPositive: true
                });
            });
            return;
        }
        else if (targetType === 'battery') {
             batteryDevices.value.forEach((bat) => {
                 path = `M 15 52 C 45 52, 45 52, 50 80`;
                 segments.push({
                   id: `grid-${targetType}-${d.id}-${bat.id}`,
                   path, power: (powerFlow / gridDevices.value.length) / batteryDevices.value.length, color, normalIsPositive: true
                 });
             });
             return;
        }

        segments.push({
          id: `grid-${targetType}-${d.id}`,
          path, power: powerFlow / gridDevices.value.length, color, normalIsPositive: true
        });
      });
    }
  };

  // 1. Solar fulfills loads first
  if (remainingSolar > 0) {
    // Solar -> Home
    const solarToHome = Math.min(remainingSolar, homeLoadRemaining);
    if (solarToHome > 0) {
      addSegment('solar', 'home', solarToHome, '#FBBF24');
      remainingSolar = Math.max(0, remainingSolar - solarToHome);
      homeLoadRemaining -= solarToHome;
    }
    // Solar -> EV
    const solarToEV = Math.min(remainingSolar, evLoadRemaining);
    if (solarToEV > 0) {
      addSegment('solar', 'ev', solarToEV, '#FBBF24');
      remainingSolar = Math.max(0, remainingSolar - solarToEV);
      evLoadRemaining -= solarToEV;
    }
    // Solar -> Battery
    const solarToBattery = Math.min(remainingSolar, batteryChargeRemaining);
    if (solarToBattery > 0) {
      addSegment('solar', 'battery', solarToBattery, '#FBBF24');
      remainingSolar = Math.max(0, remainingSolar - solarToBattery);
      batteryChargeRemaining -= solarToBattery;
    }
    // Solar -> Grid
    const solarToGrid = Math.min(remainingSolar, gridExportRemaining);
    if (solarToGrid > 0) {
      addSegment('solar', 'grid', solarToGrid, '#FBBF24');
      remainingSolar = Math.max(0, remainingSolar - solarToGrid);
      gridExportRemaining -= solarToGrid;
    }
  }

  // 2. Battery fulfills remaining loads
  if (remainingBatteryDischarge > 0) {
    // Battery -> Home
    const batteryToHome = Math.min(remainingBatteryDischarge, homeLoadRemaining);
    if (batteryToHome > 0) {
      addSegment('battery', 'home', batteryToHome, '#34D399');
      remainingBatteryDischarge = Math.max(0, remainingBatteryDischarge - batteryToHome);
      homeLoadRemaining -= batteryToHome;
    }
    // Battery -> EV
    const batteryToEV = Math.min(remainingBatteryDischarge, evLoadRemaining);
    if (batteryToEV > 0) {
      addSegment('battery', 'ev', batteryToEV, '#34D399');
      remainingBatteryDischarge = Math.max(0, remainingBatteryDischarge - batteryToEV);
      evLoadRemaining -= batteryToEV;
    }
    // Battery -> Grid (Edge case, usually avoided)
    const batteryToGrid = Math.min(remainingBatteryDischarge, gridExportRemaining);
    if (batteryToGrid > 0) {
      addSegment('battery', 'grid', batteryToGrid, '#34D399');
      remainingBatteryDischarge = Math.max(0, remainingBatteryDischarge - batteryToGrid);
      gridExportRemaining -= batteryToGrid;
    }
  }

  // 3. Grid fulfills any remaining loads
  if (remainingGridImport > 0) {
    // Grid -> Home
    const gridToHome = Math.min(remainingGridImport, homeLoadRemaining);
    if (gridToHome > 0) {
      addSegment('grid', 'home', gridToHome, '#3B82F6');
      remainingGridImport = Math.max(0, remainingGridImport - gridToHome);
      homeLoadRemaining -= gridToHome;
    }
    // Grid -> EV
    const gridToEV = Math.min(remainingGridImport, evLoadRemaining);
    if (gridToEV > 0) {
      addSegment('grid', 'ev', gridToEV, '#3B82F6');
      remainingGridImport = Math.max(0, remainingGridImport - gridToEV);
      evLoadRemaining -= gridToEV;
    }
    // Grid -> Battery
    const gridToBattery = Math.min(remainingGridImport, batteryChargeRemaining);
    if (gridToBattery > 0) {
      addSegment('grid', 'battery', gridToBattery, '#3B82F6');
      remainingGridImport = Math.max(0, remainingGridImport - gridToBattery);
      batteryChargeRemaining -= gridToBattery;
    }
  }

return segments;
});


const formatPowerSimple = (powerW: number) => {
  const absPower = Math.abs(powerW)
  if (absPower < 1000) {
    return `${Math.round(absPower)} W`
  }
  const valKw = (absPower / 1000).toFixed(1)
  return `${valKw} kW`
}

const solarChartOptions = computed(() => ({
  responsive: true,
  maintainAspectRatio: false,
  scales: {
    x: {
      type: 'time' as const,
      time: {
        unit: 'hour' as const,
        displayFormats: {
          hour: 'HH:mm',
        },
        tooltipFormat: 'PP HH:mm'
      },
      ticks: {
        color: '#9CA3AF',
      },
      grid: {
        color: '#374151',
      }
    },
    y: {
      title: {
        display: true,
        text: 'Estimated Power (W)',
        color: '#9CA3AF'
      },
      ticks: {
        color: '#9CA3AF',
      },
      grid: {
        color: '#374151',
      }
    }
  },
  plugins: {
    legend: {
      display: false
    },
    tooltip: {
      callbacks: {
        label: (context: any) => `${context.parsed.y.toFixed(0)} W`
      }
    }
  }
}))

const tariffChartOptions = computed(() => ({
  responsive: true,
  maintainAspectRatio: false,
  scales: {
    x: {
      type: 'time' as const,
      time: {
        unit: 'hour' as const,
        displayFormats: {
          hour: 'HH:mm',
        },
        tooltipFormat: 'PP HH:mm'
      },
      ticks: {
        color: '#9CA3AF',
      },
      grid: {
        color: '#374151',
      }
    },
    y: {
      title: {
        display: true,
        text: 'Price (€/kWh)',
        color: '#9CA3AF'
      },
      ticks: {
        color: '#9CA3AF',
      },
      grid: {
        color: '#374151',
      }
    }
  },
  plugins: {
    legend: {
      display: false
    },
    tooltip: {
      callbacks: {
        label: (context: any) => `${context.parsed.y.toFixed(3)} €/kWh`
      }
    }
  }
}))

const chartOptions = computed(() => ({
  responsive: true,
  maintainAspectRatio: false,
  scales: {
    x: {
      type: 'time' as const,
      time: {
        unit: selectedRange.value === 'today' || selectedRange.value === '24h' ? 'hour' as const : 'day' as const,
        displayFormats: {
          hour: 'HH:mm',
          minute: 'HH:mm'
        },
        tooltipFormat: 'HH:mm'
      },
      ticks: {
        color: '#9CA3AF',
      },
      grid: {
        color: '#374151',
      }
    },
    y: {
      title: {
        display: true,
        text: 'Power (W)',
        color: '#9CA3AF'
      },
      ticks: {
        color: '#9CA3AF',
      },
      grid: {
        color: '#374151',
      }
    }
  },
  plugins: {
    legend: {
      display: false
    },
    tooltip: {
      callbacks: {
        label: (context: any) => `${context.parsed.y.toFixed(0)} W`
      }
    }
  }
}))

const openChart = (node: string) => {
  selectedNode.value = node
  isModalOpen.value = true
  fetchHistory()
}

const closeChart = () => {
  isModalOpen.value = false
  selectedNode.value = null
  chartData.value = null
}

const setRange = (range: string) => {
  selectedRange.value = range
  fetchHistory()
}

const setEvMode = async (mode: string) => {
  for (const d of evDevices.value) {
    await setEvModeDevice(d, mode);
  }
}

const setEvModeDevice = async (d: any, mode: string) => {
  try {
    const res = await fetch(`${getApiBase()}/api/devices/${d.id}/mode`, {
      method: 'PUT',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ charge_mode: mode })
    });
    if (res.ok) {
      d.charge_mode = mode;
    }
  } catch (e) {
    console.error(`Failed to set EV mode for device ${d.id}`, e);
  }
}

const setBatteryMode = async (mode: string) => {
  for (const d of batteryDevices.value) {
    await setBatteryModeDevice(d, mode);
  }
}

const setBatteryModeDevice = async (d: any, mode: string) => {
  try {
    const res = await fetch(`${getApiBase()}/api/devices/${d.id}/mode`, {
      method: 'PUT',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ battery_mode: mode })
    });
    if (res.ok) {
      d.battery_mode = mode;
    }
  } catch (e) {
    console.error(`Failed to set battery mode for device ${d.id}`, e);
  }
}

const fetchHistory = async () => {
  if (!selectedNode.value) return
  isLoading.value = true
  chartData.value = null
  try {
    const res = await fetch(`${getApiBase()}/api/history?node=${selectedNode.value}&range=${selectedRange.value}`)
    if (res.ok) {
      const data: {timestamp: string, power_w: number}[] = await res.json()

      let lineColor = '#3B82F6' // Grid - Blue
      if (selectedNode.value === 'solar') lineColor = '#FBBF24' // Yellow
      else if (selectedNode.value === 'battery') lineColor = '#34D399' // Green
      else if (selectedNode.value === 'ev_charger') lineColor = '#A855F7' // Purple
      else if (selectedNode.value === 'home') lineColor = '#A855F7' // Purple

      if (data && data.length > 0) {
        chartData.value = {
          datasets: [
            {
              label: selectedNode.value,
              data: data.map(d => ({ x: new Date(d.timestamp), y: d.power_w })),
              borderColor: lineColor,
              backgroundColor: lineColor,
              borderWidth: 2,
              pointRadius: 1,
              pointHoverRadius: 4,
              fill: false,
              tension: 0.2
            }
          ]
        }
      } else {
        chartData.value = null // No data
      }
    }
  } catch (e) {
    console.error("Failed to fetch history:", e)
  } finally {
    isLoading.value = false
  }
}

</script>

<style scoped>
.flow-dot {
  animation: flow linear infinite;
}

@keyframes flow {
  from {
    stroke-dashoffset: 200;
  }
  to {
    stroke-dashoffset: 0;
  }
}
</style>
