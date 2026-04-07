import { createApp } from 'vue'
import './style.css'
import App from './App.vue'
import { registerSW } from 'virtual:pwa-register'
import i18n from './i18n'

const updateSW = registerSW({
  onNeedRefresh() {
    // Optionally show a prompt to user
    updateSW(true)
  },
  onOfflineReady() {
    console.log('App ready to work offline')
  },
})

createApp(App).use(i18n).mount('#app')
