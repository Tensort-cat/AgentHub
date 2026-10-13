import type { User } from "../types/domain";

const AUTH_STORAGE_KEY = "agenthub.auth.v1";

export interface StoredAuth {
  token: string;
  user: User;
}

export function readAuth(): StoredAuth | null {
  try {
    const stored = window.localStorage.getItem(AUTH_STORAGE_KEY);
    return stored ? (JSON.parse(stored) as StoredAuth) : null;
  } catch {
    return null;
  }
}

export function writeAuth(auth: StoredAuth): void {
  window.localStorage.setItem(AUTH_STORAGE_KEY, JSON.stringify(auth));
  window.dispatchEvent(new Event("agenthub:auth-change"));
}

export function clearAuth(): void {
  window.localStorage.removeItem(AUTH_STORAGE_KEY);
  window.dispatchEvent(new Event("agenthub:auth-change"));
}
