<script setup lang="ts">
import { computed, onMounted, ref, watch } from 'vue';
import Backend from 'src/client/backend';
import { useSettingsStore } from 'stores/settings';
import { PuppetNodeWithEventCount } from 'src/puppet/models/puppet-node';
import NodeTable from 'components/NodeTable.vue';
import { useRoute } from 'vue-router';

const route = useRoute();
const filter = ref('');
const nodes = ref<PuppetNodeWithEventCount[]>([]);
const settings = useSettingsStore();
const isLoading = ref(false);
const statusFilter = ref<string[]>();
const statusOptions = ['failed', 'changed', 'unchanged', 'pending'];

function loadData() {
  if (!settings.environment) return;
  isLoading.value = true;
  Backend.getViewNodeOverview(settings.environment, statusFilter.value).then(
    (result) => {
      if (result.status === 200) {
        nodes.value = result.data.Data.map((s) =>
          PuppetNodeWithEventCount.fromApi(s)
        );
      }
    }
  ).finally(() => {
    isLoading.value = false;
  });
}

const filteredNodes = computed(() => {
  return nodes.value.filter((s) => s.certname.includes(filter.value));
});

watch(statusFilter, () => {
  loadData()
})
watch(settings, () => {
  loadData();
});

onMounted(() => {
  if (route.query.status) {
    statusFilter.value = [route.query.status as string];
  }
  loadData();
});


</script>

<template>
  <q-page padding>
    <div class="row">
      <div class="col q-pr-sm">
        <q-input v-model="filter" :label="$t('LABEL_SEARCH')" />
      </div>
      <div class="col-4 q-pl-sm">
        <q-select
          label="status"
          v-model="statusFilter"
          :options="statusOptions"
          multiple
          use-chips
        />
      </div>
    </div>
    <NodeTable v-model:nodes="filteredNodes" disable_pagination />
    <q-inner-loading :showing="isLoading" />
  </q-page>
</template>

<style scoped></style>
