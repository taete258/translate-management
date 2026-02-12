<script lang="ts">
  import { page } from '$app/state';
  import { auth } from '$lib/stores/auth';

  const navItems = [
    { href: '/', label: 'Dashboard', icon: '📊' },
    { href: '/projects', label: 'Projects', icon: '📁' },
    { href: '/api-keys', label: 'API Keys', icon: '🔑' },
  ];

  let collapsed = $state(false);
</script>

<aside
  class="h-screen sticky top-0 flex flex-col bg-surface-900/80 backdrop-blur-xl border-r border-surface-700/50 transition-all duration-300 {collapsed ? 'w-16' : 'w-64'}"
>
  <!-- Logo -->
  <div class="flex items-center gap-3 px-4 h-16 border-b border-surface-700/50">
    <div class="w-8 h-8 rounded-lg bg-gradient-to-br from-primary-500 to-primary-700 flex items-center justify-center text-white font-bold text-sm shrink-0">
      T
    </div>
    {#if !collapsed}
      <span class="font-semibold text-surface-100 text-sm whitespace-nowrap">Translate Mgmt</span>
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
            ? 'bg-primary-600/20 text-primary-400 border border-primary-500/30'
            : 'text-surface-400 hover:text-surface-100 hover:bg-surface-800/60'}"
      >
        <span class="text-base shrink-0">{item.icon}</span>
        {#if !collapsed}
          <span class="whitespace-nowrap">{item.label}</span>
        {/if}
      </a>
    {/each}
  </nav>

  <!-- User section -->
  <div class="border-t border-surface-700/50 p-3">
    {#if !collapsed}
      <div class="flex items-center gap-3 mb-2 px-1">
        <div class="w-8 h-8 rounded-full bg-gradient-to-br from-primary-400 to-primary-600 flex items-center justify-center text-white text-xs font-bold shrink-0">
          {$auth.user?.name?.charAt(0).toUpperCase() || 'U'}
        </div>
        <div class="flex-1 min-w-0">
          <p class="text-sm font-medium text-surface-200 truncate">{$auth.user?.name || 'User'}</p>
          <p class="text-xs text-surface-500 truncate">{$auth.user?.email || ''}</p>
        </div>
      </div>
    {/if}
    <button
      onclick={() => auth.logout()}
      class="w-full flex items-center gap-3 px-3 py-2 rounded-lg text-sm text-surface-400 hover:text-red-400 hover:bg-red-500/10 transition-all"
    >
      <span class="shrink-0">🚪</span>
      {#if !collapsed}
        <span>Logout</span>
      {/if}
    </button>
  </div>

  <!-- Collapse toggle -->
  <button
    onclick={() => collapsed = !collapsed}
    class="absolute -right-3 top-20 w-6 h-6 rounded-full bg-surface-800 border border-surface-600 flex items-center justify-center text-surface-400 hover:text-surface-200 transition-colors text-xs"
  >
    {collapsed ? '→' : '←'}
  </button>
</aside>
