<script lang="ts">
  import Folder from '@lucide/svelte/icons/folder';
  import MoreVertical from '@lucide/svelte/icons/more-vertical';
  import Star from '@lucide/svelte/icons/star';
  import type { FolderItem } from '$lib/types';
  
  let { 
    folder, 
    onOpen, 
    onContextMenu 
  }: {
    folder: FolderItem;
    onOpen?: (id: string) => void;
    onContextMenu?: (e: MouseEvent, folder: FolderItem) => void;
  } = $props();

  function formatDate(iso: string): string {
    const d = new Date(iso);
    return d.toLocaleDateString('id-ID', { day: '2-digit', month: 'short', year: 'numeric' });
  }
</script>

<div
  role="button"
  tabindex="0"
  class="flex items-center py-2 px-3 border-b border-border/40 hover:bg-muted/30 transition-colors w-full cursor-default outline-none gap-3"
  ondblclick={() => onOpen?.(folder.id)}
  oncontextmenu={(e) => onContextMenu?.(e, folder)}
  onkeydown={(e) => { if (e.key === 'Enter') onOpen?.(folder.id); }}
>
  <Folder class="size-5 text-primary fill-primary/15 shrink-0" />
  
  <div class="text-sm flex-1 truncate font-medium text-left">
    {folder.name}
  </div>
  
  {#if folder.starred}
    <Star class="size-3.5 text-amber-500 fill-amber-500 shrink-0" />
  {/if}
  
  <div class="text-[11px] font-mono text-muted-foreground shrink-0 w-24 text-right">
    {formatDate(folder.updated_at)}
  </div>
  
  <button 
    type="button" 
    class="p-1 hover:bg-muted rounded-md text-muted-foreground transition-colors shrink-0 ml-2 cursor-pointer"
    onclick={(e) => { e.stopPropagation(); onContextMenu?.(e, folder); }}
  >
    <MoreVertical class="size-4" />
  </button>
</div>
