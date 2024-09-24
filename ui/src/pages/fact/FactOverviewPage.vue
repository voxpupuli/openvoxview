<script setup lang="ts">
import Backend from 'src/client/backend';
import {  onMounted, ref, watch } from 'vue';

interface IFactHash {
  [details: string]: string[]
}

const facts = ref<string[]>([]);
const isLoading = ref(true);
const needle = ref<string|null>(null);

const filteredFacts = ref<string[]>([]);
const groupedFilteredFacts = ref<IFactHash>({});

function loadFacts() {
  Backend.getFactNames().then(result => {
    if (result.status === 200) {
      facts.value = result.data.Data;
      refreshFilteredFacts();
    }
  }).finally(() => {
    isLoading.value = false;
  })
}

function refreshFilteredFacts() {
  let result = {} as IFactHash;

  if (needle.value) {
    filteredFacts.value = facts.value.filter(s => s.includes(needle.value as string));
  } else {
    filteredFacts.value = facts.value;
  }

  filteredFacts.value.forEach((fact) => {
    const letter = fact.charAt(0).toUpperCase();

    if (result[letter] === undefined) {
      result[letter] = [];
    }

    result[letter].push(fact)
  });

  groupedFilteredFacts.value = result;
}

watch(needle, async () => {
  refreshFilteredFacts();
})

onMounted(() => {
  loadFacts();
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
