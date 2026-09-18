<script lang="ts">
  import type { FileItem, PreviewResult } from "$lib/types";
  import * as filesApi from "$lib/api/files";
  import { formatFileSize } from "$lib/constants";
  import XIcon from "@lucide/svelte/icons/x";
  import DownloadIcon from "@lucide/svelte/icons/download";
  import ZoomInIcon from "@lucide/svelte/icons/zoom-in";
  import ZoomOutIcon from "@lucide/svelte/icons/zoom-out";
  import RotateCcwIcon from "@lucide/svelte/icons/rotate-ccw";
  import FileIcon from "@lucide/svelte/icons/file";
  import LoaderIcon from "@lucide/svelte/icons/loader";

  let {
    open = $bindable(false),
    item = null,
  }: {
    open: boolean;
    item: FileItem | null;
  } = $props();

  let preview = $state<PreviewResult | null>(null);
  let loading = $state(false);
  let error = $state(false);

  // Image zoom/pan state
  let zoom = $state(1);
  let panX = $state(0);
  let panY = $state(0);
  let isPanning = $state(false);
  let panStartX = $state(0);
  let panStartY = $state(0);

  $effect(() => {
    if (open && item) {
      loadPreview(item);
    } else {
      preview = null;
      error = false;
      resetTransform();
    }
  });

  async function loadPreview(file: FileItem) {
    loading = true;
    error = false;
    try {
      preview = await filesApi.previewFile(file.id);
    } catch {
      error = true;
    } finally {
      loading = false;
    }
  }

  function handleClose() {
    open = false;
  }

  function handleDownload() {
    if (!item) return;
    const url = filesApi.downloadUrl(item.id);
    const a = document.createElement("a");
    a.href = url;
    a.download = item.name;
    a.click();
  }

  function handleKeydown(e: KeyboardEvent) {
    if (e.key === "Escape") handleClose();
  }

  // --- Image Zoom & Pan ---
  function handleWheel(e: WheelEvent) {
    if (preview?.kind !== "image") return;
    e.preventDefault();
    const delta = e.deltaY > 0 ? -0.15 : 0.15;
    zoom = Math.min(5, Math.max(0.5, zoom + delta));
  }

  function handleMouseDown(e: MouseEvent) {
    if (preview?.kind !== "image" || zoom <= 1) return;
    isPanning = true;
    panStartX = e.clientX - panX;
    panStartY = e.clientY - panY;
  }

  function handleMouseMove(e: MouseEvent) {
    if (!isPanning) return;
    panX = e.clientX - panStartX;
    panY = e.clientY - panStartY;
  }

  function handleMouseUp() {
    isPanning = false;
  }

  function resetTransform() {
    zoom = 1;
    panX = 0;
    panY = 0;
    isPanning = false;
  }
</script>

<svelte:window onkeydown={handleKeydown} />

{#if open && item}
  <!-- svelte-ignore a11y_no_static_element_interactions -->
  <div
    class="fixed inset-0 z-50 flex flex-col bg-background/95 backdrop-blur-sm"
    onmousemove={handleMouseMove}
    onmouseup={handleMouseUp}
    onmouseleave={handleMouseUp}
  >
    <!-- Header -->
    <div class="flex items-center justify-between border-b border-border px-4 py-3 shrink-0">
      <div class="flex items-center gap-3 min-w-0">
        <h2 class="text-sm font-medium text-foreground truncate">{item.name}</h2>
        <span class="text-xs font-mono text-muted-foreground shrink-0">
          {formatFileSize(item.size)}
        </span>
      </div>
      <div class="flex items-center gap-1 shrink-0">
        {#if preview?.kind === "image" && zoom > 1}
          <button
            type="button"
            onclick={() => (zoom = Math.max(0.5, zoom - 0.25))}
            class="inline-flex items-center justify-center size-8 rounded-md text-muted-foreground hover:text-foreground hover:bg-muted transition-colors"
            aria-label="Zoom out"
          >
            <ZoomOutIcon class="size-4" />
          </button>
          <span class="text-xs font-mono text-muted-foreground min-w-[3rem] text-center">
            {Math.round(zoom * 100)}%
          </span>
          <button
            type="button"
            onclick={() => (zoom = Math.min(5, zoom + 0.25))}
            class="inline-flex items-center justify-center size-8 rounded-md text-muted-foreground hover:text-foreground hover:bg-muted transition-colors"
            aria-label="Zoom in"
          >
            <ZoomInIcon class="size-4" />
          </button>
          <button
            type="button"
            onclick={resetTransform}
            class="inline-flex items-center justify-center size-8 rounded-md text-muted-foreground hover:text-foreground hover:bg-muted transition-colors"
            aria-label="Reset zoom"
          >
            <RotateCcwIcon class="size-4" />
          </button>
          <div class="w-px h-4 bg-border mx-1"></div>
        {/if}
        <button
          type="button"
          onclick={handleDownload}
          class="inline-flex items-center justify-center size-8 rounded-md text-muted-foreground hover:text-foreground hover:bg-muted transition-colors"
          aria-label="Download"
        >
          <DownloadIcon class="size-4" />
        </button>
        <button
          type="button"
          onclick={handleClose}
          class="inline-flex items-center justify-center size-8 rounded-md text-muted-foreground hover:text-foreground hover:bg-muted transition-colors"
          aria-label="Close"
        >
          <XIcon class="size-4" />
        </button>
      </div>
    </div>

    <!-- Content -->
    <div class="flex-1 overflow-hidden flex items-center justify-center">
      {#if loading}
        <div class="flex flex-col items-center gap-3">
          <LoaderIcon class="size-6 text-primary animate-spin" />
          <span class="text-xs font-mono text-muted-foreground">Memuat preview...</span>
        </div>
      {:else if error}
        <div class="flex flex-col items-center gap-4 text-center px-4">
          <FileIcon class="size-12 text-muted-foreground" />
          <div>
            <p class="text-sm font-medium text-foreground">Gagal memuat preview</p>
            <p class="text-xs text-muted-foreground mt-1">File tidak dapat ditampilkan</p>
          </div>
          <button
            type="button"
            onclick={handleDownload}
            class="inline-flex items-center gap-2 px-4 py-2 rounded-lg bg-primary text-primary-foreground text-sm font-medium hover:opacity-90 transition-opacity"
          >
            <DownloadIcon class="size-4" />
            Download file
          </button>
        </div>
      {:else if preview}
        {#if preview.kind === "image" && preview.signed_url}
          <!-- svelte-ignore a11y_no_noninteractive_element_interactions -->
          <img
            src={preview.signed_url}
            alt={item.name}
            class="max-w-full max-h-full object-contain select-none"
            class:cursor-grab={zoom > 1}
            class:cursor-grabbing={isPanning}
            style="transform: scale({zoom}) translate({panX / zoom}px, {panY / zoom}px);"
            onwheel={handleWheel}
            onmousedown={handleMouseDown}
            draggable="false"
          />

        {:else if preview.kind === "video"}
          <video
            src={filesApi.serveUrl(item.id)}
            controls
            class="max-w-full max-h-full"
            preload="metadata"
          >
            <track kind="captions" />
          </video>

        {:else if preview.kind === "audio"}
          <div class="flex flex-col items-center gap-6 px-4">
            <div class="size-24 rounded-2xl bg-muted flex items-center justify-center">
              <svg class="size-12 text-muted-foreground" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.5">
                <path d="M9 18V5l12-2v13" />
                <circle cx="6" cy="18" r="3" />
                <circle cx="18" cy="16" r="3" />
              </svg>
            </div>
            <audio src={filesApi.serveUrl(item.id)} controls class="w-full max-w-md" preload="metadata" />
          </div>

        {:else if preview.kind === "pdf" && preview.signed_url}
          <iframe
            src={preview.signed_url}
            title={item.name}
            class="w-full h-full border-0"
          ></iframe>

        {:else if preview.kind === "text" && preview.content !== undefined}
          <div class="w-full h-full overflow-auto p-4">
            <pre class="text-sm font-mono text-foreground whitespace-pre-wrap break-words leading-relaxed">{preview.content}</pre>
          </div>

        {:else}
          <!-- Unsupported / other -->
          <div class="flex flex-col items-center gap-4 text-center px-4">
            <div class="size-20 rounded-2xl bg-muted flex items-center justify-center">
              <FileIcon class="size-10 text-muted-foreground" />
            </div>
            <div>
              <p class="text-sm font-medium text-foreground">{item.name}</p>
              <p class="text-xs text-muted-foreground mt-1">Tipe file ini belum mendukung preview</p>
            </div>
            <button
              type="button"
              onclick={handleDownload}
              class="inline-flex items-center gap-2 px-4 py-2 rounded-lg bg-primary text-primary-foreground text-sm font-medium hover:opacity-90 transition-opacity"
            >
              <DownloadIcon class="size-4" />
              Download file
            </button>
          </div>
        {/if}
      {/if}
    </div>
  </div>
{/if}
