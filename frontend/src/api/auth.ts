import { writeAuth } from "../lib/auth-storage";
import { normalizeId } from "../lib/json";
import type { JsonId, User } from "../types/domain";
import { apiRequest } from "./client";

interface LoginResponse {
  token: string;
  id: JsonId;
  name: string;
  email: string;
  avatar: string;
}

function toUser(data: Omit<LoginResponse, "token">): User {
  return { ...data, id: normalizeId(data.id) };
}

export async function login(email: string, password: string): Promise<User> {
  const data = await apiRequest<LoginResponse>("/auth/login", {
    method: "POST",
    json: { email, password },
  });
  const user = toUser(data);
  writeAuth({ token: data.token, user });
  return user;
}

export function register(input: {
  name: string;
  email: string;
  password: string;
  captcha: string;
}): Promise<null> {
  return apiRequest<null>("/auth/register", { method: "POST", json: input });
}

export function sendCaptcha(email: string): Promise<null> {
  return apiRequest<null>("/auth/sendCaptcha", {
    method: "POST",
    json: { email },
  });
}

export async function getCurrentUser(): Promise<User> {
  const data = await apiRequest<Omit<LoginResponse, "token">>("/users/me");
  return toUser(data);
}

