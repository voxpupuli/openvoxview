<script setup lang="ts">
import Backend from 'src/client/backend';
import {  onMounted, ref, watch } from 'vue';
import { useSettingsStore } from 'stores/settings';
import { type ApiPuppetFact } from 'src/puppet/models';
import PqlQuery, { PqlEntity } from 'src/puppet/query-builder';

interface IFactHash {
  [details: string]: string[]
}

const facts = ref<string[]>([]);
const isLoading = ref(true);
const needle = ref<string|null>(null);
const settings = useSettingsStore();

const filteredFacts = ref<string[]>([]);
const groupedFilteredFacts = ref<IFactHash>({});

function loadFacts() {
  const queryBuilder = new PqlQuery(PqlEntity.Facts);
  queryBuilder.addProjectionField('name');
  queryBuilder.groupBy().add('name');

  if (settings.hasEnvironment()) {
    queryBuilder.filter().and().equal('environment', settings.environment);
  }

  void Backend.getRawQueryResult<ApiPuppetFact[]>(queryBuilder.build()).then(
    (result) => {
      if (result.status === 200) {
        facts.value = result.data.Data.Data.map((s) =>
          s.name,
        );
        refreshFilteredFacts();
      }
    },
  ).finally(() => {
    isLoading.value = false;
  });
}

function refreshFilteredFacts() {
  const result = {} as IFactHash;

  if (needle.value) {
    filteredFacts.value = facts.value.filter(s => s.includes(needle.value as string));
  } else {
    filteredFacts.value = facts.value;
  }

  filteredFacts.value.sort().forEach((fact) => {
    const letter = fact.charAt(0).toUpperCase();

    if (result[letter] === undefined) {
      result[letter] = [];
    }

    result[letter].push(fact)
  });

  groupedFilteredFacts.value = result;
}

watch(needle, () => {
  refreshFilteredFacts();
})

onMounted(() => {
  watch(
    () => settings.environment,
    () => {
      loadFacts();
    },
    { immediate: true },
  );
})

</script>

<template>
  <q-page padding>
    <div v-if="!isLoading">
      <q-input debounce="500" :label="$t('LABEL_SEARCH')" v-model="needle" clearable>
        <template v-slot:append>
          <q-badge outline color="primary">{{filteredFacts.length}}</q-badge>
        </template>
      </q-input>
      <div class="column" style="max-height: 100%">
        <q-card class="q-ma-md" v-for="(groupedFacts, letter) in groupedFilteredFacts" :key="letter">
          <q-card-section class="bg-primary text-white text-h6">
            {{ letter }}
          </q-card-section>
          <q-card-section>
            <ul>
              <li v-for="fact in groupedFacts" :key="fact">
                <a class="text-primary" :href="`#/fact/${fact}`">{{fact}}</a>
              </li>
            </ul>
          </q-card-section>
        </q-card>
      </div>
    </div>
    <q-inner-loading :showing="isLoading"/>
  </q-page>
</template>

<style scoped>

</style>
