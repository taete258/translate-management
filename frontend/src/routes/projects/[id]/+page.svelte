<script lang="ts">
  import { onMount } from 'svelte';
  import { page } from '$app/state';
  import { api } from '$lib/api/client';
  import { toasts } from '$lib/stores/toast';
  import type { Project, Language, TranslationEntry, ProjectStats, CacheStatus, ProjectMemberInfo } from '$lib/types';
  import { ChevronDown, ArrowLeft, RefreshCcw, Plus, Star, X, Globe, Trash2, Download, Users, List, FolderTree } from 'lucide-svelte';
  import { fade } from 'svelte/transition';
  import KeyVisualizer from '$lib/components/KeyVisualizer.svelte';

  const projectId = $derived(page.params.id);

  let project = $state<Project | null>(null);
  let languages = $state<Language[]>([]);
  let entries = $state<TranslationEntry[]>([]);
  let stats = $state<ProjectStats | null>(null);
  let cacheStatus = $state<CacheStatus | null>(null);
  let members = $state<ProjectMemberInfo[]>([]);
  let loading = $state(true);
  let saving = $state(false);
  let search = $state('');
  let pendingChanges = $state<Map<string, string>>(new Map());
  let viewMode = $state<'table' | 'visualizer'>('table');

  // Export
  let showExportMenu = $state(false);
  let exporting = $state(false);

  // Language management
  let showAddLang = $state(false);
  let langCode = $state('');
  let langName = $state('');
  let langDefault = $state(false);

  // Key management
  let showAddKey = $state(false);
  let newKey = $state('');
  let newKeyDesc = $state('');

  // Sharing
  let showShare = $state(false);
  let inviteEmail = $state('');
  let inviteRole = $state('viewer');

  const filteredEntries = $derived(
    entries.filter((e) =>
      e.key.toLowerCase().includes(search.toLowerCase())
    )
  );

  const userRole = $derived(project?.role || 'viewer');
  const canEdit = $derived(userRole === 'owner' || userRole === 'editor');
  const isOwner = $derived(userRole === 'owner');

  onMount(() => loadAll(true));

  async function loadAll(initial = false) {
    if (initial) loading = true;
    try {
      const [p, l, t, s, m] = await Promise.all([
        api.get<Project>(`/api/projects/${projectId}`),
        api.get<Language[]>(`/api/projects/${projectId}/languages`),
        api.get<TranslationEntry[]>(`/api/projects/${projectId}/translations`),
        api.get<ProjectStats>(`/api/projects/${projectId}/stats`),
        api.get<ProjectMemberInfo[]>(`/api/projects/${projectId}/members`),
      ]);
      project = p;
      languages = l;
      entries = t;
      stats = s;
      members = m;
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
    if (!canEdit) return;
    const changeKey = `${keyId}:${langId}`;
    pendingChanges.set(changeKey, value);
    pendingChanges = new Map(pendingChanges);
  }

  // ... (keep existing functions)

  // ... in template
  
  // Header with role badge
  // <h1 class="text-3xl font-bold text-heading flex items-center gap-3">
  //   {project.name}
  //   <span class="text-xs px-2 py-1 rounded-full border font-normal
  //     {isOwner ? 'bg-indigo-500/10 text-indigo-500 border-indigo-500/20' : 
  //      canEdit ? 'bg-blue-500/10 text-blue-500 border-blue-500/20' : 
  //      'bg-slate-500/10 text-slate-500 border-slate-500/20'}">
  //     {userRole}
  //   </span>
  // </h1>

  // ...

  // Buttons visibility
  // {#if languages.length > 0 && canEdit}
  //   <button onclick={() => showAddKey = true} ... >
  
  // Translation Input
  // <input disabled={!canEdit} ... />


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

  async function inviteUser() {
    try {
      await api.post(`/api/projects/${projectId}/invitations`, {
        email: inviteEmail,
        role: inviteRole,
      });
      toasts.success(`Invitation sent to ${inviteEmail}`);
      showShare = false;
      inviteEmail = '';
      inviteRole = 'viewer';
    } catch (err: any) {
      toasts.error(err.message || 'Failed to send invitation');
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

  async function exportTranslations(langCode: string, format: 'json' | 'msgpack') {
    exporting = true;
    showExportMenu = false;
    try {
      const blob = await api.get<Blob>(
        `/api/projects/${projectId}/export/${langCode}?format=${format}`,
        { responseType: 'blob' }
      );
      const ext = format === 'msgpack' ? 'msgpack' : 'json';
      const url = URL.createObjectURL(blob);
      const a = document.createElement('a');
      a.href = url;
      a.download = `${project?.slug || 'translations'}_${langCode}.${ext}`;
      document.body.appendChild(a);
      a.click();
      document.body.removeChild(a);
      URL.revokeObjectURL(url);

      toasts.success(`Exported ${langCode} as ${format.toUpperCase()}`);
    } catch (err: any) {
      toasts.error(err.message || 'Export failed');
    } finally {
      exporting = false;
    }
  }

  function handleClickOutsideExport(event: MouseEvent) {
    const target = event.target as HTMLElement;
    if (!target.closest('.export-dropdown')) {
      showExportMenu = false;
    }
  }
</script>

<svelte:window onclick={handleClickOutsideExport} />

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
          <a href="/projects" class="text-faint hover:text-heading text-sm transition-colors flex items-center gap-1">
            <ArrowLeft size={16} /> Projects
          </a>
        </div>
        <h1 class="text-3xl font-bold text-heading flex items-center gap-3">
          {project.name}
          <span class="text-xs px-2 py-1 rounded-full border font-normal capitalize
            {project.role === 'owner' ? 'bg-indigo-500/10 text-indigo-400 border-indigo-500/20' : 
             project.role === 'editor' ? 'bg-blue-500/10 text-blue-400 border-blue-500/20' : 
             'bg-slate-500/10 text-slate-400 border-slate-500/20'}">
            {project.role || 'viewer'}
          </span>
        </h1>
        <div class="flex items-center gap-4 mt-1">
          <p class="text-subtle">{project.description || project.slug}</p>
          <div class="flex items-center -space-x-1.5">
            {#each members.slice(0, 5) as member}
              <div class="group relative">
                <div 
                  class="w-7 h-7 rounded-full border-2 border-[var(--bg-main)] bg-slate-700 flex items-center justify-center overflow-hidden shrink-0"
                >
                  {#if member.avatar_url}
                    <img src={member.avatar_url} alt={member.name} class="w-full h-full object-cover" />
                  {:else}
                    <span class="text-[9px] font-bold text-white uppercase">{member.name.slice(0, 2)}</span>
                  {/if}
                </div>
                <!-- Tooltip -->
                <div class="absolute bottom-full left-1/2 -translate-x-1/2 mb-2 hidden group-hover:block px-2 py-1 bg-slate-900 text-white text-[10px] rounded shadow-xl whitespace-nowrap z-50 border border-slate-700 pointer-events-none">
                  <div class="font-medium">{member.name}</div>
                  <div class="text-[8px] text-slate-400 capitalize">{member.role}</div>
                </div>
              </div>
            {/each}
            {#if members.length > 5}
              <div 
                class="w-7 h-7 rounded-full border-2 border-[var(--bg-main)] bg-slate-800 flex items-center justify-center text-[9px] font-bold text-faint shrink-0"
                title="{members.length - 5} more members"
              >
                +{members.length - 5}
              </div>
            {/if}
          </div>
        </div>
      </div>
      <div class="flex gap-2">
        <div class="flex items-center gap-2">
          {#if cacheStatus}
            <span class="text-xs px-2 py-1 rounded-lg {cacheStatus.cached ? 'bg-emerald-600/20 text-emerald-500' : 'text-faint'}"
              style={cacheStatus.cached ? '' : 'background: var(--bg-input);'}
            >
              {cacheStatus.cached ? `Cached (${cacheStatus.cached_keys})` : 'Not cached'}
            </span>
          {/if}

          <!-- Export Dropdown -->
          {#if languages.length > 0}
            <div class="relative export-dropdown">
              <button
                onclick={() => showExportMenu = !showExportMenu}
                disabled={exporting}
                class="px-3 py-2 bg-violet-600/20 text-violet-400 hover:bg-violet-600/30 border border-violet-500/30 rounded-xl text-sm transition-all disabled:opacity-50 flex items-center gap-1.5"
                title="Export translations"
              >
                {#if exporting}
                  <span class="animate-spin inline-block w-3.5 h-3.5 border-2 border-violet-400 border-t-transparent rounded-full"></span>
                  Exporting...
                {:else}
                  <Download size={14} /> Export
                  <ChevronDown size={14} class="transition-transform {showExportMenu ? 'rotate-180' : ''}" />
                {/if}
              </button>

              {#if showExportMenu}
                <!-- svelte-ignore a11y_no_static_element_interactions -->
                <div
                  class="absolute right-0 top-full mt-2 w-72 rounded-2xl shadow-2xl z-50 overflow-hidden"
                  style="background: var(--bg-modal); border: 1px solid var(--border-subtle);"
                >
                  <div class="px-4 py-3" style="border-bottom: 1px solid var(--border-subtle);">
                    <p class="text-sm font-semibold text-heading">Export Translations</p>
                    <p class="text-xs text-faint mt-0.5">Select language &amp; format</p>
                  </div>
                  <div class="max-h-64 overflow-y-auto">
                    {#each languages as lang}
                      <div class="px-4 py-2.5 flex items-center justify-between gap-2 transition-colors"
                        style="border-bottom: 1px solid var(--border-subtle);"
                      >
                        <div class="flex items-center gap-2 min-w-0">
                          <span class="text-sm text-body truncate">{lang.name}</span>
                          <span class="text-xs font-mono text-faint">({lang.code})</span>
                          {#if lang.is_default}
                            <span class="text-xs text-primary-500">★</span>
                          {/if}
                        </div>
                        <div class="flex gap-1.5 shrink-0">
                          <button
                            onclick={() => exportTranslations(lang.code, 'json')}
                            class="px-2.5 py-1 text-xs font-medium rounded-lg bg-blue-600/15 text-blue-400 hover:bg-blue-600/30 border border-blue-500/20 hover:border-blue-500/40 transition-all"
                            title="Export as JSON"
                          >
                            JSON
                          </button>
                          <button
                            onclick={() => exportTranslations(lang.code, 'msgpack')}
                            class="px-2.5 py-1 text-xs font-medium rounded-lg bg-orange-600/15 text-orange-400 hover:bg-orange-600/30 border border-orange-500/20 hover:border-orange-500/40 transition-all"
                            title="Export as MessagePack"
                          >
                            MsgPack
                          </button>
                        </div>
                      </div>
                    {/each}
                  </div>
                </div>
              {/if}
            </div>
          {/if}

          {#if isOwner}
            <button
              onclick={() => showShare = true}
              class="px-3 py-2 bg-blue-600/20 text-blue-400 hover:bg-blue-600/30 border border-blue-500/30 rounded-xl text-sm transition-all flex items-center gap-1.5"
              title="Share Project"
            >
              <Users size={14} /> Share
            </button>
          {/if}

          <button
            onclick={invalidateCache}
            class="px-3 py-2 bg-amber-600/20 text-amber-500 hover:bg-amber-600/30 border border-amber-500/30 rounded-xl text-sm transition-all"
            title="Force invalidate cache"
          >
            <span class="flex items-center gap-1.5">
              <RefreshCcw size={14} /> Invalidate Cache
            </span>
          </button>
        </div>
      </div>
    </div>

    <!-- Stats Bar -->
    {#if stats}
      <div class="flex gap-4 flex-wrap">
        <div class="themed-card px-4 py-2 rounded-xl">
          <span class="text-sm text-subtle">Keys:</span>
          <span class="text-sm font-medium text-heading ml-1">{stats.total_keys}</span>
        </div>
        <div class="themed-card px-4 py-2 rounded-xl">
          <span class="text-sm text-subtle">Languages:</span>
          <span class="text-sm font-medium text-heading ml-1">{stats.total_languages}</span>
        </div>
        {#each Object.entries(stats.language_progress) as [code, pct]}
          <div class="themed-card px-4 py-2 rounded-xl">
            <span class="text-sm text-subtle">{code}:</span>
            <span class="text-sm font-medium {pct === 100 ? 'text-emerald-500' : pct > 50 ? 'text-amber-500' : 'text-red-500'} ml-1">
              {pct.toFixed(0)}%
            </span>
          </div>
        {/each}
      </div>
    {/if}

    <!-- Toolbar -->
    <div class="flex items-center gap-3 flex-wrap">
      <!-- Search -->
      <input
        type="text"
        bind:value={search}
        placeholder="Search keys..."
        class="themed-input px-4 py-2 rounded-xl text-sm w-64"
      />
      
      <!-- View Toggle -->
      <div class="flex p-1 bg-[var(--bg-card)] border border-[var(--border-subtle)] rounded-xl">
        <button
            onclick={() => viewMode = 'table'}
            class="p-1.5 rounded-lg transition-all {viewMode === 'table' ? 'bg-[var(--bg-modal)] text-heading shadow-sm' : 'text-faint hover:text-subtle'}"
            title="Table View"
        >
            <List size={16} />
        </button>
        <button
            onclick={() => viewMode = 'visualizer'}
            class="p-1.5 rounded-lg transition-all {viewMode === 'visualizer' ? 'bg-[var(--bg-modal)] text-heading shadow-sm' : 'text-faint hover:text-subtle'}"
            title="Tree View"
        >
            <FolderTree size={16} />
        </button>
      </div>

      {#if languages.length > 0 && canEdit}
        <button
          onclick={() => showAddKey = true}
          class="px-3 py-2 bg-primary-600/20 text-primary-500 hover:bg-primary-600/30 border border-primary-500/30 rounded-xl text-sm transition-all flex items-center gap-1.5"
        >
          <Plus size={14} /> Add Key
        </button>
      {/if}
      {#if canEdit}
        <button
          onclick={() => showAddLang = true}
          class="px-3 py-2 bg-emerald-600/20 text-emerald-500 hover:bg-emerald-600/30 border border-emerald-500/30 rounded-xl text-sm transition-all flex items-center gap-1.5"
        >
          <Plus size={14} /> Add Language
        </button>
      {/if}

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

    <!-- Languages list (Horizontal badges - visible in both views) -->
    {#if languages.length > 0}
      <div class="flex gap-2 flex-wrap">
        {#each languages as lang}
          <span class="inline-flex items-center gap-1.5 px-3 py-1.5 rounded-lg text-sm"
            style="background: var(--bg-input); border: 1px solid var(--border-subtle);">
            <span class="text-body">{lang.name}</span>
            <span class="text-faint font-mono text-xs">({lang.code})</span>
            {#if lang.is_default}
              <Star size={12} class="text-primary-500 fill-current" />
            {/if}
            {#if canEdit}
              <button
                onclick={() => deleteLanguage(lang)}
                class="text-faint hover:text-red-500 transition-colors ml-1"
                title="Remove"
              >
                <X size={14} />
              </button>
            {/if}
          </span>
        {/each}
      </div>
    {/if}

    <!-- Content Area -->
    {#if languages.length === 0}
      <div class="themed-card text-center py-16 rounded-2xl">
        <div class="mb-3 flex justify-center text-subtle">
           <Globe size={40} />
        </div>
        <p class="text-subtle text-lg mb-2">No languages added yet</p>
        <p class="text-faint text-sm mb-4">Add at least one language before creating translation keys</p>
        <button
          onclick={() => showAddLang = true}
          class="inline-flex items-center gap-1.5 px-4 py-2 bg-emerald-600 hover:bg-emerald-500 text-white rounded-lg transition-colors text-sm"
        >
          <Plus size={14} /> Add Your First Language
        </button>
      </div>
    {:else if entries.length === 0}
      <div class="themed-card text-center py-16 rounded-2xl">
        <p class="text-subtle text-lg mb-2">No translation keys</p>
        <p class="text-faint text-sm">Add keys to start managing translations</p>
      </div>
    {:else}
        {#if viewMode === 'table'}
            <div in:fade={{ duration: 200 }} class="themed-card overflow-x-auto rounded-2xl">
                <table class="w-full text-sm">
                <thead>
                    <tr style="border-bottom: 1px solid var(--border-subtle);">
                    <th class="text-left px-4 py-3 text-subtle font-medium sticky left-0 min-w-[200px]"
                        style="background: var(--bg-modal);">Key</th>
                    {#each languages as lang}
                        <th class="text-left px-4 py-3 text-subtle font-medium min-w-[200px]">
                        {lang.name} <span class="text-faint font-normal">({lang.code})</span>
                        </th>
                    {/each}
                    <th class="w-10"></th>
                    </tr>
                </thead>
                <tbody>
                    {#each filteredEntries as entry}
                    <tr class="transition-colors" style="border-bottom: 1px solid var(--border-subtle);"
                        onmouseenter={(e) => e.currentTarget.style.background = 'var(--bg-card-hover)'}
                        onmouseleave={(e) => e.currentTarget.style.background = ''}>
                        <td class="px-4 py-2 sticky left-0" style="background: var(--bg-modal);">
                        <div class="font-mono text-heading text-xs">{entry.key}</div>
                        {#if entry.description}
                            <div class="text-xs text-faint mt-0.5">{entry.description}</div>
                        {/if}
                        </td>
                        {#each languages as lang}
                        <td class="px-4 py-2">
                            <input
                            type="text"
                            disabled={!canEdit}
                            value={pendingChanges.get(`${entry.key_id}:${lang.id}`) ?? entry.values[lang.id] ?? ''}
                            oninput={(e) => handleCellChange(entry.key_id, lang.id, (e.target as HTMLInputElement).value)}
                            class="w-full px-2.5 py-1.5 bg-transparent border border-transparent rounded-lg text-sm focus:outline-none transition-all {pendingChanges.has(`${entry.key_id}:${lang.id}`) ? 'border-amber-500/40 bg-amber-500/5' : ''} disabled:opacity-50 disabled:cursor-not-allowed"
                            style="color: var(--text-primary);"
                            placeholder={canEdit ? "—" : ""}
                            />
                        </td>
                        {/each}
                        <td class="px-2 py-2">
                        {#if canEdit}
                            <button
                            onclick={() => deleteKey(entry.key_id, entry.key)}
                            class="p-1 text-faint hover:text-red-500 rounded transition-all"
                            title="Delete key"
                            >
                            <Trash2 size={16} />
                            </button>
                        {/if}
                        </td>
                    </tr>
                    {/each}
                </tbody>
                </table>
            </div>
        {:else}
            <!-- Visualizer View -->
            <div in:fade={{ duration: 200 }} class="themed-card rounded-2xl p-2 min-h-[400px]">
                <KeyVisualizer 
                    entries={entries}
                    languages={languages}
                    search={search}
                    canEdit={canEdit}
                    pendingChanges={pendingChanges}
                    onCellChange={handleCellChange}
                    onDeleteKey={deleteKey}
                />
            </div>
        {/if}
    {/if}
  </div>

  <!-- Add Language Modal -->
  {#if showAddLang}
    <div class="fixed inset-0 z-50 flex items-center justify-center">
      <button class="themed-modal-overlay absolute inset-0 backdrop-blur-sm" aria-label="Close" onclick={() => showAddLang = false}></button>
      <div class="themed-modal relative rounded-2xl p-6 w-full max-w-sm">
        <h2 class="text-xl font-bold text-heading mb-4">Add Language</h2>
        <form onsubmit={(e) => { e.preventDefault(); addLanguage(); }}>
          <div class="space-y-4">
            <div>
              <label for="lCode" class="block text-sm font-medium text-body mb-1.5">Code</label>
              <input id="lCode" type="text" bind:value={langCode} placeholder="en, th, ja..." required class="themed-input w-full px-4 py-2.5 rounded-xl transition-all" />
            </div>
            <div>
              <label for="lName" class="block text-sm font-medium text-body mb-1.5">Name</label>
              <input id="lName" type="text" bind:value={langName} placeholder="English" required class="themed-input w-full px-4 py-2.5 rounded-xl transition-all" />
            </div>
            <label class="flex items-center gap-2 text-sm text-body cursor-pointer">
              <input type="checkbox" bind:checked={langDefault} class="accent-primary-500" />
              Set as default language
            </label>
            <div class="flex gap-3 justify-end">
              <button type="button" onclick={() => showAddLang = false} class="px-4 py-2 text-subtle hover:text-heading text-sm">Cancel</button>
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
      <button class="themed-modal-overlay absolute inset-0 backdrop-blur-sm" aria-label="Close" onclick={() => showAddKey = false}></button>
      <div class="themed-modal relative rounded-2xl p-6 w-full max-w-sm">
        <h2 class="text-xl font-bold text-heading mb-4">Add Translation Key</h2>
        <form onsubmit={(e) => { e.preventDefault(); addKey(); }}>
          <div class="space-y-4">
            <div>
              <label for="kKey" class="block text-sm font-medium text-body mb-1.5">Key</label>
              <input id="kKey" type="text" bind:value={newKey} placeholder="home.hero.title" required class="themed-input w-full px-4 py-2.5 rounded-xl font-mono text-sm transition-all" />
            </div>
            <div>
              <label for="kDesc" class="block text-sm font-medium text-body mb-1.5">Description</label>
              <input id="kDesc" type="text" bind:value={newKeyDesc} placeholder="Optional description" class="themed-input w-full px-4 py-2.5 rounded-xl transition-all" />
            </div>
            <div class="flex gap-3 justify-end">
              <button type="button" onclick={() => showAddKey = false} class="px-4 py-2 text-subtle hover:text-heading text-sm">Cancel</button>
              <button type="submit" class="px-4 py-2 bg-primary-600 hover:bg-primary-500 text-white rounded-xl text-sm">Add</button>
            </div>
          </div>
        </form>
      </div>
    </div>
  {/if}

  <!-- Share Modal -->
  {#if showShare}
    <div class="fixed inset-0 z-50 flex items-center justify-center">
      <button class="themed-modal-overlay absolute inset-0 backdrop-blur-sm" aria-label="Close" onclick={() => showShare = false}></button>
      <div class="themed-modal relative rounded-2xl p-6 w-full max-w-sm">
        <h2 class="text-xl font-bold text-heading mb-4">Share Project</h2>
        <form onsubmit={(e) => { e.preventDefault(); inviteUser(); }}>
          <div class="space-y-4">
            <div>
              <label for="iEmail" class="block text-sm font-medium text-body mb-1.5">Email Address</label>
              <input id="iEmail" type="email" bind:value={inviteEmail} placeholder="colleague@example.com" required class="themed-input w-full px-4 py-2.5 rounded-xl transition-all" />
            </div>
            <div>
              <label for="iRole" class="block text-sm font-medium text-body mb-1.5">Role</label>
              <select id="iRole" bind:value={inviteRole} class="themed-input w-full px-4 py-2.5 rounded-xl transition-all appearance-none bg-transparent">
                <option value="viewer">Viewer (Read-only)</option>
                <option value="editor">Editor (Can edit translations)</option>
                <option value="owner">Owner (Full access)</option>
              </select>
            </div>
            <div class="flex gap-3 justify-end">
              <button type="button" onclick={() => showShare = false} class="px-4 py-2 text-subtle hover:text-heading text-sm">Cancel</button>
              <button type="submit" class="px-4 py-2 bg-blue-600 hover:bg-blue-500 text-white rounded-xl text-sm">Send Invite</button>
            </div>
          </div>
        </form>
      </div>
    </div>
  {/if}
{/if}
