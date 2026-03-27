--- frontend/src/components/Settings.vue
+++ frontend/src/components/Settings.vue
@@ -32,6 +32,31 @@
                   </label>
                 </div>

+              </div>
+
+              <div class="mt-8 mb-4 border-t border-gray-200 dark:border-gray-700 pt-6">
+                <h3 class="text-lg font-medium leading-6 text-gray-900 dark:text-gray-100">
+                  Dynamic Tariffs
+                </h3>
+                <p class="mt-1 text-sm text-gray-500 dark:text-gray-400">
+                  Configure rules for dynamic energy pricing.
+                </p>
+              </div>
+
+              <div class="grid grid-cols-1 gap-y-6 gap-x-4 sm:grid-cols-6">
+                <div class="sm:col-span-6">
+                  <label for="force_charge_below_euro" class="block text-sm font-medium text-gray-700 dark:text-gray-300">Force Charge Battery if price drops below (€/kWh)</label>
+                  <div class="mt-1">
+                    <input type="number" step="0.01" id="force_charge_below_euro" v-model="siteSettings.force_charge_below_euro" class="shadow-sm focus:ring-indigo-500 focus:border-indigo-500 block w-full sm:text-sm border-gray-300 rounded-md dark:bg-gray-700 dark:border-gray-600 dark:text-white" />
+                  </div>
+                </div>
+
+                <div class="sm:col-span-6">
+                  <label for="smart_ev_cheapest_hours" class="block text-sm font-medium text-gray-700 dark:text-gray-300">Smart EV Charging: Charge during the cheapest hours of the day</label>
+                  <div class="mt-1">
+                    <input type="number" step="1" min="0" max="24" id="smart_ev_cheapest_hours" v-model="siteSettings.smart_ev_cheapest_hours" class="shadow-sm focus:ring-indigo-500 focus:border-indigo-500 block w-full sm:text-sm border-gray-300 rounded-md dark:bg-gray-700 dark:border-gray-600 dark:text-white" />
+                  </div>
+                </div>
               </div>

               <div class="pt-5 border-t border-gray-200 dark:border-gray-700">
@@ -216,7 +241,9 @@
 const siteSettings = ref({
   strategy_mode: 'eco',
   capacity_peak_limit_kw: 2.5,
-  active_inverter_curtailment: false
+  active_inverter_curtailment: false,
+  force_charge_below_euro: 0.0,
+  smart_ev_cheapest_hours: 0
 })
 const saveSettingsSuccess = ref(false)
