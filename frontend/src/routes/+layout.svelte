<script lang="ts">
  import { page } from '$app/state';
  import { browser } from '$app/environment';
  import { goto } from '$app/navigation';
  import { auth } from '$lib/stores/auth';
  import Sidebar from '$lib/components/Sidebar.svelte';
  import Toast from '$lib/components/Toast.svelte';
  import '../app.css';

  let { children } = $props();

  const publicRoutes = ['/login', '/register'];
  const isPublicRoute = $derived(publicRoutes.includes(page.url.pathname));

  $effect(() => {
    if (browser && !isPublicRoute && !$auth.user) {
      goto('/login');
    }
  });
</script>

<Toast />

{#if isPublicRoute}
  <main class="min-h-screen">
    {@render children()}
  </main>
{:else if $auth.user}
  <div class="flex min-h-screen">
    <Sidebar />
    <main class="flex-1 overflow-auto">
      <div class="p-6 max-w-full mx-auto">
        {@render children()}
      </div>
    </main>
  </div>
{:else}
  <div class="min-h-screen flex items-center justify-center">
    <div class="animate-spin w-8 h-8 border-2 border-primary-500 border-t-transparent rounded-full"></div>
  </div>
{/if}
