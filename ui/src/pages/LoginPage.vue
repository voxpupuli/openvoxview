<template>
  <q-page class="flex flex-center">
    <q-card class="login-card" flat bordered>
      <q-card-section class="text-center q-pt-lg">
        <q-img src="logo.png" width="64px" height="64px" />
        <div class="text-h5 q-mt-sm">OpenVox View</div>
        <div class="text-caption text-grey">Sign in to continue</div>
      </q-card-section>

      <q-card-section>
        <q-form @submit="onLogin" class="q-gutter-md">
          <q-input
            v-model="username"
            label="Username"
            outlined
            dense
            :disable="loading"
            autocomplete="username"
          >
            <template v-slot:prepend>
              <q-icon name="person" />
            </template>
          </q-input>

          <q-input
            v-model="password"
            label="Password"
            type="password"
            outlined
            dense
            :disable="loading"
            autocomplete="current-password"
            @keyup.enter="onLogin"
          >
            <template v-slot:prepend>
              <q-icon name="lock" />
            </template>
          </q-input>

          <q-banner v-if="errorMessage" dense rounded class="bg-negative text-white q-mb-sm">
            {{ errorMessage }}
          </q-banner>

          <q-btn
            type="submit"
            label="Login"
            color="primary"
            class="full-width"
            :loading="loading"
          />
        </q-form>

        <template v-if="samlEnabled">
          <div class="row items-center q-my-md">
            <q-separator class="col" />
            <span class="q-mx-sm text-grey">{{ $t('LABEL_LOGIN_OR') }}</span>
            <q-separator class="col" />
          </div>
          <q-btn
            :label="$t('BTN_LOGIN_SSO')"
            icon="corporate_fare"
            color="primary"
            outline
            class="full-width"
            @click="samlLogin"
          />
        </template>
      </q-card-section>
    </q-card>
  </q-page>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue';
import { useRouter, useRoute } from 'vue-router';
import { useAuthStore } from 'stores/auth';
import Backend from 'src/client/backend';

const router = useRouter();
const route = useRoute();
const auth = useAuthStore();

const username = ref('');
const password = ref('');
const loading = ref(false);
const errorMessage = ref('');
const samlEnabled = ref(false);

onMounted(async () => {
  // Handle token from SAML redirect (ADR-002)
  if (route.query.token) {
    auth.setAuth({
      access_token: route.query.token as string,
      refresh_token: (route.query.refresh as string) || '',
      expires_in: 900,
    });
    void router.replace({ name: 'Dashboard' });
    return;
  }

  // Check if auth is even enabled and if SAML is available
  try {
    const meta = await Backend.getMeta();
    auth.setAuthEnabled(meta.data.Data.AuthEnabled);
    samlEnabled.value = meta.data.Data.SamlEnabled;
    if (!meta.data.Data.AuthEnabled) {
      void router.replace({ name: 'Dashboard' });
      return;
    }
  } catch {
    // If meta fails, assume auth is enabled
  }

  // Already authenticated
  if (auth.isAuthenticated) {
    void router.replace({ name: 'Dashboard' });
  }
});

function samlLogin() {
  window.location.href = '/api/v1/auth/saml/login';
}

async function onLogin() {
  if (!username.value || !password.value) {
    errorMessage.value = 'Username and password are required';
    return;
  }

  loading.value = true;
  errorMessage.value = '';

  try {
    const result = await Backend.login(username.value, password.value);
    auth.setAuth(result.data.Data);
    void router.replace({ name: 'Dashboard' });
  } catch (error: unknown) {
    const axiosError = error as { response?: { status?: number } };
    if (axiosError.response?.status === 401) {
      errorMessage.value = 'Invalid username or password';
    } else if (axiosError.response?.status === 429) {
      errorMessage.value = 'Too many login attempts. Please try again later.';
    } else {
      errorMessage.value = 'Login failed. Please try again.';
    }
  } finally {
    loading.value = false;
  }
}
</script>

<style scoped>
.login-card {
  width: 100%;
  max-width: 400px;
}
</style>
