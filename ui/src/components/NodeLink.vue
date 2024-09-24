<script setup lang="ts">
const certname = defineModel('certname', {required: true, type: String})
const label = defineModel('label', {type: String})
</script>

<template>
  <router-link class="text-primary text-subtitle2" :to="`/node/${certname}`">{{ label ? label : certname }}</router-link>
</template>

<style scoped>
  a {
    text-decoration: none;
  }
</style>
