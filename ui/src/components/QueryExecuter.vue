<script setup lang="ts">
import { ref } from 'vue';
import Backend from 'src/client/backend';
import VueJsonPretty from 'vue-json-pretty';
import 'vue-json-pretty/lib/styles.css';
import { PuppetQueryResult } from 'src/puppet/models';

const data = ref<PuppetQueryResult<unknown[]> | null>(null);
const query = defineModel('query', { type: String });
const isLoading = ref(false);
const tab = ref('data');

function executeQuery() {
  console.log('executing: ', query.value)
  if (!query.value) return;
  isLoading.value = true;
  Backend.getRawQueryResult<unknown[]>(query.value, true)
    .then((result) => {
      if (result.status === 200) {
        data.value = result.data.Data;
      }
    })
    .finally(() => {
      isLoading.value = false;
    });
}
</script>

<template>
  <q-card>
    <q-card-section class="row items-center q-pb-none">
      <div class="text-h6">{{ $t('LABEL_EXECUTE_QUERY') }}</div>
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
          <q-table class="q-mt-lg" :rows="data.Data" :loading="isLoading" />
        </q-tab-panel>

        <q-tab-panel v-if="data.Data" name="json">
          <vue-json-pretty :data="data.Data" />
        </q-tab-panel>

        <q-tab-panel name="meta">
          <q-list>
            <q-item>
              <q-item-section>
                <q-item-label caption>{{$t('LABEL_EXECUTION_TIME')}}</q-item-label>
                <q-item-label>{{data.ExecutionTimeInMilli}}</q-item-label>
              </q-item-section>
            </q-item>
            <q-item>
              <q-item-section>
                <q-item-label caption>{{$t('LABEL_EXECUTED_ON')}}</q-item-label>
                <q-item-label>{{data.ExecutedOn}}</q-item-label>
              </q-item-section>
            </q-item>
          </q-list>
        </q-tab-panel>
      </q-tab-panels>
    </q-card-section>
  </q-card>
</template>

<style scoped></style>
