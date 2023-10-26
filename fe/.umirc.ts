import { defineConfig } from '@umijs/max';

export default defineConfig({
  antd: {
    dark: true,
    compact: true,
  },
  access: {},
  model: {},
  initialState: {},
  request: {},
  layout: {
    title: 'oneway',
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
    }
  ],
  npmClient: 'npm',
});

