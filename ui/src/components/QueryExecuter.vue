<script setup lang="ts">
/* eslint-disable  @typescript-eslint/no-explicit-any */
import { ref, watch } from 'vue';
import Backend from 'src/client/backend';
import { type PuppetQueryResult } from 'src/puppet/models';
import { JsonViewer } from 'vue3-json-viewer';
import 'vue3-json-viewer/dist/vue3-json-viewer.css';
import JsonViewDialog from 'components/JsonViewDialog.vue';
import { useQuasar } from 'quasar';
import { useI18n } from 'vue-i18n';
import {
  formatDuration,
  formatTimestamp,
  copyToClipboard,
} from 'src/helper/functions';
import { type AxiosError } from 'axios';
import { type ErrorResponse } from 'src/client/models';

interface SelectedItem {
  name: string;
  value: any;
}

const { t } = useI18n();
const data = ref<PuppetQueryResult<unknown[]>>();
const query = defineModel('query', { type: String });
const isLoading = ref(false);
const tab = ref('data');
const showJsonDialog = ref(false);
const selectedItem = ref<SelectedItem>();
const q = useQuasar();
const queryError = ref();

const queryParameters = ref<Record<string, string>>({});

watch(query, () => {
  const params: Record<string, string> = queryParameters.value;
  const regex = /\$\{([^}]+)\}/g;
  const currentNames = [];
  let match;
  while ((match = regex.exec(query.value ?? '')) !== null) {
    const name = match[1]!;
    if (!Object.keys(params).includes(name)) {
      params[name] = '';
    }
    currentNames.push(name);
  }

  for (const param of Object.keys(params)) {
    if (!currentNames.includes(param)) {
      delete params[param];
    }
  }

  queryParameters.value = params;
});

function buildQueryWithParameters(): string {
  let builtQuery = query.value ?? '';
  for (const [param, value] of Object.entries(queryParameters.value)) {
    builtQuery = builtQuery.replaceAll('${' + param + '}', value);
  }
  return builtQuery;
}

function executeQuery() {
  const queryWithParams = buildQueryWithParameters();
  console.log('executing: ', queryWithParams);
  if (!queryWithParams) return;
  isLoading.value = true;
  void Backend.getRawQueryResult<unknown[]>(queryWithParams, true)
    .then((result) => {
      if (result.status === 200) {
        data.value = result.data.Data;
        queryError.value = null;
      }
    })
    .catch((error: AxiosError<ErrorResponse>) => {
      if (error.status === 400) {
        queryError.value = error.response?.data.Error;
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
  };
  showJsonDialog.value = true;
}

function getIsoStringFromDate(date: Date): string {
  return new Date(date).toISOString();
}

function copyTimestampToClipboard(timestamp: Date) {
  copyToClipboard(getIsoStringFromDate(timestamp))
    .then(() => {
      q.notify({
        type: 'positive',
        message: t('NOTIFICATION_COPY_TO_CLIPBOARD_SUCCESSFUL'),
      });
    })
    .catch(() => {
      q.notify({
        type: 'negative',
        message: t('NOTIFICATION_COPY_TO_CLIPBOARD_FAILED'),
      });
    });
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
        @keyup.ctrl.enter="executeQuery"
      />
      <q-list
        v-if="queryParameters && Object.keys(queryParameters).length > 0"
        class="q-mt-md q-pa-none"
      >
        <q-item>
          <q-item-section>
            <q-item-label title>
              {{ t('LABEL_PARAMETER', 2) }}
            </q-item-label>
          </q-item-section>
        </q-item>
        <q-item v-for="param in Object.keys(queryParameters)" :key="param">
          <q-item-section side>
            <q-icon name="label" />
          </q-item-section>
          <q-item-section>
            <q-item-label>
              <q-input v-model="queryParameters[param]" :label="param" />
            </q-item-label>
          </q-item-section>
        </q-item>
        <q-item>
          <q-item-section>
            <q-item-label caption>{{
              $t('LABEL_RENDERED_QUERY')
            }}</q-item-label>
            <q-item-label>
              <pre>{{ buildQueryWithParameters() }}</pre>
            </q-item-label>
          </q-item-section>
        </q-item>
      </q-list>
      <q-btn
        label="Execute"
        class="q-mt-lg"
        color="primary"
        @click="executeQuery"
        :loading="isLoading"
      >
        <q-badge
          class="q-ml-md"
          color="secondary"
          :label="t('KEY_CTRL_RETURN')"
        />
      </q-btn>
    </q-card-section>
    <q-card-section v-if="queryError">
      <q-input
        :label="$t('LABEL_ERROR')"
        v-model="queryError"
        readonly
        type="textarea"
        color="negative"
        autogrow
        label-color="negative"
      />
    </q-card-section>
    <q-card-section v-if="data && !queryError">
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
                <q-item-label caption
                  >{{ $t('LABEL_EXECUTION_TIME') }}
                </q-item-label>

                <q-item-label
                  >{{ formatDuration(data.ExecutionTimeInMilli) }}
                </q-item-label>
              </q-item-section>
            </q-item>
            <q-item>
              <q-item-section>
                <q-item-label caption
                  >{{ t('LABEL_EXECUTED_ON') }}
                </q-item-label>
                <q-item-label
                  @click="copyTimestampToClipboard(data.ExecutedOn)"
                >
                  <span>
                    {{ formatTimestamp(data.ExecutedOn) }}
                    <q-tooltip anchor="center middle" self="bottom middle">
                      {{ getIsoStringFromDate(data.ExecutedOn) }}
                    </q-tooltip>
                  </span>
                </q-item-label>
              </q-item-section>
            </q-item>
          </q-list>
        </q-tab-panel>
      </q-tab-panels>
    </q-card-section>
    <JsonViewDialog
      v-if="selectedItem"
      :model-value="selectedItem.value"
      :label="selectedItem.name"
      v-model:show="showJsonDialog"
    />
  </q-card>
</template>

<style scoped>
pre {
  white-space: pre-wrap; /* Since CSS 2.1 */
  white-space: -moz-pre-wrap; /* Mozilla, since 1999 */
  white-space: -pre-wrap; /* Opera 4-6 */
  white-space: -o-pre-wrap; /* Opera 7 */
  word-wrap: break-word; /* Internet Explorer 5.5+ */
}
</style>
