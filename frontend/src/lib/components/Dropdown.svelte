<script lang="ts" generics="T">
  import { type Snippet } from 'svelte';

  interface Props {
    items?: T[];
    class?: string;          // Trigger wrapper class
    dropdownClass?: string;  // Dropdown container class
    listClass?: string;      // Items container class
    trigger: Snippet<[boolean]>; 
    header?: Snippet<[() => void]>;
    item?: Snippet<[T, () => void]>;
    footer?: Snippet<[() => void]>;
    empty?: Snippet;
  }

  let { 
    items = [], 
    class: className = '', 
    dropdownClass = 'w-56',
    listClass = 'max-h-64 overflow-y-auto',
    trigger, 
    header, 
    item, 
    footer,
    empty
  }: Props = $props();

  let isOpen = $state(false);
  let dropdownRef: HTMLDivElement;

  function toggle() {
    isOpen = !isOpen;
  }

  function close() {
    isOpen = false;
  }

  function handleOutsideClick(event: MouseEvent) {
    if (isOpen && dropdownRef && !dropdownRef.contains(event.target as Node)) {
      close();
    }
  }
</script>

<svelte:window onclick={handleOutsideClick} />

<div class="relative inline-block {className}" bind:this={dropdownRef}>
  <div role="button" tabindex="0" onclick={toggle} onkeydown={(e) => e.key === 'Enter' && toggle()} class="outline-none">
    {@render trigger(isOpen)}
  </div>

  {#if isOpen}
    <div 
      class="absolute z-50 mt-2 rounded-2xl shadow-2xl overflow-hidden bg-[var(--bg-modal)] border border-[var(--border-subtle)] {dropdownClass}"
      role="menu"
    >
      {#if header}
        {@render header(close)}
      {/if}
      
      {#if items.length > 0 && item}
        <div class="{listClass}">
          {#each items as i}
            {@render item(i, close)}
          {/each}
        </div>
      {:else if empty}
        {@render empty()}
      {/if}

      {#if footer}
        {@render footer(close)}
      {/if}
    </div>
  {/if}
</div>
