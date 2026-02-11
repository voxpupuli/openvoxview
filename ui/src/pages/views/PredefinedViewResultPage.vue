<script setup lang="ts">
import { computed, onMounted, ref } from 'vue';
import { useRoute } from 'vue-router';
import { type ApiPredefinedViewFact, PredefinedView, PredefinedViewResult } from 'src/puppet/models';
import backend from 'src/client/backend';
import NodeLink from 'components/NodeLink.vue';
import { getOsNameFromOsFact } from 'src/helper/functions';
import { type QTableColumn, type QTableProps } from 'quasar';
import { useSettingsStore } from 'stores/settings';

type Pagination = Parameters<NonNullable<QTableProps['onUpdate:pagination']>>[0];

const DEFAULT_ROWS_PER_PAGE = 20;

const route = useRoute();
const viewName = computed(() => {
  return route.params.viewName as string;
});
const viewResult = ref<PredefinedViewResult>();
const viewMeta = ref<PredefinedView>();
const pagination = ref<NonNullable<QTableProps['pagination']>>({
  rowsPerPage: DEFAULT_ROWS_PER_PAGE,
});
const settings = useSettingsStore();

const columns = computed((): QTableColumn[] => {
  return viewResult.value!.View.Facts.map((s) => {
    return {
      name: s.Name,
      field: s.Fact,
      label: s.Name,
      format: (val: string, row: never) => getProperty(row, s.Fact),
      sortable: true,
      rawSort: (a: never, b: never, rowA: never, rowB: never) => sortColumn(s, a, b, rowA, rowB),
    } as QTableColumn;
  });
});

function getProperty(obj: never, key: string): string | undefined {
  const keys = key.split('.');
  let result: never = obj;

  for (const k of keys) {
    if (Object.keys(result).findIndex((s) => s == k) < 0) return undefined;

    result = result[k];
  }

  return result;
}

function sortColumn(col: ApiPredefinedViewFact, _a: never, _b: never, rowA: never, rowB: never): number {
  let valA: string | undefined;
  let valB: string | undefined;

  switch (col.Renderer) {
    case 'hostname':
      valA = getProperty(rowA, 'trusted.hostname');
      valB = getProperty(rowB, 'trusted.hostname');
      break;
    case 'certname':
      valA = getProperty(rowA, 'trusted.certname');
      valB = getProperty(rowB, 'trusted.certname');
      break;
    case 'os_name':
      valA = getOsNameFromOsFact(getProperty(rowA, 'os'));
      valB = getOsNameFromOsFact(getProperty(rowB, 'os'));
      break;
    default:
      valA = getProperty(rowA, col.Fact);
      valB = getProperty(rowB, col.Fact);
  }

  if (valA === valB) return 0;
  if (valA === undefined) return 1;
  if (valB === undefined) return -1;

  return String(valA).localeCompare(String(valB), undefined, { numeric: true });
}

function loadViewResult() {
  void backend.getPredefinedViewsResult(viewName.value).then((result) => {
    if (result.status === 200) {
      viewResult.value = PredefinedViewResult.fromApi(result.data.Data);
    }
  });
}

function getColumnRenderer(colName: string): string | undefined {
  if (!viewResult.value) return undefined;
  return viewResult.value.View.Facts.find((s) => s.Name == colName)!.Renderer;
}

function loadViewMeta() {
  void backend.getPredefinedViewsMeta(viewName.value).then((result) => {
    if (result.status === 200) {
      viewMeta.value = PredefinedView.fromApi(result.data.Data);

      if (Object.keys(settings.viewUserSettings).includes(viewName.value)) {
        const userSettings = settings.viewUserSettings[viewName.value];
        if (userSettings && userSettings.rowsPerPage > 0) {
          pagination.value.rowsPerPage = userSettings.rowsPerPage;
        }
      } else if (viewMeta.value.RowsPerPage > 0) {
        pagination.value.rowsPerPage = viewMeta.value.RowsPerPage;
      }

      // Set default sort column to the first one
      const [firstFact] = viewMeta.value.Facts;
      if (firstFact) {
        pagination.value.sortBy = firstFact.Name;
      }
    }
  });
}

function paginationUpdate(newPagination: Pagination) {
  pagination.value = newPagination;

  settings.viewUserSettings[viewName.value] = {
    rowsPerPage: newPagination.rowsPerPage,
  };
}

onMounted(() => {
  loadViewMeta();
  loadViewResult();
});
</script>

<template>
  <q-page padding>
    <div class="text-h4 q-mb-md">{{ viewName }}</div>
    <div v-if="viewResult">
      <q-table
        :rows="viewResult.Data"
        :columns="columns"
        table-header-class="bg-primary text-white"
        :pagination="pagination"
        @update:pagination="paginationUpdate"
      >
        <template v-slot:body="props">
          <q-tr :props="props">
            <q-td v-for="col in props.cols" :key="col.name" :props="props">
              <div v-if="getColumnRenderer(col.name) == 'hostname'">
                <NodeLink
                  :certname="props.row.trusted.certname"
                  :label="props.row.trusted.hostname"
                />
              </div>
              <div v-else-if="getColumnRenderer(col.name) == 'certname'">
                <NodeLink :certname="props.row.trusted.certname" />
              </div>
              <div v-else-if="getColumnRenderer(col.name) == 'os_name'">
                {{ getOsNameFromOsFact(props.row.os) }}
              </div>
              <div v-else>
                {{ col.value }}
              </div>
            </q-td>
          </q-tr>
        </template>
      </q-table>
    </div>
  </q-page>
</template>

<style scoped></style>
