<script lang="ts">
  import { fade, scale } from 'svelte/transition';

  interface Props {
    show: boolean;
    title: string;
    message: string;
    confirmText?: string;
    cancelText?: string;
    type?: 'danger' | 'primary';
    onConfirm: () => void;
    onCancel: () => void;
  }

  let {
    show = $bindable(false),
    title,
    message,
    confirmText = 'Confirm',
    cancelText = 'Cancel',
    type = 'primary',
    onConfirm,
    onCancel
  }: Props = $props();

  function handleCancel() {
    show = false;
    onCancel();
  }

  function handleConfirm() {
    show = false;
    onConfirm();
  }
</script>

{#if show}
  <div class="fixed inset-0 z-[100] flex items-center justify-center p-4">
    <!-- Backdrop -->
    <button 
      class="absolute inset-0 bg-black/40 backdrop-blur-sm transition-opacity" 
      onclick={handleCancel}
      aria-label="Close modal"
    ></button>

    <!-- Modal Content -->
    <div 
      class="themed-modal relative w-full max-w-sm rounded-2xl p-6 shadow-2xl"
      in:scale={{ duration: 200, start: 0.95 }}
      out:scale={{ duration: 150, start: 0.95 }}
    >
      <div class="mb-5">
        <h3 class="text-xl font-bold text-heading mb-2">{title}</h3>
        <p class="text-subtle text-sm leading-relaxed">{message}</p>
      </div>

      <div class="flex gap-3 justify-end items-center">
        <button
          onclick={handleCancel}
          class="px-4 py-2 text-sm font-medium text-subtle hover:text-heading transition-colors"
        >
          {cancelText}
        </button>
        <button
          onclick={handleConfirm}
          class="px-5 py-2 text-sm font-semibold rounded-xl transition-all shadow-lg
            {type === 'danger' 
              ? 'bg-red-600 hover:bg-red-500 text-white shadow-red-500/20' 
              : 'bg-primary-600 hover:bg-primary-500 text-white shadow-primary-500/20'}"
        >
          {confirmText}
        </button>
      </div>
    </div>
  </div>
{/if}
