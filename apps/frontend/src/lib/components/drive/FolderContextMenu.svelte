<script lang="ts">
  import { onMount } from "svelte";
  import type { FolderItem } from "$lib/types";
  import FolderOpen from "@lucide/svelte/icons/folder-open";
  import Pencil from "@lucide/svelte/icons/pencil";
  import FolderInput from "@lucide/svelte/icons/folder-input";
  import Star from "@lucide/svelte/icons/star";
  import StarOff from "@lucide/svelte/icons/star-off";
  import Trash2 from "@lucide/svelte/icons/trash-2";
  import RotateCcw from "@lucide/svelte/icons/rotate-ccw";

  let {
    open = $bindable(false),
    x = 0,
    y = 0,
    item = null,
    context = "drive",
    onOpen,
    onRename,
    onMove,
    onToggleStar,
    onTrash,
    onRestore,
    onPermanentDelete,
  }: {
    open: boolean;
    x: number;
    y: number;
    item: FolderItem | null;
    context?: "drive" | "starred" | "trash";
    onOpen?: (item: FolderItem) => void;
    onRename?: (item: FolderItem) => void;
    onMove?: (item: FolderItem) => void;
    onToggleStar?: (item: FolderItem) => void;
    onTrash?: (item: FolderItem) => void;
    onRestore?: (item: FolderItem) => void;
    onPermanentDelete?: (item: FolderItem) => void;
  } = $props();

  let menuEl: HTMLDivElement = $state(null!);
  let posX = $state<number | null>(null);
  let posY = $state<number | null>(null);

  $effect(() => {
    if (!open) {
      posX = null;
      posY = null;
      return;
    }
    if (!menuEl) return;
    // baca x/y supaya effect re-run saat posisi berubah
    void x;
    void y;
    const rect = menuEl.getBoundingClientRect();
    const margin = 8;
    let nx = x;
    let ny = y;
    if (nx + rect.width > window.innerWidth - margin) {
      nx = Math.max(margin, x - rect.width);
    }
    if (ny + rect.height > window.innerHeight - margin) {
      ny = Math.max(margin, y - rect.height);
    }
    nx = Math.max(margin, Math.min(nx, window.innerWidth - rect.width - margin));
    ny = Math.max(margin, Math.min(ny, window.innerHeight - rect.height - margin));
    posX = nx;
    posY = ny;
  });

  function handleClick(e: MouseEvent) {
    if (menuEl && !menuEl.contains(e.target as Node)) {
      open = false;
    }
  }

  function handleKeydown(e: KeyboardEvent) {
    if (e.key === "Escape") open = false;
  }

  function handleItem(action: () => void) {
    return (e: MouseEvent) => {
      e.stopPropagation();
      open = false;
      action();
    };
  }

  onMount(() => {
    document.addEventListener("click", handleClick, true);
    document.addEventListener("keydown", handleKeydown);
    return () => {
      document.removeEventListener("click", handleClick, true);
      document.removeEventListener("keydown", handleKeydown);
    };
  });
</script>

{#if open && item}
  <div
    bind:this={menuEl}
    class="fixed z-50 min-w-[180px] bg-popover text-popover-foreground ring-foreground/10 rounded-md p-1 shadow-md ring-1 animate-in fade-in zoom-in-95 duration-100"
    style="left: {posX ?? x}px; top: {posY ?? y}px; {posX === null ? 'visibility:hidden;' : ''}"
  >
    {#if context === "trash"}
      <button
        type="button"
        class="focus:bg-accent focus:text-accent-foreground gap-2 rounded-sm px-2 py-1.5 text-sm relative flex cursor-default items-center outline-hidden select-none w-full"
        onclick={handleItem(() => onRestore?.(item))}
      >
        <RotateCcw class="size-4" />
        <span>Restore</span>
      </button>
      <div class="bg-border -mx-1 my-1 h-px"></div>
      <button
        type="button"
        class="focus:bg-accent focus:text-accent-foreground data-[variant=destructive]:text-destructive focus:*:[svg]:text-accent-foreground gap-2 rounded-sm px-2 py-1.5 text-sm relative flex cursor-default items-center outline-hidden select-none w-full"
        data-variant="destructive"
        onclick={handleItem(() => onPermanentDelete?.(item))}
      >
        <Trash2 class="size-4 text-destructive" />
        <span>Hapus permanen</span>
      </button>
    {:else}
      <button
        type="button"
        class="focus:bg-accent focus:text-accent-foreground gap-2 rounded-sm px-2 py-1.5 text-sm relative flex cursor-default items-center outline-hidden select-none w-full"
        onclick={handleItem(() => onOpen?.(item))}
      >
        <FolderOpen class="size-4" />
        <span>Buka</span>
      </button>
      <div class="bg-border -mx-1 my-1 h-px"></div>
      <button
        type="button"
        class="focus:bg-accent focus:text-accent-foreground gap-2 rounded-sm px-2 py-1.5 text-sm relative flex cursor-default items-center outline-hidden select-none w-full"
        onclick={handleItem(() => onRename?.(item))}
      >
        <Pencil class="size-4" />
        <span>Rename</span>
      </button>
      <button
        type="button"
        class="focus:bg-accent focus:text-accent-foreground gap-2 rounded-sm px-2 py-1.5 text-sm relative flex cursor-default items-center outline-hidden select-none w-full"
        onclick={handleItem(() => onMove?.(item))}
      >
        <FolderInput class="size-4" />
        <span>Move</span>
      </button>
      <div class="bg-border -mx-1 my-1 h-px"></div>
      <button
        type="button"
        class="focus:bg-accent focus:text-accent-foreground gap-2 rounded-sm px-2 py-1.5 text-sm relative flex cursor-default items-center outline-hidden select-none w-full"
        onclick={handleItem(() => onToggleStar?.(item))}
      >
        {#if item.starred}
          <StarOff class="size-4" />
          <span>Unstar</span>
        {:else}
          <Star class="size-4" />
          <span>Star</span>
        {/if}
      </button>
      <div class="bg-border -mx-1 my-1 h-px"></div>
      <button
        type="button"
        class="focus:bg-accent focus:text-accent-foreground data-[variant=destructive]:text-destructive focus:*:[svg]:text-accent-foreground gap-2 rounded-sm px-2 py-1.5 text-sm relative flex cursor-default items-center outline-hidden select-none w-full"
        data-variant="destructive"
        onclick={handleItem(() => onTrash?.(item))}
      >
        <Trash2 class="size-4 text-destructive" />
        <span>Trash</span>
      </button>
    {/if}
  </div>
{/if}
