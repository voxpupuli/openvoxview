<script setup lang="ts">
import { PuppetReportLog } from 'src/puppet/models/puppet-report';
import { QTableColumn } from 'quasar';
import { useI18n } from 'vue-i18n';
import { emptyPagination } from 'src/helper/objects';

const {t} = useI18n();
const logs = defineModel('logs', {type: Array<PuppetReportLog>, required: true})
const flat = defineModel('flat', {type: Boolean, default: false})

const columns: QTableColumn[] = [
  {
    name: 'timestamp',
    field: 'time',
    label: t('LABEL_TIMESTAMP'),
    align: 'left',
  },
  {
    name: 'level',
    field: 'level',
    label: t('LABEL_LEVEL'),
    align: 'left',
  },
  {
    name: 'message',
    field: 'message',
    label: t('LABEL_MESSAGE'),
    align: 'left',
  },
  {
    name: 'location',
    field: row => row.location ? ` ${row.location}:${row.line}` : '',
    label: t('LABEL_LOCATION_SHORT'),
    align: 'left',
  },
];
</script>

<template>
<q-table :columns="columns" :rows="logs" :flat="flat" :pagination="emptyPagination" hide-pagination/>
</template>

<style scoped>

</style>
