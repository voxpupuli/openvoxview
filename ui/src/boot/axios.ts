import { defineBoot } from '#q-app/wrappers';
import axios, {type AxiosError, type AxiosInstance} from 'axios';
import { Notify } from 'quasar';
import {type ErrorResponse} from "src/client/models";

declare module 'vue' {
  interface ComponentCustomProperties {
    $axios: AxiosInstance;
    $api: AxiosInstance;
  }
}


const api = axios.create({ baseURL: process.env.VUE_APP_BACKEND_BASE_ADDRESS || ''});

export default defineBoot(({ app }) => {
  // for use inside Vue files (Options API) through this.$axios and this.$api
  api.interceptors.response.use(function (response) {
    // Any status code that lie within the range of 2xx cause this function to trigger
    // Do something with response data
    return response;
  }, function (error: AxiosError<ErrorResponse>) {
    if (error.response && error.status != 400) {
      console.log('Cached Error: ', error.response.data.Error);
      Notify.create({
        message: error.response.data.Error ?? error.message,
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
