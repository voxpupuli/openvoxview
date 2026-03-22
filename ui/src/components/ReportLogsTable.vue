<script setup lang="ts">
import { type PuppetReportLog } from 'src/puppet/models/puppet-report';
import { type QTableColumn } from 'quasar';
import { useI18n } from 'vue-i18n';
import { emptyPagination } from 'src/helper/objects';
import { ref, computed } from 'vue';

const { t } = useI18n();
const logs = defineModel('logs', {
  type: Array<PuppetReportLog>,
  required: true,
});
const flat = defineModel('flat', { type: Boolean, default: false });

const props = defineProps<{
  stripPathPrefix?: string | undefined;
}>();

const wrapMessages = ref(true);

function shortenLocation(file: string): string {
  if (!props.stripPathPrefix) return file;
  return file.replace(new RegExp(props.stripPathPrefix), '…');
}

const columns = computed<QTableColumn[]>(() => [
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
    field: (row) =>
      row.source?.startsWith('/')
        ? `${row.source}: ${row.message}`
        : row.message,
    label: t('LABEL_MESSAGE'),
    align: 'left',
  },
  {
    name: 'location',
    field: (row) => (row.file ? `${shortenLocation(row.file)}:${row.line}` : ''),
    label: t('LABEL_LOCATION_SHORT'),
    align: 'left',
  },
]);
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
    <template v-slot:header-cell-message="props">
      <q-th :props="props" class="row no-wrap items-center">
        <span>{{ props.col.label }}</span>
        <q-space />
        <q-btn
          :icon="wrapMessages ? 'wrap_text' : 'notes'"
          flat
          round
          dense
          size="sm"
          @click.stop="wrapMessages = !wrapMessages"
        >
          <q-tooltip>{{ wrapMessages ? 'Wrap: on' : 'Wrap: off' }}</q-tooltip>
        </q-btn>
      </q-th>
    </template>
    <template v-slot:body="props">
      <q-tr :props="props">
        <q-td v-for="col in props.cols" :key="col.name" :props="props">
          <div v-if="col.name == 'level'">
            <q-chip :color="props.row.color" class="level-chip">{{ col.value }}</q-chip>
          </div>
          <div v-else-if="col.name == 'message'" class="message-text" :class="wrapMessages ? 'wrap' : 'nowrap'">{{ col.value }}</div>
          <div v-else-if="col.name == 'location'" class="location-text">{{ col.value }}</div>
          <div v-else>{{ col.value }}</div>
        </q-td>
      </q-tr>
    </template>
  </q-table>
</template>

<style scoped>
.level-chip {
  width: 4.5rem;
}
.level-chip :deep(.q-chip__content) {
  justify-content: center;
}
.message-text, .location-text {
  font-family: Consolas, Menlo, Courier, monospace;
}
.message-text.wrap {
  white-space: pre-wrap;
}
.message-text.nowrap {
  white-space: pre;
}
</style>
