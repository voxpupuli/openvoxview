<script setup lang="ts">
import QueryExecuter from 'components/QueryExecuter.vue';
import { onMounted, ref } from 'vue';
import Backend from 'src/client/backend';
import {
  type PuppetQueryHistoryEntry,
  PuppetQueryPredefined,
} from 'src/puppet/models';
import { useI18n } from 'vue-i18n';
import { emptyPagination } from 'src/helper/objects';
import { formatTimestamp } from 'src/helper/functions';
import { type QTableColumn } from 'quasar';

const queries = ref<string[]>([]);
const tab = ref('query');
const queryHistoryEntries = ref<PuppetQueryHistoryEntry[]>([]);
const predefinedQueries = ref<PuppetQueryPredefined[]>([]);
const { t } = useI18n();

const predefinedColumns: QTableColumn[] = [
  {
    name: 'description',
    field: 'Description',
    label: t('LABEL_DESCRIPTION'),
    align: 'left',
  },
  {
    name: 'query',
    field: 'Query',
    label: t('LABEL_QUERY'),
    align: 'left',
  },
  {
    name: 'actions',
    label: t('LABEL_ACTION', 2),
    align: 'left',
    field: '',
  },
];

const historyColumns: QTableColumn[] = [
  {
    name: 'query',
    field: (row) => row.Query.Query,
    label: t('LABEL_QUERY'),
    align: 'left',
  },
  {
    name: 'timestamp',
    field: (row) => row.Result.ExecutedOn,
    label: t('LABEL_TIMESTAMP'),
    align: 'left',
    format: (value) => formatTimestamp(value),
    sortable: true,
  },
  {
    name: 'actions',
    label: t('LABEL_ACTION', 2),
    align: 'left',
    field: '',
  },
];

function loadHistory() {
  void Backend.getQueryHistory().then((result) => {
    if (result.status === 200) {
      queryHistoryEntries.value = result.data.Data;
    }
  });
}

function loadPredefinedQueries() {
  void Backend.getQueryPredefined().then((result) => {
    predefinedQueries.value = result.data.Data.map((s) =>
      PuppetQueryPredefined.fromApi(s)
    );
  });
}

function addNewQuery(query?: string) {
  if (!query) query = '';
  queries.value.push(query);
  tab.value = 'query';
}

function removeQuery(index: number) {
  queries.value.splice(index, 1);
}

onMounted(() => {
  addNewQuery();
  loadHistory();
  loadPredefinedQueries();
});
</script>

<template>
  <q-page padding>
    <q-tabs v-model="tab">
      <q-tab name="query" label="Query" />
      <q-tab name="history" label="History" />
      <q-tab name="predefined" label="pre defined" />
    </q-tabs>
    <q-tab-panels v-model="tab">
      <q-tab-panel name="query">
        <q-btn
          class="full-width"
          color="primary"
          :label="t('BTN_ADD_NEW_QUERY')"
          @click="addNewQuery()"
        />
        <div class="q-mt-md" v-for="(query, index) in queries" :key="index">
          <QueryExecuter :query="query">
            <template v-slot:header>
              <q-btn
                color="negative"
                icon="delete"
                @click="removeQuery(index)"
              />
            </template>
          </QueryExecuter>
        </div>
      </q-tab-panel>
      <q-tab-panel name="history">
        <q-btn
          class="full-width q-mb-md"
          color="primary"
          icon="refresh"
          @click="loadHistory"
          :label="$t('BTN_REFRESH')"
        />
        <q-table :rows="queryHistoryEntries" :columns="historyColumns">
          <template v-slot:body="props">
            <q-tr :props="props">
              <q-td v-for="col in props.cols" :key="col.name" :props="props">
                <div v-if="col.name == 'query'">
                  <pre>{{ col.value }}</pre>
                </div>
                <div v-else-if="col.name == 'actions'">
                  <q-btn
                    icon="play_arrow"
                    color="primary"
                    @click="addNewQuery(props.row.Query.Query)"
                  />
                </div>
                <div v-else>{{ col.value }}</div>
              </q-td>
            </q-tr>
          </template>
        </q-table>
      </q-tab-panel>
      <q-tab-panel name="predefined">
        <q-table
          :rows="predefinedQueries"
          :pagination="emptyPagination"
          :columns="predefinedColumns"
          hide-pagination
          table-header-class="bg-primary text-white"
        >
          <template v-slot:body="props">
            <q-tr :props="props">
              <q-td v-for="col in props.cols" :key="col.name" :props="props">
                <div v-if="col.name == 'query'">
                  <pre>{{ col.value }}</pre>
                </div>
                <div v-else-if="col.name == 'actions'">
                  <q-btn
                    icon="play_arrow"
                    color="primary"
                    @click="addNewQuery(props.row.Query)"
                  />
                </div>
                <div v-else>{{ col.value }}</div>
              </q-td>
            </q-tr>
          </template>
        </q-table>
      </q-tab-panel>
    </q-tab-panels>
  </q-page>
</template>

<style scoped></style>
