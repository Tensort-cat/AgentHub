import type { User } from "../types/domain";

const AUTH_STORAGE_KEY = "agenthub.auth.v1";

export interface StoredAuth {
  token: string;
  user: User;
}

let memoryCache: StoredAuth | null | undefined;

export function readAuth(): StoredAuth | null {
  if (memoryCache !== undefined) return memoryCache;
  try {
    const stored = window.localStorage.getItem(AUTH_STORAGE_KEY);
    memoryCache = stored ? (JSON.parse(stored) as StoredAuth) : null;
  } catch {
    memoryCache = null;
  }
  return memoryCache;
}

export function writeAuth(auth: StoredAuth): void {
  memoryCache = auth;
  window.localStorage.setItem(AUTH_STORAGE_KEY, JSON.stringify(auth));
  window.dispatchEvent(new Event("agenthub:auth-change"));
}

export function clearAuth(): void {
  memoryCache = null;
  window.localStorage.removeItem(AUTH_STORAGE_KEY);
  window.dispatchEvent(new Event("agenthub:auth-change"));
}

