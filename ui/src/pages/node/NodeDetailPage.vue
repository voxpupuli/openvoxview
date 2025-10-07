<script setup lang="ts">
import { computed, onMounted, ref } from 'vue';
import { useRoute } from 'vue-router';
import Backend from 'src/client/backend';
import { type ApiPuppetFact, PuppetFact } from 'src/puppet/models';
import ReportStatus from 'components/ReportStatus.vue';
import { type ApiPuppetNode, PuppetNode } from 'src/puppet/models/puppet-node';
import { type ApiPuppetReport, PuppetReport } from 'src/puppet/models/puppet-report';
import { formatTimestamp } from 'src/helper/functions';
import PqlQuery, { PqlEntity, PqlSortOrder } from 'src/puppet/query-builder';
import { type QTableColumn, useQuasar } from 'quasar';
import { useI18n } from 'vue-i18n';
import { JsonViewer } from 'vue3-json-viewer';
import "vue3-json-viewer/dist/vue3-json-viewer.css";

const q = useQuasar();
const { t } = useI18n();
const route = useRoute();
const node = computed(() => {
  return route.params.node as string
});

const node_info = ref<PuppetNode>();
const node_facts = ref<PuppetFact[]>([]);
const reports = ref<PuppetReport[]>([]);
const needle = ref<string | null>(null);
const isLoading = ref(true);

const isReportsLoading = ref(true);

const filteredFacts = computed(() => {
  if (!needle.value) return node_facts.value;

  return node_facts.value.filter((s) => {
    const needleLower = needle.value!.toLowerCase();
    return (
      s.name.toLowerCase().includes(needleLower) ||
      (s.value && JSON.stringify(s.value).toLowerCase().includes(needleLower))
    );
  });
});

const columns: QTableColumn[] = [
  {
    name: 'end_time',
    field: 'endTimeFormatted',
    label: t('LABEL_END_TIME'),
    align: 'left',
    format: (val: string, row: PuppetReport) => row.endTimeFormatted,
  },
  {
    name: 'status',
    field: 'status',
    label: t('LABEL_STATUS'),
    align: 'left',
  },
];

const pagination = ref({
  sortBy: 'end_time',
  descending: false,
  page: 1,
  rowsPerPage: 100,
  rowsNumber: 10,
});

function loadFacts() {
  const query = `facts {certname = '${node.value}' }`;

  void Backend.getRawQueryResult<ApiPuppetFact[]>(query).then((result) => {
    if (result.status === 200) {
      node_facts.value = result.data.Data.Data.map((s) =>
        PuppetFact.fromApi(s)
      );
    }
  });
}

function loadNodeInfo() {
  const query = `nodes { certname = '${node.value}' }`;
  isLoading.value = true;

  void Backend.getRawQueryResult<ApiPuppetNode[]>(query)
    .then((result) => {
      if (result.status === 200) {
        node_info.value = PuppetNode.fromApi(result.data.Data.Data[0]!);
      }
    })
    .finally(() => {
      isLoading.value = false;
    });
}

function loadReports() {
  isReportsLoading.value = true;
  const query = new PqlQuery(PqlEntity.Reports);
  query.filter().and().equal('certname', node.value);
  query.sortBy().add('end_time', PqlSortOrder.Descending);
  query.limit(10);

  void Backend.getQueryResult<ApiPuppetReport[]>(query)
    .then((result) => {
      if (result.status === 200) {
        reports.value = result.data.Data.Data.map((s) =>
          PuppetReport.fromApi(s)
        );
      }
    })
    .finally(() => {
      isReportsLoading.value = false;
    });
}

onMounted(() => {
  loadNodeInfo();
  loadReports();
  loadFacts();
});
</script>

<template>
  <q-page padding>
    <div v-if="!isLoading" class="row">
      <div class="col q-pr-md">
        <q-card>
          <q-card-section class="bg-primary text-white text-h6">
            {{ $t('LABEL_DETAIL', 2) }}
          </q-card-section>
          <q-card-section class="q-pa-none">
            <q-markup-table v-if="node_info" flat>
              <tbody>
                <tr>
                  <td class="text-left text-bold">Certname</td>
                  <td class="text-left text-bold">{{ node_info.certname }}</td>
                </tr>
                <tr>
                  <td class="text-left text-bold">Facts</td>
                  <td class="text-left">
                    {{ formatTimestamp(node_info.facts_timestamp) }}
                  </td>
                </tr>
                <tr>
                  <td class="text-left text-bold">Catalog</td>
                  <td class="text-left">
                    {{ formatTimestamp(node_info.catalog_timestamp) }}
                  </td>
                </tr>
                <tr>
                  <td class="text-left text-bold">Report</td>
                  <td class="text-left">
                    {{ formatTimestamp(node_info.report_timestamp) }}
                  </td>
                </tr>
              </tbody>
            </q-markup-table>
          </q-card-section>
        </q-card>
        <q-card class="q-mt-md">
          <q-card-section class="bg-primary text-white text-h6">
            {{ $t('LABEL_REPORT', 2) }}
          </q-card-section>
          <q-card-section class="q-pa-none">
            <q-table
              v-if="!isReportsLoading"
              :rows="reports"
              :columns="columns"
              row-key="hash"
              v-model:pagination="pagination"
              wrap-cells
              :loading="isReportsLoading"
              binary-state-sort
              table-header-class="bg-primary text-white"
              flat
              square
            >
              <template v-slot:body="props">
                <q-tr :props="props">
                  <q-td
                    v-for="col in props.cols"
                    :key="col.name"
                    :props="props"
                  >
                    <div v-if="col.name == 'status'">
                      <ReportStatus
                        :report="props.row"
                        :inline="q.screen.gt.md"
                      />
                    </div>
                    <div v-else class="text-subtitle1">
                      {{ col.value }}
                    </div>
                  </q-td>
                </q-tr>
              </template>
            </q-table>
            <q-inner-loading :showing="isReportsLoading" />
          </q-card-section>
        </q-card>
      </div>
      <q-card class="col-8">
        <q-card-section class="bg-primary text-white text-h6">
          {{ $t('LABEL_FACT', 2) }}
        </q-card-section>
        <q-card-section>
          <q-input v-model="needle" :label="$t('LABEL_SEARCH')" />
          <q-markup-table class="q-mt-lg" wrap-cells flat>
            <thead>
              <tr>
                <th class="text-left">Name</th>
                <th class="text-left">Value</th>
              </tr>
            </thead>
            <tbody>
              <tr v-for="fact in filteredFacts" :key="fact.name">
                <td>{{ fact.name }}</td>
                <td class="text-left">
                  <JsonViewer
                    :value="fact.value"
                    expanded
                    :expand-depth="-1"
                    :theme="q.dark.isActive ? 'dark' : 'light'"
                    preview-mode
                  />
                </td>
              </tr>
            </tbody>
          </q-markup-table>
        </q-card-section>
      </q-card>
    </div>
    <q-inner-loading :showing="isLoading" />
  </q-page>
</template>

<style scoped>
</style>
