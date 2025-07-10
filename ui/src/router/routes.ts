import { RouteRecordRaw } from 'vue-router';

const routes: RouteRecordRaw[] = [
  {
    path: '/',
    redirect: { name: 'Dashboard' },
    component: () => import('layouts/MainLayout.vue'),
    children: [
      {
        name: 'Dashboard',
        path: 'dashboard',
        component: () => import('pages/DashboardPage.vue'),
      },
      {
        name: 'Query',
        path: 'query',
        component: () => import('pages/QueryPage.vue'),
      },
      {
        name: 'FactOverview',
        path: 'facts',
        component: () => import('pages/fact/FactOverviewPage.vue'),
      },
      {
        name: 'FactDetail',
        path: 'fact/:fact',
        component: () => import('pages/fact/FactDetailPage.vue'),
      },
      {
        name: 'NodeOverview',
        path: 'nodes',
        component: () => import('pages/node/NodeOverviewPage.vue'),
      },
      {
        name: 'NodeDetail',
        path: 'node/:node',
        component: () => import('pages/node/NodeDetailPage.vue'),
      },
      {
        name: 'ReportOverview',
        path: 'reports',
        component: () => import('pages/report/ReportOverviewPage.vue'),
      },
      {
        name: 'ReportDetail',
        path: 'report/:certname/:report_hash',
        component: () => import('pages/report/ReportDetailPage.vue'),
      },
      {
        name: 'PredefinedViewResult',
        path: 'view/:viewName',
        component: () => import('pages/views/PredefinedViewResultPage.vue'),
      },
    ],
  },
];

export default routes;
