import type { User } from "$lib/types";
import * as authApi from "$lib/api/auth";

interface AuthState {
  user: User | null;
  loading: boolean;
}

const state: AuthState = $state({
  user: null,
  loading: true,
});

export const auth = {
  get user() {
    return state.user;
  },
  get loading() {
    return state.loading;
  },
  get authenticated() {
    return state.user !== null;
  },

  async fetchUser() {
    state.loading = true;
    try {
      state.user = await authApi.getMe();
    } catch {
      state.user = null;
    } finally {
      state.loading = false;
    }
  },

  async login(email: string, password: string) {
    const user = await authApi.login({ email, password });
    state.user = user;
    return user;
  },

  async register(email: string, name: string, password: string) {
    const user = await authApi.register({ email, name, password });
    state.user = user;
    return user;
  },

  async logout() {
    await authApi.logout();
    state.user = null;
  },

  clear() {
    state.user = null;
    state.loading = false;
  },
};
