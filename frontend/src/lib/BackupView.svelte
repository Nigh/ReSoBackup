<script>
import { BackupService } from '../../bindings/reed-solomon-backup';
import { getT } from './i18n.svelte.js';

const MIN_SHARES = 3;
const MAX_SHARES = 128;
const SLIDER_MAX = 1000;

const SPLIT_POS = 800;
const SPLIT_VAL = 64;

function valueFromPos(pos) {
  if (pos <= SPLIT_POS) {
    return Math.round(MIN_SHARES + (SPLIT_VAL - MIN_SHARES) * (pos / SPLIT_POS));
  }
  return Math.round(SPLIT_VAL + (MAX_SHARES - SPLIT_VAL) * ((pos - SPLIT_POS) / (SLIDER_MAX - SPLIT_POS)));
}

function posFromValue(val) {
  if (val <= SPLIT_VAL) {
    return Math.round(SPLIT_POS * (val - MIN_SHARES) / (SPLIT_VAL - MIN_SHARES));
  }
  return Math.round(SPLIT_POS + (SLIDER_MAX - SPLIT_POS) * (val - SPLIT_VAL) / (MAX_SHARES - SPLIT_VAL));
}

let inputPath = $state('');
let inputInfo = $state('');
let shares = $state(8);
let threshold = $state(5);
let storedRatio = $state(8 / 5);
let sharesPos = $state(posFromValue(8));
let encrypt = $state(false);
let encryptFilename = $state(true);
let password = $state('');
let showPassword = $state(false);
let outputDir = $state('');
let warnings = $state([]);
let loading = $state(false);
let status = $state(null);
let statusMsg = $state('');

function onSharesSliderChange(e) {
  sharesPos = Number(e.target.value);
  const newShares = valueFromPos(sharesPos);
  shares = newShares;
  threshold = Math.max(1, Math.min(shares, Math.round(shares / storedRatio)));
}

function onSharesInputChange(e) {
  let val = Number(e.target.value);
  val = Math.max(MIN_SHARES, Math.min(MAX_SHARES, val));
  shares = val;
  sharesPos = posFromValue(val);
  threshold = Math.max(1, Math.min(shares, Math.round(shares / storedRatio)));
}

function onThresholdSliderChange(e) {
  threshold = Number(e.target.value);
  storedRatio = shares / threshold;
}

function onThresholdInputChange(e) {
  let val = Number(e.target.value);
  val = Math.max(1, Math.min(shares, val));
  threshold = val;
  storedRatio = shares / threshold;
}

async function pickFile() {
  try {
    const result = await BackupService.SelectInputFile();
    if (result) {
      inputPath = result;
      outputDir = '';
      try {
        inputInfo = await BackupService.GetFileInfo(result);
      } catch {
        inputInfo = '';
      }
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

async function checkWarnings() {
  try {
    warnings = await BackupService.GetBackupWarnings(shares, threshold);
  } catch {
    warnings = [];
  }
}

$effect(() => {
  const _ = shares + threshold;
  checkWarnings();
});

function getRedundancy() {
  if (threshold <= 0) return 0;
  return ((shares - threshold) / shares * 100).toFixed(1);
}

function getMaxLoss() {
  return shares - threshold;
}

function fmt(key, n) {
  return key.replace('{n}', n);
}

async function doBackup() {
  const tr = getT();
  if (!inputPath) {
    status = 'error';
    statusMsg = tr.errSelectFile;
    return;
  }
  if (encrypt && !password) {
    status = 'error';
    statusMsg = tr.errEnterPwd;
    return;
  }

  if (warnings.length > 0) {
    const confirmed = confirm(
      tr.warnTitle + '\n\n' + warnings.join('\n') + '\n\n' + tr.confirmContinue
    );
    if (!confirmed) return;
  }

  loading = true;
  status = null;
  statusMsg = '';

  try {
    await BackupService.RunBackup(inputPath, password, outputDir, shares, threshold, encrypt, encrypt && encryptFilename);
    status = 'success';
    if (encrypt && encryptFilename) {
      const prefix = await BackupService.GetLastBackupPrefix();
      statusMsg = fmt(tr.backupSuccessEncFN, prefix);
    } else {
      statusMsg = tr.backupSuccess;
    }
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
          <path stroke-linecap="round" stroke-linejoin="round" d="M9 12h6m-6 4h6m2 5H7a2 2 0 01-2-2V5a2 2 0 012-2h5.586a1 1 0 01.707.293l5.414 5.414a1 1 0 01.293.707V19a2 2 0 01-2 2z" />
        </svg>
        {getT().inputFile}
      </h2>

      <div class="flex gap-2">
        <input
          type="text"
          class="input input-bordered flex-1"
          placeholder={getT().selectFile}
          value={inputPath}
          readonly
        />
        <button class="btn btn-primary" onclick={pickFile}>
          {getT().browse}
        </button>
      </div>
      {#if inputInfo}
        <p class="text-sm text-base-content/60">{inputInfo}</p>
      {/if}
    </div>
  </div>

  <div class="card bg-base-200 shadow-xl">
    <div class="card-body space-y-4">
      <h2 class="card-title text-secondary">
        <svg xmlns="http://www.w3.org/2000/svg" class="h-5 w-5" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2">
          <path stroke-linecap="round" stroke-linejoin="round" d="M10.325 4.317c.426-1.756 2.924-1.756 3.35 0a1.724 1.724 0 002.573 1.066c1.543-.94 3.31.826 2.37 2.37a1.724 1.724 0 001.066 2.573c1.756.426 1.756 2.924 0 3.35a1.724 1.724 0 00-1.066 2.573c.94 1.543-.826 3.31-2.37 2.37a1.724 1.724 0 00-2.573 1.066c-.426 1.756-2.924 1.756-3.35 0a1.724 1.724 0 00-2.573-1.066c-1.543.94-3.31-.826-2.37-2.37a1.724 1.724 0 00-1.066-2.573c-1.756-.426-1.756-2.924 0-3.35a1.724 1.724 0 001.066-2.573c-.94-1.543.826-3.31 2.37-2.37.996.608 2.296.07 2.572-1.065z" />
          <path stroke-linecap="round" stroke-linejoin="round" d="M15 12a3 3 0 11-6 0 3 3 0 016 0z" />
        </svg>
        {getT().rsParams}
      </h2>

      <div class="form-control">
        <div class="flex justify-between items-center">
          <span class="label-text">{getT().totalShares}</span>
          <div class="flex items-center gap-2">
            <input
              type="number"
              class="input input-xs input-bordered w-16 text-center"
              min={MIN_SHARES}
              max={MAX_SHARES}
              value={shares}
              onchange={onSharesInputChange}
            />
            <span class="text-xs text-base-content/40 w-14 text-right">{MIN_SHARES}~{MAX_SHARES}</span>
          </div>
        </div>
        <input
          type="range"
          min="0"
          max={SLIDER_MAX}
          value={sharesPos}
          oninput={onSharesSliderChange}
          class="range range-primary range-sm mt-2"
        />
      </div>

      <div class="form-control">
        <div class="flex justify-between items-center">
          <span class="label-text">{getT().threshold}</span>
          <div class="flex items-center gap-2">
            <input
              type="number"
              class="input input-xs input-bordered w-16 text-center"
              min="1"
              max={shares}
              value={threshold}
              onchange={onThresholdInputChange}
            />
            <span class="text-xs text-base-content/40 w-14 text-right">1~{shares}</span>
          </div>
        </div>
        <input
          type="range"
          min="1"
          max={shares}
          value={threshold}
          oninput={onThresholdSliderChange}
          class="range range-secondary range-sm mt-2"
        />
      </div>

      <div class="stats stats-horizontal shadow w-full">
        <div class="stat">
          <div class="stat-title">{getT().redundancy}</div>
          <div class="stat-value text-accent text-2xl">{getRedundancy()}%</div>
          <div class="stat-desc">{fmt(getT().canLose, getMaxLoss())}</div>
        </div>
        <div class="stat">
          <div class="stat-title">{getT().storageOverhead}</div>
          <div class="stat-value text-info text-2xl">{(shares / threshold).toFixed(2)}x</div>
          <div class="stat-desc">{fmt(getT().sharesTotal, shares)}</div>
        </div>
      </div>

      {#if warnings.length > 0}
        <div class="space-y-1">
          {#each warnings as warning}
            <div class="alert alert-warning alert-sm">
              <svg xmlns="http://www.w3.org/2000/svg" class="h-4 w-4 shrink-0" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2">
                <path stroke-linecap="round" stroke-linejoin="round" d="M12 9v2m0 4h.01m-6.938 4h13.856c1.54 0 2.502-1.667 1.732-3L13.732 4c-.77-1.333-2.694-1.333-3.464 0L3.34 16c-.77 1.333.192 3 1.732 3z" />
              </svg>
              <span class="text-sm">{warning}</span>
            </div>
          {/each}
        </div>
      {/if}
    </div>
  </div>

  <div class="card bg-base-200 shadow-xl">
    <div class="card-body space-y-4">
      <h2 class="card-title text-accent">
        <svg xmlns="http://www.w3.org/2000/svg" class="h-5 w-5" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2">
          <path stroke-linecap="round" stroke-linejoin="round" d="M15 7a2 2 0 012 2m4 0a6 6 0 01-7.743 5.743L11 17H9v2H7v2H4a1 1 0 01-1-1v-2.586a1 1 0 01.293-.707l5.964-5.964A6 6 0 1121 9z" />
        </svg>
        {getT().encryption}
      </h2>

      <div class="form-control">
        <label class="label cursor-pointer justify-start gap-3">
          <input type="checkbox" class="checkbox checkbox-sm" bind:checked={encrypt} />
          <div>
            <span class="label-text font-medium">{getT().enableEncryption}</span>
            <p class="text-xs text-base-content/50">{getT().enableEncryptionDesc}</p>
          </div>
        </label>
      </div>

      {#if encrypt}
        <div class="form-control ml-6">
          <label class="label cursor-pointer justify-start gap-3">
            <input type="checkbox" class="checkbox checkbox-sm checkbox-accent" bind:checked={encryptFilename} />
            <div>
              <span class="label-text">{getT().encryptFilename}</span>
              <p class="text-xs text-base-content/50">{getT().encryptFilenameDesc}</p>
            </div>
          </label>
        </div>

        <div class="form-control">
          <!-- svelte-ignore a11y_label_has_associated_control -->
          <label class="label">
            <span class="label-text">{getT().password}</span>
          </label>
          <div class="join w-full">
            <input
              type={showPassword ? 'text' : 'password'}
              class="input input-bordered join-item flex-1"
              placeholder={getT().enterPassword}
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
      {/if}
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
          placeholder={getT().outputDirPlaceholder}
          bind:value={outputDir}
        />
        <button class="btn btn-info btn-outline" onclick={pickOutputDir}>
          {getT().browse}
        </button>
      </div>
      <p class="text-xs text-base-content/40">{getT().outputDirHint}</p>
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
    class="btn btn-primary btn-lg w-full"
    onclick={doBackup}
    disabled={loading || !inputPath || (encrypt && !password)}
  >
    {#if loading}
      <span class="loading loading-spinner"></span>
      {getT().creatingBackup}
    {:else}
      <svg xmlns="http://www.w3.org/2000/svg" class="h-5 w-5" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2">
        <path stroke-linecap="round" stroke-linejoin="round" d="M12 15v2m-6 4h12a2 2 0 002-2v-6a2 2 0 00-2-2H6a2 2 0 00-2 2v6a2 2 0 002 2zm10-10V7a4 4 0 00-8 0v4h8z" />
      </svg>
      {getT().createBackup}
    {/if}
  </button>
</div>
