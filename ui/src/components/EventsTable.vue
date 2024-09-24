<script setup lang="ts">
import { PuppetEvent } from 'src/puppet/models/puppet-event';
import { QTableColumn } from 'quasar';
import { useI18n } from 'vue-i18n';

const {t} = useI18n();
const events = defineModel('events', {required: true, type: Array<PuppetEvent>})
const flat = defineModel('flat', {type: Boolean, default: false})

const columns: QTableColumn[]  = [
  {
    name: 'resource',
    field: 'resource_type',
    label: t('LABEL_RESOURCE'),
    align: 'left',
    format: (val, row) => `${val}[${row['resource_title']}]`
  },
  {
    name: 'status',
    field: 'status',
    label: t('LABEL_STATUS'),
    align: 'left',
  },
  {
    name: 'changed_from',
    field: 'changed_from',
    label: t('LABEL_CHANGED_FROM'),
    align: 'left',
  },
  {
    name: 'changed_to',
    field: 'changed_to',
    label: t('LABEL_CHANGED_TO'),
    align: 'left',
  },
];
</script>

<template>
<q-table :columns="columns" :rows="events" :flat="flat"/>
</template>

<style scoped>

</style>
