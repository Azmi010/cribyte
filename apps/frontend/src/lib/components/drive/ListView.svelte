<script lang="ts">
	import type { FileItem, FolderItem } from '$lib/types';
	import FolderListItem from './FolderListItem.svelte';
	import FileListItem from './FileListItem.svelte';

	let {
		folders = [],
		files = [],
		selectedIds = new Set<string>(),
		onOpenFolder,
		onOpenFile,
		onSelect,
		onFolderContextMenu,
		onFileContextMenu
	} = $props<{
		folders?: FolderItem[];
		files?: FileItem[];
		selectedIds?: Set<string> | import('svelte/reactivity').SvelteSet<string>;
		onOpenFolder?: (id: string) => void;
		onOpenFile?: (id: string) => void;
		onSelect?: (id: string, e: MouseEvent) => void;
		onFolderContextMenu?: (e: MouseEvent, folder: FolderItem) => void;
		onFileContextMenu?: (e: MouseEvent, file: FileItem) => void;
	}>();
</script>

<div class="select-none border rounded-md bg-card overflow-hidden">
	<div class="sticky top-0 z-10 flex items-center px-4 py-2 border-b bg-muted/30 text-[11px] font-mono uppercase tracking-wider text-muted-foreground">
		<div class="flex-1">Name</div>
		<div class="w-8"></div>
		<div class="w-24 text-right">Size</div>
		<div class="w-32 text-right">Modified</div>
		<div class="w-8"></div>
	</div>
	
	<div class="flex flex-col">
		{#each folders as folder (folder.id)}
			<FolderListItem 
				{folder} 
				selected={selectedIds.has(folder.id)}
				onOpen={() => onOpenFolder?.(folder.id)}
				onSelect={(id: string, e: MouseEvent) => onSelect?.(id, e)}
				onContextMenu={(e: MouseEvent) => onFolderContextMenu?.(e, folder)}
			/>
		{/each}
		
		{#each files as file (file.id)}
			<FileListItem 
				{file} 
				selected={selectedIds.has(file.id)}
				onOpen={() => onOpenFile?.(file.id)}
				onSelect={(id: string, e: MouseEvent) => onSelect?.(id, e)}
				onContextMenu={(e: MouseEvent) => onFileContextMenu?.(e, file)}
			/>
		{/each}
	</div>
</div>
