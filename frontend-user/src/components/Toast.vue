<script setup lang="ts">
defineProps<{ message: string }>()
const emit = defineEmits<{ close: [] }>()
</script>

<template>
  <div v-if="message" class="fixed right-6 top-6 z-50 flex max-w-sm items-start gap-3 rounded-xl border border-crit/40 bg-panel px-4 py-3 shadow-xl">
    <p class="text-sm text-white">{{ message }}</p>
    <button class="text-mute hover:text-white" @click="emit('close')">×</button>
  </div>
</template>
