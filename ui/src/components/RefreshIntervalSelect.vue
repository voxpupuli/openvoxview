<script setup lang="ts">
import Backend from 'src/client/backend';
import { computed, onMounted, ref, watch } from 'vue';
import { useI18n } from 'vue-i18n';

const refreshInterval = ref<number>(300);
const emit = defineEmits(['refresh']);
const lastRefresh = ref<Date>(new Date());
const { t } = useI18n();

const options = computed(() => {
  const opts = [
    { label: t('LABEL_OFF'), value: 0 },
    { label: `30 ${t('SECOND', 2)}`, value: 30 },
    { label: `1 ${t('MINUTE', 2)}`, value: 60 },
    { label: `5 ${t('MINUTE', 2)}`, value: 300 },
    { label: `10 ${t('MINUTE', 2)}`, value: 600 },
    { label: `30 ${t('MINUTE', 2)}`, value: 1800 },
    { label: `1 ${t('HOUR')}`, value: 3600 },
  ];
  if (!opts.some((o) => o.value === refreshInterval.value)) {
    opts.unshift({
      label: `(${refreshInterval.value} ${t('SECOND', 2)})`,
      value: refreshInterval.value,
    });
  }
  return opts;
});

function loadDefaultInterval() {
  void Backend.getMeta().then((result) => {
    refreshInterval.value =
      result.data.Data.UiDefaultRefreshIntervalInSeconds ?? refreshInterval.value;
  });
}

watch(
  refreshInterval,
  (seconds, _previous, onCleanup) => {
    if (seconds <= 0) return;

    const interval = setInterval(() => {
      emit('refresh');
      lastRefresh.value = new Date();
    }, seconds * 1000);

    onCleanup(() => clearInterval(interval));
  },
  { immediate: true },
);

onMounted(() => {
  loadDefaultInterval();
});
</script>

<template>
  <q-select v-model="refreshInterval" :options="options" emit-value map-options>
    <template v-slot:prepend>
      <q-icon name="autorenew" />
    </template>
    <template v-slot:default>
      <q-tooltip anchor="bottom middle" self="top middle">
        {{
          refreshInterval === 0
            ? `${t('LABEL_AUTO_REFRESH_OFF')}`
            : `${t('LABEL_LAST_REFRESH')}: ${lastRefresh}`
        }}
      </q-tooltip>
    </template>
  </q-select>
</template>

<style scoped></style>
