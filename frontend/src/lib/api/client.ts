import { browser } from "$app/environment";

import { env } from "$env/dynamic/public";

const API_URL = browser
  ? env.PUBLIC_API_URL || "http://localhost:3000"
  : // @ts-ignore
    (typeof process !== "undefined" && process?.env?.INTERNAL_API_URL) ||
    "http://backend:3000";

interface RequestOptions {
  method?: string;
  body?: unknown;
  headers?: Record<string, string>;
  responseType?: "json" | "blob" | "text";
}

class ApiClient {
  private baseUrl: string;

  constructor(baseUrl: string) {
    this.baseUrl = baseUrl;
  }

  private getToken(): string | null {
    if (!browser) return null;
    return localStorage.getItem("token");
  }

  async request<T>(endpoint: string, options: RequestOptions = {}): Promise<T> {
    const {
      method = "GET",
      body,
      headers = {},
      responseType = "json",
    } = options;
    const token = this.getToken();

    const config: RequestInit = {
      method,
      headers: {
        "Content-Type": "application/json",
        ...headers,
        ...(token ? { Authorization: `Bearer ${token}` } : {}),
      },
    };

    if (body) {
      config.body = JSON.stringify(body);
    }

    const response = await fetch(`${this.baseUrl}${endpoint}`, config);

    if (!response.ok) {
      const error = await response
        .json()
        .catch(() => ({ error: "Request failed" }));
      throw new ApiError(
        response.status,
        error.error || error.message || "Request failed",
      );
    }

    if (response.status === 204) {
      return {} as T;
    }

    if (responseType === "blob") {
      return response.blob() as Promise<T>;
    }

    if (responseType === "text") {
      return response.text() as Promise<T>;
    }

    return response.json();
  }

  get<T>(endpoint: string, options: RequestOptions = {}) {
    return this.request<T>(endpoint, { ...options, method: "GET" });
  }

  post<T>(endpoint: string, body?: unknown) {
    return this.request<T>(endpoint, { method: "POST", body });
  }

  put<T>(endpoint: string, body?: unknown) {
    return this.request<T>(endpoint, { method: "PUT", body });
  }

  delete<T>(endpoint: string) {
    return this.request<T>(endpoint, { method: "DELETE" });
  }
}

export class ApiError extends Error {
  status: number;

  constructor(status: number, message: string) {
    super(message);
    this.status = status;
    this.name = "ApiError";
  }
}

export const api = new ApiClient(API_URL);
