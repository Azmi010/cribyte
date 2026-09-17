<script lang="ts">
	import type { FileItem, FolderItem } from '$lib/types';
	import { formatFileSize } from '$lib/constants';
	import FolderGridItem from './FolderGridItem.svelte';
	import FileGridItem from './FileGridItem.svelte';

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

	let totalFileSize = $derived(
		files.reduce((acc: number, file: FileItem) => acc + (file.size || 0), 0)
	);
</script>

<div class="space-y-6">
	{#if folders.length > 0}
		<section>
			<div class="mb-3 text-xs font-mono font-semibold tracking-wider text-muted-foreground uppercase">
				FOLDERS ({folders.length})
			</div>
			<div class="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-4 gap-3">
				{#each folders as folder (folder.id)}
					<FolderGridItem 
						{folder} 
						onOpen={() => onOpenFolder?.(folder.id)}
						onContextMenu={(e: MouseEvent) => onFolderContextMenu?.(e, folder)}
					/>
				{/each}
			</div>
		</section>
	{/if}

	{#if files.length > 0}
		<section>
			<div class="flex items-center justify-between mb-3">
				<div class="text-xs font-mono font-semibold tracking-wider text-muted-foreground uppercase">
					FILES ({files.length})
				</div>
				<div class="text-xs font-mono tracking-wider text-muted-foreground">
					{formatFileSize(totalFileSize)}
				</div>
			</div>
			<div class="grid grid-cols-2 sm:grid-cols-3 lg:grid-cols-6 gap-3">
				{#each files as file (file.id)}
					<FileGridItem 
						{file} 
						onOpen={() => onOpenFile?.(file.id)}
						onContextMenu={(e: MouseEvent) => onFileContextMenu?.(e, file)}
					/>
				{/each}
			</div>
		</section>
	{/if}
</div>
