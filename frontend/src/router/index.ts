import { createRouter, createWebHashHistory } from 'vue-router'
import MainLayout from '@/views/MainLayout.vue'

const router = createRouter({
  history: createWebHashHistory(),
  routes: [
    {
      path: '/',
      name: 'main',
      component: MainLayout,
    },
  ],
})

export default router
