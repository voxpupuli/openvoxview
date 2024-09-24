<script setup lang="ts">
import { PuppetEventCount } from 'src/puppet/models/puppet-event-count';

const eventCount = defineModel('event_count', {
  type: PuppetEventCount,
  required: true,
});
</script>

<template>
  <div class="row">
    <div class="col q-mr-sm event bg-negative rounded-borders text-center content-center">
      {{ eventCount.failures }}
      <q-tooltip>{{ $t('LABEL_FAILURE', 2) }}</q-tooltip>
    </div>
    <div class="col q-mr-sm event bg-positive rounded-borders text-center content-center">
      {{ eventCount.successes }}
      <q-tooltip>{{ $t('LABEL_SUCCESS', 2) }}</q-tooltip>
    </div>
    <div class="col q-mr-sm event bg-primary rounded-borders text-center content-center">
      {{ eventCount.skips }}
      <q-tooltip>{{ $t('LABEL_SKIP', 2) }}</q-tooltip>
    </div>
    <div class="col event bg-grey rounded-borders text-center content-center">
      {{ eventCount.noops }}
      <q-tooltip>{{ $t('LABEL_NOOP', 2) }}</q-tooltip>
    </div>
  </div>
</template>

<style scoped>
.event {
  min-width: 32px;
  min-height: 32px;
}
</style>
