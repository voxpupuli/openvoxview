import { defineStore } from 'pinia';
import { ref } from 'vue';

export interface ViewUserSetting {
  rowsPerPage: number;
}

export interface ViewUserSettings {
  [viewName: string]: ViewUserSetting;
}

export const useSettingsStore = defineStore(
  'settings',
  () => {
    const environment = ref<string | null>();
    const darkMode = ref(false);
    const viewUserSettings = ref<ViewUserSettings>({});

    function hasEnvironment() {
      if (!environment.value) return false;
      return environment.value != '*' && environment.value != '';
    }

    return { environment, darkMode, hasEnvironment, viewUserSettings };
  },
  { persist: true }
);
