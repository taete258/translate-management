<script lang="ts">
  import { onMount } from 'svelte';
  import { page } from '$app/state';
  import { api } from '$lib/api/client';
  import { toasts } from '$lib/stores/toast';
  import type { Project, Language, TranslationEntry, ProjectStats, CacheStatus } from '$lib/types';

  const projectId = $derived(page.params.id);

  let project = $state<Project | null>(null);
  let languages = $state<Language[]>([]);
  let entries = $state<TranslationEntry[]>([]);
  let stats = $state<ProjectStats | null>(null);
  let cacheStatus = $state<CacheStatus | null>(null);
  let loading = $state(true);
  let saving = $state(false);
  let search = $state('');
  let pendingChanges = $state<Map<string, string>>(new Map());

  // Language management
  let showAddLang = $state(false);
  let langCode = $state('');
  let langName = $state('');
  let langDefault = $state(false);

  // Key management
  let showAddKey = $state(false);
  let newKey = $state('');
  let newKeyDesc = $state('');

  const filteredEntries = $derived(
    entries.filter((e) =>
      e.key.toLowerCase().includes(search.toLowerCase())
    )
  );

  onMount(() => loadAll(true));

  async function loadAll(initial = false) {
    if (initial) loading = true;
    try {
      const [p, l, t, s] = await Promise.all([
        api.get<Project>(`/api/projects/${projectId}`),
        api.get<Language[]>(`/api/projects/${projectId}/languages`),
        api.get<TranslationEntry[]>(`/api/projects/${projectId}/translations`),
        api.get<ProjectStats>(`/api/projects/${projectId}/stats`),
      ]);
      project = p;
      languages = l;
      entries = t;
      stats = s;
      try {
        cacheStatus = await api.get<CacheStatus>(`/api/projects/${projectId}/cache/status`);
      } catch { /* ok */ }
    } catch {
      toasts.error('Failed to load project');
    } finally {
      loading = false;
    }
  }

  function handleCellChange(keyId: string, langId: string, value: string) {
    const changeKey = `${keyId}:${langId}`;
    pendingChanges.set(changeKey, value);
    pendingChanges = new Map(pendingChanges);
  }

  async function saveChanges() {
    if (pendingChanges.size === 0) return;
    saving = true;
    try {
      const translations = Array.from(pendingChanges.entries()).map(([key, value]) => {
        const [keyId, langId] = key.split(':');
        return { key_id: keyId, language_id: langId, value };
      });
      await api.put(`/api/projects/${projectId}/translations`, { translations });
      toasts.success(`Saved ${translations.length} translations`);
      pendingChanges = new Map();
      // Refresh stats
      stats = await api.get<ProjectStats>(`/api/projects/${projectId}/stats`);
      entries = await api.get<TranslationEntry[]>(`/api/projects/${projectId}/translations`);
    } catch (err: any) {
      toasts.error(err.message || 'Save failed');
    } finally {
      saving = false;
    }
  }

  async function addLanguage() {
    try {
      await api.post(`/api/projects/${projectId}/languages`, {
        code: langCode,
        name: langName,
        is_default: langDefault,
      });
      toasts.success('Language added');
      showAddLang = false;
      langCode = '';
      langName = '';
      langDefault = false;
      await loadAll();
    } catch (err: any) {
      toasts.error(err.message || 'Failed to add language');
    }
  }

  async function deleteLanguage(lang: Language) {
    if (!confirm(`Remove "${lang.name}" and all its translations?`)) return;
    try {
      await api.delete(`/api/projects/${projectId}/languages/${lang.id}`);
      toasts.success('Language removed');
      await loadAll();
    } catch (err: any) {
      toasts.error(err.message || 'Failed to remove language');
    }
  }

  async function addKey() {
    try {
      await api.post(`/api/projects/${projectId}/keys`, {
        key: newKey,
        description: newKeyDesc,
      });
      toasts.success('Key added');
      showAddKey = false;
      newKey = '';
      newKeyDesc = '';
      await loadAll();
    } catch (err: any) {
      toasts.error(err.message || 'Failed to add key');
    }
  }

  async function deleteKey(keyId: string, keyName: string) {
    if (!confirm(`Delete key "${keyName}"?`)) return;
    try {
      await api.delete(`/api/projects/${projectId}/keys/${keyId}`);
      toasts.success('Key deleted');
      await loadAll();
    } catch (err: any) {
      toasts.error(err.message || 'Failed to delete key');
    }
  }

  async function invalidateCache() {
    if (!confirm('Force invalidate all cached translations for this project?')) return;
    try {
      await api.post(`/api/projects/${projectId}/cache/invalidate`);
      toasts.success('Cache invalidated');
      cacheStatus = await api.get<CacheStatus>(`/api/projects/${projectId}/cache/status`);
    } catch (err: any) {
      toasts.error(err.message || 'Failed to invalidate cache');
    }
  }
</script>

{#if loading}
  <div class="flex items-center justify-center py-20">
    <div class="animate-spin w-8 h-8 border-2 border-primary-500 border-t-transparent rounded-full"></div>
  </div>
{:else if project}
  <div class="space-y-6">
    <!-- Header -->
    <div class="flex items-start justify-between">
      <div>
        <div class="flex items-center gap-2 mb-1">
          <a href="/projects" class="text-surface-500 hover:text-surface-300 text-sm transition-colors">← Projects</a>
        </div>
        <h1 class="text-3xl font-bold text-surface-100">{project.name}</h1>
        <p class="text-surface-400 mt-1">{project.description || project.slug}</p>
      </div>
      <div class="flex gap-2">
        <!-- Cache Status & Invalidate -->
        <div class="flex items-center gap-2">
          {#if cacheStatus}
            <span class="text-xs px-2 py-1 rounded-lg {cacheStatus.cached ? 'bg-emerald-600/20 text-emerald-400' : 'bg-surface-700/50 text-surface-400'}">
              {cacheStatus.cached ? `Cached (${cacheStatus.cached_keys})` : 'Not cached'}
            </span>
          {/if}
          <button
            onclick={invalidateCache}
            class="px-3 py-2 bg-amber-600/20 text-amber-400 hover:bg-amber-600/30 border border-amber-500/30 rounded-xl text-sm transition-all"
            title="Force invalidate cache"
          >
            🔄 Invalidate Cache
          </button>
        </div>
      </div>
    </div>

    <!-- Stats Bar -->
    {#if stats}
      <div class="flex gap-4 flex-wrap">
        <div class="px-4 py-2 bg-surface-900/60 border border-surface-700/50 rounded-xl">
          <span class="text-sm text-surface-400">Keys:</span>
          <span class="text-sm font-medium text-surface-200 ml-1">{stats.total_keys}</span>
        </div>
        <div class="px-4 py-2 bg-surface-900/60 border border-surface-700/50 rounded-xl">
          <span class="text-sm text-surface-400">Languages:</span>
          <span class="text-sm font-medium text-surface-200 ml-1">{stats.total_languages}</span>
        </div>
        {#each Object.entries(stats.language_progress) as [code, pct]}
          <div class="px-4 py-2 bg-surface-900/60 border border-surface-700/50 rounded-xl">
            <span class="text-sm text-surface-400">{code}:</span>
            <span class="text-sm font-medium {pct === 100 ? 'text-emerald-400' : pct > 50 ? 'text-amber-400' : 'text-red-400'} ml-1">
              {pct.toFixed(0)}%
            </span>
          </div>
        {/each}
      </div>
    {/if}

    <!-- Toolbar -->
    <div class="flex items-center gap-3 flex-wrap">
      <input
        type="text"
        bind:value={search}
        placeholder="Search keys..."
        class="px-4 py-2 bg-surface-800/50 border border-surface-700/50 rounded-xl text-surface-100 placeholder-surface-500 focus:outline-none focus:ring-2 focus:ring-primary-500/50 transition-all text-sm w-64"
      />
      {#if languages.length > 0}
        <button
          onclick={() => showAddKey = true}
          class="px-3 py-2 bg-primary-600/20 text-primary-400 hover:bg-primary-600/30 border border-primary-500/30 rounded-xl text-sm transition-all"
        >+ Add Key</button>
      {/if}
      <button
        onclick={() => showAddLang = true}
        class="px-3 py-2 bg-emerald-600/20 text-emerald-400 hover:bg-emerald-600/30 border border-emerald-500/30 rounded-xl text-sm transition-all"
      >+ Add Language</button>

      {#if pendingChanges.size > 0}
        <button
          onclick={saveChanges}
          disabled={saving}
          class="px-4 py-2 bg-gradient-to-r from-primary-600 to-primary-700 hover:from-primary-500 hover:to-primary-600 text-white rounded-xl text-sm transition-all shadow-lg shadow-primary-500/20 disabled:opacity-50 ml-auto"
        >
          {saving ? 'Saving...' : `Save ${pendingChanges.size} changes`}
        </button>
      {/if}
    </div>

    <!-- Languages list -->
    {#if languages.length > 0}
      <div class="flex gap-2 flex-wrap">
        {#each languages as lang}
          <span class="inline-flex items-center gap-1.5 px-3 py-1.5 bg-surface-800/50 border border-surface-700/50 rounded-lg text-sm">
            <span class="text-surface-300">{lang.name}</span>
            <span class="text-surface-500 font-mono text-xs">({lang.code})</span>
            {#if lang.is_default}
              <span class="text-xs text-primary-400">★</span>
            {/if}
            <button
              onclick={() => deleteLanguage(lang)}
              class="text-surface-500 hover:text-red-400 transition-colors ml-1"
              title="Remove"
            >×</button>
          </span>
        {/each}
      </div>
    {/if}

    <!-- Translation Grid -->
    {#if languages.length === 0}
      <div class="text-center py-16 bg-surface-900/40 rounded-2xl border border-surface-700/30">
        <p class="text-4xl mb-3">🌐</p>
        <p class="text-surface-400 text-lg mb-2">No languages added yet</p>
        <p class="text-surface-500 text-sm mb-4">Add at least one language before creating translation keys</p>
        <button
          onclick={() => showAddLang = true}
          class="inline-flex px-4 py-2 bg-emerald-600 hover:bg-emerald-500 text-white rounded-lg transition-colors text-sm"
        >
          + Add Your First Language
        </button>
      </div>
    {:else if entries.length === 0}
      <div class="text-center py-16 bg-surface-900/40 rounded-2xl border border-surface-700/30">
        <p class="text-surface-400 text-lg mb-2">No translation keys</p>
        <p class="text-surface-500 text-sm">Add keys to start managing translations</p>
      </div>
    {:else}
      <div class="overflow-x-auto bg-surface-900/40 rounded-2xl border border-surface-700/30">
        <table class="w-full text-sm">
          <thead>
            <tr class="border-b border-surface-700/50">
              <th class="text-left px-4 py-3 text-surface-400 font-medium sticky left-0 bg-surface-900/90 backdrop-blur-sm min-w-[200px]">Key</th>
              {#each languages as lang}
                <th class="text-left px-4 py-3 text-surface-400 font-medium min-w-[200px]">
                  {lang.name} <span class="text-surface-500 font-normal">({lang.code})</span>
                </th>
              {/each}
              <th class="w-10"></th>
            </tr>
          </thead>
          <tbody>
            {#each filteredEntries as entry}
              <tr class="border-b border-surface-700/30 hover:bg-surface-800/30 transition-colors">
                <td class="px-4 py-2 sticky left-0 bg-surface-900/90 backdrop-blur-sm">
                  <div class="font-mono text-surface-200 text-xs">{entry.key}</div>
                  {#if entry.description}
                    <div class="text-xs text-surface-500 mt-0.5">{entry.description}</div>
                  {/if}
                </td>
                {#each languages as lang}
                  <td class="px-4 py-2">
                    <input
                      type="text"
                      value={pendingChanges.get(`${entry.key_id}:${lang.id}`) ?? entry.values[lang.id] ?? ''}
                      oninput={(e) => handleCellChange(entry.key_id, lang.id, (e.target as HTMLInputElement).value)}
                      class="w-full px-2.5 py-1.5 bg-transparent border border-transparent hover:border-surface-600/50 focus:border-primary-500/50 focus:bg-surface-800/50 rounded-lg text-surface-100 text-sm focus:outline-none transition-all {pendingChanges.has(`${entry.key_id}:${lang.id}`) ? 'border-amber-500/40 bg-amber-500/5' : ''}"
                      placeholder="—"
                    />
                  </td>
                {/each}
                <td class="px-2 py-2">
                  <button
                    onclick={() => deleteKey(entry.key_id, entry.key)}
                    class="p-1 text-surface-500 hover:text-red-400 hover:bg-red-500/10 rounded transition-all"
                    title="Delete key"
                  >🗑️</button>
                </td>
              </tr>
            {/each}
          </tbody>
        </table>
      </div>
    {/if}
  </div>

  <!-- Add Language Modal -->
  {#if showAddLang}
    <div class="fixed inset-0 z-50 flex items-center justify-center">
      <div class="absolute inset-0 bg-black/60 backdrop-blur-sm" onclick={() => showAddLang = false}></div>
      <div class="relative bg-surface-900 border border-surface-700/50 rounded-2xl p-6 w-full max-w-sm shadow-2xl">
        <h2 class="text-xl font-bold text-surface-100 mb-4">Add Language</h2>
        <form onsubmit={(e) => { e.preventDefault(); addLanguage(); }}>
          <div class="space-y-4">
            <div>
              <label for="lCode" class="block text-sm font-medium text-surface-300 mb-1.5">Code</label>
              <input id="lCode" type="text" bind:value={langCode} placeholder="en, th, ja..." required class="w-full px-4 py-2.5 bg-surface-800/50 border border-surface-600/50 rounded-xl text-surface-100 focus:outline-none focus:ring-2 focus:ring-primary-500/50 transition-all" />
            </div>
            <div>
              <label for="lName" class="block text-sm font-medium text-surface-300 mb-1.5">Name</label>
              <input id="lName" type="text" bind:value={langName} placeholder="English" required class="w-full px-4 py-2.5 bg-surface-800/50 border border-surface-600/50 rounded-xl text-surface-100 focus:outline-none focus:ring-2 focus:ring-primary-500/50 transition-all" />
            </div>
            <label class="flex items-center gap-2 text-sm text-surface-300 cursor-pointer">
              <input type="checkbox" bind:checked={langDefault} class="accent-primary-500" />
              Set as default language
            </label>
            <div class="flex gap-3 justify-end">
              <button type="button" onclick={() => showAddLang = false} class="px-4 py-2 text-surface-400 hover:text-surface-200 text-sm">Cancel</button>
              <button type="submit" class="px-4 py-2 bg-primary-600 hover:bg-primary-500 text-white rounded-xl text-sm">Add</button>
            </div>
          </div>
        </form>
      </div>
    </div>
  {/if}

  <!-- Add Key Modal -->
  {#if showAddKey}
    <div class="fixed inset-0 z-50 flex items-center justify-center">
      <div class="absolute inset-0 bg-black/60 backdrop-blur-sm" onclick={() => showAddKey = false}></div>
      <div class="relative bg-surface-900 border border-surface-700/50 rounded-2xl p-6 w-full max-w-sm shadow-2xl">
        <h2 class="text-xl font-bold text-surface-100 mb-4">Add Translation Key</h2>
        <form onsubmit={(e) => { e.preventDefault(); addKey(); }}>
          <div class="space-y-4">
            <div>
              <label for="kKey" class="block text-sm font-medium text-surface-300 mb-1.5">Key</label>
              <input id="kKey" type="text" bind:value={newKey} placeholder="home.hero.title" required class="w-full px-4 py-2.5 bg-surface-800/50 border border-surface-600/50 rounded-xl text-surface-100 font-mono text-sm focus:outline-none focus:ring-2 focus:ring-primary-500/50 transition-all" />
            </div>
            <div>
              <label for="kDesc" class="block text-sm font-medium text-surface-300 mb-1.5">Description</label>
              <input id="kDesc" type="text" bind:value={newKeyDesc} placeholder="Optional description" class="w-full px-4 py-2.5 bg-surface-800/50 border border-surface-600/50 rounded-xl text-surface-100 focus:outline-none focus:ring-2 focus:ring-primary-500/50 transition-all" />
            </div>
            <div class="flex gap-3 justify-end">
              <button type="button" onclick={() => showAddKey = false} class="px-4 py-2 text-surface-400 hover:text-surface-200 text-sm">Cancel</button>
              <button type="submit" class="px-4 py-2 bg-primary-600 hover:bg-primary-500 text-white rounded-xl text-sm">Add</button>
            </div>
          </div>
        </form>
      </div>
    </div>
  {/if}
{/if}
