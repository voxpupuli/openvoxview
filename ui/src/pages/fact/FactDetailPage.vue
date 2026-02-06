<script setup lang="ts">
import { useRoute, useRouter } from 'vue-router';
import { computed, onMounted, ref, watch } from 'vue';
import { PuppetFact, type ApiPuppetFact } from 'src/puppet/models';
import Backend from 'src/client/backend';
import { useI18n } from 'vue-i18n';
import { type QTableColumn } from 'quasar';
import JsonViewDialog from 'components/JsonViewDialog.vue';
import { useSettingsStore } from 'stores/settings';
import PqlQuery, { PqlEntity } from 'src/puppet/query-builder';

const route = useRoute();
const router = useRouter();
const facts = ref([] as PuppetFact[]);
const needle = ref<string | null>(null);
const { t } = useI18n();
const showJsonView = ref(false);
const selectedFact = ref<PuppetFact | null>();
const settings = useSettingsStore();

const fact = computed(() => {
  return route.params.fact as string;
})

const columns: QTableColumn<PuppetFact>[] = [
  {
    name: 'certname',
    label: t('LABEL_CERTNAME'),
    field: 'certname',
    sortable: true,
    align: 'left',
  },
  {
    name: 'value',
    label: t('LABEL_VALUE'),
    field: 'value',
    sortable: true,
  },
];

function loadFacts() {
  const queryBuilder = new PqlQuery(PqlEntity.Facts);
  queryBuilder.filter().and().equal('name', fact.value);

  if (settings.hasEnvironment()) {
    queryBuilder.filter().and().equal('environment', settings.environment);
  }

  void Backend.getRawQueryResult<ApiPuppetFact[]>(queryBuilder.build()).then(
    (result) => {
      if (result.status === 200) {
        facts.value = result.data.Data.Data.map((s) => PuppetFact.fromApi(s));
      }
    }
  );
}

const filteredFacts = computed(() => {
  if (needle.value == null) return facts.value;
  return facts.value.filter(
    (s) =>
      s.value.toString().includes(needle.value ?? '') ||
      s.certname.toLowerCase().includes(needle.value!.toLowerCase())
  );
});

function jumpToNode(event: unknown, row: PuppetFact) {
  void router.push({ name: 'NodeDetail', params: { node: row.certname } });
}

function showJson(fact: PuppetFact) {
  selectedFact.value = fact;
  showJsonView.value = true;
}

onMounted(() => {
  watch(
    () => settings.environment,
    () => {
      loadFacts();
    },
    { immediate: true },
  );
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
    >
      <template v-slot:body="props">
        <q-tr :props="props">
          <q-td v-for="col in props.cols" :key="col.name" :props="props">
            <div v-if="col.name == 'value' && props.row.isJson">
              <q-btn
                color="primary"
                icon="visibility"
                @click="showJson(props.row)"
              />
            </div>
            <div v-else>{{ col.value }}</div>
          </q-td>
        </q-tr>
      </template>
    </q-table>
    <JsonViewDialog v-if="selectedFact" v-model:show="showJsonView" :model-value="selectedFact.value" :label="selectedFact?.name"/>
  </q-page>
</template>

<style scoped></style>
