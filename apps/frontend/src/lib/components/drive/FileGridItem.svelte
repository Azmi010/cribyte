<script lang="ts">
  import MoreVertical from '@lucide/svelte/icons/more-vertical';
  import Star from '@lucide/svelte/icons/star';
  import FileText from '@lucide/svelte/icons/file-text';
  import Image from '@lucide/svelte/icons/image';
  import PlayCircle from '@lucide/svelte/icons/play-circle';
  import FileIcon from '@lucide/svelte/icons/file';
  
  import type { FileItem } from '$lib/types';
  import { formatFileSize, getPreviewKind } from '$lib/constants';
  import * as filesApi from '$lib/api/files';
  
  let {
    file,
    selected = false,
    onOpen,
    onSelect,
    onContextMenu
  }: {
    file: FileItem;
    selected?: boolean;
    onOpen?: (id: string) => void;
    onSelect?: (id: string, e: MouseEvent) => void;
    onContextMenu?: (e: MouseEvent, file: FileItem) => void;
  } = $props();

  function formatDate(iso: string): string {
    const d = new Date(iso);
    return d.toLocaleDateString('id-ID', { day: '2-digit', month: 'short', year: 'numeric' });
  }

  let previewKind = $derived(getPreviewKind(file.mime_type));
  let extLabel = $derived(file.extension ? file.extension.toUpperCase() : 'FILE');

  let hasThumbnail = $derived(file.has_thumbnail);
  let thumbError = $state(false);

  $effect(() => {
    void file.id;
    thumbError = false;
  });
</script>

<div
  role="button"
  tabindex="0"
  aria-pressed={selected}
  data-select-id={file.id}
  data-select-type="file"
  class="select-none border bg-card rounded-lg flex flex-col w-full text-left overflow-hidden outline-none cursor-default transition-colors {selected ? 'border-primary ring-1 ring-primary/40' : 'border-border/70 hover:border-primary/50'}"
  onclick={(e) => onSelect?.(file.id, e)}
  ondblclick={() => onOpen?.(file.id)}
  oncontextmenu={(e) => { onSelect?.(file.id, e); onContextMenu?.(e, file); }}
  onkeydown={(e) => { if (e.key === 'Enter') onOpen?.(file.id); }}
>
  <div class="h-28 bg-muted/30 border-b relative flex items-center justify-center">
    {#if file.starred}
      <div class="absolute top-2 right-2 z-10">
        <Star class="size-4 text-amber-500 fill-amber-500" />
      </div>
    {/if}

    {#if hasThumbnail && !thumbError}
      <img
        src={filesApi.thumbnailUrl(file.id)}
        alt={file.name}
        loading="lazy"
        class="w-full h-full object-cover"
        onerror={() => (thumbError = true)}
        draggable="false"
      />
      {#if previewKind === 'video'}
        <div class="absolute inset-0 flex items-center justify-center pointer-events-none">
          <PlayCircle class="size-10 text-white drop-shadow-lg opacity-90" />
        </div>
      {/if}
    {:else if previewKind === 'pdf'}
      <div class="flex flex-col items-center gap-2">
        <FileText class="size-8 text-red-500" />
        <span class="text-[10px] tracking-wider px-1.5 py-0.5 rounded bg-muted/60 border border-border/50 uppercase font-mono font-semibold text-red-500">PDF</span>
      </div>
    {:else if previewKind === 'image'}
      <div class="flex flex-col items-center gap-2">
        <Image class="size-8 text-blue-500" />
        <span class="text-[10px] tracking-wider px-1.5 py-0.5 rounded bg-muted/60 border border-border/50 uppercase font-mono font-semibold text-blue-500">{extLabel}</span>
      </div>
    {:else if previewKind === 'video'}
      <div class="flex flex-col items-center gap-2">
        <PlayCircle class="size-10 text-primary opacity-80" />
        <span class="text-[10px] tracking-wider px-1.5 py-0.5 rounded bg-muted/60 border border-border/50 uppercase font-mono text-muted-foreground">—</span>
      </div>
    {:else if previewKind === 'text'}
      <div class="flex flex-col items-center gap-2">
        <FileText class="size-8 text-muted-foreground" />
        <span class="text-[10px] tracking-wider px-1.5 py-0.5 rounded bg-muted/60 border border-border/50 uppercase font-mono font-semibold text-muted-foreground">{extLabel}</span>
      </div>
    {:else}
      <div class="flex flex-col items-center gap-2">
        <FileIcon class="size-8 text-muted-foreground" />
        <span class="text-[10px] tracking-wider px-1.5 py-0.5 rounded bg-muted/60 border border-border/50 uppercase font-mono font-semibold text-muted-foreground">{extLabel}</span>
      </div>
    {/if}
  </div>
  
  <div class="p-2.5 flex flex-col gap-1 relative">
    <div class="flex items-start justify-between">
      <div class="truncate text-xs font-medium w-full pr-4" title={file.name}>
        {file.name}
      </div>
      <button 
        type="button" 
        class="absolute right-1 top-2 p-1 hover:bg-muted rounded-md text-muted-foreground transition-colors cursor-pointer"
        onclick={(e) => { e.stopPropagation(); onContextMenu?.(e, file); }}
      >
        <MoreVertical class="size-3.5" />
      </button>
    </div>
    
    <div class="text-[10px] font-mono text-muted-foreground w-full flex items-center justify-between mt-1 pr-4">
      <span>{formatFileSize(file.size)}</span>
      <span>{formatDate(file.updated_at)}</span>
    </div>
  </div>
</div>
