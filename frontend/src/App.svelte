<script>
import BackupView from './lib/BackupView.svelte';
import RestoreView from './lib/RestoreView.svelte';
import { getT, getLang, toggleLang } from './lib/i18n.svelte.js';

let activeTab = $state('backup');
</script>

<div class="min-h-screen bg-base-100 flex flex-col">
  <header class="navbar bg-base-200 shadow-lg px-6">
    <div class="flex-1 flex items-center gap-3">
      <svg xmlns="http://www.w3.org/2000/svg" class="h-8 w-8 text-primary" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2">
        <path stroke-linecap="round" stroke-linejoin="round" d="M12 15v2m-6 4h12a2 2 0 002-2v-6a2 2 0 00-2-2H6a2 2 0 00-2 2v6a2 2 0 002 2zm10-10V7a4 4 0 00-8 0v4h8z" />
      </svg>
      <h1 class="text-xl font-bold text-base-content">{getT().appTitle}</h1>
    </div>
    <div class="flex-none">
      <button class="btn btn-ghost btn-sm" onclick={toggleLang}>
        {getLang() === 'en' ? '中文' : 'EN'}
      </button>
    </div>
  </header>

  <main class="flex-1 flex flex-col">
    <div class="tabs tabs-border justify-center px-6 pt-4">
      <button
        class="tab {activeTab === 'backup' ? 'tab-active' : ''}"
        onclick={() => activeTab = 'backup'}
      >
        <svg xmlns="http://www.w3.org/2000/svg" class="h-4 w-4 mr-1" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2">
          <path stroke-linecap="round" stroke-linejoin="round" d="M4 16v1a3 3 0 003 3h10a3 3 0 003-3v-1m-4-8l-4-4m0 0L8 8m4-4v12" />
        </svg>
        {getT().backupTab}
      </button>
      <button
        class="tab {activeTab === 'restore' ? 'tab-active' : ''}"
        onclick={() => activeTab = 'restore'}
      >
        <svg xmlns="http://www.w3.org/2000/svg" class="h-4 w-4 mr-1" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2">
          <path stroke-linecap="round" stroke-linejoin="round" d="M4 16v1a3 3 0 003 3h10a3 3 0 003-3v-1m-4-4l-4 4m0 0l-4-4m4 4V4" />
        </svg>
        {getT().restoreTab}
      </button>
    </div>

    <div class="flex-1 p-6">
      {#if activeTab === 'backup'}
        <BackupView />
      {:else}
        <RestoreView />
      {/if}
    </div>
  </main>

  <footer class="footer footer-center p-4 bg-base-200 text-base-content/60 text-sm">
    <p>{getT().footer}</p>
  </footer>
</div>
