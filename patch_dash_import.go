--- frontend/src/components/Dashboard.vue
+++ frontend/src/components/Dashboard.vue
@@ -127,7 +127,7 @@
 import PowerFlow from './PowerFlow.vue'
 import { Bar } from 'vue-chartjs'
 import { Chart as ChartJS, Title, Tooltip, Legend, BarElement, CategoryScale, LinearScale } from 'chart.js'
-import { format, parseISO } from 'date-fns'
+import { format } from 'date-fns'

 // Register ChartJS components
 ChartJS.register(Title, Tooltip, Legend, BarElement, CategoryScale, LinearScale)
