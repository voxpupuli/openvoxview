import { defineStore } from 'pinia';
import { ref } from 'vue';

export const useSettingsStore = defineStore(
  'settings',
  () => {
    const environment = ref<string | null>();
    const darkMode = ref(false);

    return { environment, darkMode };
  },
  { persist: true }
);
