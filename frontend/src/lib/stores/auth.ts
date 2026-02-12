import { writable } from "svelte/store";
import { browser } from "$app/environment";
import { goto } from "$app/navigation";
import { api } from "$lib/api/client";
import type { User, AuthResponse } from "$lib/types";

function createAuthStore() {
  const storedToken = browser ? localStorage.getItem("token") : null;
  const storedUser = browser ? localStorage.getItem("user") : null;

  const { subscribe, set, update } = writable<{
    user: User | null;
    token: string | null;
    loading: boolean;
  }>({
    user: storedUser ? JSON.parse(storedUser) : null,
    token: storedToken,
    loading: false,
  });

  return {
    subscribe,
    login: async (username: string, password: string) => {
      update((s) => ({ ...s, loading: true }));
      try {
        const res = await api.post<AuthResponse>("/api/auth/login", {
          username,
          password,
        });
        if (browser) {
          localStorage.setItem("token", res.token);
          localStorage.setItem("user", JSON.stringify(res.user));
        }
        set({ user: res.user, token: res.token, loading: false });
        goto("/");
      } catch (err) {
        update((s) => ({ ...s, loading: false }));
        throw err;
      }
    },
    register: async (
      email: string,
      username: string,
      password: string,
      name: string,
    ) => {
      update((s) => ({ ...s, loading: true }));
      try {
        const res = await api.post<AuthResponse>("/api/auth/register", {
          email,
          username,
          password,
          name,
        });
        if (browser) {
          localStorage.setItem("token", res.token);
          localStorage.setItem("user", JSON.stringify(res.user));
        }
        set({ user: res.user, token: res.token, loading: false });
        goto("/");
      } catch (err) {
        update((s) => ({ ...s, loading: false }));
        throw err;
      }
    },
    logout: () => {
      if (browser) {
        localStorage.removeItem("token");
        localStorage.removeItem("user");
      }
      set({ user: null, token: null, loading: false });
      goto("/login");
    },
    checkAuth: async () => {
      const token = browser ? localStorage.getItem("token") : null;
      if (!token) return;
      try {
        const user = await api.get<User>("/api/auth/me");
        update((s) => ({ ...s, user }));
      } catch {
        if (browser) {
          localStorage.removeItem("token");
          localStorage.removeItem("user");
        }
        set({ user: null, token: null, loading: false });
      }
    },
  };
}

export const auth = createAuthStore();
