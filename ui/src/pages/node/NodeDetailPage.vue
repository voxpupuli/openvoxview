<script setup lang="ts">
import { computed, onMounted, ref } from 'vue';
import { useRoute } from 'vue-router';
import Backend from 'src/client/backend';
import { PuppetFact } from 'src/puppet/models';
import ReportStatus from 'components/ReportStatus.vue';
import { ApiPuppetNode, PuppetNode } from 'src/puppet/models/puppet-node';
import { ApiPuppetReport, PuppetReport } from 'src/puppet/models/puppet-report';
import { formatTimestamp } from 'src/helper/functions';
import PqlQuery, { PqlEntity, PqlSortOrder } from 'src/puppet/query-builder';
import VueJsonPretty from 'vue-json-pretty';
import 'vue-json-pretty/lib/styles.css';
import { QTableColumn, useQuasar } from 'quasar';
import { useI18n } from 'vue-i18n';

const q = useQuasar();
const { t } = useI18n();
const route = useRoute();
const node = route.params.node;

const node_info = ref<PuppetNode>();
const node_facts = ref<PuppetFact[]>([]);
const reports = ref<PuppetReport[]>([]);
const needle = ref<string | null>(null);
const isLoading = ref(true);

const isReportsLoading = ref(true);

const filteredFacts = computed(() => {
  if (!needle.value) return node_facts.value;

  return node_facts.value.filter((s) =>
    s.name.toLowerCase().includes(needle.value!.toLowerCase())
  );
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
  const query = `facts {certname = '${node}' }`;

  Backend.getRawQueryResult<PuppetFact[]>(query).then((result) => {
    if (result.status === 200) {
      node_facts.value = result.data.Data.Data;
    }
  });
}

function loadNodeInfo() {
  const query = `nodes { certname = '${node}' }`;
  isLoading.value = true;

  Backend.getRawQueryResult<ApiPuppetNode[]>(query)
    .then((result) => {
      if (result.status === 200) {
        node_info.value = PuppetNode.fromApi(result.data.Data.Data[0]);
      }
    })
    .finally(() => {
      isLoading.value = false;
    });
}

function loadReports() {
  isReportsLoading.value = true;
  const query = new PqlQuery(PqlEntity.Reports);
  query.filter().and().equal('certname', node);
  query.sortBy().add('end_time', PqlSortOrder.Descending);
  query.limit(10);

  Backend.getQueryResult<ApiPuppetReport[]>(query)
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

onMounted(async () => {
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
                  <vue-json-pretty
                    :data="fact.value"
                    :theme="q.dark.isActive ? 'dark' : 'light'"
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

<style scoped></style>
