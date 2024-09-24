<script setup lang="ts">
import { useRoute, useRouter } from 'vue-router';
import { computed, onMounted, ref } from 'vue';
import { PuppetFact } from 'src/puppet/models';
import Backend from 'src/client/backend';
import { useI18n } from 'vue-i18n';
import { QTableColumn } from 'quasar';

const route = useRoute();
const router = useRouter();
const fact = route.params.fact;
const facts = ref([] as PuppetFact[]);
const needle = ref<string | null>(null);
const { t } = useI18n();

const columns: QTableColumn<PuppetFact>[] = [
  {
    name: 'certname',
    label: t('LABEL_CERTNAME'),
    field: 'certname',
    sortable: true,
    align: 'left'
  },
  {
    name: 'value',
    label: t('LABEL_VALUE'),
    field: 'value',
    sortable: true,
  },
];

function loadFacts() {
  Backend.getRawQueryResult<PuppetFact[]>(`facts { name = '${fact}'}`).then((result) => {
    if (result.status === 200) {
      facts.value = result.data.Data.Data;
    }
  });
}

const filteredFacts = computed(() => {
  if (needle.value == null) return facts.value;
  return facts.value.filter(
    (s) =>
      s.value.toString().includes(needle.value) ||
      s.certname.toLowerCase().includes(needle.value!.toLowerCase())
  );
});

function jumpToNode(event: unknown, row: PuppetFact) {
  router.push({name: 'NodeDetail', params: {node: row.certname}})
}

onMounted(() => {
  loadFacts();
});
</script>

<template>
  <q-page padding>
    <div class="text-h3">{{ fact }}</div>
    <q-input v-model="needle" :label="$t('LABEL_SEARCH')" />
    <q-table
      class="q-mt-md"
      :columns="columns"
      :rows="filteredFacts"
      :pagination="{ rowsPerPage: 0 }"
      row-key="certname"
      @row-click="jumpToNode"
    />
  </q-page>
</template>

<style scoped></style>
