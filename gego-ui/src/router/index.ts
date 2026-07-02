import { createRouter, createWebHistory } from 'vue-router'

import AppLayout from '@/layouts/AppLayout.vue'
import { authGuard } from '@/router/guards/auth-guard'
import { permissionGuard } from '@/router/guards/permission-guard'
import { upgradeGuard } from '@/router/guards/upgrade-guard'

const router = createRouter({
  history: createWebHistory(import.meta.env.BASE_URL),
  routes: [
    {
      path: '/upgrade',
      name: 'upgrade',
      component: () => import('@/views/MajorUpgradeView.vue'),
    },
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
          component: () => import('@/views/DashboardView.vue'),
          meta: { permissions: ['dashboard'] },
          beforeEnter: [permissionGuard],
        },
        {
          path: 'search',
          name: 'search',
          component: () => import('@/views/SearchView.vue'),
          meta: { permissions: ['search'] },
          beforeEnter: [permissionGuard],
        },
        {
          path: 'admin/models',
          name: 'models',
          component: () => import('@/views/ModelsView.vue'),
          meta: { permissions: ['models'] },
          beforeEnter: [permissionGuard],
        },
        {
          path: 'admin/prompts',
          name: 'prompts',
          component: () => import('@/views/PromptsView.vue'),
          meta: { permissions: ['prompts'] },
          beforeEnter: [permissionGuard],
        },
        {
          path: 'admin/scheduler',
          name: 'scheduler',
          component: () => import('@/views/SchedulerView.vue'),
          meta: { permissions: ['scheduler'] },
          beforeEnter: [permissionGuard],
        },
        {
          path: 'admin/exclusion-words',
          name: 'exclusionWords',
          component: () => import('@/views/ExclusionWordsView.vue'),
          meta: { permissions: ['exclusionWords'] },
          beforeEnter: [permissionGuard],
        },
        {
          path: 'admin/logs',
          name: 'logs',
          component: () => import('@/views/LogsView.vue'),
          meta: { permissions: ['logs'] },
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

router.beforeEach(upgradeGuard)
router.beforeEach(authGuard)

export default router
