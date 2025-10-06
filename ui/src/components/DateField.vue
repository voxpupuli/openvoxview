<script setup lang="ts">
import moment from 'moment';
import { type PropType } from 'vue';

const value = defineModel('value', { type: Object as PropType<string | Date>, required: true })

</script>

<template>
<div v-if="value" class="text-subtitle1">{{ moment(value).format('DD. MMM. YYYY - HH:mm:ss') }}</div>
</template>

<style scoped>

</style>
