<script setup lang="ts">
import { onMounted, ref } from 'vue';
import PqlQuery, { PqlEntity } from 'src/puppet/query-builder';
import Backend from 'src/client/backend';
import { type PuppetEnvironment } from 'src/puppet/models';
import { useSettingsStore } from 'stores/settings';

const environments = ref<string[]>([]);
const settings = useSettingsStore();

function loadEnvironments() {
  const query = new PqlQuery(PqlEntity.Environments);

  void Backend.getQueryResult<PuppetEnvironment[]>(query).then((result) => {
    if (result.status === 200) {
      environments.value = result.data.Data.Data.map((s) => s.name);
      environments.value.push('*');

      if (!settings.environment) {
        settings.environment = environments.value[0];
      }
    }
  });
}

onMounted(() => {
  loadEnvironments();
});
</script>

<template>
  <q-select v-model="settings.environment" :options="environments" />
</template>

<style scoped></style>
