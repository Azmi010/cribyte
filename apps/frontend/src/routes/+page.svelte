<script lang="ts">
  import Toolbar from "$lib/components/layout/Toolbar.svelte";
  import CloudUploadIcon from "@lucide/svelte/icons/cloud-upload";
  import FolderIcon from "@lucide/svelte/icons/folder";
  import MoreVerticalIcon from "@lucide/svelte/icons/more-vertical";
  import StarIcon from "@lucide/svelte/icons/star";
  import PlayIcon from "@lucide/svelte/icons/play";
  import FileTextIcon from "@lucide/svelte/icons/file-text";
  import TableIcon from "@lucide/svelte/icons/table";
  import RefreshCwIcon from "@lucide/svelte/icons/refresh-cw";
  import XIcon from "@lucide/svelte/icons/x";
  import { uploadStore } from "$lib/stores/upload.svelte";

  const breadcrumbs = [
    { id: null, name: "Home" },
    { id: "proj", name: "Projects" },
    { id: "assets", name: "design-assets" },
  ];

  const folders = [
    { name: "Client Deliverables", items: 14, tag: "EXT4" },
    { name: "Docker Configs & Scripts", items: 8, tag: "MOUNTED" },
    { name: "Raw Footage & Media", items: 29, tag: "ZFS-POOL" },
    { name: "Personal Archives", items: 4, tag: "AES-256" },
  ];

  const files = [
    {
      name: "quarterly-report-2026....",
      size: "4.8 MB",
      date: "14 Sep 2026",
      type: "pdf",
      starred: true,
    },
    {
      name: "architecture_diagram_v...",
      size: "2.1 MB",
      date: "12 Sep 2026",
      type: "png",
      starred: true,
    },
    {
      name: "production-compose.yml",
      size: "14.2 KB",
      date: "08 Sep 2026",
      type: "yml",
      starred: false,
    },
    {
      name: "team_meeting_recordi...",
      size: "640.5 MB",
      date: "01 Sep 2026",
      type: "video",
      starred: false,
      duration: "42:15",
    },
    {
      name: "notes_brainstorm.md",
      size: "3.8 KB",
      date: "28 Aug 2026",
      type: "markdown",
      starred: false,
    },
    {
      name: "dataset_user_metrics.csv",
      size: "38.4 MB",
      date: "20 Aug 2026",
      type: "csv",
      starred: false,
    },
  ];

  let showMockUpload = $state(true);
</script>

<Toolbar {breadcrumbs} />

<div class="flex-1 p-4 md:p-6 space-y-6 overflow-y-auto">
  <!-- Hero Drag & Drop Banner -->
  <div
    class="relative rounded-lg border border-dashed border-border/80 bg-muted/15 p-5 flex flex-col sm:flex-row items-center justify-between gap-4 transition-colors hover:border-primary/50"
    style="background-image: radial-gradient(circle, var(--border) 1px, transparent 1px); background-size: 20px 20px;"
  >
    <div class="flex items-center gap-4">
      <div class="size-10 rounded-lg bg-muted/60 border border-border/70 flex items-center justify-center text-primary shrink-0">
        <CloudUploadIcon class="size-5" />
      </div>
      <div>
        <div class="text-xs font-semibold text-foreground">
          Drag & drop files here to upload directly to this folder
        </div>
        <div class="text-[11px] text-muted-foreground font-mono mt-0.5">
          Instant encryption · Direct stream ingestion · Max single file: 50 GB
        </div>
      </div>
    </div>
    <div class="text-[11px] font-mono text-muted-foreground bg-muted/50 border border-border/70 px-2 py-1 rounded shrink-0">
      Shift + U
    </div>
  </div>

  <!-- Folders Section -->
  <section class="space-y-3">
    <div class="flex items-center justify-between">
      <h2 class="text-xs font-mono font-semibold tracking-wider text-muted-foreground uppercase">
        FOLDERS ({folders.length})
      </h2>
      <span class="text-[11px] font-mono text-muted-foreground">Storage priority: High</span>
    </div>

    <div class="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-4 gap-3">
      {#each folders as folder (folder.name)}
        <div class="group p-3.5 rounded-lg border border-border/70 bg-card hover:border-primary/50 transition-all cursor-pointer">
          <div class="flex items-start justify-between mb-3">
            <FolderIcon class="size-6 text-primary fill-primary/15" />
            <button
              type="button"
              class="p-1 rounded text-muted-foreground hover:text-foreground hover:bg-muted/60 transition-colors"
              aria-label="Folder options"
            >
              <MoreVerticalIcon class="size-3.5" />
            </button>
          </div>
          <div class="text-xs font-medium text-foreground truncate mb-1.5" title={folder.name}>
            {folder.name}
          </div>
          <div class="flex items-center justify-between text-[11px] font-mono text-muted-foreground">
            <span>{folder.items} items</span>
            <span class="text-[10px] tracking-wider px-1.5 py-0.2 rounded bg-muted/60 border border-border/50 uppercase">
              {folder.tag}
            </span>
          </div>
        </div>
      {/each}
    </div>
  </section>

  <!-- Files Section -->
  <section class="space-y-3">
    <div class="flex items-center justify-between">
      <h2 class="text-xs font-mono font-semibold tracking-wider text-muted-foreground uppercase">
        FILES ({files.length})
      </h2>
      <span class="text-[11px] font-mono text-muted-foreground">685.8 MB total</span>
    </div>

    <div class="grid grid-cols-2 sm:grid-cols-3 lg:grid-cols-6 gap-3">
      {#each files as file (file.name)}
        <div class="group rounded-lg border border-border/70 bg-card overflow-hidden hover:border-primary/50 transition-all cursor-pointer flex flex-col">
          <!-- Preview Box -->
          <div class="relative h-28 bg-muted/30 border-b border-border/50 flex flex-col items-center justify-center p-2 select-none">
            {#if file.starred}
              <div class="absolute top-2 right-2 text-primary">
                <StarIcon class="size-3.5 fill-primary" />
              </div>
            {/if}

            {#if file.type === "pdf"}
              <div class="flex flex-col items-center gap-1 text-red-500">
                <span class="text-[10px] font-bold tracking-widest uppercase">PDF</span>
                <FileTextIcon class="size-8 stroke-[1.5]" />
              </div>
            {:else if file.type === "png"}
              <div class="size-full rounded bg-muted/50 border border-border/60 flex items-center justify-center relative overflow-hidden">
                <div class="absolute inset-0 bg-gradient-to-tr from-amber-500/10 to-transparent"></div>
                <span class="text-[10px] font-mono font-bold text-muted-foreground tracking-wider uppercase">PNG</span>
              </div>
            {:else if file.type === "yml"}
              <div class="w-full text-left font-mono text-[9px] text-muted-foreground/80 leading-tight space-y-0.5">
                <div class="text-primary font-bold">version: '3.9'</div>
                <div>services:</div>
                <div class="pl-2">app-core:</div>
                <div class="pl-3">restart: always</div>
              </div>
            {:else if file.type === "video"}
              <div class="size-full rounded bg-black/40 flex items-center justify-center relative">
                <div class="size-7 rounded-full bg-black/60 border border-white/20 flex items-center justify-center text-white">
                  <PlayIcon class="size-3 fill-white ml-0.5" />
                </div>
                {#if file.duration}
                  <span class="absolute bottom-1.5 right-1.5 text-[9px] font-mono bg-black/70 px-1 py-0.2 rounded text-white/90">
                    {file.duration}
                  </span>
                {/if}
              </div>
            {:else if file.type === "markdown"}
              <div class="flex flex-col items-center gap-1 text-muted-foreground">
                <FileTextIcon class="size-7 stroke-[1.5]" />
                <span class="text-[9px] font-mono tracking-wider font-semibold uppercase">MARKDOWN</span>
              </div>
            {:else if file.type === "csv"}
              <div class="flex flex-col items-center gap-1 text-primary">
                <TableIcon class="size-7 stroke-[1.5]" />
                <span class="text-[9px] font-mono tracking-wider font-semibold uppercase">CSV DATA</span>
              </div>
            {/if}
          </div>

          <!-- Metadata -->
          <div class="p-2.5 space-y-1">
            <div class="text-xs font-medium text-foreground truncate" title={file.name}>
              {file.name}
            </div>
            <div class="flex items-center justify-between text-[10px] font-mono text-muted-foreground">
              <span>{file.size}</span>
              <span>{file.date}</span>
            </div>
          </div>
        </div>
      {/each}
    </div>
  </section>
</div>

<!-- Floating Upload Progress Bar Widget -->
{#if showMockUpload || uploadStore.hasActive}
  <div class="fixed bottom-4 right-4 z-50 w-80 rounded-lg border border-border/90 bg-card shadow-xl p-3.5 space-y-2.5 animate-in fade-in slide-in-from-bottom-3 duration-200">
    <div class="flex items-center justify-between text-xs">
      <div class="flex items-center gap-2 font-medium text-foreground truncate">
        <RefreshCwIcon class="size-3.5 text-primary animate-spin" />
        <span class="truncate">Uploading backup-db-2026.sql.gz</span>
      </div>
      <button
        type="button"
        onclick={() => (showMockUpload = false)}
        class="text-muted-foreground hover:text-foreground cursor-pointer p-0.5"
        aria-label="Dismiss upload notification"
      >
        <XIcon class="size-3.5" />
      </button>
    </div>

    <!-- Golden Progress Bar -->
    <div class="h-1.5 w-full bg-muted rounded-full overflow-hidden">
      <div class="h-full bg-primary rounded-full transition-all duration-300" style="width: 78%"></div>
    </div>

    <div class="flex items-center justify-between text-[11px] font-mono text-muted-foreground">
      <span>78% of 142 MB</span>
      <span>12s remaining</span>
    </div>
  </div>
{/if}
