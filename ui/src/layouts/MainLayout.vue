<template>
  <q-layout view="lHh Lpr lFf">
    <q-header elevated>
      <q-toolbar>
        <q-btn
          flat
          dense
          round
          icon="menu"
          aria-label="Menu"
          @click="toggleLeftDrawer"
        />

        <q-toolbar-title>
          <q-img height="48px" width="48px" src="logo.png"/>
          OpenVox View
        </q-toolbar-title>

        <EnvironmentSelector/>
        <q-toggle unchecked-icon="light_mode" checked-icon="dark_mode" v-model="settings.darkMode" color="positive"/>
        <div>{{version}}</div>
      </q-toolbar>
    </q-header>

    <q-drawer
      v-model="leftDrawerOpen"
      show-if-above
      bordered
    >
      <q-list>
        <q-item clickable :to="{name: 'Dashboard'}">
          <q-item-section avatar>
            <q-icon name="dashboard"/>
          </q-item-section>

          <q-item-section>
            <q-item-label>{{$t('MENU_DASHBOARD')}}</q-item-label>
          </q-item-section>
        </q-item>

        <q-item clickable :to="{name: 'FactOverview'}">
          <q-item-section avatar>
            <q-icon name="list"/>
          </q-item-section>

          <q-item-section>
            <q-item-label>{{$t('MENU_FACTS')}}</q-item-label>
          </q-item-section>
        </q-item>
        <q-item clickable :to="{name: 'NodeOverview'}">
          <q-item-section avatar>
            <q-icon name="computer"/>
          </q-item-section>

          <q-item-section>
            <q-item-label>{{$t('MENU_NODES')}}</q-item-label>
          </q-item-section>
        </q-item>
        <q-item clickable :to="{name: 'ReportOverview'}">
          <q-item-section avatar>
            <q-icon name="analytics"/>
          </q-item-section>

          <q-item-section>
            <q-item-label>{{$t('MENU_REPORTS')}}</q-item-label>
          </q-item-section>
        </q-item>
        <q-expansion-item :label="$t('MENU_VIEWS')" v-if="predefinedViews.length > 0" :content-inset-level="0.5">
          <q-list>
            <q-item clickable :key="view.Name" v-for="view in predefinedViews" :to="{name: 'PredefinedViewResult', params: {viewName: view.Name}}">
              <q-item-section>
                <q-item-label>{{view.Name}}</q-item-label>
              </q-item-section>
            </q-item>
          </q-list>
        </q-expansion-item>
        <q-item clickable :to="{name: 'Query'}">
          <q-item-section avatar>
            <q-icon name="terminal"/>
          </q-item-section>

          <q-item-section>
            <q-item-label>{{$t('MENU_QUERY')}}</q-item-label>
          </q-item-section>
        </q-item>
      </q-list>
    </q-drawer>

    <q-page-container>
      <router-view />
    </q-page-container>
  </q-layout>
</template>

<script setup lang="ts">
import { onMounted, ref, watch } from 'vue';
import { useQuasar } from 'quasar';
import EnvironmentSelector from 'components/EnvironmentSelector.vue';
import { useSettingsStore } from 'stores/settings';
import Backend from 'src/client/backend';
import { PredefinedView } from 'src/puppet/models';

const leftDrawerOpen = ref(false);
const settings = useSettingsStore();
const version = ref('dirty')
const predefinedViews = ref<PredefinedView[]>([]);

function toggleLeftDrawer () {
  leftDrawerOpen.value = !leftDrawerOpen.value;
}

function loadVersion() {
  Backend.getVersion().then(result => {
    if (result.status === 200) {
      version.value = result.data.Data.Version;
    }
  })
}

function loadPredefinedViews() {
  Backend.getPredefinedViews().then(result => {
    if (result.status === 200) {
      predefinedViews.value = result.data.Data.map(s => PredefinedView.fromApi(s))
    }
  })
}

const q = useQuasar();

watch(settings, () => {
  q.dark.set(settings.darkMode);
})

onMounted(() => {
  q.dark.set(settings.darkMode);
  loadVersion();
  loadPredefinedViews();
})

</script>
