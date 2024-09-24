<script setup lang="ts">
import { onMounted, ref, watch } from 'vue';
import Backend from 'src/client/backend';
import ReportStatus from 'components/ReportStatus.vue';
import { useI18n } from 'vue-i18n';
import { QTableColumn } from 'quasar';
import NodeLink from 'components/NodeLink.vue';
import PqlQuery, { PqlEntity, PqlSortOrder } from 'src/puppet/query-builder';
import { ApiPuppetReport, PuppetReport } from 'src/puppet/models/puppet-report';

const { t } = useI18n();
const reports = ref<PuppetReport[]>([]);
const isLoading = ref(false);
const filter = ref('');
const filterEndTimeStart = ref<string | null>(null);
const filterEndTimeEnd = ref<string | null>(null);

const filterStatus = ref(['failed', 'changed', 'unchanged', 'noop']);
const filterOptions = ref([
  {
    label: t('LABEL_FAILED'),
    value: 'failed',
  },
  {
    label: t('LABEL_CHANGED'),
    value: 'changed',
  },
  {
    label: t('LABEL_UNCHANGED'),
    value: 'unchanged',
  },
  {
    label: t('LABEL_NOOP'),
    value: 'noop',
  },
]);

const pagination = ref({
  sortBy: 'end_time',
  descending: false,
  page: 1,
  rowsPerPage: 100,
  rowsNumber: 10,
});

const columns: QTableColumn[] = [
  {
    name: 'end_time',
    field: 'endTimeFormatted',
    label: t('LABEL_END_TIME'),
    align: 'left',
  },
  {
    name: 'status',
    field: 'status',
    label: t('LABEL_STATUS'),
    align: 'left',
  },
  {
    name: 'certname',
    field: 'certname',
    label: t('LABEL_CERTNAME'),
    align: 'left',
    sortable: true,
  },
  {
    name: 'configuartion_version',
    field: 'configuration_version',
    label: t('LABEL_CONFIGURATION_VERSION'),
    align: 'left',
  },
  {
    name: 'agent_version',
    field: 'puppet_version',
    label: t('LABEL_AGENT_VERSION'),
    align: 'left',
    sortable: true,
  },
];

function loadReports() {
  isLoading.value = true;

  const { page, rowsPerPage, sortBy, descending } = pagination.value;
  const query = new PqlQuery(PqlEntity.Reports);
  if (filter.value && filter.value != '') {
    query.filter()
      .newGroup()
      .or()
      .regex('certname', filter.value)
      .or()
      .regex('puppet_version', filter.value);
  }

  if (filterEndTimeStart.value || filterEndTimeEnd.value) {
    const group = query.filter().newGroup()

    if (filterEndTimeStart.value) {
      group.and().greaterThanEqual('end_time', filterEndTimeStart.value);
    }

    if (filterEndTimeEnd.value) {
      group.and().lowerThanEqual('end_time', filterEndTimeEnd.value);
    }
  }

  if (filterStatus.value.length > 0) {
    query.filter().newGroup().and()
      .in('status', filterStatus.value);
  }

  const start = (page - 1) * rowsPerPage;

  query
    .sortBy()
    .add(sortBy, descending ? PqlSortOrder.Descending : PqlSortOrder.Ascending);
  query.limit(rowsPerPage);
  query.offset(start);

  Backend.getQueryResult<ApiPuppetReport[]>(query)
    .then((result) => {
      if (result.status === 200) {
        reports.value = result.data.Data.Data.map(s => PuppetReport.fromApi(s));
      }
    })
    .finally(() => {
      isLoading.value = false;
    });
}

/* eslint-disable-next-line @typescript-eslint/no-explicit-any */
function onRequest() {
  loadReports();
}

watch(filterStatus, () => {
  loadReports();
})

watch(filterEndTimeStart, () => {
  loadReports();
})

watch(filterEndTimeEnd, () => {
  loadReports();
})

onMounted(() => {
  loadReports();
});
</script>

<template>
  <q-page padding>
    <q-card>
      <q-card-section class="bg-primary text-white text-h6">
        {{ $t('LABEL_FILTER') }}
      </q-card-section>
      <q-card-section>
        <q-input
          debounce="300"
          v-model="filter"
          :placeholder="$t('LABEL_SEARCH')"
          class="full-width"
        />
        <q-select
          :label="$t('LABEL_STATUS')"
          v-model="filterStatus"
          :options="filterOptions"
          multiple
          use-chips
          map-options
          emit-value
          class="full-width"
        >
          <template v-slot:option="{ itemProps, opt, selected, toggleOption }">
            <q-item v-bind="itemProps">
              <q-item-section>
                <q-item-label>{{ opt.label }}</q-item-label>
              </q-item-section>
              <q-item-section side>
                <q-toggle
                  :model-value="selected"
                  @update:model-value="toggleOption(opt)"
                />
              </q-item-section>
            </q-item>
          </template>
        </q-select>
        <div class="row">
          <div class="col q-py-md q-pr-md">
            <q-input
              type="date"
              v-model="filterEndTimeStart"
              :label="$t('LABEL_END_TIME_START')"
            />
          </div>
          <div class="col q-py-md">
            <q-input
              type="date"
              v-model="filterEndTimeEnd"
              :label="$t('LABEL_END_TIME_START')"
            />
          </div>
        </div>
      </q-card-section>
    </q-card>
    <q-table
      :rows="reports"
      :columns="columns"
      row-key="hash"
      v-model:pagination="pagination"
      wrap-cells
      :loading="isLoading"
      :filter="filter"
      binary-state-sort
      @request="onRequest"
      class="q-mt-md"
      table-header-class="bg-primary text-white"
    >
      <template v-slot:body="props">
        <q-tr :props="props">
          <q-td v-for="col in props.cols" :key="col.name" :props="props">
            <div v-if="col.name == 'certname'">
              <NodeLink :certname="col.value" />
            </div>
            <div v-else-if="col.name == 'status'">
              <ReportStatus :report="props.row"/>
            </div>
            <div v-else class="text-subtitle1">
              {{ col.value }}
            </div>
          </q-td>
        </q-tr>
      </template>
    </q-table>
  </q-page>
</template>

<style scoped>

</style>
