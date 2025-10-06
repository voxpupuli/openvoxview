<script setup lang="ts">
import { computed, onMounted, ref } from 'vue';
import { useRoute } from 'vue-router';
import { PredefinedViewResult } from 'src/puppet/models';
import backend from 'src/client/backend';
import NodeLink from 'components/NodeLink.vue';
import { getOsNameFromOsFact } from 'src/helper/functions';
import { type QTableColumn } from 'quasar';

const route = useRoute();
const viewName = computed(() => {
  return route.params.viewName as string;
});
const viewResult = ref<PredefinedViewResult>();

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

onMounted(() => {
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
