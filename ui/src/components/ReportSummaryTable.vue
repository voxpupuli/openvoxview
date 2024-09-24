<script setup lang="ts">
import { QTableColumn } from 'quasar';
import { PuppetReport } from 'src/puppet/models/puppet-report';
import { useI18n } from 'vue-i18n';
import { emptyPagination } from 'src/helper/objects';
import { formatTimestamp } from 'src/helper/functions';

const {t} = useI18n();
const reports = defineModel('reports', {type: Array<PuppetReport>, required: true})
const flat = defineModel('flat', {type: Boolean, default: false})

const columns: QTableColumn[]  = [
  {
    name: 'certname',
    field: 'certname',
    label: t('LABEL_CERTNAME'),
    align: 'left',
  },
  {
    name: 'version',
    field: 'configuration_version',
    label: t('LABEL_CONFIGURATION_VERSION'),
    align: 'left',
  },
  {
    name: 'start_time',
    field: 'start_time',
    label: t('LABEL_START_TIME'),
    align: 'left',
    format: val => formatTimestamp(val, true),
  },
  {
    name: 'end_time',
    field: 'end_time',
    label: t('LABEL_END_TIME'),
    align: 'left',
    format: val => formatTimestamp(val, true),
  },
  {
    name: 'duration',
    field: 'durationFormatted',
    label: t('LABEL_DURATION'),
    align: 'left',
  },
];

</script>

<template>
  <q-table :columns="columns" :rows="reports" :flat="flat" :pagination="emptyPagination" hide-pagination/>
</template>

<style scoped>

</style>
