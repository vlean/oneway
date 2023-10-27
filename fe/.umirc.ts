import { defineConfig } from '@umijs/max';

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
  request: {
    dataField: 'data',
    errorThrower: (res: any) => {
      if (res.code != 0) {
        const error: any = new Error(res.msg);
        error.name = 'BizError';
        throw error;
      }
    },
    errorHandler: (error: any, opts: any) => {
      if (opts?.skipErrorHandler) throw error;
      // 我们的 errorThrower 抛出的错误。
      if (error.name === 'BizError') {
        message.error(error.errorMessage);
      }
    }
  },
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

