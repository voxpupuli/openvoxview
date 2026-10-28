<script setup lang="ts">
import { type PuppetCatalog } from 'src/puppet/models/puppet-catalog';
import { useI18n } from 'vue-i18n';

const { t } = useI18n();
const catalog = defineModel<PuppetCatalog>('catalog', { required: true });
const title = defineModel('title', { type: String, required: true });
</script>

<template>
  <q-card>
    <q-card-section class="bg-primary text-white">
      {{ title }}
    </q-card-section>
    <q-card-section class="q-pa-none">
      <q-list>
        <q-item>
          <q-item-section>
            <q-item-label caption>{{ t('LABEL_CERTNAME') }}</q-item-label>
            <q-item-label>{{ catalog.certname }}</q-item-label>
          </q-item-section>
        </q-item>
        <q-item>
          <q-item-section>
            <q-item-label caption>{{
              t('LABEL_CATALOG_VERSION')
            }}</q-item-label>
            <q-item-label>{{ catalog.version }}</q-item-label>
          </q-item-section>
        </q-item>
        <q-item>
          <q-item-section>
            <q-item-label caption>{{
              t('LABEL_PRODUCER_TIMESTAMP')
            }}</q-item-label>
            <q-item-label>{{ catalog.producer_timestamp }}</q-item-label>
          </q-item-section>
        </q-item>
      </q-list>
    </q-card-section>
  </q-card>
</template>

<style scoped></style>
