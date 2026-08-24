<script setup lang="ts">
import { useRoute, useRouter } from 'vue-router'
const route = useRoute()
const router = useRouter()
</script>

<template>
  <div class="radar-bg min-h-screen w-full relative overflow-hidden">
    <div class="pointer-events-none absolute -right-24 -top-24 h-[420px] w-[420px] rounded-full scan-line opacity-70" />
    <header class="relative z-10 flex w-full items-center justify-between border-b border-line px-6 py-4">
      <button class="text-left" @click="router.push('/')">
        <div class="font-display text-xl font-semibold tracking-wide text-phosphor">GoAPIScanner</div>
        <div class="text-xs text-mute">上线前 API 缺陷扫描 · GMT+8</div>
      </button>
      <nav class="flex gap-4 text-sm">
        <router-link to="/" class="hover:text-phosphor" :class="route.path === '/' ? 'text-phosphor' : 'text-mute'">扫描控制台</router-link>
        <span class="text-line">|</span>
        <span class="text-mute">看板 / 报告随任务进入</span>
      </nav>
    </header>
    <main class="relative z-10 w-full px-6 py-6">
      <router-view />
    </main>
  </div>
</template>
