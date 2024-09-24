import { boot } from 'quasar/wrappers';
import axios, { AxiosInstance } from 'axios';
import { Notify } from 'quasar';

declare module 'vue' {
  interface ComponentCustomProperties {
    $axios: AxiosInstance;
    $api: AxiosInstance;
  }
}


const api = axios.create({ baseURL: process.env.VUE_APP_BACKEND_BASE_ADDRESS });

export default boot(({ app }) => {
  // for use inside Vue files (Options API) through this.$axios and this.$api
  api.interceptors.response.use(function (response) {
    // Any status code that lie within the range of 2xx cause this function to trigger
    // Do something with response data
    return response;
  }, function (error) {
    if (error.response) {
      console.log('Cached Error: ', error.response.data.Error);
      Notify.create({
        message: error.response.data.Error,
        color: 'negative',
        multiLine: true,
        closeBtn: true,
      })
    }

    return Promise.reject(error);
  });


  app.config.globalProperties.$axios = axios;
  // ^ ^ ^ this will allow you to use this.$axios (for Vue Options API form)
  //       so you won't necessarily have to import axios in each vue file

  app.config.globalProperties.$api = api;
  // ^ ^ ^ this will allow you to use this.$api (for Vue Options API form)
  //       so you can easily perform requests against your app's API
});

export { api };
