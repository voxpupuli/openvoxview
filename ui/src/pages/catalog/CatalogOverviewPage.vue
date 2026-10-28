<script setup lang="ts">
import { computed, onMounted, ref, watch } from 'vue';
import Backend from 'src/client/backend';
import { useI18n } from 'vue-i18n';
import { type QTableColumn } from 'quasar';
import NodeLink from 'components/NodeLink.vue';
import PqlQuery, { PqlEntity, PqlSortOrder } from 'src/puppet/query-builder';
import { useSettingsStore } from 'stores/settings';
import RefreshIntervalSelect from 'components/RefreshIntervalSelect.vue';
import { exportQTableAsCsv } from 'src/helper/csv';
import {
  type ApiPuppetCatalog,
  PuppetCatalog,
} from 'src/puppet/models/puppet-catalog';

interface PaginationInterface {
  sortBy?: string | null;
  descending?: boolean;
  page?: number;
  rowsPerPage?: number;
  rowsNumber?: number;
}

const { t } = useI18n();
const catalogs = ref<PuppetCatalog[]>([]);
const isLoading = ref(false);
const filter = ref('');
const settings = useSettingsStore();
const catalogA = ref<PuppetCatalog | null>();
const diffStart = computed(() => catalogA.value != null);

const pagination = ref<PaginationInterface>({
  sortBy: 'producer_timestamp',
  descending: true,
  page: 1,
  rowsPerPage: 100,
  rowsNumber: 10,
});

const columns: QTableColumn[] = [
  {
    name: 'certname',
    field: 'certname',
    label: t('LABEL_CERTNAME'),
    align: 'left',
    sortable: true,
  },
  {
    name: 'producer_timestamp',
    field: 'producer_timestamp',
    label: t('LABEL_PRODUCER_TIMESTAMP'),
    align: 'left',
    sortable: true,
  },
  {
    name: 'version',
    field: 'version',
    label: t('LABEL_VERSION'),
    align: 'left',
    sortable: true,
  },
  {
    name: 'environment',
    field: 'environment',
    label: t('LABEL_ENVIRONMENT'),
    align: 'left',
    sortable: true,
  },
  {
    name: 'actions',
    label: t('LABEL_ACTION', 2),
    align: 'right',
  },
] as QTableColumn[];

function loadCatalogs() {
  isLoading.value = true;

  const { page, rowsPerPage, sortBy, descending } = pagination.value;
  const query = new PqlQuery(PqlEntity.Catalogs);

  if (settings.hasEnvironment()) {
    query.filter().and().equal('environment', settings.environment);
  }

  const start = ((page ?? 1) - 1) * (rowsPerPage ?? 0);

  if (sortBy) {
    query
      .sortBy()
      .add(
        sortBy,
        descending ? PqlSortOrder.Descending : PqlSortOrder.Ascending,
      );
  }

  if (rowsPerPage) {
    query.limit(rowsPerPage);
    query.offset(start);
  }

  void Backend.getQueryResult<ApiPuppetCatalog[]>(query)
    .then((result) => {
      if (result.status === 200) {
        catalogs.value = result.data.Data.Data.map((s) =>
          PuppetCatalog.fromApi(s),
        );
      }
    })
    .finally(() => {
      isLoading.value = false;
    });
}

function onRequest(props: { pagination: PaginationInterface }) {
  pagination.value = props.pagination;
  loadCatalogs();
}

onMounted(() => {
  watch(
    () => settings.environment,
    () => {
      loadCatalogs();
    },
    { immediate: true },
  );
});
</script>

<template>
  <q-page padding>
    <q-table
      :rows="catalogs"
      :columns="columns"
      row-key="hash"
      v-model:pagination="pagination"
      wrap-cells
      :loading="isLoading"
      :filter="filter"
      binary-state-sort
      @request="onRequest"
      class="q-mt-md"
      :title="$t('LABEL_CATALOG', 2)"
    >
      <template v-slot:top-right>
        <div class="row q-gutter-sm">
          <RefreshIntervalSelect @refresh="loadCatalogs" />
          <q-btn icon="refresh" color="secondary" @click="loadCatalogs()">
            <q-tooltip>{{ $t('LABEL_REFRESH') }}</q-tooltip>
          </q-btn>
          <q-btn
            v-if="!diffStart"
            color="primary"
            icon-right="archive"
            :label="$t('EXPORT_AS_CSV')"
            no-caps
            @click="() => exportQTableAsCsv(catalogs, columns, t)"
          />
        </div>
      </template>
      <template v-slot:body="props">
        <q-tr :props="props">
          <q-td v-for="col in props.cols" :key="col.name" :props="props">
            <div v-if="col.name == 'certname'">
              <NodeLink :certname="col.value" />
            </div>
            <div v-else-if="col.name == 'producer_timestamp'">
              <router-link
                class="text-primary text-subtitle2"
                :to="{
                  name: 'CatalogDetail',
                  params: { certname: props.row.certname },
                }"
                >{{ props.row.producer_timestamp }}</router-link
              >
            </div>
            <div v-else-if="col.name == 'actions'">
              <q-btn
                icon="compare_arrows"
                color="primary"
                v-if="catalogA == null"
                @click="catalogA = props.row as PuppetCatalog"
              >
                <q-tooltip>{{ t('LABEL_COMPARE') }}</q-tooltip>
              </q-btn>
              <q-btn
                icon="compare_arrows"
                color="primary"
                v-if="
                  catalogA != null && props.row.certname != catalogA.certname
                "
                :to="{
                  name: 'CatalogDiffView',
                  params: {
                    source: catalogA.certname,
                    against: props.row.certname,
                  },
                }"
              >
                <q-tooltip>{{ t('LABEL_COMPARE_WITH', {certname: catalogA.certname}) }}</q-tooltip>
              </q-btn>
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

<style scoped></style>
