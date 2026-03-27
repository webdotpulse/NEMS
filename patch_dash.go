--- frontend/src/components/Dashboard.vue
+++ frontend/src/components/Dashboard.vue
@@ -107,6 +107,17 @@
           </p>
         </div>
       </div>
+
+      <!-- Tariffs Section -->
+      <div class="mt-8 mb-8" v-if="chartData">
+        <h2 class="text-2xl font-bold leading-7 text-gray-900 dark:text-white sm:text-3xl sm:truncate mb-4">
+          Tariffs
+        </h2>
+        <div class="bg-white dark:bg-gray-800 shadow sm:rounded-lg p-6">
+          <Bar :data="chartData" :options="chartOptions" class="h-64" />
+        </div>
+      </div>
+
     </div>
   </main>
 </template>
@@ -114,6 +125,12 @@
 <script setup lang="ts">
 import { ref, onMounted, onUnmounted } from 'vue'
 import PowerFlow from './PowerFlow.vue'
+import { Bar } from 'vue-chartjs'
+import { Chart as ChartJS, Title, Tooltip, Legend, BarElement, CategoryScale, LinearScale } from 'chart.js'
+import { format, parseISO } from 'date-fns'
+
+// Register ChartJS components
+ChartJS.register(Title, Tooltip, Legend, BarElement, CategoryScale, LinearScale)

 interface SiteState {
   grid_power_w: number | null
@@ -124,11 +141,61 @@
   device_health?: Record<number, string>
 }

+interface PricePoint {
+  timestamp: string
+  price_per_kwh: number
+}
+
 const state = ref<SiteState | null>(null)
 let eventSource: EventSource | null = null

-onMounted(() => {
+const chartData = ref<any>(null)
+const chartOptions = ref<any>({
+  responsive: true,
+  maintainAspectRatio: false,
+  scales: {
+    y: {
+      title: {
+        display: true,
+        text: 'Price (cents/kWh)',
+        color: '#9CA3AF'
+      },
+      grid: {
+        color: '#374151'
+      },
+      ticks: {
+        color: '#9CA3AF'
+      }
+    },
+    x: {
+      grid: {
+        display: false
+      },
+      ticks: {
+        color: '#9CA3AF',
+        maxRotation: 45,
+        minRotation: 45
+      }
+    }
+  },
+  plugins: {
+    legend: {
+      display: false
+    },
+    tooltip: {
+      callbacks: {
+        label: function(context: any) {
+          return `${context.parsed.y.toFixed(2)} cents/kWh`
+        }
+      }
+    }
+  }
+})
+
+onMounted(async () => {
   const host = window.location.hostname
+
+  // SSE Connection
   eventSource = new EventSource(`http://${host}:8080/api/live`)

   eventSource.onopen = () => {
@@ -148,6 +215,64 @@
     console.error('SSE Error:', error)
     // Don't null out state on every error, wait to see if it reconnects
   }
+
+  // Fetch settings to know thresholds
+  let siteSettings: any = null
+  try {
+    const res = await fetch(`http://${host}:8080/api/settings`)
+    if (res.ok) {
+      siteSettings = await res.json()
+    }
+  } catch (e) {
+    console.error("Failed to fetch settings:", e)
+  }
+
+  // Fetch today's prices
+  try {
+    const res = await fetch(`http://${host}:8080/api/tariffs/today`)
+    if (res.ok) {
+      const prices: PricePoint[] = await res.json()
+      if (prices && prices.length > 0) {
+
+        // Filter strictly to current day's 24 hours (today 00:00 to 23:00)
+        // The API might return more, depending on local time. Just grab the first 24 entries that match today.
+        const labels = prices.slice(0, 24).map(p => {
+          // Try to extract HH:mm from ISO
+          const date = new Date(p.timestamp)
+          return format(date, 'HH:mm')
+        })
+
+        const data = prices.slice(0, 24).map(p => p.price_per_kwh * 100) // Convert to cents
+
+        // Sort prices to find 90th percentile (top 10% most expensive)
+        const sortedData = [...data].sort((a, b) => a - b)
+        const p90Index = Math.floor(sortedData.length * 0.9)
+        const p90Value = sortedData[p90Index]
+
+        const thresholdEuro = siteSettings?.force_charge_below_euro || 0
+        const thresholdCents = thresholdEuro * 100
+
+        const backgroundColors = data.map(price => {
+          if (price < 0 || price < thresholdCents) return 'rgba(34, 197, 94, 0.8)' // Green
+          if (price >= p90Value) return 'rgba(239, 68, 68, 0.8)' // Red
+          return 'rgba(156, 163, 175, 0.8)' // Grey
+        })
+
+        chartData.value = {
+          labels,
+          datasets: [
+            {
+              label: 'Price (cents/kWh)',
+              backgroundColor: backgroundColors,
+              data
+            }
+          ]
+        }
+      }
+    }
+  } catch (e) {
+    console.error("Failed to fetch tariffs:", e)
+  }
 })

 onUnmounted(() => {
