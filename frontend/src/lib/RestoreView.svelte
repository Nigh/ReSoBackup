<script>
import { BackupService } from '../../bindings/reed-solomon-backup';
import { getT } from './i18n.svelte.js';

let inputPath = $state('');
let password = $state('');
let showPassword = $state(false);
let outputDir = $state('');
let loading = $state(false);
let status = $state(null);
let statusMsg = $state('');

async function pickFile() {
  try {
    const result = await BackupService.SelectRestoreFile();
    if (result) {
      inputPath = result;
    }
  } catch (e) {
    console.error(e);
  }
}

async function pickOutputDir() {
  try {
    const result = await BackupService.SelectOutputDirectory();
    if (result) {
      outputDir = result;
    }
  } catch (e) {
    console.error(e);
  }
}

async function doRestore() {
  const tr = getT();
  if (!inputPath) {
    status = 'error';
    statusMsg = tr.errSelectRestore;
    return;
  }
  if (!password) {
    status = 'error';
    statusMsg = tr.errEnterRestorePwd;
    return;
  }

  loading = true;
  status = null;
  statusMsg = '';

  try {
    await BackupService.RunRestore(inputPath, password, outputDir);
    status = 'success';
    statusMsg = tr.restoreSuccess;
  } catch (e) {
    status = 'error';
    statusMsg = e.message || String(e);
  } finally {
    loading = false;
  }
}
</script>

<div class="space-y-6 max-w-2xl mx-auto">
  <div class="card bg-base-200 shadow-xl">
    <div class="card-body space-y-4">
      <h2 class="card-title text-primary">
        <svg xmlns="http://www.w3.org/2000/svg" class="h-5 w-5" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2">
          <path stroke-linecap="round" stroke-linejoin="round" d="M4 16v1a3 3 0 003 3h10a3 3 0 003-3v-1m-4-8l-4-4m0 0L8 8m4-4v12" />
        </svg>
        {getT().shareFile}
      </h2>

      <div class="flex gap-2">
        <input
          type="text"
          class="input input-bordered flex-1"
          placeholder={getT().selectRestore}
          value={inputPath}
          readonly
        />
        <button class="btn btn-primary" onclick={pickFile}>
          {getT().browse}
        </button>
      </div>
      <p class="text-xs text-base-content/40">
        {getT().restoreHint}
      </p>
    </div>
  </div>

  <div class="card bg-base-200 shadow-xl">
    <div class="card-body space-y-4">
      <h2 class="card-title text-accent">
        <svg xmlns="http://www.w3.org/2000/svg" class="h-5 w-5" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2">
          <path stroke-linecap="round" stroke-linejoin="round" d="M15 7a2 2 0 012 2m4 0a6 6 0 01-7.743 5.743L11 17H9v2H7v2H4a1 1 0 01-1-1v-2.586a1 1 0 01.293-.707l5.964-5.964A6 6 0 1121 9z" />
        </svg>
        {getT().decryptPwd}
      </h2>

      <div class="form-control">
        <!-- svelte-ignore a11y_label_has_associated_control -->
        <label class="label">
          <span class="label-text">{getT().password}</span>
        </label>
        <div class="join w-full">
          <input
            type={showPassword ? 'text' : 'password'}
            class="input input-bordered join-item flex-1"
            placeholder={getT().enterRestorePwd}
            bind:value={password}
          />
          <button
            class="btn btn-square join-item"
            onclick={() => showPassword = !showPassword}
          >
            {#if showPassword}
              <svg xmlns="http://www.w3.org/2000/svg" class="h-5 w-5" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2">
                <path stroke-linecap="round" stroke-linejoin="round" d="M13.875 18.825A10.05 10.05 0 0112 19c-4.478 0-8.268-2.943-9.543-7a9.97 9.97 0 011.563-3.029m5.858.908a3 3 0 114.243 4.243M9.878 9.878l4.242 4.242M9.878 9.878L3 3m6.878 6.878L21 21" />
              </svg>
            {:else}
              <svg xmlns="http://www.w3.org/2000/svg" class="h-5 w-5" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2">
                <path stroke-linecap="round" stroke-linejoin="round" d="M15 12a3 3 0 11-6 0 3 3 0 016 0z" />
                <path stroke-linecap="round" stroke-linejoin="round" d="M2.458 12C3.732 7.943 7.523 5 12 5c4.478 0 8.268 2.943 9.542 7-1.274 4.057-5.064 7-9.542 7-4.477 0-8.268-2.943-9.542-7z" />
              </svg>
            {/if}
          </button>
        </div>
      </div>
    </div>
  </div>

  <div class="card bg-base-200 shadow-xl">
    <div class="card-body space-y-4">
      <h2 class="card-title text-info">
        <svg xmlns="http://www.w3.org/2000/svg" class="h-5 w-5" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2">
          <path stroke-linecap="round" stroke-linejoin="round" d="M3 7v10a2 2 0 002 2h14a2 2 0 002-2V9a2 2 0 00-2-2h-6l-2-2H5a2 2 0 00-2 2z" />
        </svg>
        {getT().outputDir}
      </h2>

      <div class="flex gap-2">
        <input
          type="text"
          class="input input-bordered flex-1"
          placeholder={getT().restoreOutputPlaceholder}
          bind:value={outputDir}
        />
        <button class="btn btn-info btn-outline" onclick={pickOutputDir}>
          {getT().browse}
        </button>
      </div>
      <p class="text-xs text-base-content/40">{getT().restoreOutputHint}</p>
    </div>
  </div>

  {#if status}
    <div class="alert {status === 'success' ? 'alert-success' : 'alert-error'}">
      {#if status === 'success'}
        <svg xmlns="http://www.w3.org/2000/svg" class="h-5 w-5 shrink-0" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2">
          <path stroke-linecap="round" stroke-linejoin="round" d="M9 12l2 2 4-4m6 2a9 9 0 11-18 0 9 9 0 0118 0z" />
        </svg>
      {:else}
        <svg xmlns="http://www.w3.org/2000/svg" class="h-5 w-5 shrink-0" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2">
          <path stroke-linecap="round" stroke-linejoin="round" d="M10 14l2-2m0 0l2-2m-2 2l-2-2m2 2l2 2m7-2a9 9 0 11-18 0 9 9 0 0118 0z" />
        </svg>
      {/if}
      <span>{statusMsg}</span>
    </div>
  {/if}

  <button
    class="btn btn-secondary btn-lg w-full"
    onclick={doRestore}
    disabled={loading || !inputPath || !password}
  >
    {#if loading}
      <span class="loading loading-spinner"></span>
      {getT().restoring}
    {:else}
      <svg xmlns="http://www.w3.org/2000/svg" class="h-5 w-5" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2">
        <path stroke-linecap="round" stroke-linejoin="round" d="M4 16v1a3 3 0 003 3h10a3 3 0 003-3v-1m-4-4l-4 4m0 0l-4-4m4 4V4" />
      </svg>
      {getT().restoreFile}
    {/if}
  </button>
</div>
