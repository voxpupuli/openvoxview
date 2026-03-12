<script setup lang="ts">
import StatusButton from 'components/StatusButton.vue';
import { type PropType } from 'vue';
import { type PuppetReport } from 'src/puppet/models/puppet-report';
import EventCountBlock from 'components/EventCountBlock.vue';

const report = defineModel('report', {
  type: Object as PropType<PuppetReport>,
  required: true,
});
</script>

<template>
  <div class="row">
    <StatusButton
      class="col-auto q-mr-sm"
      :status="report.status"
      :hash="report.hash"
      :certname="report.certname"
    />
    <div
      class="col event bg-grey-7 rounded-borders text-center content-center"
    >
      {{ report.getMetricsValue('resources', 'total') }}
      <q-tooltip>{{ $t('LABEL_RESOURCE', 2) }}</q-tooltip>
    </div>
  </div>
  <EventCountBlock class="q-mt-sm" :event_count="report.getEventCounts()" />
</template>

<style scoped>
.event {
  min-width: 32px;
}
</style>
