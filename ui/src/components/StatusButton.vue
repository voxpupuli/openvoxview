<script setup lang="ts">

const certname = defineModel('certname', {required: true, type: String})
const hash = defineModel('hash', {required: true, type: String})
const node_status = defineModel('status', {required: true, type: String})

function getStatusColor(status: string) {
  switch (status.toLowerCase()) {
    case 'unchanged':
      return 'positive';
    case 'failed':
      return 'negative';
    case 'unreported':
      return 'accent';
    default:
      return 'primary';
  }
}
</script>

<template>
  <q-btn :to="`/report/${certname}/${hash}`" :color="getStatusColor(node_status)" :label="node_status"
    class="col q-mr-sm status-btn">
    <q-tooltip>{{ node_status }}</q-tooltip>
  </q-btn>
</template>

<style scoped>
.status-btn {
  max-width: 15em;
}
</style>
