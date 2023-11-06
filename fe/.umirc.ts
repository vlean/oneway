import { defineConfig } from '@umijs/max';
import { message } from 'antd';

export default defineConfig({
  antd: {
    dark: true,
    compact: true,
  },
  access: {},
  model: {},
  initialState: {},
  proxy: {
    '/api': {
      'target': 'http://127.0.0.1:8080'
    }
  },
  request: {},
  layout: {
    title: 'oneway'
  },
  routes: [
    {
      path: '/',
      redirect: '/dashboard',
    },
    {
      name: "首页",
      path: "/dashboard",
      component: "./Dashboard",
      icon: 'DashboardOutlined'
    },
    {
      name: '代理',
      path: '/proxy',
      component: './Proxy',
      icon: 'DeploymentUnitOutlined'
    },
    {
      name: '设置',
      path: '/setting',
      component: "./Setting",
      icon: 'SettingOutlined'
    },
    {
      path: '/login', 
      component: "Login",
      hideInMenu: true, 
      layout: false,
    },
    { path: '/*', component: '@/pages/404.tsx', layout: false }
  ],
  npmClient: 'npm',
});

