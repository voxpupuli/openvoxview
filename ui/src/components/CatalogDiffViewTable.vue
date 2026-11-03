<script setup lang="ts">
import type { DiffItem, DiffStyle } from 'src/helper/diff';

const rows = defineModel<DiffItem<unknown>[]>('rows')

function getDiffChar(style: DiffStyle) {
  switch (style) {
    case 'empty':
      return '';
    case 'normal':
      return '';
    case 'removed':
      return '-';
    case 'added':
      return '+';
  }
}

function getDiffColor(style: DiffStyle) {
  const textColor = 'black';

  switch (style) {
    case 'empty':
      return `bg-red-3 text-${textColor}`;
    case 'normal':
      return '';
    case 'removed':
      return `bg-red-3 text-${textColor}`;
    case 'added':
      return `bg-green-3 text-${textColor}`;
  }
}

</script>

<template>
  <q-markup-table>
    <tbody>
      <tr
        v-for="(row, index) in rows"
        :key="index"
        :class="getDiffColor(row.style)"
      >
        <td>{{ getDiffChar(row.style) }}</td>
        <slot name="columns" :row="row"></slot>
      </tr>
    </tbody>
  </q-markup-table>
</template>

<style scoped></style>
