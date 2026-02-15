<script lang="ts">
  import { slide } from 'svelte/transition';
  import { ChevronRight, ChevronDown, Folder, FileText, Trash2 } from 'lucide-svelte';
  import type { TranslationEntry, Language } from '$lib/types';

  interface Props {
    entries: TranslationEntry[];
    languages: Language[];
    search?: string;
    canEdit: boolean;
    pendingChanges: Map<string, string>;
    onCellChange: (keyId: string, langId: string, value: string) => void;
    onDeleteKey: (keyId: string, keyName: string) => void;
  }

  let { 
    entries, 
    languages, 
    search = '', 
    canEdit, 
    pendingChanges, 
    onCellChange, 
    onDeleteKey 
  }: Props = $props();

  interface TreeNode {
    name: string;
    fullPath: string;
    children: Map<string, TreeNode>;
    entry?: TranslationEntry;
    isOpen: boolean;
  }

  function buildTree(entries: TranslationEntry[], filter: string): Map<string, TreeNode> {
    const root = new Map<string, TreeNode>();

    for (const entry of entries) {
      if (filter && !entry.key.toLowerCase().includes(filter.toLowerCase())) {
        continue;
      }

      const parts = entry.key.split('.');
      let currentLevel = root;
      let pathSoFar = '';

      for (let i = 0; i < parts.length; i++) {
        const part = parts[i];
        pathSoFar = pathSoFar ? `${pathSoFar}.${part}` : part;

        if (!currentLevel.has(part)) {
          currentLevel.set(part, {
            name: part,
            fullPath: pathSoFar,
            children: new Map(),
            isOpen: true, // Default open for now
          });
        }

        const node = currentLevel.get(part)!;
        
        // If it's the leaf node (last part), attach the entry
        if (i === parts.length - 1) {
          node.entry = entry;
        }

        currentLevel = node.children;
      }
    }
    return root;
  }

  // Reactive tree generation
  let tree = $derived(buildTree(entries, search));

  // Grid column definition: Key column + 1 column per language
  let gridStyle = $derived(`display: grid; grid-template-columns: minmax(300px, 1.5fr) repeat(${languages.length}, minmax(200px, 1fr)); gap: 0.5rem;`);
</script>

<!-- Header Row -->
<div class="mb-2 px-2 py-2 border-b border-[var(--border-subtle)] font-medium text-subtle text-sm items-center" style={gridStyle}>
    <div>Key</div>
    {#each languages as lang}
        <div class="truncate" title={lang.name}>
            {lang.name} <span class="text-faint font-normal">({lang.code})</span>
        </div>
    {/each}
</div>

{#snippet TreeNodeItem({ node, languages, canEdit, pendingChanges, onCellChange, onDeleteKey, depth }: { 
    node: TreeNode, 
    languages: Language[], 
    canEdit: boolean, 
    pendingChanges: Map<string, string>, 
    onCellChange: (k: string, l: string, v: string) => void,
    onDeleteKey: (k: string, n: string) => void,
    depth: number
  })}
  
  <div class="select-none">
    <!-- Row Content -->
    <div 
      class="items-center py-1.5 px-2 hover:bg-slate-100/50 dark:hover:bg-slate-800/50 rounded-lg group transition-colors"
      style={gridStyle}
    >
      <!-- Column 1: Key Tree -->
      <div class="flex items-center gap-2 min-w-0" style="padding-left: {depth * 1.5}rem">
          <!-- Toggle Button (if has children) -->
          {#if node.children.size > 0}
              <button 
                onclick={() => node.isOpen = !node.isOpen}
                class="p-0.5 text-faint hover:text-heading transition-colors shrink-0"
              >
                {#if node.isOpen}
                  <ChevronDown size={14} />
                {:else}
                  <ChevronRight size={14} />
                {/if}
              </button>
          {:else}
              <div class="w-4.5 shrink-0"></div> <!-- Spacer -->
          {/if}

          <!-- Icon -->
          {#if node.children.size > 0}
              <Folder size={16} class="text-blue-400 shrink-0" />
          {:else}
              <FileText size={16} class="text-emerald-400 shrink-0" />
          {/if}

          <!-- Name & Meta -->
          <div class="flex flex-col min-w-0 flex-1">
            <div class="flex items-center gap-2">
                <span class="text-sm font-medium text-heading truncate" title={node.name}>{node.name}</span>
                {#if node.children.size > 0}
                    <span class="text-xs text-faint">({node.children.size})</span>
                {/if}
            </div>
          </div>

          <!-- Actions (Delete) -->
          {#if canEdit && node.entry}
            <button
                onclick={() => onDeleteKey(node.entry!.key_id, node.entry!.key)}
                class="opacity-0 group-hover:opacity-100 p-1 text-faint hover:text-red-500 transition-all ml-auto shrink-0"
                title="Delete key"
            >
                <Trash2 size={14} />
            </button>
          {/if}
      </div>

      <!-- Columns 2..N: Languages -->
      {#each languages as lang}
          <div class="min-w-0">
             {#if node.entry}
                <input
                    type="text"
                    disabled={!canEdit}
                    value={pendingChanges.get(`${node.entry.key_id}:${lang.id}`) ?? node.entry.values[lang.id] ?? ''}
                    oninput={(e) => onCellChange(node.entry!.key_id, lang.id, (e.currentTarget as HTMLInputElement).value)}
                    class="w-full px-2.5 py-1.5 bg-transparent border border-transparent hover:border-slate-200 dark:hover:border-slate-700 rounded-lg text-sm focus:outline-none focus:border-primary-500 hover:bg-[var(--bg-card)] focus:bg-[var(--bg-card)] transition-all {pendingChanges.has(`${node.entry.key_id}:${lang.id}`) ? 'border-amber-500/40 bg-amber-500/5' : ''}"
                    placeholder={canEdit ? "-" : ""}
                />
             {/if}
          </div>
      {/each}
    </div>

    <!-- Children (if open) -->
    {#if node.isOpen && node.children.size > 0}
        <div transition:slide={{ duration: 200 }}>
            {#each node.children.values() as child (child.fullPath)}
                {@render TreeNodeItem({ node: child, languages, canEdit, pendingChanges, onCellChange, onDeleteKey, depth: depth + 1 })}
            {/each}
        </div>
    {/if}
  </div>
{/snippet}

<div class="space-y-0.5">
  {#each tree.values() as node (node.fullPath)}
    {@render TreeNodeItem({ node, languages, canEdit, pendingChanges, onCellChange, onDeleteKey, depth: 0 })}
  {/each}
</div>
