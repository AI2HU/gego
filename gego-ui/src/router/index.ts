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
      path: '/set-password',
      name: 'set-password',
      component: () => import('@/views/SetPasswordView.vue'),
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
          path: 'admin/words',
          name: 'words',
          component: () => import('@/views/WordsView.vue'),
          meta: { permissions: ['words'] },
          beforeEnter: [permissionGuard],
        },
        {
          path: 'admin/exclusion-words',
          redirect: '/admin/words',
        },
        {
          path: 'admin/logs',
          name: 'logs',
          component: () => import('@/views/LogsView.vue'),
          meta: { permissions: ['logs'] },
          beforeEnter: [permissionGuard],
        },
        {
          path: 'admin/users',
          name: 'users',
          component: () => import('@/views/UsersView.vue'),
          meta: { permissions: ['users'] },
          beforeEnter: [permissionGuard],
        },
        {
          path: 'admin/configuration',
          name: 'configuration',
          component: () => import('@/views/ConfigurationView.vue'),
          meta: { permissions: ['configuration'] },
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
