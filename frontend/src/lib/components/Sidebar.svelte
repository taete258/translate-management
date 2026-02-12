<script lang="ts">
  import { page } from '$app/state';
  import { auth } from '$lib/stores/auth';
  import ThemeToggle from '$lib/components/ThemeToggle.svelte';
  import { LayoutDashboard, Folder, Key, LogOut, ChevronLeft, ChevronRight } from 'lucide-svelte';

  const navItems = [
    { href: '/', label: 'Dashboard', icon: LayoutDashboard },
    { href: '/projects', label: 'Projects', icon: Folder },
    { href: '/api-keys', label: 'API Keys', icon: Key },
  ];

  let collapsed = $state(false);
</script>

<aside
  class="themed-sidebar h-screen sticky top-0 flex flex-col backdrop-blur-xl transition-all duration-300 {collapsed ? 'w-16' : 'w-64'}"
>
  <!-- Logo -->
  <div class="flex items-center gap-3 px-4 h-16 border-b" style="border-color: var(--border-subtle);">
    <div class="w-8 h-8 rounded-lg bg-gradient-to-br from-primary-500 to-primary-700 flex items-center justify-center text-white font-bold text-sm shrink-0">
      T
    </div>
    {#if !collapsed}
      <span class="font-semibold text-heading text-sm whitespace-nowrap">Translate Mgmt</span>
    {/if}
  </div>

  <!-- Navigation -->
  <nav class="flex-1 py-4 px-2 space-y-1">
    {#each navItems as item}
      {@const isActive = page.url.pathname === item.href || (item.href !== '/' && page.url.pathname.startsWith(item.href))}
      <a
        href={item.href}
        class="flex items-center gap-3 px-3 py-2.5 rounded-lg text-sm transition-all duration-200
          {isActive
            ? 'bg-primary-600/20 text-primary-500 font-medium'
            : 'text-subtle hover:text-heading'}"
        style={isActive ? '' : ''}
        onmouseenter={(e) => { if (!isActive) e.currentTarget.style.background = 'var(--bg-card-hover)'; }}
        onmouseleave={(e) => { if (!isActive) e.currentTarget.style.background = ''; }}
      >
        <span class="text-base shrink-0">
          <item.icon size={20} />
        </span>
        {#if !collapsed}
          <span class="whitespace-nowrap">{item.label}</span>
        {/if}
      </a>
    {/each}
  </nav>

  <!-- Theme toggle + User section -->
  <div class="border-t p-3" style="border-color: var(--border-subtle);">
    {#if !collapsed}
      <!-- Theme toggle -->
      <div class="flex items-center justify-between mb-3 px-1">
        <span class="text-xs text-subtle font-medium uppercase tracking-wider">Theme</span>
        <ThemeToggle />
      </div>

      <div class="flex items-center gap-3 mb-2 px-1">
        <div class="w-8 h-8 rounded-full bg-gradient-to-br from-primary-400 to-primary-600 flex items-center justify-center text-white text-xs font-bold shrink-0">
          {$auth.user?.name?.charAt(0).toUpperCase() || 'U'}
        </div>
        <div class="flex-1 min-w-0">
          <p class="text-sm font-medium text-heading truncate">{$auth.user?.name || 'User'}</p>
          <p class="text-xs text-faint truncate">{$auth.user?.email || ''}</p>
        </div>
      </div>
    {:else}
      <div class="flex justify-center mb-2">
        <ThemeToggle />
      </div>
    {/if}
    <button
      onclick={() => auth.logout()}
      class="w-full flex items-center gap-3 px-3 py-2 rounded-lg text-sm text-subtle hover:text-red-500 transition-all"
      onmouseenter={(e) => e.currentTarget.style.background = 'rgba(239, 68, 68, 0.1)'}
      onmouseleave={(e) => e.currentTarget.style.background = ''}
    >
      <span class="shrink-0">
        <LogOut size={20} />
      </span>
      {#if !collapsed}
        <span>Logout</span>
      {/if}
    </button>
  </div>

  <!-- Collapse toggle -->
  <button
    onclick={() => collapsed = !collapsed}
    class="absolute -right-3 top-20 w-6 h-6 rounded-full flex items-center justify-center text-subtle hover:text-heading transition-colors text-xs"
    style="background: var(--bg-elevated); border: 1px solid var(--border-default);"
  >
    {#if collapsed}
      <ChevronRight size={14} />
    {:else}
      <ChevronLeft size={14} />
    {/if}
  </button>
</aside>
