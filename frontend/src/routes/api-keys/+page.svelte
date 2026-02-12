<script lang="ts">
  import { onMount } from 'svelte';
  import { api } from '$lib/api/client';
  import { toasts } from '$lib/stores/toast';
  import type { Project, APIKey, CreateAPIKeyResponse } from '$lib/types';

  let projects = $state<Project[]>([]);
  let selectedProjectId = $state('');
  let apiKeys = $state<APIKey[]>([]);
  let loading = $state(true);
  let showCreate = $state(false);
  let newKeyName = $state('');
  let newRawKey = $state('');

  onMount(async () => {
    try {
      projects = await api.get<Project[]>('/api/projects');
      if (projects.length > 0) {
        selectedProjectId = projects[0].id;
        await loadKeys();
      }
    } catch {
      toasts.error('Failed to load projects');
    } finally {
      loading = false;
    }
  });

  async function loadKeys() {
    if (!selectedProjectId) return;
    try {
      apiKeys = await api.get<APIKey[]>(`/api/projects/${selectedProjectId}/api-keys`);
    } catch {
      toasts.error('Failed to load API keys');
    }
  }

  async function handleProjectChange() {
    await loadKeys();
  }

  async function createKey() {
    try {
      const res = await api.post<CreateAPIKeyResponse>(`/api/projects/${selectedProjectId}/api-keys`, {
        name: newKeyName,
        scopes: ['read'],
      });
      newRawKey = res.raw_key;
      toasts.success('API key created! Copy it now — it won\'t be shown again.');
      newKeyName = '';
      await loadKeys();
    } catch (err: any) {
      toasts.error(err.message || 'Failed to create key');
    }
  }

  async function deactivateKey(keyId: string) {
    if (!confirm('Deactivate this API key? It will no longer work for export.')) return;
    try {
      await api.delete(`/api/projects/${selectedProjectId}/api-keys/${keyId}`);
      toasts.success('Key deactivated');
      await loadKeys();
    } catch (err: any) {
      toasts.error(err.message || 'Failed to deactivate key');
    }
  }

  function copyToClipboard(text: string) {
    navigator.clipboard.writeText(text);
    toasts.success('Copied to clipboard');
  }
</script>

<div class="space-y-6">
  <div>
    <h1 class="text-3xl font-bold text-surface-100">API Keys</h1>
    <p class="text-surface-400 mt-1">Generate API keys to access translations from your applications</p>
  </div>

  <!-- Project Selector -->
  <div class="flex items-center gap-4">
    <label for="projSelect" class="text-sm text-surface-400">Project:</label>
    <select
      id="projSelect"
      bind:value={selectedProjectId}
      onchange={handleProjectChange}
      class="px-4 py-2 bg-surface-800/50 border border-surface-700/50 rounded-xl text-surface-100 focus:outline-none focus:ring-2 focus:ring-primary-500/50 transition-all"
    >
      {#each projects as p}
        <option value={p.id}>{p.name}</option>
      {/each}
    </select>
    <button
      onclick={() => { showCreate = true; newRawKey = ''; }}
      class="px-4 py-2 bg-primary-600 hover:bg-primary-500 text-white rounded-xl text-sm transition-colors ml-auto"
    >
      + Generate Key
    </button>
  </div>

  <!-- New key display -->
  {#if newRawKey}
    <div class="bg-emerald-900/30 border border-emerald-500/30 rounded-2xl p-4">
      <p class="text-sm text-emerald-400 font-medium mb-2">🔑 New API Key (copy it now — it won't be shown again)</p>
      <div class="flex items-center gap-2">
        <code class="flex-1 px-3 py-2 bg-surface-900/80 rounded-lg text-emerald-300 text-sm font-mono break-all">{newRawKey}</code>
        <button
          onclick={() => copyToClipboard(newRawKey)}
          class="px-3 py-2 bg-emerald-600/20 text-emerald-400 hover:bg-emerald-600/30 rounded-xl text-sm transition-all"
        >Copy</button>
      </div>
      <div class="mt-3 text-xs text-surface-400">
        <p><strong>Usage:</strong> Add <code class="text-surface-300">X-API-Key: {newRawKey.slice(0, 12)}...</code> header to your requests</p>
        <p class="mt-1"><strong>Endpoint:</strong> <code class="text-surface-300">GET /api/export/&#123;slug&#125;/&#123;lang_code&#125;?format=json</code></p>
      </div>
    </div>
  {/if}

  <!-- Keys List -->
  {#if loading}
    <div class="flex items-center justify-center py-12">
      <div class="animate-spin w-6 h-6 border-2 border-primary-500 border-t-transparent rounded-full"></div>
    </div>
  {:else if apiKeys.length === 0}
    <div class="text-center py-16 bg-surface-900/40 rounded-2xl border border-surface-700/30">
      <p class="text-surface-400 text-lg mb-2">No API keys</p>
      <p class="text-surface-500 text-sm">Generate a key to access translations via API</p>
    </div>
  {:else}
    <div class="space-y-3">
      {#each apiKeys as key}
        <div class="bg-surface-900/60 backdrop-blur-xl border border-surface-700/50 rounded-2xl p-5 flex items-center gap-4">
          <div class="w-10 h-10 rounded-xl bg-primary-600/20 flex items-center justify-center text-lg shrink-0">
            🔑
          </div>
          <div class="flex-1 min-w-0">
            <div class="flex items-center gap-2">
              <h3 class="font-medium text-surface-200">{key.name}</h3>
              <span class="text-xs px-2 py-0.5 rounded-full {key.is_active ? 'bg-emerald-600/20 text-emerald-400' : 'bg-red-600/20 text-red-400'}">
                {key.is_active ? 'Active' : 'Inactive'}
              </span>
            </div>
            <div class="flex items-center gap-4 mt-1">
              <span class="text-xs text-surface-500 font-mono">{key.key_prefix}...</span>
              <span class="text-xs text-surface-500">Scopes: {key.scopes?.join(', ') || 'read'}</span>
              {#if key.last_used_at}
                <span class="text-xs text-surface-500">Last used: {new Date(key.last_used_at).toLocaleDateString()}</span>
              {/if}
              <span class="text-xs text-surface-500">Created: {new Date(key.created_at).toLocaleDateString()}</span>
            </div>
          </div>
          {#if key.is_active}
            <button
              onclick={() => deactivateKey(key.id)}
              class="px-3 py-1.5 text-red-400 hover:bg-red-500/10 border border-red-500/20 rounded-lg text-sm transition-all"
            >
              Deactivate
            </button>
          {/if}
        </div>
      {/each}
    </div>
  {/if}
</div>

<!-- Create Modal -->
{#if showCreate}
  <div class="fixed inset-0 z-50 flex items-center justify-center">
    <div class="absolute inset-0 bg-black/60 backdrop-blur-sm" onclick={() => showCreate = false}></div>
    <div class="relative bg-surface-900 border border-surface-700/50 rounded-2xl p-6 w-full max-w-sm shadow-2xl">
      <h2 class="text-xl font-bold text-surface-100 mb-4">Generate API Key</h2>
      <form onsubmit={(e) => { e.preventDefault(); createKey(); showCreate = false; }}>
        <div class="space-y-4">
          <div>
            <label for="keyName" class="block text-sm font-medium text-surface-300 mb-1.5">Key Name</label>
            <input
              id="keyName"
              type="text"
              bind:value={newKeyName}
              placeholder="e.g. Production Frontend"
              required
              class="w-full px-4 py-2.5 bg-surface-800/50 border border-surface-600/50 rounded-xl text-surface-100 focus:outline-none focus:ring-2 focus:ring-primary-500/50 transition-all"
            />
          </div>
          <div class="flex gap-3 justify-end">
            <button type="button" onclick={() => showCreate = false} class="px-4 py-2 text-surface-400 hover:text-surface-200 text-sm">Cancel</button>
            <button type="submit" class="px-4 py-2 bg-primary-600 hover:bg-primary-500 text-white rounded-xl text-sm">Generate</button>
          </div>
        </div>
      </form>
    </div>
  </div>
{/if}
