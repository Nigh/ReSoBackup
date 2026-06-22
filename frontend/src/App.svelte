<script>
import BackupView from './lib/BackupView.svelte';
import RestoreView from './lib/RestoreView.svelte';
import { getT, getLang, toggleLang } from './lib/i18n.svelte.js';
import { Browser } from '@wailsio/runtime';

let activeTab = $state('backup');

function openGitHub() {
  Browser.OpenURL('https://github.com/Nigh/ReSoBackup');
}
</script>

<div class="min-h-screen bg-base-100 flex flex-col">
  <header class="navbar bg-base-200 shadow-lg px-6">
    <div class="flex-1 flex items-center gap-3">
      <svg xmlns="http://www.w3.org/2000/svg" class="h-8 w-8 text-primary" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2">
        <path stroke-linecap="round" stroke-linejoin="round" d="M12 15v2m-6 4h12a2 2 0 002-2v-6a2 2 0 00-2-2H6a2 2 0 00-2 2v6a2 2 0 002 2zm10-10V7a4 4 0 00-8 0v4h8z" />
      </svg>
      <h1 class="text-xl font-bold text-base-content">{getT().appTitle}</h1>
    </div>
    <div class="flex-none flex items-center gap-2">
      <button class="btn btn-ghost btn-sm gap-1.5" onclick={toggleLang}>
        <svg xmlns="http://www.w3.org/2000/svg" class="h-4 w-4" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2">
          <path stroke-linecap="round" stroke-linejoin="round" d="M21 12a9 9 0 01-9 9m9-9a9 9 0 00-9-9m9 9H3m9 9a9 9 0 01-9-9m9 9c1.657 0 3-4.03 3-9s-1.343-9-3-9m0 18c-1.657 0-3-4.03-3-9s1.343-9 3-9m-9 9a9 9 0 019-9" />
        </svg>
        <span class="text-xs font-medium">{getLang() === 'en' ? '中文' : 'EN'}</span>
      </button>
      <div class="divider divider-horizontal mx-0 h-6"></div>
      <button class="btn btn-ghost btn-sm btn-circle" onclick={openGitHub} title="GitHub">
        <svg xmlns="http://www.w3.org/2000/svg" class="h-5 w-5" viewBox="0 0 24 24" fill="currentColor">
          <path d="M12 0c-6.626 0-12 5.373-12 12 0 5.302 3.438 9.8 8.207 11.387.599.111.793-.261.793-.577v-2.234c-3.338.726-4.033-1.416-4.033-1.416-.546-1.387-1.333-1.756-1.333-1.756-1.089-.745.083-.729.083-.729 1.205.084 1.839 1.237 1.839 1.237 1.07 1.834 2.807 1.304 3.492.997.107-.775.418-1.305.762-1.604-2.665-.305-5.467-1.334-5.467-5.931 0-1.311.469-2.381 1.236-3.221-.124-.303-.535-1.524.117-3.176 0 0 1.008-.322 3.301 1.23.957-.266 1.983-.399 3.003-.404 1.02.005 2.047.138 3.006.404 2.291-1.552 3.297-1.23 3.297-1.23.653 1.653.242 2.874.118 3.176.77.84 1.235 1.911 1.235 3.221 0 4.609-2.807 5.624-5.479 5.921.43.372.823 1.102.823 2.222v3.293c0 .319.192.694.801.576 4.765-1.589 8.199-6.086 8.199-11.386 0-6.627-5.373-12-12-12z"/>
        </svg>
      </button>
    </div>
  </header>

  <main class="flex-1 flex flex-col">
    <div class="flex justify-center pt-6 pb-2">
      <div class="join bg-base-200 p-1 rounded-box shadow-md">
        <button
          class="btn join-item btn-lg {activeTab === 'backup' ? 'btn-primary' : 'btn-ghost text-base-content/60'} min-w-36"
          onclick={() => activeTab = 'backup'}
        >
          <svg xmlns="http://www.w3.org/2000/svg" class="h-5 w-5" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2">
            <path stroke-linecap="round" stroke-linejoin="round" d="M4 16v1a3 3 0 003 3h10a3 3 0 003-3v-1m-4-8l-4-4m0 0L8 8m4-4v12" />
          </svg>
          {getT().backupTab}
        </button>
        <button
          class="btn join-item btn-lg {activeTab === 'restore' ? 'btn-secondary' : 'btn-ghost text-base-content/60'} min-w-36"
          onclick={() => activeTab = 'restore'}
        >
          <svg xmlns="http://www.w3.org/2000/svg" class="h-5 w-5" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2">
            <path stroke-linecap="round" stroke-linejoin="round" d="M4 16v1a3 3 0 003 3h10a3 3 0 003-3v-1m-4-4l-4 4m0 0l-4-4m4 4V4" />
          </svg>
          {getT().restoreTab}
        </button>
      </div>
    </div>

    <div class="flex-1 p-6 pt-2">
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
