<script lang="ts">
  import MoreVertical from '@lucide/svelte/icons/more-vertical';
  import Star from '@lucide/svelte/icons/star';
  import FileText from '@lucide/svelte/icons/file-text';
  import Image from '@lucide/svelte/icons/image';
  import Film from '@lucide/svelte/icons/film';
  import Music from '@lucide/svelte/icons/music';
  import FileCode from '@lucide/svelte/icons/file-code';
  import FileIcon from '@lucide/svelte/icons/file';
  
  import type { FileItem } from '$lib/types';
  import { formatFileSize, getPreviewKind } from '$lib/constants';
  
  let { 
    file, 
    onOpen, 
    onContextMenu 
  }: {
    file: FileItem;
    onOpen?: (id: string) => void;
    onContextMenu?: (e: MouseEvent, file: FileItem) => void;
  } = $props();

  function formatDate(iso: string): string {
    const d = new Date(iso);
    return d.toLocaleDateString('id-ID', { day: '2-digit', month: 'short', year: 'numeric' });
  }

  let previewKind = $derived(getPreviewKind(file.mime_type));
</script>

<div
  role="button"
  tabindex="0"
  class="flex items-center py-2 px-3 border-b border-border/40 hover:bg-muted/30 transition-colors w-full cursor-default outline-none gap-3"
  ondblclick={() => onOpen?.(file.id)}
  oncontextmenu={(e) => onContextMenu?.(e, file)}
  onkeydown={(e) => { if (e.key === 'Enter') onOpen?.(file.id); }}
>
  <div class="shrink-0 flex items-center justify-center w-5">
    {#if previewKind === 'pdf'}
      <FileText class="size-5 text-red-500" />
    {:else if previewKind === 'image'}
      <Image class="size-5 text-blue-500" />
    {:else if previewKind === 'video'}
      <Film class="size-5 text-purple-500" />
    {:else if previewKind === 'audio'}
      <Music class="size-5 text-amber-500" />
    {:else if previewKind === 'text'}
      <FileCode class="size-5 text-muted-foreground" />
    {:else}
      <FileIcon class="size-5 text-muted-foreground" />
    {/if}
  </div>
  
  <div class="text-sm flex-1 truncate font-medium text-left">
    {file.name}
  </div>
  
  {#if file.starred}
    <Star class="size-3.5 text-amber-500 fill-amber-500 shrink-0" />
  {/if}
  
  <div class="text-[11px] font-mono text-muted-foreground shrink-0 w-20 text-right">
    {formatFileSize(file.size)}
  </div>
  
  <div class="text-[11px] font-mono text-muted-foreground shrink-0 w-24 text-right">
    {formatDate(file.updated_at)}
  </div>
  
  <button 
    type="button" 
    class="p-1 hover:bg-muted rounded-md text-muted-foreground transition-colors shrink-0 ml-2 cursor-pointer"
    onclick={(e) => { e.stopPropagation(); onContextMenu?.(e, file); }}
  >
    <MoreVertical class="size-4" />
  </button>
</div>
