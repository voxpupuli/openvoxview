<script setup lang="ts">
import VueJsonPretty from 'vue-json-pretty';
import 'vue-json-pretty/lib/styles.css';
import { useRoute, useRouter } from 'vue-router';
import { computed, onMounted, ref } from 'vue';
import { PuppetFact, ApiPuppetFact } from 'src/puppet/models';
import Backend from 'src/client/backend';
import { useI18n } from 'vue-i18n';
import { QTableColumn, useQuasar } from 'quasar';

const route = useRoute();
const router = useRouter();
const fact = route.params.fact;
const facts = ref([] as PuppetFact[]);
const needle = ref<string | null>(null);
const { t } = useI18n();
const showJsonView = ref(false);
const selectedFact = ref<PuppetFact | null>();
const q = useQuasar();

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
  Backend.getRawQueryResult<ApiPuppetFact[]>(`facts { name = '${fact}'}`).then(
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
      s.value.toString().includes(needle.value) ||
      s.certname.toLowerCase().includes(needle.value!.toLowerCase())
  );
});

function jumpToNode(event: unknown, row: PuppetFact) {
  router.push({ name: 'NodeDetail', params: { node: row.certname } });
}

function showJson(fact: PuppetFact) {
  selectedFact.value = fact;
  showJsonView.value = true;
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
  </q-page>
  <q-dialog v-model="showJsonView" v-if="selectedFact">
    <q-card>
      <q-card-section class="bg-primary text-white text-h6">
        {{ selectedFact.name }}
      </q-card-section>
      <q-card-section style="max-height: 50vh" class="scroll">
        <vue-json-pretty
          :data="selectedFact.value"
          :theme="q.dark.isActive ? 'dark' : 'light'"
        />
      </q-card-section>
      <q-card-section>
        <q-card-actions>
          <q-btn
            color="negative"
            class="full-width"
            :label="t('BTN_CLOSE')"
            v-close-popup
          />
        </q-card-actions>
      </q-card-section>
    </q-card>
  </q-dialog>
</template>

<style scoped></style>
