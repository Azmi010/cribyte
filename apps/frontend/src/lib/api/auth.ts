import type { User } from "$lib/types";
import { api } from "./client";

interface RegisterPayload {
  email: string;
  name: string;
  password: string;
}

interface LoginPayload {
  email: string;
  password: string;
}

export function register(payload: RegisterPayload): Promise<User> {
  return api<User>("/auth/register", {
    method: "POST",
    body: payload,
  });
}

export function login(payload: LoginPayload): Promise<User> {
  return api<User>("/auth/login", {
    method: "POST",
    body: payload,
  });
}

export function logout(): Promise<void> {
  return api<void>("/auth/logout", {
    method: "POST",
  });
}

export function getMe(): Promise<User> {
  return api<User>("/auth/me");
}
