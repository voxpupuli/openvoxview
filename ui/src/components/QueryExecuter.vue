<script setup lang="ts">
/* eslint-disable  @typescript-eslint/no-explicit-any */
import { ref } from 'vue';
import Backend from 'src/client/backend';
import {
  type PuppetQueryResult,
} from 'src/puppet/models';
import { JsonViewer } from 'vue3-json-viewer';
import 'vue3-json-viewer/dist/vue3-json-viewer.css';
import JsonViewDialog from 'components/JsonViewDialog.vue';
import { useQuasar } from 'quasar';
import { useI18n } from 'vue-i18n';

interface SelectedItem {
  name: string;
  value: any;
}

const {t} = useI18n();
const data = ref<PuppetQueryResult<unknown[]>>();
const query = defineModel('query', { type: String });
const isLoading = ref(false);
const tab = ref('data');
const showJsonDialog = ref(false);
const selectedItem = ref<SelectedItem>();
const q = useQuasar();

function executeQuery() {
  console.log('executing: ', query.value);
  if (!query.value) return;
  isLoading.value = true;
  void Backend.getRawQueryResult<unknown[]>(query.value, true)
    .then((result) => {
      if (result.status === 200) {
        data.value = result.data.Data;
      }
    })
    .finally(() => {
      isLoading.value = false;
    });
}

function showJson(label: string, value: any) {
  selectedItem.value = {
    name: label,
    value: value,
  }
  showJsonDialog.value = true;
}
</script>

<template>
  <q-card>
    <q-card-section class="row items-center q-pb-none">
      <div class="text-h6">{{ t('LABEL_EXECUTE_QUERY') }}</div>
      <q-space />
      <slot name="header" />
    </q-card-section>
    <q-card-section>
      <q-input
        v-model="query"
        type="textarea"
        label="Query"
        placeholder="nodes {}"
      />
      <q-btn
        label="Execute"
        class="q-mt-lg"
        color="primary"
        @click="executeQuery"
        :loading="isLoading"
      />
    </q-card-section>
    <q-card-section v-if="data">
      <q-tabs
        v-model="tab"
        dense
        class="text-grey"
        active-color="primary"
        indicator-color="primary"
        align="justify"
        narrow-indicator
      >
        <q-tab name="data" :label="$t('LABEL_DATA')" />
        <q-tab name="json" :label="$t('LABEL_JSON')" />
        <q-tab name="meta" :label="$t('LABEL_META')" />
      </q-tabs>

      <q-separator />

      <q-tab-panels v-model="tab" animated>
        <q-tab-panel v-if="data.Data" name="data">
          <q-table :rows="data.Data">
            <template v-slot:body="props">
              <q-tr :props="props">
                <q-td v-for="col in props.cols" :key="col.name" :props="props">
                  <div v-if="col.value instanceof Object">
                    <q-btn
                      color="primary"
                      icon="visibility"
                      label="json"
                      @click="showJson(col.name, col.value)"
                    />
                  </div>
                  <div v-else>{{ col.value }}</div>
                </q-td>
              </q-tr>
            </template>
          </q-table>
        </q-tab-panel>

        <q-tab-panel v-if="data.Data" name="json">
          <JsonViewer
            :value="data.Data"
            expanded
            :expand-depth="-1"
            :theme="q.dark.isActive ? 'dark' : 'light'"
            preview-mode
          />
        </q-tab-panel>

        <q-tab-panel name="meta">
          <q-list>
            <q-item>
              <q-item-section>
                <q-item-label caption>{{
                  $t('LABEL_EXECUTION_TIME')
                }}</q-item-label>
                <q-item-label>{{ data.ExecutionTimeInMilli }}</q-item-label>
              </q-item-section>
            </q-item>
            <q-item>
              <q-item-section>
                <q-item-label caption>{{
                  $t('LABEL_EXECUTED_ON')
                }}</q-item-label>
                <q-item-label>{{ data.ExecutedOn }}</q-item-label>
              </q-item-section>
            </q-item>
          </q-list>
        </q-tab-panel>
      </q-tab-panels>
    </q-card-section>
    <JsonViewDialog v-if="selectedItem" :model-value="selectedItem.value" :label="selectedItem.name" v-model:show="showJsonDialog"/>
  </q-card>
</template>

<style scoped></style>
