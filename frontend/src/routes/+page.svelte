<script lang="ts">
  import { onMount } from 'svelte';
  import { api } from '$lib/api/client';
  import type { Project, ProjectInvitation } from '$lib/types';
  import { Folder, Globe, Key, ArrowRight, Mail, Check, X } from 'lucide-svelte';
  import { toasts } from '$lib/stores/toast';

  let projects = $state<Project[]>([]);
  let invitations = $state<ProjectInvitation[]>([]);
  let loading = $state(true);

  onMount(async () => {
    try {
      const [p, i] = await Promise.all([
        api.get<Project[]>('/api/projects'),
        api.get<ProjectInvitation[]>('/api/invitations')
      ]);
      projects = p;
      invitations = i;
    } catch {
      // silent
    } finally {
      loading = false;
    }
  });

  async function respond(invitation: ProjectInvitation, accept: boolean) {
    try {
      await api.post(`/api/invitations/${invitation.id}/respond`, { accept });
      toasts.success(`Invitation ${accept ? 'accepted' : 'rejected'}`);
      // Refresh
      const [p, i] = await Promise.all([
        api.get<Project[]>('/api/projects'),
        api.get<ProjectInvitation[]>('/api/invitations')
      ]);
      projects = p;
      invitations = i;
    } catch (err: any) {
      toasts.error(err.message || 'Failed to respond');
    }
  }
</script>

<div class="space-y-8">
  <!-- Header -->
  <div>
    <h1 class="text-3xl font-bold text-heading">Dashboard</h1>
    <p class="text-subtle mt-1">Overview of your translation projects</p>
  </div>

  {#if invitations.length > 0}
    <div class="themed-card backdrop-blur-xl rounded-2xl p-6 border-l-4 border-l-blue-500">
      <h2 class="text-lg font-semibold text-heading mb-4 flex items-center gap-2">
        <Mail size={20} class="text-blue-500" /> Pending Invitations
      </h2>
      <div class="space-y-3">
        {#each invitations as inv}
          <div class="flex items-center justify-between p-4 rounded-xl" style="background: var(--bg-input);">
            <div>
              <p class="text-sm font-medium text-heading">
                You were invited to join <span class="text-primary-500">{inv.project_name}</span>
              </p>
              <p class="text-xs text-faint mt-0.5">
                Invited by {inv.inviter_name} as <span class="capitalize">{inv.role}</span> · {new Date(inv.created_at).toLocaleDateString()}
              </p>
            </div>
            <div class="flex gap-2">
              <button
                onclick={() => respond(inv, false)}
                class="p-2 text-red-500 hover:bg-red-500/10 rounded-lg transition-colors"
                title="Reject"
              >
                <X size={18} />
              </button>
              <button
                onclick={() => respond(inv, true)}
                class="px-3 py-1.5 bg-primary-600 hover:bg-primary-500 text-white text-sm font-medium rounded-lg transition-colors flex items-center gap-1.5"
                title="Accept"
              >
                <Check size={16} /> Accept
              </button>
            </div>
          </div>
        {/each}
      </div>
    </div>
  {/if}

  <!-- Stats Grid -->
  <div class="grid grid-cols-1 md:grid-cols-3 gap-4">
    <div class="themed-card backdrop-blur-xl rounded-2xl p-6">
      <div class="flex items-center gap-4">
        <div class="w-12 h-12 rounded-xl bg-primary-600/20 flex items-center justify-center text-primary-600">
          <Folder size={24} />
        </div>
        <div>
          <p class="text-sm text-subtle">Total Projects</p>
          <p class="text-2xl font-bold text-heading">{loading ? '...' : projects.length}</p>
        </div>
      </div>
    </div>

    <div class="themed-card backdrop-blur-xl rounded-2xl p-6">
      <div class="flex items-center gap-4">
        <div class="w-12 h-12 rounded-xl bg-emerald-600/20 flex items-center justify-center text-emerald-600">
          <Globe size={24} />
        </div>
        <div>
          <p class="text-sm text-subtle">Active Translations</p>
          <p class="text-2xl font-bold text-heading">{loading ? '...' : projects.length > 0 ? 'Active' : 'None'}</p>
        </div>
      </div>
    </div>

    <div class="themed-card backdrop-blur-xl rounded-2xl p-6">
      <div class="flex items-center gap-4">
        <div class="w-12 h-12 rounded-xl bg-amber-600/20 flex items-center justify-center text-amber-600">
          <Key size={24} />
        </div>
        <div>
          <p class="text-sm text-subtle">API Integration</p>
          <p class="text-2xl font-bold text-heading">Ready</p>
        </div>
      </div>
    </div>
  </div>

  <!-- Recent Projects -->
  <div class="themed-card backdrop-blur-xl rounded-2xl p-6">
    <div class="flex items-center justify-between mb-4">
      <h2 class="text-lg font-semibold text-heading">Recent Projects</h2>
      <a
        href="/projects"
        class="text-sm text-primary-500 hover:text-primary-400 transition-colors flex items-center gap-1 group"
      >View all <ArrowRight size={14} class="group-hover:translate-x-0.5 transition-transform" /></a>
    </div>

    {#if loading}
      <div class="flex items-center justify-center py-12">
        <div class="animate-spin w-6 h-6 border-2 border-primary-500 border-t-transparent rounded-full"></div>
      </div>
    {:else if projects.length === 0}
      <div class="text-center py-12">
        <p class="text-subtle text-lg mb-2">No projects yet</p>
        <p class="text-faint text-sm mb-4">Create your first translation project to get started</p>
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
            class="flex items-center justify-between p-4 rounded-xl transition-all group"
            style="background: var(--bg-input); border: 1px solid var(--border-subtle);"
            onmouseenter={(e) => { e.currentTarget.style.background = 'var(--bg-card-hover)'; }}
            onmouseleave={(e) => { e.currentTarget.style.background = 'var(--bg-input)'; }}
          >
            <div>
              <h3 class="font-medium text-heading">{project.name}</h3>
              <p class="text-sm text-faint">{project.slug} · {project.description || 'No description'}</p>
            </div>
            <span class="text-faint group-hover:text-heading transition-colors">
              <ArrowRight size={16} />
            </span>
          </a>
        {/each}
      </div>
    {/if}
  </div>
</div>
