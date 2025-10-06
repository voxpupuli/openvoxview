<script setup lang="ts">
import ReportSummaryTable from 'components/ReportSummaryTable.vue';
import PqlQuery, { PqlEntity } from 'src/puppet/query-builder';
import { computed, onMounted, ref } from 'vue';
import { useRoute } from 'vue-router';
import Backend from 'src/client/backend';
import { type ApiPuppetReport, PuppetReport } from 'src/puppet/models/puppet-report';
import { type ApiPuppetEvent, PuppetEvent } from 'src/puppet/models/puppet-event';
import EventsTable from 'components/EventsTable.vue';
import ReportLogsTable from 'components/ReportLogsTable.vue';
import MetricsTable from 'components/MetricsTable.vue';

const route = useRoute();
const report = ref<PuppetReport>();
const events = ref<PuppetEvent[]>();

const report_hash = computed(() => {
  return route.params.report_hash;
});
const certname = computed(() => {
  return route.params.certname;
});

function loadReport() {
  const query = new PqlQuery(PqlEntity.Reports);

  query
    .filter()
    .and()
    .equal('certname', certname.value)
    .and()
    .equal('hash', report_hash.value);

  void Backend.getQueryResult<ApiPuppetReport[]>(query).then((result) => {
    if (result.status === 200) {
      report.value = PuppetReport.fromApi(result.data.Data.Data[0]!);
    }
  });
}

function loadEvents() {
  const eventQuery = new PqlQuery(PqlEntity.Events);

  eventQuery
    .filter()
    .and()
    .equal('report', report_hash.value)
    .and()
    .equal('certname', certname.value);

  void Backend.getQueryResult<ApiPuppetEvent[]>(eventQuery).then((result) => {
    if (result.status === 200) {
      events.value = result.data.Data.Data.map((s) => PuppetEvent.fromApi(s));
    }
  });
}

onMounted(() => {
  loadReport();
  loadEvents();
});
</script>

<template>
  <q-page padding>
    <q-card v-if="report">
      <q-card-section class="bg-primary text-white text-h6">
        {{ $t('LABEL_SUMMARY') }}
      </q-card-section>
      <q-card-section class="q-pa-none">
        <ReportSummaryTable flat :reports="[report]" />
      </q-card-section>
    </q-card>

    <q-card class="q-mt-lg" v-if="report">
      <q-card-section class="bg-primary text-white text-h6">
        {{ $t('LABEL_LOG', 2) }}
      </q-card-section>
      <q-card-section class="q-pa-none">
        <ReportLogsTable :logs="report.logsMapped" flat />
      </q-card-section>
    </q-card>

    <q-card v-if="events" class="q-mt-lg">
      <q-card-section class="bg-primary text-white text-h6">
        {{ $t('LABEL_EVENT', 2) }}
      </q-card-section>
      <q-card-section class="q-pa-none">
        <EventsTable flat :events="events" />
      </q-card-section>
    </q-card>

    <q-card v-if="report" class="q-mt-lg">
      <q-card-section class="bg-primary text-white text-h6">
        {{ $t('LABEL_METRIC', 2) }}
      </q-card-section>
      <q-card-section>
        <MetricsTable flat :metrics="report.metrics.data" />
      </q-card-section>
    </q-card>
  </q-page>
</template>

<style scoped></style>
