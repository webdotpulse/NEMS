import { createApp } from 'vue'
import './style.css'
import App from './App.vue'
import { registerSW } from 'virtual:pwa-register'

const updateSW = registerSW({
  onNeedRefresh() {
    // Optionally show a prompt to user
    updateSW(true)
  },
  onOfflineReady() {
    console.log('App ready to work offline')
  },
})

createApp(App).mount('#app')
