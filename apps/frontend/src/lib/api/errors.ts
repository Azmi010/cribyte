import { toast } from "svelte-sonner";
import { ApiRequestError } from "./client";

/**
 * Map error API ke toast yang actionable + konsisten.
 * 401 sudah ditangani otomatis di client (redirect ke login),
 * jadi di sini fokus ke 404/409/413/500 dan network error.
 */
export function handleApiError(err: unknown, fallback = "Terjadi kesalahan"): void {
  if (err instanceof ApiRequestError) {
    switch (err.status) {
      case 401:
        // sudah di-handle client (redirect). jangan spam toast.
        return;
      case 403:
        toast.error("Kamu tidak punya akses ke item ini");
        return;
      case 404:
        toast.error("File atau folder tidak ditemukan");
        return;
      case 409:
        toast.error("Nama sudah digunakan, pilih nama lain");
        return;
      case 413:
        toast.error("Ukuran file melebihi batas");
        return;
      default:
        if (err.status >= 500) {
          toast.error("Server sedang bermasalah, coba lagi nanti");
          return;
        }
        toast.error(err.message || fallback);
        return;
    }
  }

  if (err instanceof TypeError) {
    // fetch gagal total = network error
    toast.error("Gagal terhubung ke server, cek koneksi kamu");
    return;
  }

  toast.error(err instanceof Error ? err.message : fallback);
}
