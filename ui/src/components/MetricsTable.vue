<script setup lang="ts">
import { type QTableColumn } from 'quasar';
import { useI18n } from 'vue-i18n';
import { type PuppetReportMetric } from 'src/puppet/models/puppet-report';
import { emptyPagination } from 'src/helper/objects';

const {t} = useI18n();
const metrics = defineModel('metrics', {required: true, type: Array<PuppetReportMetric>})
const flat = defineModel('flat', {type: Boolean, default: false})

const columns: QTableColumn[]  = [
  {
    name: 'category',
    field: 'category',
    label: t('LABEL_CATEGORY'),
    align: 'left',
  },
  {
    name: 'name',
    field: 'name',
    label: t('LABEL_NAME'),
    align: 'left',
  },
  {
    name: 'value',
    field: 'value',
    label: t('LABEL_VALUE'),
    align: 'left',
    format: (val => `${val.toFixed(2)}`)
  },
];
</script>

<template>
<q-table :columns="columns" :rows="metrics" :flat="flat" :pagination="emptyPagination" hide-pagination/>
</template>

<style scoped>

</style>
