<script setup lang="ts">
import DashboardItem from 'components/DashboardItem.vue';
import NodeTable from 'components/NodeTable.vue';
import Backend from 'src/client/backend';
import { PuppetNodeWithEventCount } from 'src/puppet/models/puppet-node';
import { computed, onMounted, ref, watch } from 'vue';
import { useSettingsStore } from 'stores/settings';
import PqlQuery, { PqlEntity } from 'src/puppet/query-builder';
import { useQuasar } from 'quasar';

const q = useQuasar();
const nodes = ref<PuppetNodeWithEventCount[]>([]);
const settings = useSettingsStore();
const population = ref(0);
const resources = ref(0);

const avg_resources_per_node = computed(() => {
  return resources.value / population.value;
});

const unchanged = computed(() => {
  return nodes.value.filter((s) => s.latest_report_status == 'unchanged')
    .length;
});

const changed = computed(() => {
  return nodes.value.filter((s) => s.latest_report_status == 'changed').length;
});

const failed = computed(() => {
  return nodes.value.filter((s) => s.latest_report_status == 'failed').length;
});

const pending = computed(() => {
  return nodes.value.filter((s) => s.latest_report_status == 'pending').length;
});

const nodesNotEqualUnchaged = computed(() => {
  return nodes.value.filter((s) => s.latest_report_status != 'unchanged');
});

type CountResult = {
  count: number;
};

function loadPopulation() {
  if (!settings.environment) return;
  const builder = new PqlQuery(PqlEntity.Nodes);
  if (settings.hasEnvironment()) {
    builder.filter().and().equal('catalog_environment', settings.environment);
  }
  builder.addProjectionField('count()');

  void Backend.getQueryResult<CountResult[]>(builder).then((result) => {
    if (result.status === 200) {
      population.value = result.data.Data.Data[0]!.count;
    }
  });
}

function loadResources() {
  if (!settings.environment) return;
  const builder = new PqlQuery(PqlEntity.Resources);
  if (settings.hasEnvironment()) {
    builder.filter().and().equal('environment', settings.environment);
  }
  builder.addProjectionField('count()');

  void Backend.getQueryResult<CountResult[]>(builder).then((result) => {
    if (result.status === 200) {
      resources.value = result.data.Data.Data[0]!.count;
    }
  });
}

function loadData() {
  if (!settings.environment) return;
  void Backend.getViewNodeOverview(settings.environment).then((result) => {
    if (result.status === 200) {
      nodes.value = result.data.Data.map((s) =>
        PuppetNodeWithEventCount.fromApi(s),
      );
    }
  });
}

function load() {
  loadPopulation();
  loadResources();
  loadData();
}

onMounted(() => {
  watch(
    () => settings.environment,
    () => {
      load();
    },
    { immediate: true },
  );
});
</script>

<template>
  <q-page padding>
    <div class="row q-pa-sm">
      <DashboardItem
        v-model="failed"
        :suffix="$t('LABEL_NODE', failed)"
        caption="with status failed"
        title_color="negative"
        :to="{ name: 'NodeOverview', query: { status: 'failed' } }"
      />
      <DashboardItem
        v-model="pending"
        :suffix="$t('LABEL_NODE', pending)"
        caption="with status pending"
        title_color="warning"
        :to="{ name: 'NodeOverview', query: { status: 'pending' } }"
      />
      <DashboardItem
        v-model="changed"
        b
        :suffix="$t('LABEL_NODE', changed)"
        caption="with status changed"
        title_color="primary"
        :to="{ name: 'NodeOverview', query: { status: 'changed' } }"
      />
      <DashboardItem
        v-model="unchanged"
        :suffix="$t('LABEL_NODE', unchanged)"
        caption="with status unchanged"
        title_color="positive"
        :to="{ name: 'NodeOverview', query: { status: 'unchanged' } }"
      />
      <DashboardItem
        :model-value="0"
        suffix="nodes"
        caption="unreported in last 3 hours"
        title_color="secondary"
      />
      <DashboardItem
        v-model="population"
        caption="Population"
        :title_color="q.dark ? 'text-white' : ''"
      />
      <DashboardItem
        v-model="resources"
        caption="Resources managed"
        :title_color="q.dark ? 'text-white' : ''"
      />
      <DashboardItem
        v-model="avg_resources_per_node"
        caption="Avg. resources/node"
        :title_color="q.dark ? 'text-white' : ''"
        :decimal_places="2"
      />
    </div>
    <div class="row">
      <NodeTable
        class="q-ma-md col"
        v-model:nodes="nodesNotEqualUnchaged"
        disable_pagination
      />
    </div>
  </q-page>
</template>

<style scoped></style>
