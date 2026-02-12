<script lang="ts">
  import { auth } from '$lib/stores/auth';
  import { toasts } from '$lib/stores/toast';

  let username = $state('');
  let password = $state('');
  let error = $state('');

  async function handleLogin() {
    error = '';
    try {
      await auth.login(username, password);
      toasts.success('Welcome back!');
    } catch (err: any) {
      error = err.message || 'Login failed';
    }
  }
</script>

<div class="min-h-screen flex items-center justify-center bg-gradient-to-br from-surface-950 via-surface-900 to-primary-950/30 px-4">
  <div class="w-full max-w-md">
    <!-- Logo -->
    <div class="text-center mb-8">
      <div class="w-16 h-16 rounded-2xl bg-gradient-to-br from-primary-500 to-primary-700 flex items-center justify-center text-white font-bold text-2xl mx-auto mb-4 shadow-lg shadow-primary-500/20">
        T
      </div>
      <h1 class="text-2xl font-bold text-surface-100">Welcome Back</h1>
      <p class="text-surface-400 mt-1">Sign in to your translation management account</p>
    </div>

    <!-- Form -->
    <form
      onsubmit={(e) => { e.preventDefault(); handleLogin(); }}
      class="bg-surface-900/60 backdrop-blur-xl border border-surface-700/50 rounded-2xl p-8 shadow-2xl"
    >
      {#if error}
        <div class="mb-4 p-3 rounded-lg bg-red-500/10 border border-red-500/30 text-red-400 text-sm">
          {error}
        </div>
      {/if}

      <div class="space-y-5">
        <div>
          <label for="username" class="block text-sm font-medium text-surface-300 mb-1.5">Username</label>
          <input
            id="username"
            type="text"
            bind:value={username}
            placeholder="Enter your username"
            required
            class="w-full px-4 py-2.5 bg-surface-800/50 border border-surface-600/50 rounded-xl text-surface-100 placeholder-surface-500 focus:outline-none focus:ring-2 focus:ring-primary-500/50 focus:border-primary-500/50 transition-all"
          />
        </div>

        <div>
          <label for="password" class="block text-sm font-medium text-surface-300 mb-1.5">Password</label>
          <input
            id="password"
            type="password"
            bind:value={password}
            placeholder="Enter your password"
            required
            class="w-full px-4 py-2.5 bg-surface-800/50 border border-surface-600/50 rounded-xl text-surface-100 placeholder-surface-500 focus:outline-none focus:ring-2 focus:ring-primary-500/50 focus:border-primary-500/50 transition-all"
          />
        </div>

        <button
          type="submit"
          disabled={$auth.loading}
          class="w-full py-2.5 px-4 bg-gradient-to-r from-primary-600 to-primary-700 hover:from-primary-500 hover:to-primary-600 text-white font-medium rounded-xl transition-all duration-200 shadow-lg shadow-primary-500/20 hover:shadow-primary-500/30 disabled:opacity-50 disabled:cursor-not-allowed"
        >
          {$auth.loading ? 'Signing in...' : 'Sign In'}
        </button>
      </div>

      <p class="mt-6 text-center text-sm text-surface-400">
        Don't have an account?
        <a href="/register" class="text-primary-400 hover:text-primary-300 transition-colors font-medium">Create one</a>
      </p>
    </form>
  </div>
</div>
