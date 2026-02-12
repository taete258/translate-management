<script lang="ts">
  import { onMount } from 'svelte';
  import { api } from '$lib/api/client';
  import type { Project } from '$lib/types';

  let projects = $state<Project[]>([]);
  let loading = $state(true);

  onMount(async () => {
    try {
      projects = await api.get<Project[]>('/api/projects');
    } catch {
      // silent
    } finally {
      loading = false;
    }
  });
</script>

<div class="space-y-8">
  <!-- Header -->
  <div>
    <h1 class="text-3xl font-bold text-surface-100">Dashboard</h1>
    <p class="text-surface-400 mt-1">Overview of your translation projects</p>
  </div>

  <!-- Stats Grid -->
  <div class="grid grid-cols-1 md:grid-cols-3 gap-4">
    <div class="bg-surface-900/60 backdrop-blur-xl border border-surface-700/50 rounded-2xl p-6">
      <div class="flex items-center gap-4">
        <div class="w-12 h-12 rounded-xl bg-primary-600/20 flex items-center justify-center text-2xl">📁</div>
        <div>
          <p class="text-sm text-surface-400">Total Projects</p>
          <p class="text-2xl font-bold text-surface-100">{loading ? '...' : projects.length}</p>
        </div>
      </div>
    </div>

    <div class="bg-surface-900/60 backdrop-blur-xl border border-surface-700/50 rounded-2xl p-6">
      <div class="flex items-center gap-4">
        <div class="w-12 h-12 rounded-xl bg-emerald-600/20 flex items-center justify-center text-2xl">🌐</div>
        <div>
          <p class="text-sm text-surface-400">Active Translations</p>
          <p class="text-2xl font-bold text-surface-100">{loading ? '...' : projects.length > 0 ? 'Active' : 'None'}</p>
        </div>
      </div>
    </div>

    <div class="bg-surface-900/60 backdrop-blur-xl border border-surface-700/50 rounded-2xl p-6">
      <div class="flex items-center gap-4">
        <div class="w-12 h-12 rounded-xl bg-amber-600/20 flex items-center justify-center text-2xl">🔑</div>
        <div>
          <p class="text-sm text-surface-400">API Integration</p>
          <p class="text-2xl font-bold text-surface-100">Ready</p>
        </div>
      </div>
    </div>
  </div>

  <!-- Recent Projects -->
  <div class="bg-surface-900/60 backdrop-blur-xl border border-surface-700/50 rounded-2xl p-6">
    <div class="flex items-center justify-between mb-4">
      <h2 class="text-lg font-semibold text-surface-100">Recent Projects</h2>
      <a
        href="/projects"
        class="text-sm text-primary-400 hover:text-primary-300 transition-colors"
      >View all →</a>
    </div>

    {#if loading}
      <div class="flex items-center justify-center py-12">
        <div class="animate-spin w-6 h-6 border-2 border-primary-500 border-t-transparent rounded-full"></div>
      </div>
    {:else if projects.length === 0}
      <div class="text-center py-12">
        <p class="text-surface-400 text-lg mb-2">No projects yet</p>
        <p class="text-surface-500 text-sm mb-4">Create your first translation project to get started</p>
        <a
          href="/projects"
          class="inline-flex px-4 py-2 bg-primary-600 hover:bg-primary-500 text-white rounded-lg transition-colors text-sm"
        >
          Create Project
        </a>
      </div>
    {:else}
      <div class="space-y-3">
        {#each projects.slice(0, 5) as project}
          <a
            href="/projects/{project.id}"
            class="flex items-center justify-between p-4 rounded-xl bg-surface-800/40 hover:bg-surface-800/70 border border-surface-700/30 hover:border-surface-600/50 transition-all group"
          >
            <div>
              <h3 class="font-medium text-surface-200 group-hover:text-surface-100">{project.name}</h3>
              <p class="text-sm text-surface-500">{project.slug} · {project.description || 'No description'}</p>
            </div>
            <span class="text-surface-500 group-hover:text-surface-300 transition-colors">→</span>
          </a>
        {/each}
      </div>
    {/if}
  </div>
</div>
