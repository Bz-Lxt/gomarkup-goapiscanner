import { createRouter, createWebHistory } from 'vue-router'
import Console from './views/Console.vue'
import Monitor from './views/Monitor.vue'
import Report from './views/Report.vue'

export const router = createRouter({
  history: createWebHistory(),
  routes: [
    { path: '/', name: 'console', component: Console },
    { path: '/monitor/:id', name: 'monitor', component: Monitor },
    { path: '/report/:id', name: 'report', component: Report },
  ],
})
