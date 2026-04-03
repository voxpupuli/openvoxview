import { defineStore } from 'pinia';
import { computed, ref } from 'vue';
import type { LoginResponse } from 'src/client/models';

export const useAuthStore = defineStore(
  'auth',
  () => {
    const accessToken = ref<string | null>(null);
    const refreshToken = ref<string | null>(null);
    const username = ref<string | null>(null);
    const email = ref<string | null>(null);
    const displayName = ref<string | null>(null);
    const expiresAt = ref<number | null>(null);
    const authEnabled = ref<boolean | null>(null);

    const isAuthenticated = computed(() => {
      if (authEnabled.value === false) return true;
      return (
        !!accessToken.value &&
        !!expiresAt.value &&
        Date.now() < expiresAt.value * 1000
      );
    });

    const needsRefresh = computed(() => {
      if (!accessToken.value || !expiresAt.value) return false;
      // Token expires within 60 seconds
      return Date.now() > (expiresAt.value - 60) * 1000;
    });

    function setAuth(data: LoginResponse) {
      accessToken.value = data.access_token;
      refreshToken.value = data.refresh_token;
      expiresAt.value = Math.floor(Date.now() / 1000) + data.expires_in;

      // Decode username from JWT payload
      try {
        const parts = data.access_token.split('.');
        const payload = JSON.parse(atob(parts[1] ?? ''));
        username.value = payload.username || null;
        email.value = payload.email || null;
        displayName.value = payload.display_name || null;
      } catch {
        // If decoding fails, keep existing values
      }
    }

    function clearAuth() {
      accessToken.value = null;
      refreshToken.value = null;
      username.value = null;
      email.value = null;
      displayName.value = null;
      expiresAt.value = null;
    }

    function setAuthEnabled(enabled: boolean) {
      authEnabled.value = enabled;
    }

    return {
      accessToken,
      refreshToken,
      username,
      email,
      displayName,
      expiresAt,
      authEnabled,
      isAuthenticated,
      needsRefresh,
      setAuth,
      clearAuth,
      setAuthEnabled,
    };
  },
  { persist: true },
);
