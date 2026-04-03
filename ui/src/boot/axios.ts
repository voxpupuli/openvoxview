import { defineBoot } from '#q-app/wrappers';
import axios, { type AxiosError, type AxiosInstance, type InternalAxiosRequestConfig } from 'axios';
import { Notify } from 'quasar';
import { type ErrorResponse } from 'src/client/models';
import { useAuthStore } from 'stores/auth';

declare module 'vue' {
  interface ComponentCustomProperties {
    $axios: AxiosInstance;
    $api: AxiosInstance;
  }
}

const api = axios.create({ baseURL: process.env.VUE_APP_BACKEND_BASE_ADDRESS || '' });

let isRefreshing = false;
let failedQueue: Array<{
  resolve: (config: InternalAxiosRequestConfig) => void;
  reject: (error: unknown) => void;
}> = [];

function processQueue(error: unknown) {
  failedQueue.forEach((prom) => {
    if (error) {
      prom.reject(error);
    }
  });
  failedQueue = [];
}

export default defineBoot(({ app, router }) => {
  // Request interceptor: inject Authorization header
  api.interceptors.request.use((config) => {
    const auth = useAuthStore();
    if (auth.accessToken) {
      config.headers.Authorization = `Bearer ${auth.accessToken}`;
    }
    return config;
  });

  // Response interceptor: handle errors and token refresh
  api.interceptors.response.use(
    function (response) {
      return response;
    },
    async function (error: AxiosError<ErrorResponse>) {
      const originalRequest = error.config;

      // Handle 401 with silent token refresh
      if (
        error.response?.status === 401 &&
        originalRequest &&
        !originalRequest.url?.includes('/auth/login') &&
        !originalRequest.url?.includes('/auth/refresh')
      ) {
        const auth = useAuthStore();

        if (!auth.refreshToken) {
          auth.clearAuth();
          void router.push({ name: 'Login' });
          return Promise.reject(error);
        }

        if (isRefreshing) {
          return new Promise((resolve, reject) => {
            failedQueue.push({
              resolve: () => {
                // Re-add the updated auth header
                const authStore = useAuthStore();
                if (originalRequest.headers && authStore.accessToken) {
                  originalRequest.headers.Authorization = `Bearer ${authStore.accessToken}`;
                }
                resolve(api(originalRequest));
              },
              reject,
            });
          });
        }

        isRefreshing = true;

        try {
          const { default: Backend } = await import('src/client/backend');
          const res = await Backend.refreshToken(auth.refreshToken);
          auth.setAuth(res.data.Data);

          // Retry queued requests
          failedQueue.forEach((prom) => {
            prom.resolve(originalRequest);
          });
          failedQueue = [];

          // Retry original request
          if (originalRequest.headers) {
            originalRequest.headers.Authorization = `Bearer ${auth.accessToken}`;
          }
          return api(originalRequest);
        } catch (refreshError: unknown) {
          processQueue(refreshError);
          auth.clearAuth();
          void router.push({ name: 'Login' });
          return Promise.reject(refreshError instanceof Error ? refreshError : new Error('Token refresh failed'));
        } finally {
          isRefreshing = false;
        }
      }

      // Show notification for non-400, non-401 errors
      if (error.response && error.response.status !== 400 && error.response.status !== 401) {
        Notify.create({
          message: error.response.data.Error ?? error.message,
          color: 'negative',
          multiLine: true,
          closeBtn: true,
        });
      }

      return Promise.reject(error);
    },
  );

  app.config.globalProperties.$axios = axios;
  app.config.globalProperties.$api = api;
});

export { api };
