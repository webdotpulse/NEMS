<template>
  <main class="max-w-7xl mx-auto py-6 sm:px-6 lg:px-8">
    <div class="px-4 py-6 sm:px-0 max-w-3xl mx-auto">
        <!-- Header -->
        <div class="mb-8">
            <h1 class="text-3xl font-bold flex items-center gap-3 text-gray-900 dark:text-white">
                <svg xmlns="http://www.w3.org/2000/svg" width="24" height="24" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round" class="w-8 h-8 text-gray-900 dark:text-white"><rect width="20" height="8" x="2" y="2" rx="2" ry="2"/><rect width="20" height="8" x="2" y="14" rx="2" ry="2"/><line x1="6" x2="6.01" y1="6" y2="6"/><line x1="6" x2="6.01" y1="18" y2="18"/></svg>
                System Update
            </h1>
            <p class="text-gray-500 dark:text-gray-400 mt-2">Manage your NEMS software version and install updates.</p>
        </div>

        <!-- Main Card -->
        <div class="bg-white dark:bg-gray-800 rounded-2xl shadow-sm border border-gray-200 dark:border-gray-700 overflow-hidden">

            <!-- Status Header -->
            <div class="p-6 border-b border-gray-100 dark:border-gray-700 flex flex-col md:flex-row items-start md:items-center justify-between gap-4"
                 :class="{'bg-blue-50/50 dark:bg-blue-900/10': isUpdateAvailable && !loading && !error, 'bg-emerald-50/50 dark:bg-emerald-900/10': !isUpdateAvailable && !loading && !error && latestRelease}">

                <div class="flex items-center gap-4">
                    <div class="w-12 h-12 rounded-full flex items-center justify-center shrink-0"
                         :class="{
                            'bg-blue-100 text-blue-600 dark:bg-blue-900/30 dark:text-blue-400': isUpdateAvailable && !loading && !error,
                            'bg-emerald-100 text-emerald-600 dark:bg-emerald-900/30 dark:text-emerald-400': !isUpdateAvailable && !loading && !error && latestRelease,
                            'bg-gray-100 text-gray-400 dark:bg-gray-700 dark:text-gray-500': loading,
                            'bg-red-100 text-red-600 dark:bg-red-900/30 dark:text-red-400': error
                         }">

                        <svg v-if="loading" xmlns="http://www.w3.org/2000/svg" width="24" height="24" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round" class="w-6 h-6 animate-spin"><path d="M3 12a9 9 0 1 0 9-9 9.75 9.75 0 0 0-6.74 2.74L3 8"/><path d="M3 3v5h5"/></svg>

                        <svg v-else-if="isUpdateAvailable && !error" xmlns="http://www.w3.org/2000/svg" width="24" height="24" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round" class="w-6 h-6"><circle cx="12" cy="12" r="10"/><path d="m16 12-4-4-4 4"/><path d="M12 16V8"/></svg>

                        <svg v-else-if="!isUpdateAvailable && !error && latestRelease" xmlns="http://www.w3.org/2000/svg" width="24" height="24" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round" class="w-6 h-6"><circle cx="12" cy="12" r="10"/><path d="m9 12 2 2 4-4"/></svg>

                        <svg v-else-if="error" xmlns="http://www.w3.org/2000/svg" width="24" height="24" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round" class="w-6 h-6"><circle cx="12" cy="12" r="10"/><line x1="12" x2="12" y1="8" y2="12"/><line x1="12" x2="12.01" y1="16" y2="16"/></svg>
                    </div>

                    <div>
                        <h2 class="text-lg font-semibold text-gray-900 dark:text-white">
                            <span v-if="loading">Checking for updates...</span>
                            <span v-else-if="error">Update Check Failed</span>
                            <span v-else-if="isUpdateAvailable">New Update Available!</span>
                            <span v-else-if="!isUpdateAvailable && latestRelease">System is up to date</span>
                            <span v-else>No releases found</span>
                        </h2>
                        <p class="text-sm text-gray-500 dark:text-gray-400" v-if="!loading && !error && latestRelease">
                            Last checked: {{ lastChecked }}
                        </p>
                    </div>
                </div>

                <button @click="checkUpdate" :disabled="loading || installing"
                        class="px-4 py-2 bg-white dark:bg-gray-800 border border-gray-200 dark:border-gray-600 text-gray-700 dark:text-gray-300 rounded-lg hover:bg-gray-50 dark:hover:bg-gray-700 transition flex items-center gap-2 disabled:opacity-50 text-sm font-medium">
                    <svg xmlns="http://www.w3.org/2000/svg" width="24" height="24" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round" class="w-4 h-4" :class="{'animate-spin': loading}"><path d="M3 12a9 9 0 1 0 9-9 9.75 9.75 0 0 0-6.74 2.74L3 8"/><path d="M3 3v5h5"/></svg>
                    Check Again
                </button>
            </div>

            <!-- Version Comparison -->
            <div class="p-6 grid grid-cols-1 md:grid-cols-2 gap-6 bg-gray-50 dark:bg-gray-900/50 border-b border-gray-100 dark:border-gray-700">
                <div class="bg-white dark:bg-gray-800 p-4 rounded-xl border border-gray-200 dark:border-gray-700 shadow-sm">
                    <div class="text-sm text-gray-500 dark:text-gray-400 mb-1 font-medium">Current Installed Version</div>
                    <div class="text-2xl font-mono font-semibold text-gray-900 dark:text-white">{{ currentVersion }}</div>
                </div>
                <div class="bg-white dark:bg-gray-800 p-4 rounded-xl border border-gray-200 dark:border-gray-700 shadow-sm relative overflow-hidden">
                    <div class="absolute top-0 left-0 w-1 h-full" :class="isUpdateAvailable ? 'bg-blue-500' : 'bg-emerald-500'"></div>
                    <div class="text-sm text-gray-500 dark:text-gray-400 mb-1 font-medium">Latest Available Version</div>
                    <div class="text-2xl font-mono font-semibold text-gray-900 dark:text-white flex items-center gap-3">
                        <span v-if="loading" class="text-gray-300 dark:text-gray-600">...</span>
                        <span v-else-if="latestRelease">{{ latestRelease.tag_name }}</span>
                        <span v-else class="text-gray-400 dark:text-gray-500 text-lg">N/A</span>

                        <span v-if="isUpdateAvailable && !loading" class="px-2.5 py-0.5 text-xs font-sans bg-blue-100 text-blue-700 dark:bg-blue-900/30 dark:text-blue-400 rounded-full font-bold">
                            NEW
                        </span>
                    </div>
                </div>
            </div>

            <!-- Error State -->
            <div v-if="error" class="p-6 text-red-600 bg-red-50 dark:bg-red-900/10 border-b border-red-100 dark:border-red-900/20">
                <p><strong>Error:</strong> {{ error }}</p>
                <p class="text-sm mt-2">Make sure the repository is public or you have provided a valid GitHub token.</p>
            </div>

            <!-- Release Details -->
            <div v-if="latestRelease && !loading && !error" class="p-6">
                <div class="flex items-center justify-between mb-4">
                    <h3 class="text-lg font-bold text-gray-800 dark:text-gray-100">{{ latestRelease.name || latestRelease.tag_name }}</h3>
                    <div class="text-sm text-gray-500 dark:text-gray-400">Published {{ formatDate(latestRelease.published_at) }}</div>
                </div>

                <!-- Markdown Content rendered with Tailwind Typography -->
                <div class="prose prose-sm prose-slate dark:prose-invert max-w-none bg-gray-50 dark:bg-gray-900/50 p-6 rounded-xl border border-gray-200 dark:border-gray-700"
                     v-html="parsedBody">
                </div>

                <!-- Actions -->
                <div class="mt-6 pt-6 border-t border-gray-100 dark:border-gray-700 flex flex-col sm:flex-row gap-3">
                    <button v-if="isUpdateAvailable" @click="installUpdate" :disabled="installing" class="px-6 py-2.5 bg-blue-600 hover:bg-blue-700 disabled:bg-blue-400 text-white font-medium rounded-lg shadow-sm transition flex items-center justify-center gap-2">
                        <svg v-if="!installing" xmlns="http://www.w3.org/2000/svg" width="24" height="24" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round" class="w-5 h-5"><path d="M12 13v8l-4-4"/><path d="m12 21 4-4"/><path d="M4.393 15.269A7 7 0 1 1 15.71 8h1.79a4.5 4.5 0 0 1 2.436 8.284"/></svg>
                        <svg v-else xmlns="http://www.w3.org/2000/svg" width="24" height="24" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round" class="w-5 h-5 animate-spin"><path d="M3 12a9 9 0 1 0 9-9 9.75 9.75 0 0 0-6.74 2.74L3 8"/><path d="M3 3v5h5"/></svg>
                        {{ installing ? 'Installing (System will reboot)...' : 'Install Update Automatically' }}
                    </button>

                    <a :href="latestRelease.html_url" target="_blank"
                       class="px-6 py-2.5 bg-white dark:bg-gray-800 border border-gray-200 dark:border-gray-600 text-gray-700 dark:text-gray-300 hover:bg-gray-50 dark:hover:bg-gray-700 font-medium rounded-lg shadow-sm transition flex items-center justify-center gap-2">
                        <svg xmlns="http://www.w3.org/2000/svg" width="24" height="24" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round" class="w-5 h-5"><path d="M15 3h6v6"/><path d="M10 14 21 3"/><path d="M18 13v6a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2V8a2 2 0 0 1 2-2h6"/></svg>
                        View on GitHub
                    </a>
                </div>

                <!-- Assets List -->
                <div v-if="latestRelease.assets && latestRelease.assets.length > 0" class="mt-6">
                    <h4 class="text-sm font-semibold text-gray-700 dark:text-gray-300 mb-3 uppercase tracking-wider">Download Assets</h4>
                    <div class="flex flex-col gap-2">
                        <a v-for="asset in latestRelease.assets" :key="asset.id" :href="asset.browser_download_url"
                           class="flex items-center justify-between p-3 rounded-lg border border-gray-200 dark:border-gray-700 hover:border-blue-300 dark:hover:border-blue-700 hover:bg-blue-50 dark:hover:bg-blue-900/10 transition text-sm">
                            <div class="flex items-center gap-3">
                                <svg xmlns="http://www.w3.org/2000/svg" width="24" height="24" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round" class="w-4 h-4 text-gray-400"><path d="m7.5 4.27 9 5.15"/><path d="M21 8a2 2 0 0 0-1-1.73l-7-4a2 2 0 0 0-2 0l-7 4A2 2 0 0 0 3 8v8a2 2 0 0 0 1 1.73l7 4a2 2 0 0 0 2 0l7-4A2 2 0 0 0 21 16Z"/><path d="m3.3 7 8.7 5 8.7-5"/><path d="M12 22V12"/></svg>
                                <span class="font-medium text-blue-600 dark:text-blue-400">{{ asset.name }}</span>
                            </div>
                            <span class="text-gray-500 dark:text-gray-400">{{ formatBytes(asset.size) }}</span>
                        </a>
                    </div>
                </div>
            </div>
        </div>
    </div>
  </main>
</template>

<script setup lang="ts">
import { ref, onMounted, computed } from 'vue';
import { marked } from 'marked';
import { getApiBase } from '../api';

const repository = 'webdotpulse/nems'; // Based on github url in memory
const currentVersion = ref('v0.0.0');

const latestRelease = ref<any>(null);
const loading = ref(true);
const error = ref<string | null>(null);
const lastChecked = ref('');
const installing = ref(false);
const githubToken = ref('');

const fetchSystemInfo = async () => {
  try {
    const res = await fetch(`${getApiBase()}/api/system/info`);
    if (res.ok) {
      const data = await res.json();
      if (data.build) {
        currentVersion.value = data.build;
      }
    }
  } catch (e) {
    console.error("Failed to fetch system info:", e);
  }
};

const fetchSettings = async () => {
    try {
        const res = await fetch(`${getApiBase()}/api/settings`);
        if (res.ok) {
            const data = await res.json();
            if (data.github_token) {
                githubToken.value = data.github_token;
            }
        }
    } catch (e) {
        console.error("Failed to fetch settings for github token:", e);
    }
};

const checkUpdate = async () => {
    loading.value = true;
    error.value = null;

    try {
        const headers: Record<string, string> = {
            'Accept': 'application/vnd.github.v3+json'
        };
        if (githubToken.value) {
            headers['Authorization'] = `Bearer ${githubToken.value}`;
        }
        const response = await fetch(`https://api.github.com/repos/${repository}/releases`, {
            headers
        });

        if (response.status === 404) {
            throw new Error(`No releases found for ${repository} yet.`);
        }

        if (!response.ok) {
            throw new Error(`API returned ${response.status}: ${response.statusText}`);
        }

        const data = await response.json();
        if (data.length > 0) {
            latestRelease.value = data[0];
        } else {
            throw new Error(`No releases found for ${repository} yet.`);
        }

        lastChecked.value = new Date().toLocaleTimeString();
    } catch (e: any) {
        error.value = e.message;
        latestRelease.value = null;
    } finally {
        loading.value = false;
    }
};

const installUpdate = async () => {
  if (confirm("Are you sure you want to download and install the new update? The system will reboot automatically when finished.")) {
    installing.value = true;
    try {
      const res = await fetch(`${getApiBase()}/api/system/update/install`, { method: 'POST' });
      if (!res.ok) {
        error.value = "Failed to start update process.";
        installing.value = false;
      }
    } catch (e) {
      console.error("Failed to install update:", e);
      error.value = "Failed to start update process.";
      installing.value = false;
    }
  }
};

onMounted(async () => {
    await fetchSystemInfo();
    await fetchSettings();
    checkUpdate();
});

const parsedBody = computed(() => {
    if (!latestRelease.value || !latestRelease.value.body) return '<p><i>No release notes provided.</i></p>';
    return marked.parse(latestRelease.value.body) as string;
});

const isUpdateAvailable = computed(() => {
    if (!latestRelease.value) return false;
    return currentVersion.value !== latestRelease.value.tag_name && currentVersion.value !== 'development';
});

const formatDate = (dateString: string) => {
    if (!dateString) return '';
    return new Date(dateString).toLocaleDateString(undefined, {
        year: 'numeric', month: 'long', day: 'numeric', hour: '2-digit', minute:'2-digit'
    });
};

const formatBytes = (bytes: number, decimals = 2) => {
    if (!+bytes) return '0 Bytes';
    const k = 1024;
    const dm = decimals < 0 ? 0 : decimals;
    const sizes = ['Bytes', 'KB', 'MB', 'GB'];
    const i = Math.floor(Math.log(bytes) / Math.log(k));
    return `${parseFloat((bytes / Math.pow(k, i)).toFixed(dm))} ${sizes[i]}`;
};

</script>