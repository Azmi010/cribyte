<script lang="ts">
  import { goto } from "$app/navigation";
  import { auth } from "$lib/stores/auth.svelte";
  import { Sidebar, ContentArea } from "$lib/components/layout";

  let { children } = $props();

  let mobileSidebarOpen = $state(false);

  // Redirect to login if not authenticated (after loading completes)
  $effect(() => {
    if (!auth.loading && !auth.authenticated) {
      goto("/login");
    }
  });
</script>

{#if auth.loading}
  <!-- Loading spinner while checking auth -->
  <div class="flex h-screen items-center justify-center bg-background">
    <div class="flex flex-col items-center gap-3">
      <div class="size-8 border-2 border-primary border-t-transparent rounded-full animate-spin"></div>
      <span class="text-xs font-mono text-muted-foreground">Memuat...</span>
    </div>
  </div>
{:else if auth.authenticated}
  <div class="flex h-screen w-full overflow-hidden bg-background text-foreground">
    <div class="hidden md:flex h-full">
      <Sidebar />
    </div>

    {#if mobileSidebarOpen}
      <div
        class="fixed inset-0 z-40 bg-black/60 backdrop-blur-xs md:hidden"
        onclick={() => (mobileSidebarOpen = false)}
        onkeydown={(e) => e.key === "Escape" && (mobileSidebarOpen = false)}
        role="button"
        tabindex="0"
        aria-label="Close sidebar overlay"
      ></div>
      <div class="fixed inset-y-0 left-0 z-50 md:hidden animate-in slide-in-from-left duration-200">
        <Sidebar isMobile onClose={() => (mobileSidebarOpen = false)} />
      </div>
    {/if}

    <ContentArea onMenuClick={() => (mobileSidebarOpen = true)}>
      {@render children()}
    </ContentArea>
  </div>
{/if}

