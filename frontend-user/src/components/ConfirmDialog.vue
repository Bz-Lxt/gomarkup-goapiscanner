<script setup lang="ts">
defineProps<{ open: boolean; title: string; body: string }>()
const emit = defineEmits<{ confirm: []; cancel: [] }>()
</script>

<template>
  <div v-if="open" class="fixed inset-0 z-40 flex items-center justify-center bg-black/60 px-4">
    <div class="w-full max-w-md rounded-2xl border border-line bg-panel p-6 shadow-2xl">
      <h3 class="font-display text-lg">{{ title }}</h3>
      <p class="mt-2 text-sm text-mute">{{ body }}</p>
      <div class="mt-6 flex justify-end gap-3">
        <button class="rounded-lg border border-line px-4 py-2 text-sm" @click="emit('cancel')">取消</button>
        <button class="rounded-lg bg-phosphor px-4 py-2 text-sm font-medium text-ink" @click="emit('confirm')">确认启动</button>
      </div>
    </div>
  </div>
</template>
