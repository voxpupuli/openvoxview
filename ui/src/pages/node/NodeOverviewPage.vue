<script setup lang="ts">
import { computed, onMounted, ref, watch } from 'vue';
import Backend from 'src/client/backend';
import { useSettingsStore } from 'stores/settings';
import { PuppetNodeWithEventCount } from 'src/puppet/models/puppet-node';
import NodeTable from 'components/NodeTable.vue';
import { useRoute, useRouter } from 'vue-router';
import moment from 'moment';

const route = useRoute();
const router = useRouter();
const filter = ref('');
const nodes = ref<PuppetNodeWithEventCount[]>([]);
const settings = useSettingsStore();
const isLoading = ref(false);
const statusFilter = ref<string[]>();
const statusOptions = ['failed', 'changed', 'unchanged', 'pending', 'unreported'];
const unreportedDate = ref<moment.Moment>();

function loadMeta() {
  void Backend.getMeta().then((result) => {
    if (result.status === 200 && result.data.Data.UnreportedHours) {
      unreportedDate.value = moment().subtract(
        moment.duration(result.data.Data.UnreportedHours, 'hours'),
      );
    }
  });
}

function loadData() {
  if (!settings.environment) return;
  const env = settings.hasEnvironment() ? settings.environment : undefined;
  const hasUnreported = statusFilter.value?.includes('unreported');
  const apiStatuses = hasUnreported ? undefined : statusFilter.value;
  isLoading.value = true;
  void Backend.getViewNodeOverview(env, apiStatuses)
    .then((result) => {
      if (result.status === 200) {
        nodes.value = result.data.Data.map((s) =>
          PuppetNodeWithEventCount.fromApi(s),
        );
      }
    })
    .finally(() => {
      isLoading.value = false;
    });
}

const filteredNodes = computed(() => {
  let result = nodes.value.filter((s) => s.certname.includes(filter.value));

  if (statusFilter.value?.includes('unreported')) {
    const ud = unreportedDate.value;
    const otherStatuses = statusFilter.value.filter((s) => s !== 'unreported');
    result = result.filter((s) => {
      const isUnreported = !s.report_timestamp || (ud ? ud.isAfter(s.report_timestamp) : false);
      if (otherStatuses.length === 0) return isUnreported;
      return isUnreported || otherStatuses.includes(s.latest_report_status);
    });
  }

  return result;
});

function updateRoute() {
  void router.replace({
    name: route.name,
    query: {
      filter: filter.value,
      status: statusFilter.value,
    },
  });
}

watch(filter, () => {
  updateRoute();
});

watch(statusFilter, () => {
  loadData();
  updateRoute();
});

watch(settings, () => {
  loadData();
});

onMounted(() => {
  loadMeta();

  if (route.query.status) {
    const s = route.query.status;
    statusFilter.value = (Array.isArray(s) ? s : [s]).filter((v): v is string => v !== null);
  }

  if (route.query.filter) {
    filter.value = route.query.filter as string;
  }

  watch(
    () => settings.environment,
    () => {
      loadData();
    },
    { immediate: true },
  );
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
