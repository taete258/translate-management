<script lang="ts">
  import type { Toast } from '$lib/stores/toast';
  import { toasts } from '$lib/stores/toast';

  const typeStyles: Record<Toast['type'], string> = {
    success: 'bg-emerald-600/90 border-emerald-400/50 text-white',
    error: 'bg-red-600/90 border-red-400/50 text-white',
    info: 'bg-blue-600/90 border-blue-400/50 text-white',
    warning: 'bg-amber-600/90 border-amber-400/50 text-white',
  };

  const icons: Record<Toast['type'], string> = {
    success: '✓',
    error: '✕',
    info: 'ℹ',
    warning: '⚠',
  };
</script>

<div class="fixed top-4 right-4 z-50 flex flex-col gap-2 min-w-[320px]">
  {#each $toasts as toast (toast.id)}
    <div
      class="flex items-center gap-3 px-4 py-3 rounded-xl border backdrop-blur-sm shadow-2xl animate-slide-in {typeStyles[toast.type]}"
      role="alert"
    >
      <span class="text-lg font-bold">{icons[toast.type]}</span>
      <span class="flex-1 text-sm font-medium">{toast.message}</span>
      <button
        class="opacity-60 hover:opacity-100 transition-opacity text-lg"
        onclick={() => toasts.dismiss(toast.id)}
      >
        ✕
      </button>
    </div>
  {/each}
</div>

<style>
  @keyframes slide-in {
    from {
      transform: translateX(100%);
      opacity: 0;
    }
    to {
      transform: translateX(0);
      opacity: 1;
    }
  }

  .animate-slide-in {
    animation: slide-in 0.3s ease-out;
  }
</style>
