<script setup lang="ts">
import DashboardItem from 'components/DashboardItem.vue';
import NodeTable from 'components/NodeTable.vue';
import Backend from 'src/client/backend';
import { PuppetNodeWithEventCount } from 'src/puppet/models/puppet-node';
import { computed, onMounted, ref, watch } from 'vue';
import { useSettingsStore } from 'stores/settings';
import PqlQuery, { PqlEntity } from 'src/puppet/query-builder';
import { useQuasar } from 'quasar';
import { type ApiMeta } from 'src/client/models';
import moment from 'moment';

const q = useQuasar();
const nodes = ref<PuppetNodeWithEventCount[]>([]);
const settings = useSettingsStore();
const population = ref(0);
const resources = ref(0);
const meta = ref<ApiMeta>();

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

const nodesNotEqualUnchanged = computed(() => {
  const unreportedDate = meta.value ? moment().subtract(meta.value.UnreportedHours, 'hours') : null;
  return nodes.value.filter((s) =>
    s.latest_report_status != 'unchanged' || !s.report_timestamp || (
      unreportedDate && unreportedDate.isAfter(s.report_timestamp)
    )
  );
});

const unreported = computed(() => {
  if (!meta.value) return 0;
  const unreportedDate = moment().subtract(meta.value.UnreportedHours, 'hours');
  return nodes.value.filter((s) => !s.report_timestamp || unreportedDate.isAfter(s.report_timestamp)).length;
});

const unreportedDuration = computed(() => {
  if (!meta.value) return '...';
  return moment.duration(meta.value.UnreportedHours, 'hours').humanize();
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
  const env = settings.hasEnvironment() ? settings.environment : undefined;
  void Backend.getViewNodeOverview(env).then((result) => {
    if (result.status === 200) {
      nodes.value = result.data.Data.map((s) =>
        PuppetNodeWithEventCount.fromApi(s),
      );
    }
  });
}

function loadMeta() {
  void Backend.getMeta().then((result) => {
    if (result.status === 200) {
      meta.value = result.data.Data;
    }
  });
}

function load() {
  loadPopulation();
  loadResources();
  loadData();
}

onMounted(() => {
  loadMeta();

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
        :model-value="unreported"
        suffix="nodes"
        :caption="$t('LABEL_UNREPORTED', { dur: unreportedDuration })"
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
        v-model:nodes="nodesNotEqualUnchanged"
        disable_pagination
      />
    </div>
  </q-page>
</template>

<style scoped></style>
