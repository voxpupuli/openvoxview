<script setup lang="ts">
import { useRoute } from 'vue-router';
import { onMounted, ref } from 'vue';
import {
  type ApiPuppetCatalog,
  PuppetCatalog,
} from 'src/puppet/models/puppet-catalog';
import Backend from 'src/client/backend';
import PqlQuery, { PqlEntity } from 'src/puppet/query-builder';
import NodeLink from 'components/NodeLink.vue';

const route = useRoute();
const certname = route.params.certname!.toString();
const isLoading = ref(true);
const tab = ref('resources');
const catalog = ref<PuppetCatalog>();

function loadCatalog() {
  const query = new PqlQuery(PqlEntity.Catalogs);
  query.filter().and().equal('certname', certname);
  void Backend.getQueryResult<ApiPuppetCatalog[]>(query)
    .then((result) => {
      if (result.status === 200) {
        console.log(result.data.Data.Data);
        catalog.value = PuppetCatalog.fromApi(result.data.Data.Data[0]!);
      }
    })
    .finally(() => {
      isLoading.value = false;
    });
}

onMounted(() => {
  loadCatalog();
});
</script>

<template>
  <q-page padding>
    <div v-if="!isLoading" class="row">
      <div class="col q-pr-md">
        <q-card v-if="catalog">
          <q-card-section class="bg-primary text-white text-h6">
            {{ $t('LABEL_DETAIL', 2) }}
          </q-card-section>
          <q-card-section class="q-pa-none">
            <q-markup-table v-if="certname" flat>
              <tbody>
                <tr>
                  <td class="text-left text-bold">Certname</td>
                  <td class="text-left text-bold">
                    <NodeLink :certname="certname" />
                  </td>
                </tr>
                <tr>
                  <td class="text-left text-bold">Version</td>
                  <td class="text-left">
                    {{ catalog.version }}
                  </td>
                </tr>
                <tr>
                  <td class="text-left text-bold">Transaction UUID</td>
                  <td class="text-left">
                    {{ catalog.transaction_uuid }}
                  </td>
                </tr>
                <tr v-if="catalog.code_id">
                  <td class="text-left text-bold">Code ID</td>
                  <td class="text-left">
                    {{ catalog.code_id }}
                  </td>
                </tr>
              </tbody>
            </q-markup-table>
          </q-card-section>
        </q-card>
        <q-card class="q-mt-md">
          <q-card-section class="q-pa-none">
            <q-tab-panels v-model="tab" v-if="catalog">
              <q-tab-panel name="resources" class="q-pa-none">
                <q-markup-table v-if="certname" flat>
                  <thead>
                    <tr>
                      <th>{{ $t('LABEL_RESOURCE') }}</th>
                      <th>{{ $t('LABEL_LOCATION') }}</th>
                    </tr>
                  </thead>
                  <tbody>
                    <tr
                      v-for="resource in catalog.resources.data"
                      :key="resource.title"
                    >
                      <td class="text-left">
                        {{ resource.type }}[{{ resource.title }}]
                      </td>
                      <td class="text-left">
                        {{ resource.file }}:{{ resource.line }}
                      </td>
                    </tr>
                  </tbody>
                </q-markup-table>
              </q-tab-panel>
            </q-tab-panels>
          </q-card-section>
        </q-card>
      </div>
    </div>
  </q-page>
</template>

<style scoped></style>
