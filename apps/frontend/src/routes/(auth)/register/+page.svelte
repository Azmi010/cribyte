<script lang="ts">
  import { goto } from "$app/navigation";
  import { auth } from "$lib/stores/auth.svelte";
  import { toast } from "svelte-sonner";
  import { ApiRequestError } from "$lib/api/client";
  import { Button } from "$lib/components/ui/button";
  import { Input } from "$lib/components/ui/input";
  import Logo from "$lib/components/Logo.svelte";
  import LockIcon from "@lucide/svelte/icons/lock";
  import MailIcon from "@lucide/svelte/icons/mail";
  import UserIcon from "@lucide/svelte/icons/user";
  import EyeIcon from "@lucide/svelte/icons/eye";
  import EyeOffIcon from "@lucide/svelte/icons/eye-off";
  import ArrowRightIcon from "@lucide/svelte/icons/arrow-right";
  import LoaderIcon from "@lucide/svelte/icons/loader";

  let email = $state("");
  let name = $state("");
  let password = $state("");
  let confirmPassword = $state("");
  let showPassword = $state(false);
  let showConfirmPassword = $state(false);
  let loading = $state(false);
  let errors = $state({ email: "", name: "", password: "", confirmPassword: "" });

  function validate(): boolean {
    errors = { email: "", name: "", password: "", confirmPassword: "" };
    let valid = true;

    if (!email.trim()) {
      errors.email = "Email wajib diisi";
      valid = false;
    } else if (!/^[^\s@]+@[^\s@]+\.[^\s@]+$/.test(email.trim())) {
      errors.email = "Format email tidak valid";
      valid = false;
    }

    if (!name.trim()) {
      errors.name = "Nama wajib diisi";
      valid = false;
    } else if (name.trim().length < 2) {
      errors.name = "Nama minimal 2 karakter";
      valid = false;
    }

    if (!password) {
      errors.password = "Password wajib diisi";
      valid = false;
    } else if (password.length < 8) {
      errors.password = "Password minimal 8 karakter";
      valid = false;
    }

    if (!confirmPassword) {
      errors.confirmPassword = "Konfirmasi password wajib diisi";
      valid = false;
    } else if (password !== confirmPassword) {
      errors.confirmPassword = "Password tidak cocok";
      valid = false;
    }

    return valid;
  }

  async function handleSubmit(e: Event) {
    e.preventDefault();
    if (!validate()) return;

    loading = true;
    try {
      await auth.register(email.trim().toLowerCase(), name.trim(), password);
      toast.success("Akun berhasil dibuat");
      goto("/");
    } catch (err) {
      if (err instanceof ApiRequestError) {
        if (err.status === 409) {
          toast.error("Email sudah terdaftar");
        } else {
          toast.error(err.message);
        }
      } else {
        toast.error("Gagal terhubung ke server");
      }
    } finally {
      loading = false;
    }
  }
</script>

<svelte:head>
  <title>Daftar — CriByte</title>
</svelte:head>

<!-- Top Header Bar -->
<header class="h-14 flex items-center justify-between px-4 md:px-6 border-b border-border/60 bg-background/80 backdrop-blur-sm">
  <div class="flex items-center gap-2.5">
    <Logo class="size-6" />
    <span class="text-sm font-semibold text-foreground tracking-tight">CriByte</span>
    <span class="text-[10px] font-mono text-primary/90 border border-primary/30 rounded px-1.5 py-0.5 bg-primary/10 select-none">
      v0.4.2
    </span>
  </div>
  <div class="flex items-center gap-3 text-xs font-mono text-muted-foreground">
    <div class="flex items-center gap-1.5 px-2 py-0.5 rounded border border-emerald-500/30 bg-emerald-500/10 text-emerald-400">
      <span class="size-1.5 rounded-full bg-emerald-400 animate-pulse"></span>
      <span>daemon: healthy</span>
    </div>
  </div>
</header>

<!-- Main Content -->
<main class="flex items-center justify-center px-4 py-12 md:py-20 min-h-[calc(100vh-3rem)]">
  <div class="w-full max-w-md">
    <!-- Card -->
    <div class="rounded-xl border border-border/80 bg-card p-6 md:p-8 space-y-6 shadow-sm">
      <!-- Badge -->
      <div class="inline-flex items-center gap-1.5 rounded-md border border-primary/30 bg-primary/10 px-2.5 py-1 text-[11px] font-mono font-medium text-primary uppercase tracking-wider select-none">
        <LockIcon class="size-3" />
        <span>create_account</span>
      </div>

      <!-- Title -->
      <div class="space-y-1.5">
        <h1 class="text-xl md:text-2xl font-semibold text-foreground tracking-tight">
          Buat akun CriByte
        </h1>
        <p class="text-sm text-muted-foreground leading-relaxed">
          Daftarkan akun untuk mengakses instance penyimpanan privat ini.
        </p>
      </div>

      <!-- Form -->
      <form onsubmit={handleSubmit} class="space-y-4">
        <!-- Email Field -->
        <div class="space-y-1.5">
          <label for="email" class="text-sm font-medium text-foreground">Email</label>
          <div class="relative">
            <Input
              id="email"
              type="email"
              placeholder="admin@local"
              bind:value={email}
              class="h-11 pr-10 font-mono text-sm bg-muted/20 border-border/70 focus-visible:border-primary focus-visible:ring-primary/30"
              aria-invalid={!!errors.email}
              disabled={loading}
            />
            <div class="absolute right-3 top-1/2 -translate-y-1/2 text-muted-foreground">
              <MailIcon class="size-4" />
            </div>
          </div>
          {#if errors.email}
            <p class="text-xs text-destructive mt-1">{errors.email}</p>
          {/if}
        </div>

        <!-- Name Field -->
        <div class="space-y-1.5">
          <label for="name" class="text-sm font-medium text-foreground">Nama</label>
          <div class="relative">
            <Input
              id="name"
              type="text"
              placeholder="Nama lengkap"
              bind:value={name}
              class="h-11 pr-10 text-sm bg-muted/20 border-border/70 focus-visible:border-primary focus-visible:ring-primary/30"
              aria-invalid={!!errors.name}
              disabled={loading}
            />
            <div class="absolute right-3 top-1/2 -translate-y-1/2 text-muted-foreground">
              <UserIcon class="size-4" />
            </div>
          </div>
          {#if errors.name}
            <p class="text-xs text-destructive mt-1">{errors.name}</p>
          {/if}
        </div>

        <!-- Password Field -->
        <div class="space-y-1.5">
          <label for="password" class="text-sm font-medium text-foreground">Password</label>
          <div class="relative">
            <Input
              id="password"
              type={showPassword ? "text" : "password"}
              placeholder="Minimal 8 karakter"
              bind:value={password}
              class="h-11 pr-10 font-mono text-sm bg-muted/20 border-border/70 focus-visible:border-primary focus-visible:ring-primary/30"
              aria-invalid={!!errors.password}
              disabled={loading}
            />
            <button
              type="button"
              class="absolute right-3 top-1/2 -translate-y-1/2 text-muted-foreground hover:text-foreground transition-colors cursor-pointer"
              onclick={() => (showPassword = !showPassword)}
              tabindex={-1}
              aria-label={showPassword ? "Sembunyikan password" : "Tampilkan password"}
            >
              {#if showPassword}
                <EyeOffIcon class="size-4" />
              {:else}
                <EyeIcon class="size-4" />
              {/if}
            </button>
          </div>
          {#if errors.password}
            <p class="text-xs text-destructive mt-1">{errors.password}</p>
          {/if}
        </div>

        <!-- Confirm Password Field -->
        <div class="space-y-1.5">
          <label for="confirm-password" class="text-sm font-medium text-foreground">Konfirmasi password</label>
          <div class="relative">
            <Input
              id="confirm-password"
              type={showConfirmPassword ? "text" : "password"}
              placeholder="Ulangi password"
              bind:value={confirmPassword}
              class="h-11 pr-10 font-mono text-sm bg-muted/20 border-border/70 focus-visible:border-primary focus-visible:ring-primary/30"
              aria-invalid={!!errors.confirmPassword}
              disabled={loading}
            />
            <button
              type="button"
              class="absolute right-3 top-1/2 -translate-y-1/2 text-muted-foreground hover:text-foreground transition-colors cursor-pointer"
              onclick={() => (showConfirmPassword = !showConfirmPassword)}
              tabindex={-1}
              aria-label={showConfirmPassword ? "Sembunyikan password" : "Tampilkan password"}
            >
              {#if showConfirmPassword}
                <EyeOffIcon class="size-4" />
              {:else}
                <EyeIcon class="size-4" />
              {/if}
            </button>
          </div>
          {#if errors.confirmPassword}
            <p class="text-xs text-destructive mt-1">{errors.confirmPassword}</p>
          {/if}
        </div>

        <!-- Submit Button -->
        <div class="pt-1">
          <Button
            type="submit"
            disabled={loading}
            class="w-full h-11 text-sm font-semibold gap-2"
          >
            {#if loading}
              <LoaderIcon class="size-4 animate-spin" />
              <span>Memproses...</span>
            {:else}
              <span>Buat Akun</span>
              <ArrowRightIcon class="size-4" />
            {/if}
          </Button>
        </div>
      </form>

      <!-- Separator -->
      <div class="relative">
        <div class="absolute inset-0 flex items-center">
          <div class="w-full border-t border-border/60"></div>
        </div>
        <div class="relative flex justify-center">
          <span class="bg-card px-3 text-[10px] font-mono tracking-widest text-muted-foreground uppercase">
            atau
          </span>
        </div>
      </div>

      <!-- Login Link -->
      <div class="text-center">
        <p class="text-sm text-muted-foreground">
          Sudah punya akun?
          <a
            href="/login"
            class="text-primary hover:text-primary/80 font-medium transition-colors"
          >
            Masuk
          </a>
        </p>
      </div>
    </div>
  </div>
</main>

