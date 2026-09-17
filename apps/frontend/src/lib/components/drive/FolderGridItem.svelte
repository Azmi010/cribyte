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
</script>

<div
  role="button"
  tabindex="0"
  class="border border-border/70 bg-card rounded-lg p-3.5 hover:border-primary/50 relative flex flex-col gap-2 w-full text-left outline-none cursor-default transition-colors"
  ondblclick={() => onOpen?.(folder.id)}
  oncontextmenu={(e) => onContextMenu?.(e, folder)}
  onkeydown={(e) => { if (e.key === 'Enter') onOpen?.(folder.id); }}
>
  {#if folder.starred}
    <div class="absolute top-3 right-10">
      <Star class="size-3.5 text-amber-500 fill-amber-500" />
    </div>
  {/if}
  
  <div class="flex items-start justify-between">
    <Folder class="size-6 text-primary fill-primary/15" />
    <button 
      type="button" 
      class="p-1 hover:bg-muted rounded-md text-muted-foreground transition-colors -mr-1 -mt-1 cursor-pointer"
      onclick={(e) => { e.stopPropagation(); onContextMenu?.(e, folder); }}
    >
      <MoreVertical class="size-4" />
    </button>
  </div>
  
  <div class="truncate text-xs font-medium w-full mt-1">
    {folder.name}
  </div>
  
  <div class="text-[10px] font-mono text-muted-foreground w-full">
    —
  </div>
</div>
