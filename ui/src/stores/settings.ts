import { defineStore } from 'pinia';
import { ref } from 'vue';

export const useSettingsStore = defineStore(
  'settings',
  () => {
    const environment = ref<string | null>();
    const darkMode = ref(false);

    function hasEnvironment() {
      if (!environment.value) return false;
      return environment.value != '*' && environment.value != '';
    }

    return { environment, darkMode, hasEnvironment };
  },
  { persist: true }
);
