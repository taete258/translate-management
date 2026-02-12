export interface User {
  id: string;
  email: string;
  username: string;
  name: string;
  avatar_url: string;
  created_at: string;
  updated_at: string;
}

export interface AuthResponse {
  token: string;
  user: User;
}

export interface Project {
  id: string;
  name: string;
  slug: string;
  description: string;
  created_by?: string;
  created_at: string;
  updated_at: string;
}

export interface Language {
  id: string;
  project_id: string;
  code: string;
  name: string;
  is_default: boolean;
  created_at: string;
}

export interface TranslationKey {
  id: string;
  project_id: string;
  key: string;
  description: string;
  created_at: string;
  updated_at: string;
}

export interface TranslationEntry {
  key_id: string;
  key: string;
  description: string;
  values: Record<string, string>; // language_id -> value
}

export interface TranslationUpdate {
  key_id: string;
  language_id: string;
  value: string;
}

export interface APIKey {
  id: string;
  project_id: string;
  name: string;
  key_prefix: string;
  scopes: string[];
  is_active: boolean;
  last_used_at?: string;
  created_at: string;
}

export interface CreateAPIKeyResponse {
  api_key: APIKey;
  raw_key: string;
}

export interface ProjectStats {
  total_keys: number;
  total_languages: number;
  language_progress: Record<string, number>;
}

export interface CacheStatus {
  project: string;
  cached: boolean;
  cached_keys: number;
}
