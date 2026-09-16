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
    throw new ApiRequestError(response.status, message);
  }

  const contentType = response.headers.get("Content-Type");
  if (response.status === 204 || !contentType?.includes("application/json")) {
    return undefined as T;
  }

  return response.json() as Promise<T>;
}
