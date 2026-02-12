<script lang="ts">
  import { onMount } from 'svelte';
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
  <div class="flex items-center justify-between">
    <div>
      <h1 class="text-3xl font-bold text-surface-100">Projects</h1>
      <p class="text-surface-400 mt-1">Manage your translation projects</p>
    </div>
    <button
      onclick={openCreate}
      class="px-4 py-2.5 bg-gradient-to-r from-primary-600 to-primary-700 hover:from-primary-500 hover:to-primary-600 text-white font-medium rounded-xl transition-all shadow-lg shadow-primary-500/20 text-sm"
    >
      + New Project
    </button>
  </div>

  <!-- Search -->
  <input
    type="text"
    bind:value={search}
    placeholder="Search projects..."
    class="w-full max-w-md px-4 py-2.5 bg-surface-800/50 border border-surface-700/50 rounded-xl text-surface-100 placeholder-surface-500 focus:outline-none focus:ring-2 focus:ring-primary-500/50 transition-all"
  />

  <!-- Projects Grid -->
  {#if loading}
    <div class="flex items-center justify-center py-20">
      <div class="animate-spin w-8 h-8 border-2 border-primary-500 border-t-transparent rounded-full"></div>
    </div>
  {:else if filteredProjects.length === 0}
    <div class="text-center py-20 bg-surface-900/40 rounded-2xl border border-surface-700/30">
      <p class="text-surface-400 text-lg">No projects found</p>
      <p class="text-surface-500 text-sm mt-1">Create your first project to get started</p>
    </div>
  {:else}
    <div class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4">
      {#each filteredProjects as project}
        <div class="bg-surface-900/60 backdrop-blur-xl border border-surface-700/50 rounded-2xl p-5 hover:border-surface-600/50 transition-all group">
          <div class="flex items-start justify-between mb-3">
            <a href="/projects/{project.id}" class="flex-1">
              <h3 class="font-semibold text-surface-200 group-hover:text-primary-400 transition-colors">{project.name}</h3>
              <p class="text-xs text-surface-500 mt-0.5 font-mono">{project.slug}</p>
            </a>
            <div class="flex gap-1">
              <button
                onclick={() => openEdit(project)}
                class="p-1.5 text-surface-500 hover:text-surface-300 hover:bg-surface-700/50 rounded-lg transition-all text-sm"
                title="Edit"
              >✏️</button>
              <button
                onclick={() => handleDelete(project)}
                class="p-1.5 text-surface-500 hover:text-red-400 hover:bg-red-500/10 rounded-lg transition-all text-sm"
                title="Delete"
              >🗑️</button>
            </div>
          </div>
          <p class="text-sm text-surface-400 mb-4 line-clamp-2">{project.description || 'No description'}</p>
          <div class="flex items-center justify-between">
            <span class="text-xs text-surface-500">{new Date(project.created_at).toLocaleDateString()}</span>
            <a
              href="/projects/{project.id}"
              class="text-xs text-primary-400 hover:text-primary-300 transition-colors"
            >Open →</a>
          </div>
        </div>
      {/each}
    </div>
  {/if}
</div>

<!-- Create/Edit Modal -->
{#if showCreate}
  <div class="fixed inset-0 z-50 flex items-center justify-center">
    <div class="absolute inset-0 bg-black/60 backdrop-blur-sm" onclick={() => showCreate = false}></div>
    <div class="relative bg-surface-900 border border-surface-700/50 rounded-2xl p-6 w-full max-w-md shadow-2xl">
      <h2 class="text-xl font-bold text-surface-100 mb-4">{editingProject ? 'Edit' : 'New'} Project</h2>
      <form onsubmit={(e) => { e.preventDefault(); handleSubmit(); }}>
        <div class="space-y-4">
          <div>
            <label for="projName" class="block text-sm font-medium text-surface-300 mb-1.5">Name</label>
            <input
              id="projName"
              type="text"
              bind:value={formName}
              required
              class="w-full px-4 py-2.5 bg-surface-800/50 border border-surface-600/50 rounded-xl text-surface-100 focus:outline-none focus:ring-2 focus:ring-primary-500/50 transition-all"
            />
          </div>
          <div>
            <label for="projDesc" class="block text-sm font-medium text-surface-300 mb-1.5">Description</label>
            <textarea
              id="projDesc"
              bind:value={formDescription}
              rows={3}
              class="w-full px-4 py-2.5 bg-surface-800/50 border border-surface-600/50 rounded-xl text-surface-100 focus:outline-none focus:ring-2 focus:ring-primary-500/50 transition-all resize-none"
            ></textarea>
          </div>
          <div class="flex gap-3 justify-end">
            <button
              type="button"
              onclick={() => showCreate = false}
              class="px-4 py-2 text-surface-400 hover:text-surface-200 transition-colors text-sm"
            >Cancel</button>
            <button
              type="submit"
              class="px-4 py-2 bg-primary-600 hover:bg-primary-500 text-white rounded-xl transition-colors text-sm"
            >{editingProject ? 'Update' : 'Create'}</button>
          </div>
        </div>
      </form>
    </div>
  </div>
{/if}
