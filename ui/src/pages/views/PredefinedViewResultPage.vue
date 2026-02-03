<script setup lang="ts">
import { computed, onMounted, ref } from 'vue';
import { useRoute } from 'vue-router';
import { PredefinedView, PredefinedViewResult } from 'src/puppet/models';
import backend from 'src/client/backend';
import NodeLink from 'components/NodeLink.vue';
import { getOsNameFromOsFact } from 'src/helper/functions';
import { type QTableColumn } from 'quasar';
import { useSettingsStore } from 'stores/settings';

const route = useRoute();
const viewName = computed(() => {
  return route.params.viewName as string;
});
const viewResult = ref<PredefinedViewResult>();
const viewMeta = ref<PredefinedView>();
const pagination = ref({
  rowsPerPage: 20,
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
    }
  });
}

function paginationUpdate(newPagination: typeof pagination.value) {
  pagination.value = newPagination;

  settings.viewUserSettings[viewName.value] = {
    rowsPerPage: pagination.value.rowsPerPage,
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
