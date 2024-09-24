<script setup lang="ts">
import EventCountBlock from 'components/EventCountBlock.vue';
import NodeLink from 'components/NodeLink.vue';
import { QTableColumn } from 'quasar';
import { useI18n } from 'vue-i18n';
import { PropType } from 'vue';
import { PuppetNodeWithEventCount } from 'src/puppet/models/puppet-node';
import { emptyPagination } from 'src/helper/objects';
import StatusButton from 'components/StatusButton.vue';

const { t } = useI18n();

const nodes = defineModel('nodes', {
  type: Array as PropType<PuppetNodeWithEventCount[]>,
  required: true,
});
const disablePagination = defineModel('disable_pagination', {
  type: Boolean,
  default: false,
});

const columns: QTableColumn[] = [
  {
    name: 'events',
    field: 'events',
    label: t('LABEL_EVENT', 2),
    align: 'left',
  },
  {
    name: 'certname',
    field: 'certname',
    label: t('LABEL_CERTNAME'),
    align: 'left',
  },
  {
    name: 'catalog',
    field: 'catalog_timestamp',
    label: t('LABEL_CATALOG'),
    align: 'left',
  },
  {
    name: 'report',
    field: 'report_timestamp',
    label: t('LABEL_REPORT'),
    align: 'left',
  },
];
</script>

<template>
  <q-table
    class="q-mt-lg"
    flat
    bordered
    :columns="columns"
    :rows="nodes"
    :pagination="disablePagination ? emptyPagination : {}"
    :hide-pagination="disablePagination"
  >
    <template v-slot:header="props">
      <q-tr :props="props">
        <q-th v-for="col in props.cols" :key="col.name" :props="props">
          {{ col.label }}
        </q-th>
      </q-tr>
    </template>

    <template v-slot:body="props">
      <q-tr :props="props">
        <q-td v-for="col in props.cols" :key="col.name" :props="props">
          <div v-if="col.name == 'events'">
            <StatusButton v-if="col.value.latest_report_status"
              :status="col.value.latest_report_status"
              :hash="col.value.latest_report_hash"
              :certname="col.value.certname"
            />
            <EventCountBlock :event_count="props.row.eventsMapped" />
          </div>
          <NodeLink v-else-if="col.name == 'certname'" :certname="col.value" />
          <div v-else>{{ col.value }}</div>
        </q-td>
      </q-tr>
    </template>
  </q-table>
</template>

<style scoped></style>
