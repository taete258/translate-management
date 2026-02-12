import { writable } from "svelte/store";
import { browser } from "$app/environment";

type Theme = "dark" | "light";

function getInitialTheme(): Theme {
  if (!browser) return "dark";
  const stored = localStorage.getItem("theme") as Theme | null;
  if (stored === "light" || stored === "dark") return stored;
  // Check system preference
  if (window.matchMedia("(prefers-color-scheme: light)").matches)
    return "light";
  return "dark";
}

function createThemeStore() {
  const { subscribe, set, update } = writable<Theme>(getInitialTheme());

  function applyTheme(theme: Theme) {
    if (!browser) return;
    const html = document.documentElement;
    html.classList.toggle("dark", theme === "dark");
    html.classList.toggle("light", theme === "light");
    localStorage.setItem("theme", theme);
  }

  // Apply on creation
  if (browser) {
    const initial = getInitialTheme();
    applyTheme(initial);
  }

  return {
    subscribe,
    toggle() {
      update((current) => {
        const next = current === "dark" ? "light" : "dark";
        applyTheme(next);
        return next;
      });
    },
    set(theme: Theme) {
      applyTheme(theme);
      set(theme);
    },
  };
}

export const theme = createThemeStore();
