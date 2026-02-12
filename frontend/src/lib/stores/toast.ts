import { writable } from "svelte/store";

export interface Toast {
  id: number;
  message: string;
  type: "success" | "error" | "info" | "warning";
}

let nextId = 0;

function createToastStore() {
  const { subscribe, update } = writable<Toast[]>([]);

  return {
    subscribe,
    show: (message: string, type: Toast["type"] = "info") => {
      const id = nextId++;
      update((toasts) => [...toasts, { id, message, type }]);
      setTimeout(() => {
        update((toasts) => toasts.filter((t) => t.id !== id));
      }, 4000);
    },
    success: (message: string) => {
      const id = nextId++;
      update((toasts) => [...toasts, { id, message, type: "success" }]);
      setTimeout(() => {
        update((toasts) => toasts.filter((t) => t.id !== id));
      }, 4000);
    },
    error: (message: string) => {
      const id = nextId++;
      update((toasts) => [...toasts, { id, message, type: "error" }]);
      setTimeout(() => {
        update((toasts) => toasts.filter((t) => t.id !== id));
      }, 5000);
    },
    dismiss: (id: number) => {
      update((toasts) => toasts.filter((t) => t.id !== id));
    },
  };
}

export const toasts = createToastStore();
