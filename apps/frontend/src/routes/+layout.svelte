<script lang="ts">
  import "../app.css";
  import favicon from "$lib/assets/favicon.svg";
  import { onMount } from "svelte";
  import { page } from "$app/state";
  import { ModeWatcher } from "mode-watcher";
  import { Toaster } from "$lib/components/ui/sonner";
  import { Sidebar, ContentArea } from "$lib/components/layout";
  import { auth } from "$lib/stores/auth.svelte";

  let { children } = $props();

  let mobileSidebarOpen = $state(false);

  const isAuthRoute = $derived(
    page.url.pathname.startsWith("/login") || page.url.pathname.startsWith("/register")
  );

  onMount(() => {
    auth.fetchUser();
  });
</script>

<svelte:head>
  <link rel="icon" href={favicon} />
  <title>CriByte - Self-Hosted Cloud Storage</title>
</svelte:head>

<ModeWatcher defaultMode="dark" />
<Toaster position="bottom-right" />

{#if isAuthRoute}
  <main class="min-h-screen bg-background text-foreground">
    {@render children()}
  </main>
{:else}
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
