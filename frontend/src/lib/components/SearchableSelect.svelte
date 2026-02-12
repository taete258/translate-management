<script lang="ts">
  import { onMount } from 'svelte';
  
  interface Item {
    [key: string]: any;
  }

  let { 
    items = [] as Item[], 
    value = $bindable(''), 
    placeholder = 'Select...', 
    labelKey = 'name', 
    valueKey = 'id',
    className = ''
  } = $props();

  let isOpen = $state(false);
  let searchTerm = $state('');
  let containerRef: HTMLDivElement;

  // Derived filtered items
  let filteredItems = $derived(
    items.filter(item => 
      String(item[labelKey]).toLowerCase().includes(searchTerm.toLowerCase())
    )
  );

  let selectedItem = $derived(items.find(i => i[valueKey] === value));

  function toggle() {
    isOpen = !isOpen;
    if (isOpen) {
      setTimeout(() => {
        const input = containerRef?.querySelector('input');
        input?.focus();
      }, 50);
    } else {
        searchTerm = '';
    }
  }

  function select(item: Item) {
    value = item[valueKey];
    isOpen = false;
    searchTerm = '';
  }

  function handleClickOutside(event: MouseEvent) {
    if (containerRef && !containerRef.contains(event.target as Node)) {
      isOpen = false;
    }
  }

  onMount(() => {
    document.addEventListener('click', handleClickOutside);
    return () => {
      document.removeEventListener('click', handleClickOutside);
    };
  });
</script>

<div class="relative {className}" bind:this={containerRef}>
  <button
    type="button"
    onclick={toggle}
    class="themed-input w-full px-4 py-2 rounded-xl text-left flex items-center justify-between transition-all"
    aria-expanded={isOpen}
  >
    <span class={!selectedItem ? "text-muted" : "text-heading"}>
      {selectedItem ? selectedItem[labelKey] : placeholder}
    </span>
    <span class="text-subtle text-xs transition-transform {isOpen ? 'rotate-180' : ''}">▼</span>
  </button>

  {#if isOpen}
    <div class="absolute z-50 left-0 right-0 top-full mt-2 themed-card rounded-xl shadow-elevated overflow-hidden max-h-60 flex flex-col">
      <div class="p-2 border-b border-subtle">
        <input
          type="text"
          bind:value={searchTerm}
          placeholder="Search..."
          class="w-full px-3 py-1.5 rounded-lg themed-input text-sm focus:outline-none focus:ring-1 focus:ring-primary-500 text-heading placeholder:text-muted"
        />
      </div>
      <div class="overflow-y-auto flex-1">
        {#if filteredItems.length === 0}
          <div class="px-4 py-3 text-sm text-faint text-center">No results found</div>
        {:else}
          {#each filteredItems as item}
            <button
              type="button"
              onclick={() => select(item)}
              class="w-full text-left px-4 py-2 text-sm themed-card transition-colors {item[valueKey] === value ? 'text-primary-600 font-medium' : 'text-body'}"
            >
              {item[labelKey]}
            </button>
          {/each}
        {/if}
      </div>
    </div>
  {/if}
</div>

<style>
    /* Custom scrollbar for the dropdown */
    .overflow-y-auto::-webkit-scrollbar {
        width: 6px;
    }
    .overflow-y-auto::-webkit-scrollbar-track {
        background: transparent;
    }
    .overflow-y-auto::-webkit-scrollbar-thumb {
        background-color: var(--color-surface-300);
        border-radius: 3px;
    }
    :global(.dark) .overflow-y-auto::-webkit-scrollbar-thumb {
        background-color: var(--color-surface-600);
    }
</style>
