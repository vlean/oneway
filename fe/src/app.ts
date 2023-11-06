// 运行时配置
import React from 'react';
import { UserOutlined } from '@ant-design/icons';
import { Avatar, Space, message } from 'antd';
import {useLocation} from 'umi';
import { RequestConfig } from '@umijs/max';
import { useModel } from '@umijs/max';
import { userinfo } from './services/controller';
// 全局初始化数据配置，用于 Layout 用户信息和权限初始化
// 更多信息见文档：https://umijs.org/docs/api/runtime-config#getinitialstate
export async function getInitialState(): Promise<{ name: string }> {
  return { name: localStorage.getItem("email") ?? "" };
}

export const layout = () => {
  const location  = useLocation();
  console.log(location);
  if (location.pathname == "/login") {
    
  }
  return {
    logo: 'https://img.alicdn.com/tfs/TB1YHEpwUT1gK0jSZFhXXaAtVXa-28-27.svg',
    menu: {
      locale: false,
    },
    logout: (state: any) => {
      localStorage.removeItem("email");
    },
    
  };
};

export const request  = () => {
  return {
    dataField: 'data',
    timeout: 10000,
    errorConfig: {
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
      },
    },
    requestInterceptors: [
        (url: any, options: any) => {
          // 鉴权
          // const email = localStorage.getItem("email") || ""
          // if (url != "/api/user" && email == "") {
          //   message.warning("跳转登录ing");
          //   window.location.href = '/login';
          //   return {}
          // }
          return {url, options}
        }
    ],
    responseInterceptors: [((response : any) => {
        response.data.success = (response.data?.code || 1) === 0;
        console.log("response", response);
        return response
    })]
  };
};
