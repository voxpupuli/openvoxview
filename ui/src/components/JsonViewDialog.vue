<script setup lang="ts">

import { useI18n } from 'vue-i18n';
import { JsonViewer } from 'vue3-json-viewer';
import { useQuasar } from 'quasar';
import 'vue3-json-viewer/dist/vue3-json-viewer.css';

const show = defineModel<boolean>('show', {default: false})
const value = defineModel<unknown>({required: true})
const label = defineModel<string>('label', {default: ''})

const {t} = useI18n();
const q = useQuasar();
</script>

<template>
  <q-dialog v-model="show" v-if="value">
    <q-card>
      <q-card-section class="bg-primary text-white text-h6">
        {{ label }}
      </q-card-section>
      <q-card-section style="max-height: 50vh" class="scroll">
        <JsonViewer
          :value="value"
          expanded
          :expand-depth="-1"
          preview-mode
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
