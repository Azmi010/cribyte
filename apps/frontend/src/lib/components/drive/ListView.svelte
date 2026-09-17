<script lang="ts">
	import type { FileItem, FolderItem } from '$lib/types';
	import FolderListItem from './FolderListItem.svelte';
	import FileListItem from './FileListItem.svelte';

	let {
		folders = [],
		files = [],
		onOpenFolder,
		onOpenFile,
		onFolderContextMenu,
		onFileContextMenu
	} = $props<{
		folders?: FolderItem[];
		files?: FileItem[];
		onOpenFolder?: (id: string) => void;
		onOpenFile?: (id: string) => void;
		onFolderContextMenu?: (e: MouseEvent, folder: FolderItem) => void;
		onFileContextMenu?: (e: MouseEvent, file: FileItem) => void;
	}>();
</script>

<div class="border rounded-md bg-card overflow-hidden">
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
				onOpen={() => onOpenFolder?.(folder.id)}
				onContextMenu={(e: MouseEvent) => onFolderContextMenu?.(e, folder)}
			/>
		{/each}
		
		{#each files as file (file.id)}
			<FileListItem 
				{file} 
				onOpen={() => onOpenFile?.(file.id)}
				onContextMenu={(e: MouseEvent) => onFileContextMenu?.(e, file)}
			/>
		{/each}
	</div>
</div>
