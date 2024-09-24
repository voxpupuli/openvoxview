<script setup lang="ts">
import QueryExecuter from 'components/QueryExecuter.vue';
import { onMounted, ref } from 'vue';
import Backend from 'src/client/backend';
import { PuppetQueryHistoryEntry, PuppetQueryPredefined } from 'src/puppet/models';

const queries = ref<string[]>([]);
const tab = ref('query');
const queryHistoryEntries = ref<PuppetQueryHistoryEntry[]>([]);
const predefinedQueries = ref<PuppetQueryPredefined[]>([]);

function loadHistory() {
  Backend.getQueryHistory().then((result) => {
    if (result.status === 200) {
      queryHistoryEntries.value = result.data.Data;
    }
  });
}

function loadPredefinedQueries() {
  Backend.getQueryPredefined().then(result => {
    predefinedQueries.value = result.data.Data.map(s => PuppetQueryPredefined.fromApi(s));
  })
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
          :label="$t('BTN_ADD_NEW_QUERY')"
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
        <q-btn icon="refresh" @click="loadHistory" />
        <q-markup-table>
          <thead>
            <tr>
              <th>{{ $t('LABEL_QUERY') }}</th>
              <th></th>
            </tr>
          </thead>
          <tbody>
            <tr
              v-for="historyEntry in queryHistoryEntries"
              :key="historyEntry.Query.Query"
            >
              <td>{{ historyEntry.Query.Query }}</td>
              <td>
                <q-btn
                  icon="play_arrow"
                  color="primary"
                  @click="addNewQuery(historyEntry.Query.Query)"
                />
              </td>
            </tr>
          </tbody>
        </q-markup-table>
      </q-tab-panel>
      <q-tab-panel name="predefined">
        <q-markup-table>
          <thead>
          <tr>
            <th>{{ $t('LABEL_DESCRIPTION') }}</th>
            <th>{{ $t('LABEL_QUERY') }}</th>
            <th></th>
          </tr>
          </thead>
          <tbody>
          <tr
            v-for="query in predefinedQueries"
            :key="query.Query"
          >
            <td>{{ query.Description }}</td>
            <td><pre>{{ query.Query }}</pre></td>
            <td>
              <q-btn
                icon="play_arrow"
                color="primary"
                @click="addNewQuery(query.Query)"
              />
            </td>
          </tr>
          </tbody>
        </q-markup-table>
      </q-tab-panel>
    </q-tab-panels>
  </q-page>
</template>

<style scoped></style>
