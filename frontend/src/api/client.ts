import { clearAuth, readAuth } from "../lib/auth-storage";
import { parseJson, stringifyJson } from "../lib/json";

export const API_BASE_URL = import.meta.env.VITE_API_BASE_URL ?? "/api/v1";

interface ApiEnvelope<T> {
  code: number;
  msg: string;
  data: T;
}

interface RequestOptions extends Omit<RequestInit, "body"> {
  json?: unknown;
  body?: BodyInit;
}

export class ApiError extends Error {
  constructor(
    message: string,
    readonly code: number,
    readonly status: number,
  ) {
    super(message);
    this.name = "ApiError";
  }
}

export function apiUrl(path: string): string {
  return `${API_BASE_URL}${path}`;
}

export async function apiRequest<T>(
  path: string,
  options: RequestOptions = {},
): Promise<T> {
  const headers = new Headers(options.headers);
  const token = readAuth()?.token;
  if (token) headers.set("Authorization", token);

  let body = options.body;
  if (options.json !== undefined) {
    headers.set("Content-Type", "application/json");
    body = stringifyJson(options.json);
  }

  const response = await fetch(apiUrl(path), { ...options, headers, body });
  const text = await response.text();
  let envelope: ApiEnvelope<T> | null = null;
  if (text) {
    try {
      envelope = parseJson<ApiEnvelope<T>>(text);
    } catch {
      throw new ApiError("The server returned an unreadable response.", 4001, response.status);
    }
  }

  if (response.status === 401 || envelope?.code === 1001) {
    clearAuth();
    if (window.location.pathname !== "/login") window.location.assign("/login");
  }

  if (!response.ok || !envelope || envelope.code !== 0) {
    const fallback = response.status === 403
      ? "You do not have permission to perform this action."
      : "The request could not be completed.";
    throw new ApiError(envelope?.msg || fallback, envelope?.code ?? 4001, response.status);
  }

  return envelope.data;
}

export function errorMessage(error: unknown, fallback: string): string {
  return error instanceof Error && error.message ? error.message : fallback;
}

