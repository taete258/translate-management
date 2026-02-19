<script lang="ts">
  import { onMount } from 'svelte';
  import { api } from '$lib/api/client';
  import { toasts } from '$lib/stores/toast';
  import type { Project, APIKey, CreateAPIKeyResponse } from '$lib/types';
  import SearchableSelect from '$lib/components/SearchableSelect.svelte';
  import { Key } from 'lucide-svelte';

  let projects = $state<Project[]>([]);
  let selectedProjectId = $state('');
  let apiKeys = $state<APIKey[]>([]);
  let loading = $state(true);
  let showCreate = $state(false);
  let newKeyName = $state('');
  let newRawKey = $state('');

  // Effect to load keys when project changes
  $effect(() => {
    if (selectedProjectId) {
      loadKeys();
    }
  });

  onMount(async () => {
    try {
      const allProjects = await api.get<Project[]>('/api/projects');
      projects = allProjects.filter(p => p.role === 'owner');
      if (projects.length > 0) {
        // This assignment will trigger the effect above
        selectedProjectId = projects[0].id;
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
    <h1 class="text-3xl font-bold text-heading">API Keys</h1>
    <p class="text-subtle mt-1">Generate API keys to access translations from your applications</p>
  </div>

  <!-- Project Selector -->
  <div class="flex flex-col sm:flex-row sm:items-center gap-4">
    <div class="flex items-center gap-3 w-full sm:w-auto flex-1 max-w-md">
      <label for="projSelect" class="text-sm text-subtle whitespace-nowrap">Project:</label>
      <div class="flex-1 min-w-[200px]">
        <SearchableSelect 
          items={projects} 
          bind:value={selectedProjectId} 
          labelKey="name" 
          valueKey="id" 
          placeholder="Select a project" 
        />
      </div>
    </div>
    <button
      onclick={() => { showCreate = true; newRawKey = ''; }}
      class="px-4 py-2 bg-primary-600 hover:bg-primary-500 text-white rounded-xl text-sm transition-colors sm:ml-auto w-full sm:w-auto"
    >
      + Generate Key
    </button>
  </div>

  <!-- New key display -->
  {#if newRawKey}
    <div class="bg-white dark:bg-emerald-900/20 border border-emerald-200 dark:border-emerald-500/30 rounded-2xl p-6 shadow-sm">
      <div class="flex items-start gap-4">
        <div class="p-2 bg-emerald-100 dark:bg-emerald-900/50 rounded-lg hidden sm:flex items-center justify-center text-emerald-700 dark:text-emerald-500">
           <Key size={24} />
        </div>
        <div class="flex-1 min-w-0">
           <p class="text-base text-heading dark:text-emerald-400 font-semibold mb-1">API Key Generated</p>
           <p class="text-sm text-emerald-600 dark:text-emerald-500 mb-4">Copy this key now. It will not be shown again.</p>
           
           <div class="flex items-center gap-2 mb-4">
             <code class="flex-1 px-4 py-3 rounded-xl bg-surface-50 dark:bg-black/40 border border-emerald-100 dark:border-emerald-500/20 text-emerald-700 dark:text-emerald-400 text-sm font-mono break-all tracking-wide">{newRawKey}</code>
             <button
               onclick={() => copyToClipboard(newRawKey)}
               class="px-4 py-3 bg-white dark:bg-emerald-600/20 hover:bg-emerald-50 dark:hover:bg-emerald-600/30 border border-emerald-200 dark:border-transparent text-emerald-700 dark:text-emerald-400 rounded-xl text-sm font-medium transition-all shadow-sm"
             >Copy</button>
           </div>
           
           <div class="text-xs text-subtle bg-surface-50 dark:bg-surface-800/50 p-3 rounded-lg border border-subtle">
             <p class="mb-1"><strong class="text-heading">Usage:</strong> Add <code class="bg-surface-200 dark:bg-surface-700 px-1 py-0.5 rounded text-heading">X-API-Key: {newRawKey.slice(0, 12)}...</code> header to your requests</p>
             <p><strong>Endpoint:</strong> <code class="bg-surface-200 dark:bg-surface-700 px-1 py-0.5 rounded text-heading">GET /api/export/&#123;slug&#125;/&#123;lang_code&#125;?format=json</code></p>
           </div>
        </div>
      </div>
    </div>
  {/if}

  <!-- Keys List -->
  {#if loading}
    <div class="flex items-center justify-center py-12">
      <div class="animate-spin w-6 h-6 border-2 border-primary-500 border-t-transparent rounded-full"></div>
    </div>
  {:else if apiKeys.length === 0}
    <div class="themed-card text-center py-16 rounded-2xl">
      <p class="text-subtle text-lg mb-2">No API keys</p>
      {#if !selectedProjectId}
        <p class="text-faint text-sm">Select a project to view keys</p>
      {:else}
        <p class="text-faint text-sm">Generate a key to access translations via API</p>
      {/if}
    </div>
  {:else}
    <div class="grid gap-3">
      {#each apiKeys as key}
        <div class="themed-card backdrop-blur-xl rounded-2xl p-5 flex flex-col sm:flex-row sm:items-center gap-4">
          <div class="w-10 h-10 rounded-xl bg-primary-600/20 flex items-center justify-center text-primary-600 shrink-0 hidden sm:flex">
            <Key size={20} />
          </div>
          <div class="flex-1 min-w-0">
            <div class="flex items-center gap-2 mb-1">
              <h3 class="font-medium text-heading text-base">{key.name}</h3>
              <span class="text-xs px-2.5 py-0.5 rounded-full font-medium {key.is_active ? 'bg-emerald-100 text-emerald-700 dark:bg-emerald-600/20 dark:text-emerald-500' : 'bg-red-100 text-red-700 dark:bg-red-600/20 dark:text-red-500'}">
                {key.is_active ? 'Active' : 'Inactive'}
              </span>
            </div>
            <div class="flex flex-wrap items-center gap-x-4 gap-y-1 mt-1 text-sm">
              <span class="text-faint font-mono px-1.5 py-0.5 rounded">{key.key_prefix}...</span>
              <span class="text-subtle">Scopes: {key.scopes?.join(', ') || 'read'}</span>
              {#if key.last_used_at}
                <span class="text-subtle">Last used: {new Date(key.last_used_at).toLocaleDateString()}</span>
              {/if}
              <span class="text-subtle">Created: {new Date(key.created_at).toLocaleDateString()}</span>
            </div>
          </div>
          {#if key.is_active}
            <button
              onclick={() => deactivateKey(key.id)}
              class="px-3 py-1.5 text-red-600 dark:text-red-500 hover:bg-red-50 dark:hover:bg-red-500/10 border border-red-200 dark:border-red-500/20 rounded-lg text-sm transition-all sm:self-center self-start"
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
  <div class="fixed inset-0 z-50 flex items-center justify-center p-4">
    <button class="absolute inset-0 bg-black/50 backdrop-blur-sm transition-opacity" aria-label="Close modal" onclick={() => showCreate = false}></button>
    <div class="themed-modal relative rounded-2xl p-6 w-full max-w-sm animate-in fade-in zoom-in-95 duration-200 z-10">
      <h2 class="text-xl font-bold text-heading mb-1">Generate API Key</h2>
      <p class="text-subtle text-sm mb-5">Create a scoped key for accessing translations.</p>
      
      <form onsubmit={(e) => { e.preventDefault(); createKey(); showCreate = false; }}>
        <div class="space-y-4">
           <div>
             <span class="text-xs text-faint uppercase font-bold tracking-wider block mb-1">Project</span>
             <p class="text-heading font-medium truncate">
                {projects.find(p => p.id === selectedProjectId)?.name || 'None selected'}
             </p>
           </div>

          <div>
            <label for="keyName" class="block text-sm font-medium text-body mb-1.5">Key Name</label>
            <input
              id="keyName"
              type="text"
              bind:value={newKeyName}
              placeholder="e.g. Production Frontend"
              required
              class="themed-input w-full px-4 py-2.5 rounded-xl transition-all focus:ring-2 focus:ring-primary-500/20"
            />
          </div>
          <div class="flex gap-3 justify-end pt-2">
            <button type="button" onclick={() => showCreate = false} class="px-4 py-2 text-subtle hover:text-heading text-sm font-medium transition-colors">Cancel</button>
            <button type="submit" class="px-4 py-2 bg-primary-600 hover:bg-primary-500 text-white rounded-xl text-sm font-medium shadow-lg shadow-primary-500/20 transition-all active:scale-95">Generate Key</button>
          </div>
        </div>
      </form>
    </div>
  </div>
{/if}
