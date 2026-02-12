<script lang="ts">
  import { onMount } from 'svelte';
  import { fade } from 'svelte/transition';
  import { api } from '$lib/api/client';
  import { toasts } from '$lib/stores/toast';
  import type { Project } from '$lib/types';

  let projects = $state<Project[]>([]);
  let loading = $state(true);
  let showCreate = $state(false);
  let editingProject = $state<Project | null>(null);
  let formName = $state('');
  let formDescription = $state('');
  let search = $state('');
  let viewMode = $state<'grid' | 'table'>('grid');

  const filteredProjects = $derived(
    projects.filter((p) =>
      p.name.toLowerCase().includes(search.toLowerCase()) ||
      p.slug.toLowerCase().includes(search.toLowerCase())
    )
  );

  onMount(loadProjects);

  async function loadProjects() {
    loading = true;
    try {
      projects = await api.get<Project[]>('/api/projects');
    } catch {
      toasts.error('Failed to load projects');
    } finally {
      loading = false;
    }
  }

  function openCreate() {
    formName = '';
    formDescription = '';
    editingProject = null;
    showCreate = true;
  }

  function openEdit(project: Project) {
    formName = project.name;
    formDescription = project.description;
    editingProject = project;
    showCreate = true;
  }

  async function handleSubmit() {
    try {
      if (editingProject) {
        await api.put(`/api/projects/${editingProject.id}`, {
          name: formName,
          description: formDescription,
        });
        toasts.success('Project updated');
      } else {
        await api.post('/api/projects', {
          name: formName,
          description: formDescription,
        });
        toasts.success('Project created');
      }
      showCreate = false;
      await loadProjects();
    } catch (err: any) {
      toasts.error(err.message || 'Operation failed');
    }
  }

  async function handleDelete(project: Project) {
    if (!confirm(`Delete "${project.name}"? This will remove all translations.`)) return;
    try {
      await api.delete(`/api/projects/${project.id}`);
      toasts.success('Project deleted');
      await loadProjects();
    } catch (err: any) {
      toasts.error(err.message || 'Delete failed');
    }
  }
</script>

<div class="space-y-6">
  <div class="flex flex-col sm:flex-row sm:items-center justify-between gap-4">
    <div>
      <h1 class="text-3xl font-bold text-heading">Projects</h1>
      <p class="text-subtle mt-1">Manage your translation projects</p>
    </div>
    <div class="flex items-center gap-3 self-end sm:self-auto">
      <!-- View Toggle -->
      <div class="flex bg-[var(--bg-input)] rounded-xl p-1">
        <button
          onclick={() => viewMode = 'grid'}
          class="p-2 rounded-lg transition-all {viewMode === 'grid' ? 'bg-[var(--bg-elevated)] shadow-sm text-primary-600' : 'text-subtle hover:text-heading'}"
          title="Grid View"
        >
          <svg xmlns="http://www.w3.org/2000/svg" width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><rect x="3" y="3" width="7" height="7"></rect><rect x="14" y="3" width="7" height="7"></rect><rect x="14" y="14" width="7" height="7"></rect><rect x="3" y="14" width="7" height="7"></rect></svg>
        </button>
        <button
          onclick={() => viewMode = 'table'}
          class="p-2 rounded-lg transition-all {viewMode === 'table' ? 'bg-[var(--bg-elevated)] shadow-sm text-primary-600' : 'text-subtle hover:text-heading'}"
          title="Table View"
        >
          <svg xmlns="http://www.w3.org/2000/svg" width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><line x1="8" y1="6" x2="21" y2="6"></line><line x1="8" y1="12" x2="21" y2="12"></line><line x1="8" y1="18" x2="21" y2="18"></line><line x1="3" y1="6" x2="3.01" y2="6"></line><line x1="3" y1="12" x2="3.01" y2="12"></line><line x1="3" y1="18" x2="3.01" y2="18"></line></svg>
        </button>
      </div>

      <button
        onclick={openCreate}
        class="px-4 py-2.5 bg-gradient-to-r from-primary-600 to-primary-700 hover:from-primary-500 hover:to-primary-600 text-white font-medium rounded-xl transition-all shadow-lg shadow-primary-500/20 text-sm whitespace-nowrap"
      >
        + New Project
      </button>
    </div>
  </div>

  <!-- Search -->
  <input
    type="text"
    bind:value={search}
    placeholder="Search projects..."
    class="themed-input w-full max-w-md px-4 py-2.5 rounded-xl transition-all focus:ring-2 focus:ring-primary-500/20"
  />

  <!-- Content -->
  {#if loading}
    <div class="flex items-center justify-center py-20">
      <div class="animate-spin w-8 h-8 border-2 border-primary-500 border-t-transparent rounded-full"></div>
    </div>
  {:else if filteredProjects.length === 0}
    <div class="themed-card text-center py-20 rounded-2xl border-dashed">
      <div class="text-4xl mb-3">📂</div>
      <p class="text-subtle text-lg font-medium">No projects found</p>
      {#if search}
        <p class="text-faint text-sm mt-1">Try adjusting your search terms</p>
      {:else}
        <p class="text-faint text-sm mt-1">Create your first translation project to get started</p>
        <button 
          onclick={openCreate}
          class="mt-4 px-4 py-2 text-primary-600 hover:bg-primary-500/10 rounded-lg text-sm font-medium transition-colors"
        >
          Create Project
        </button>
      {/if}
    </div>
  {:else}
    {#if viewMode === 'grid'}
      <!-- Grid View -->
      <div class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4" in:fade={{ duration: 200 }}>
        {#each filteredProjects as project (project.id)}
          <div class="themed-card backdrop-blur-xl rounded-2xl p-5 transition-all hover:shadow-elevated hover:border-primary-500/30 group relative overflow-hidden flex flex-col h-full">
            <div class="absolute top-2 right-2 p-2 flex gap-1 z-10">
              <button
                onclick={(e) => { e.preventDefault(); e.stopPropagation(); openEdit(project); }}
                class="p-2 backdrop-blur-sm text-primary-600 bg-[var(--bg-elevated)] rounded-lg shadow-sm"
                title="Edit"
              >
                <svg xmlns="http://www.w3.org/2000/svg" width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="M11 4H4a2 2 0 0 0-2 2v14a2 2 0 0 0 2 2h14a2 2 0 0 0 2-2v-7"></path><path d="M18.5 2.5a2.121 2.121 0 0 1 3 3L12 15l-4 1 1-4 9.5-9.5z"></path></svg>
              </button>
              <button
                onclick={(e) => { e.preventDefault(); e.stopPropagation(); handleDelete(project); }}
                class="p-2 backdrop-blur-sm text-red-500 bg-[var(--bg-elevated)] rounded-lg shadow-sm"
                title="Delete"
              >
                <svg xmlns="http://www.w3.org/2000/svg" width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><polyline points="3 6 5 6 21 6"></polyline><path d="M19 6v14a2 2 0 0 1-2 2H7a2 2 0 0 1-2-2V6m3 0V4a2 2 0 0 1 2-2h4a2 2 0 0 1 2 2v2"></path></svg>
              </button>
            </div>

            <a href="/projects/{project.id}" class="flex-1 block">
              <h3 class="font-bold text-lg text-heading group-hover:text-primary-600 transition-colors leading-tight mb-2 pr-12">{project.name}</h3>
              <p class="text-xs text-subtle font-mono bg-[var(--bg-input)] px-2 py-1 rounded-md inline-block mb-3">{project.slug}</p>
              <p class="text-sm text-subtle line-clamp-3 mb-4">{project.description || 'No description provided.'}</p>
            </a>
            
            <div class="flex items-center justify-between pt-4 border-t border-subtle mt-auto">
              <span class="text-xs text-faint">{new Date(project.created_at).toLocaleDateString()}</span>
              <a
                href="/projects/{project.id}"
                class="text-xs font-medium text-primary-600 hover:text-primary-500 transition-colors flex items-center gap-1 group/link"
              >
                Open <span class="group-hover/link:translate-x-0.5 transition-transform">→</span>
              </a>
            </div>
          </div>
        {/each}
      </div>
    {:else}
      <!-- Table View -->
      <div class="themed-card rounded-2xl overflow-hidden shadow-sm" in:fade={{ duration: 200 }}>
        <div class="overflow-x-auto">
          <table class="w-full text-left border-collapse">
            <thead>
              <tr class="border-b border-subtle bg-[var(--bg-input)]">
                <th class="px-6 py-4 text-xs font-semibold text-subtle uppercase tracking-wider w-1/3">Project</th>
                <th class="px-6 py-4 text-xs font-semibold text-subtle uppercase tracking-wider hidden sm:table-cell">Description</th>
                <th class="px-6 py-4 text-xs font-semibold text-subtle uppercase tracking-wider whitespace-nowrap">Created</th>
                <th class="px-6 py-4 text-xs font-semibold text-subtle uppercase tracking-wider text-right">Actions</th>
              </tr>
            </thead>
            <tbody class="divide-y divide-subtle">
              {#each filteredProjects as project (project.id)}
                <tr class="group hover:bg-[var(--bg-card-hover)] transition-colors">
                  <td class="px-6 py-4">
                    <a href="/projects/{project.id}" class="block">
                      <div class="font-medium text-heading group-hover:text-primary-600 transition-colors">{project.name}</div>
                      <div class="text-xs text-faint font-mono mt-0.5">{project.slug}</div>
                    </a>
                  </td>
                  <td class="px-6 py-4 hidden sm:table-cell">
                    <div class="text-sm text-subtle line-clamp-1 max-w-xs">{project.description || '-'}</div>
                  </td>
                  <td class="px-6 py-4 whitespace-nowrap">
                    <div class="text-sm text-subtle">{new Date(project.created_at).toLocaleDateString()}</div>
                    <div class="text-xs text-faint">{new Date(project.created_at).toLocaleTimeString([], {hour: '2-digit', minute:'2-digit'})}</div>
                  </td>
                  <td class="px-6 py-4 text-right whitespace-nowrap">
                    <div class="flex items-center justify-end gap-2">
                      <button
                        onclick={() => openEdit(project)}
                        class="p-2 text-primary-600 bg-primary-500/10 rounded-lg"
                        title="Edit Project"
                      >
                        <svg xmlns="http://www.w3.org/2000/svg" width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="M11 4H4a2 2 0 0 0-2 2v14a2 2 0 0 0 2 2h14a2 2 0 0 0 2-2v-7"></path><path d="M18.5 2.5a2.121 2.121 0 0 1 3 3L12 15l-4 1 1-4 9.5-9.5z"></path></svg>
                      </button>
                      <button
                        onclick={() => handleDelete(project)}
                        class="p-2 text-red-500 bg-red-500/10 rounded-lg"
                        title="Delete Project"
                      >
                        <svg xmlns="http://www.w3.org/2000/svg" width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><polyline points="3 6 5 6 21 6"></polyline><path d="M19 6v14a2 2 0 0 1-2 2H7a2 2 0 0 1-2-2V6m3 0V4a2 2 0 0 1 2-2h4a2 2 0 0 1 2 2v2"></path></svg>
                      </button>
                      <a 
                        href="/projects/{project.id}"
                        class="p-2 text-subtle hover:text-primary-600 hover:bg-primary-500/10 rounded-lg transition-all"
                        title="Open Project"
                      >
                       <svg xmlns="http://www.w3.org/2000/svg" width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="M18 13v6a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2V8a2 2 0 0 1 2-2h6"></path><polyline points="15 3 21 3 21 9"></polyline><line x1="10" y1="14" x2="21" y2="3"></line></svg>
                      </a>
                    </div>
                  </td>
                </tr>
              {/each}
            </tbody>
          </table>
        </div>
      </div>
    {/if}
  {/if}
</div>

<!-- Create/Edit Modal -->
{#if showCreate}
  <div class="fixed inset-0 z-50 flex items-center justify-center">
    <button class="themed-modal-overlay absolute inset-0 backdrop-blur-sm" aria-label="Hide create modal" onclick={() => showCreate = false}></button>
    <div class="themed-modal relative rounded-2xl p-6 w-full max-w-md animate-in fade-in zoom-in-95 duration-200">
      <h2 class="text-xl font-bold text-heading mb-4">{editingProject ? 'Edit' : 'New'} Project</h2>
      <form onsubmit={(e) => { e.preventDefault(); handleSubmit(); }}>
        <div class="space-y-4">
          <div>
            <label for="projName" class="block text-sm font-medium text-body mb-1.5">Name</label>
            <input
              id="projName"
              type="text"
              bind:value={formName}
              required
              class="themed-input w-full px-4 py-2.5 rounded-xl transition-all focus:ring-2 focus:ring-primary-500/20"
            />
          </div>
          <div>
            <label for="projDesc" class="block text-sm font-medium text-body mb-1.5">Description</label>
            <textarea
              id="projDesc"
              bind:value={formDescription}
              rows={3}
              class="themed-input w-full px-4 py-2.5 rounded-xl transition-all resize-none focus:ring-2 focus:ring-primary-500/20"
            ></textarea>
          </div>
          <div class="flex gap-3 justify-end pt-2">
            <button
              type="button"
              onclick={() => showCreate = false}
              class="px-4 py-2 text-subtle hover:text-heading transition-colors text-sm font-medium"
            >Cancel</button>
            <button
              type="submit"
              class="px-4 py-2 bg-primary-600 hover:bg-primary-500 text-white rounded-xl transition-colors text-sm font-medium shadow-lg shadow-primary-500/20"
            >{editingProject ? 'Update' : 'Create'}</button>
          </div>
        </div>
      </form>
    </div>
  </div>
{/if}
