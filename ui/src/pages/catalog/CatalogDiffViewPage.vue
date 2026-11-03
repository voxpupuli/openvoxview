<script setup lang="ts">
import { useRoute } from 'vue-router';
import type {
  CatalogEdgeData,
  CatalogResourceData,
} from 'src/puppet/models/puppet-catalog';
import {
  type ApiPuppetCatalog,
  PuppetCatalog,
} from 'src/puppet/models/puppet-catalog';
import { onMounted, ref } from 'vue';
import Backend from 'src/client/backend';
import PqlQuery, { PqlEntity } from 'src/puppet/query-builder';
import { useI18n } from 'vue-i18n';
import CatalogDetailCard from 'components/CatalogDetailCard.vue';
import type { DiffItem } from 'src/helper/diff';
import { diffObjectsWithoutContext } from 'src/helper/diff';
import CatalogDiffViewTable from 'components/CatalogDiffViewTable.vue';

const { t } = useI18n();
const route = useRoute();
const source = route.params.source as string;
const against = route.params.against as string;
const tab = ref('resources');

const catalogA = ref<PuppetCatalog>();
const catalogB = ref<PuppetCatalog>();

const resourcesA = ref<DiffItem<CatalogResourceData>[]>([]);
const resourcesB = ref<DiffItem<CatalogResourceData>[]>([]);

const edgesA = ref<DiffItem<CatalogEdgeData>[]>([]);
const edgesB = ref<DiffItem<CatalogEdgeData>[]>([]);

const isLoading = ref(true);

const keyOfResources = (x: CatalogResourceData) => `${x.type}[${x.title}]`;
const keyOfEdges = (x: CatalogEdgeData) =>
  `${x.source_title}[${x.source_type}]`;

type result<T> = {
  resultA: DiffItem<T>[];
  resultB: DiffItem<T>[];
};

function buildDiff<T>(
  sourceA: T[],
  sourceB: T[],
  keyOfFunction: (item: T) => string,
): result<T> {
  const rows = diffObjectsWithoutContext<T>(sourceA, sourceB, keyOfFunction);

  const resultA: DiffItem<T>[] = [];
  const resultB: DiffItem<T>[] = [];

  rows.forEach((row) => {
    switch (row.type) {
      case 'same':
        resultA.push({
          item: row.oldItem,
          style: 'normal',
        });
        resultB.push({
          item: row.newItem,
          style: 'normal',
        });
        break;
      case 'add':
        resultA.push({
          item: row.oldItem,
          style: 'empty',
        });
        resultB.push({
          item: row.newItem,
          style: 'added',
        });
        break;
      case 'del':
        resultA.push({
          item: row.oldItem,
          style: 'removed',
        });
        resultB.push({
          item: row.newItem,
          style: 'empty',
        });
        break;
    }
  });

  return {
    resultA: resultA,
    resultB: resultB,
  };
}

async function loadCatalog(certname: string): Promise<PuppetCatalog> {
  const query = new PqlQuery(PqlEntity.Catalogs);
  query.filter().and().equal('certname', certname);

  const result = await Backend.getQueryResult<ApiPuppetCatalog[]>(query);
  return PuppetCatalog.fromApi(result.data.Data.Data[0]!);
}

async function loadCatalogs() {
  isLoading.value = true;

  catalogA.value = await loadCatalog(source);
  catalogB.value = await loadCatalog(against);

  const resultResources = buildDiff<CatalogResourceData>(
    catalogA.value.resources.data ?? [],
    catalogB.value.resources.data ?? [],
    keyOfResources,
  );
  resourcesA.value = resultResources.resultA;
  resourcesB.value = resultResources.resultB;

  const resultEdges = buildDiff<CatalogEdgeData>(
    catalogA.value.edges.data ?? [],
    catalogB.value.edges.data ?? [],
    keyOfEdges,
  );
  edgesA.value = resultEdges.resultA;
  edgesB.value = resultEdges.resultB;

  isLoading.value = false;
}

onMounted(() => loadCatalogs());
</script>

<template>
  <q-page padding>
    <div v-if="!isLoading && catalogA != null && catalogB != null">
      <div class="row">
        <div class="col q-pr-sm">
          <CatalogDetailCard :title="t('LABEL_SOURCE')" :catalog="catalogA" />
        </div>
        <div class="col q-pl-sm">
          <CatalogDetailCard :title="t('LABEL_AGAINST')" :catalog="catalogB" />
        </div>
      </div>

      <q-tabs
        v-model="tab"
        class="text-grey q-mt-lg bg-primary"
        active-color="white"
        indicator-color="white"
      >
        <q-tab name="resources" :label="t('LABEL_RESOURCE', 2)" />
        <q-tab name="edges" :label="t('LABEL_EDGE', 2)" />
      </q-tabs>

      <q-separator />

      <q-tab-panels v-model="tab" animated class="q-mt-sm">
        <q-tab-panel name="resources" class="q-pa-none">
          <div class="row">
            <div class="col q-pr-sm">
              <CatalogDiffViewTable v-model:rows="resourcesA">
                <template #columns="{ row }">
                  <td v-if="row.item" class="line">
                    {{ (row.item as CatalogResourceData).type }}[{{
                      (row.item as CatalogResourceData).title
                    }}]
                  </td>
                  <td v-else></td>
                </template>
              </CatalogDiffViewTable>
            </div>
            <div class="col q-pl-sm">
              <CatalogDiffViewTable v-model:rows="resourcesB">
                <template #columns="{ row }">
                  <td v-if="row.item" class="line">
                    {{ (row.item as CatalogResourceData).type }}[{{
                      (row.item as CatalogResourceData).title
                    }}]
                  </td>
                  <td v-else></td>
                </template>
              </CatalogDiffViewTable>
            </div>
          </div>
        </q-tab-panel>

        <q-tab-panel name="edges" class="q-pa-none">
          <div class="row">
            <div class="col q-pr-sm">
              <CatalogDiffViewTable v-model:rows="edgesA">
                <template #columns="{ row }">
                  <template v-if="row.item">
                    <td class="line">
                      {{ (row.item as CatalogEdgeData).source_type }}[{{
                        (row.item as CatalogEdgeData).source_title
                      }}]
                    </td>
                    <td class="line">
                      {{ (row.item as CatalogEdgeData).relationship }}
                    </td>
                    <td class="line">
                      {{ (row.item as CatalogEdgeData).target_type }}[{{
                        (row.item as CatalogEdgeData).target_title
                      }}]
                    </td>
                  </template>
                  <template v-else>
                    <td colspan="3"></td>
                  </template>
                </template>
              </CatalogDiffViewTable>
            </div>
            <div class="col q-pl-sm">
              <CatalogDiffViewTable v-model:rows="edgesB">
                <template #columns="{ row }">
                  <template v-if="row.item">
                    <td class="line">
                      {{ (row.item as CatalogEdgeData).source_type }}[{{
                        (row.item as CatalogEdgeData).source_title
                      }}]
                    </td>
                    <td class="line">
                      {{ (row.item as CatalogEdgeData).relationship }}
                    </td>
                    <td class="line">
                      {{ (row.item as CatalogEdgeData).target_type }}[{{
                        (row.item as CatalogEdgeData).target_title
                      }}]
                    </td>
                  </template>
                  <template v-else>
                    <td colspan="3"></td>
                  </template>
                </template>
              </CatalogDiffViewTable>
            </div>
          </div>
        </q-tab-panel>
      </q-tab-panels>
    </div>
    <q-inner-loading :showing="isLoading" />
  </q-page>
</template>

<style scoped>
td.line {
  font-family: monospace;
}
</style>
