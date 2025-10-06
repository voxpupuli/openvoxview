<script setup lang="ts">
import { type PuppetReportLog } from 'src/puppet/models/puppet-report';
import { type QTableColumn } from 'quasar';
import { useI18n } from 'vue-i18n';
import { emptyPagination } from 'src/helper/objects';

const { t } = useI18n();
const logs = defineModel('logs', {
  type: Array<PuppetReportLog>,
  required: true,
});
const flat = defineModel('flat', { type: Boolean, default: false });

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
    field: (row) => (row.location ? ` ${row.location}:${row.line}` : ''),
    label: t('LABEL_LOCATION_SHORT'),
    align: 'left',
  },
];
</script>

<template>
  <q-table
    :columns="columns"
    :rows="logs"
    :flat="flat"
    :pagination="emptyPagination"
    hide-pagination
    dense
  >
    <template v-slot:body="props">
      <q-tr :props="props">
        <q-td v-for="col in props.cols" :key="col.name" :props="props">
          <div v-if="col.name == 'level'">
            <q-chip :color="props.row.color">{{ col.value }}</q-chip>
          </div>
          <div v-else>{{ col.value }}</div>
        </q-td>
      </q-tr>
    </template>
  </q-table>
</template>

<style scoped></style>
