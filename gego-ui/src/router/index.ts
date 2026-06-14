import { createRouter, createWebHistory } from 'vue-router'

import AppLayout from '@/layouts/AppLayout.vue'
import { authGuard } from '@/router/guards/auth-guard'
import { permissionGuard } from '@/router/guards/permission-guard'

const router = createRouter({
  history: createWebHistory(import.meta.env.BASE_URL),
  routes: [
    {
      path: '/login',
      name: 'login',
      component: () => import('@/views/LoginView.vue'),
      meta: { guestOnly: true },
    },
    {
      path: '/',
      component: AppLayout,
      meta: { requiresAuth: true },
      children: [
        {
          path: '',
          name: 'dashboard',
          component: () => import('@/views/PlaceholderView.vue'),
          meta: { permissions: ['dashboard'] },
          beforeEnter: [permissionGuard],
        },
        {
          path: 'forbidden',
          name: 'forbidden',
          component: () => import('@/views/ForbiddenView.vue'),
        },
      ],
    },
  ],
})

router.beforeEach(authGuard)

export default router
