import { API_BASE } from "$lib/constants";
import type { ApiError } from "$lib/types";

export class ApiRequestError extends Error {
  constructor(
    public status: number,
    message: string,
  ) {
    super(message);
    this.name = "ApiRequestError";
  }
}

type FetchOptions = Omit<RequestInit, "body"> & {
  body?: object | FormData;
};

// Callback yang dipanggil saat session expired (401). Di-set oleh app shell
// supaya modul client tetap bebas dependency ke store/router.
let onSessionExpired: (() => void) | undefined;

export function setSessionExpiredHandler(handler: () => void): void {
  onSessionExpired = handler;
}

export async function api<T>(path: string, options: FetchOptions = {}): Promise<T> {
  const { body, headers: customHeaders, ...rest } = options;

  const headers = new Headers(customHeaders);

  let processedBody: BodyInit | undefined;

  if (body instanceof FormData) {
    processedBody = body;
  } else if (body !== undefined) {
    headers.set("Content-Type", "application/json");
    processedBody = JSON.stringify(body);
  }

  const response = await fetch(`${API_BASE}${path}`, {
    ...rest,
    headers,
    body: processedBody,
    credentials: "include",
  });

  if (!response.ok) {
    let message = `Request failed with status ${response.status}`;
    try {
      const err: ApiError = await response.json();
      if (err.error) message = err.error;
    } catch {
      // response body bukan JSON
    }

    // Session expired: redirect ke login kecuali request memang endpoint auth
    // (login/register/me) supaya form auth tetap bisa nampilin error sendiri.
    if (response.status === 401 && typeof window !== "undefined") {
      const isAuthEndpoint = path.startsWith("/auth/");
      const onLoginPage = window.location.pathname.startsWith("/login");
      if (!isAuthEndpoint && !onLoginPage) {
        onSessionExpired?.();
      }
    }

    throw new ApiRequestError(response.status, message);
  }

  const contentType = response.headers.get("Content-Type");
  if (response.status === 204 || !contentType?.includes("application/json")) {
    return undefined as T;
  }

  return response.json() as Promise<T>;
}
