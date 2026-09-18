<script lang="ts">
	import FolderOpen from '@lucide/svelte/icons/folder-open';
	import SearchX from '@lucide/svelte/icons/search-x';
	import Star from '@lucide/svelte/icons/star';
	import Trash2 from '@lucide/svelte/icons/trash-2';

	let {
		variant,
		query = ''
	} = $props<{
		variant: 'folder' | 'search' | 'starred' | 'trash';
		query?: string;
	}>();
</script>

<div class="flex flex-col items-center justify-center py-20 text-center">
	<div class="mb-4 text-muted-foreground/30">
		{#if variant === 'folder'}
			<FolderOpen size={64} strokeWidth={1} />
		{:else if variant === 'search'}
			<SearchX size={64} strokeWidth={1} />
		{:else if variant === 'starred'}
			<Star size={64} strokeWidth={1} />
		{:else if variant === 'trash'}
			<Trash2 size={64} strokeWidth={1} />
		{/if}
	</div>

	<h3 class="mb-1 text-sm font-medium text-foreground">
		{#if variant === 'folder'}
			Belum ada file di sini
		{:else if variant === 'search'}
			Tidak ada hasil
		{:else if variant === 'starred'}
			Belum ada yang ditandai
		{:else if variant === 'trash'}
			Trash kosong
		{/if}
	</h3>

	<p class="text-xs text-muted-foreground font-mono">
		{#if variant === 'folder'}
			Seret file ke sini atau klik Upload
		{:else if variant === 'search'}
			Tidak ada hasil untuk "{query}"
		{:else if variant === 'starred'}
			Tandai file atau folder dengan bintang untuk akses cepat
		{:else if variant === 'trash'}
			File yang dihapus akan muncul di sini
		{/if}
	</p>
</div>
