<script lang="ts">
  import { theme } from '$lib/stores/theme';
</script>

<button
  onclick={() => theme.toggle()}
  class="theme-toggle"
  title={$theme === 'dark' ? 'Switch to light mode' : 'Switch to dark mode'}
  aria-label="Toggle theme"
>
  <div class="toggle-track" class:dark={$theme === 'dark'}>
    <span class="toggle-icon sun">☀️</span>
    <span class="toggle-icon moon">🌙</span>
    <div class="toggle-thumb" class:dark={$theme === 'dark'}></div>
  </div>
</button>

<style>
  .theme-toggle {
    background: none;
    border: none;
    cursor: pointer;
    padding: 0;
    outline: none;
  }

  .toggle-track {
    position: relative;
    width: 52px;
    height: 28px;
    border-radius: 14px;
    background: linear-gradient(135deg, #60a5fa, #3b82f6);
    transition: background 0.3s ease;
    display: flex;
    align-items: center;
    justify-content: space-between;
    padding: 0 6px;
  }

  .toggle-track.dark {
    background: linear-gradient(135deg, #334155, #1e293b);
  }

  .toggle-icon {
    font-size: 14px;
    line-height: 1;
    z-index: 1;
    transition: opacity 0.3s ease;
  }

  .toggle-track.dark .sun {
    opacity: 0.3;
  }

  .toggle-track:not(.dark) .moon {
    opacity: 0.3;
  }

  .toggle-thumb {
    position: absolute;
    top: 3px;
    left: 3px;
    width: 22px;
    height: 22px;
    border-radius: 50%;
    background: white;
    box-shadow: 0 1px 3px rgba(0, 0, 0, 0.3);
    transition: transform 0.3s cubic-bezier(0.68, -0.55, 0.27, 1.55);
  }

  .toggle-thumb.dark {
    transform: translateX(24px);
  }
</style>
