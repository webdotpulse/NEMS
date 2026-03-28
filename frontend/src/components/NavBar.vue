<template>
  <nav class="bg-white border-b border-gray-200 dark:bg-gray-800 dark:border-gray-700">
    <div class="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
      <div class="flex justify-between h-16">
        <div class="flex">
          <div class="-ml-2 mr-2 flex items-center sm:hidden">
            <!-- Mobile menu button -->
            <button @click="mobileMenuOpen = !mobileMenuOpen" type="button" class="inline-flex items-center justify-center p-2 rounded-md text-gray-400 hover:text-gray-500 hover:bg-gray-100 focus:outline-none focus:ring-2 focus:ring-inset focus:ring-indigo-500 dark:hover:text-gray-300 dark:hover:bg-gray-700" aria-expanded="false">
              <span class="sr-only">Open main menu</span>
              <svg v-if="!mobileMenuOpen" class="block h-6 w-6" xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" stroke="currentColor" aria-hidden="true">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M4 6h16M4 12h16M4 18h16" />
              </svg>
              <svg v-else class="block h-6 w-6" xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" stroke="currentColor" aria-hidden="true">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12" />
              </svg>
            </button>
          </div>
          <div class="flex-shrink-0 flex items-center">
            <span class="text-xl font-bold text-gray-900 dark:text-white mr-8">Pulse EMS</span>
          </div>
          <div class="hidden sm:-my-px sm:ml-6 sm:flex sm:space-x-8">
            <a href="#"
               @click.prevent="navigate('dashboard')"
               :class="[currentView === 'dashboard' ? 'border-indigo-500 text-gray-900 dark:text-white' : 'border-transparent text-gray-500 hover:text-gray-700 hover:border-gray-300 dark:text-gray-400 dark:hover:text-gray-300', 'inline-flex items-center px-1 pt-1 border-b-2 text-sm font-medium']">
              Dashboard
            </a>
            <a href="#"
               @click.prevent="navigate('settings')"
               :class="[currentView === 'settings' ? 'border-indigo-500 text-gray-900 dark:text-white' : 'border-transparent text-gray-500 hover:text-gray-700 hover:border-gray-300 dark:text-gray-400 dark:hover:text-gray-300', 'inline-flex items-center px-1 pt-1 border-b-2 text-sm font-medium']">
              Settings
            </a>
            <a href="#"
               @click.prevent="navigate('logger')"
               :class="[currentView === 'logger' ? 'border-indigo-500 text-gray-900 dark:text-white' : 'border-transparent text-gray-500 hover:text-gray-700 hover:border-gray-300 dark:text-gray-400 dark:hover:text-gray-300', 'inline-flex items-center px-1 pt-1 border-b-2 text-sm font-medium']">
              Logger
            </a>
            <a href="#"
               @click.prevent="navigate('scanner')"
               :class="[currentView === 'scanner' ? 'border-indigo-500 text-gray-900 dark:text-white' : 'border-transparent text-gray-500 hover:text-gray-700 hover:border-gray-300 dark:text-gray-400 dark:hover:text-gray-300', 'inline-flex items-center px-1 pt-1 border-b-2 text-sm font-medium']">
              Scanner
            </a>
          </div>
        </div>
        <div class="flex items-center space-x-4">
          <!-- Connection Status -->
          <div class="flex items-center space-x-2">
            <span class="relative flex h-3 w-3">
              <span v-if="connected" class="animate-ping absolute inline-flex h-full w-full rounded-full bg-green-400 opacity-75"></span>
              <span class="relative inline-flex rounded-full h-3 w-3" :class="connected ? 'bg-green-500' : 'bg-red-500'"></span>
            </span>
            <span class="text-sm font-medium text-gray-700 dark:text-gray-300">
              {{ connected ? 'Connected' : 'Disconnected' }}
            </span>
          </div>

          <!-- Theme Toggle -->
          <button @click="toggleTheme" class="p-2 rounded-md text-gray-400 hover:text-gray-500 focus:outline-none focus:ring-2 focus:ring-inset focus:ring-indigo-500 dark:hover:text-gray-300">
            <span class="sr-only">Toggle dark mode</span>
            <svg v-if="isDark" class="h-6 w-6" xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" stroke="currentColor">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 3v1m0 16v1m9-9h-1M4 12H3m15.364 6.364l-.707-.707M6.343 6.343l-.707-.707m12.728 0l-.707.707M6.343 17.657l-.707.707M16 12a4 4 0 11-8 0 4 4 0 018 0z" />
            </svg>
            <svg v-else class="h-6 w-6" xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" stroke="currentColor">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M20.354 15.354A9 9 0 018.646 3.646 9.003 9.003 0 0012 21a9.003 9.003 0 008.354-5.646z" />
            </svg>
          </button>
        </div>
      </div>
    </div>

    <!-- Mobile menu, show/hide based on menu state. -->
    <div v-show="mobileMenuOpen" class="sm:hidden">
      <div class="pt-2 pb-3 space-y-1">
        <a href="#"
           @click.prevent="navigate('dashboard')"
           :class="[currentView === 'dashboard' ? 'bg-indigo-50 border-indigo-500 text-indigo-700 dark:bg-gray-700 dark:border-indigo-500 dark:text-white' : 'border-transparent text-gray-500 hover:bg-gray-50 hover:border-gray-300 hover:text-gray-700 dark:text-gray-400 dark:hover:bg-gray-700 dark:hover:text-gray-300', 'block pl-3 pr-4 py-2 border-l-4 text-base font-medium']">
          Dashboard
        </a>
        <a href="#"
           @click.prevent="navigate('settings')"
           :class="[currentView === 'settings' ? 'bg-indigo-50 border-indigo-500 text-indigo-700 dark:bg-gray-700 dark:border-indigo-500 dark:text-white' : 'border-transparent text-gray-500 hover:bg-gray-50 hover:border-gray-300 hover:text-gray-700 dark:text-gray-400 dark:hover:bg-gray-700 dark:hover:text-gray-300', 'block pl-3 pr-4 py-2 border-l-4 text-base font-medium']">
          Settings
        </a>
        <a href="#"
           @click.prevent="navigate('logger')"
           :class="[currentView === 'logger' ? 'bg-indigo-50 border-indigo-500 text-indigo-700 dark:bg-gray-700 dark:border-indigo-500 dark:text-white' : 'border-transparent text-gray-500 hover:bg-gray-50 hover:border-gray-300 hover:text-gray-700 dark:text-gray-400 dark:hover:bg-gray-700 dark:hover:text-gray-300', 'block pl-3 pr-4 py-2 border-l-4 text-base font-medium']">
          Logger
        </a>
        <a href="#"
           @click.prevent="navigate('scanner')"
           :class="[currentView === 'scanner' ? 'bg-indigo-50 border-indigo-500 text-indigo-700 dark:bg-gray-700 dark:border-indigo-500 dark:text-white' : 'border-transparent text-gray-500 hover:bg-gray-50 hover:border-gray-300 hover:text-gray-700 dark:text-gray-400 dark:hover:bg-gray-700 dark:hover:text-gray-300', 'block pl-3 pr-4 py-2 border-l-4 text-base font-medium']">
          Scanner
        </a>
      </div>
    </div>
  </nav>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue';

const props = defineProps<{
  connected: boolean;
  currentView: string;
}>();

const emit = defineEmits(['navigate']);

const mobileMenuOpen = ref(false);

const navigate = (view: string) => {
  emit('navigate', view);
  mobileMenuOpen.value = false;
};

const isDark = ref(false);

const toggleTheme = () => {
  isDark.value = !isDark.value;
  if (isDark.value) {
    document.documentElement.classList.add('dark');
  } else {
    document.documentElement.classList.remove('dark');
  }
};

onMounted(() => {
  // Check system preference on load
  if (window.matchMedia && window.matchMedia('(prefers-color-scheme: dark)').matches) {
    isDark.value = true;
    document.documentElement.classList.add('dark');
  }
});
</script>
