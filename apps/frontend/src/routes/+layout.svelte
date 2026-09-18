<script lang="ts">
  import "../app.css";
  import favicon from "$lib/assets/favicon.svg";
  import { onMount } from "svelte";
  import { goto } from "$app/navigation";
  import { ModeWatcher } from "mode-watcher";
  import { toast } from "svelte-sonner";
  import { Toaster } from "$lib/components/ui/sonner";
  import { auth } from "$lib/stores/auth.svelte";
  import { setSessionExpiredHandler } from "$lib/api/client";

  let { children } = $props();

  let redirecting = false;

  onMount(() => {
    setSessionExpiredHandler(() => {
      if (redirecting) return;
      redirecting = true;
      auth.clear();
      toast.error("Sesi berakhir, silakan login lagi");
      goto("/login").finally(() => {
        redirecting = false;
      });
    });
    auth.fetchUser();
  });
</script>

<svelte:head>
  <link rel="icon" href={favicon} />
  <title>CriByte - Self-Hosted Cloud Storage</title>
</svelte:head>

<ModeWatcher defaultMode="system" />
<Toaster position="bottom-right" />

{@render children()}
