import {
  AlipayCircleOutlined,
  LockOutlined,
  MobileOutlined,
  TaobaoCircleOutlined,
  UserOutlined,
  WeiboCircleOutlined,
} from '@ant-design/icons';
import {
  LoginForm,
  LoginFormPage,
  ProConfigProvider,
  ProFormCaptcha,
  ProFormCheckbox,
  ProFormText,
  setAlpha,
} from '@ant-design/pro-components';
import { Space, Tabs, message, theme } from 'antd';
import type { CSSProperties } from 'react';
import { useEffect, useState } from 'react';
import styles from './Login.less';
import { Link, useLocation } from '@umijs/max';
import { useSearchParams } from '@umijs/max';
type LoginType = 'phone' | 'account';

export default () => {
  const { token } = theme.useToken();
  const [loginType, setLoginType] = useState<LoginType>('phone');

  const iconStyles: CSSProperties = {
    marginInlineStart: '16px',
    color: setAlpha(token.colorTextBase, 0.2),
    fontSize: '24px',
    verticalAlign: 'middle',
    cursor: 'pointer',
  };
  const loc = useLocation();
  let [searchParams, setSearchParams] = useSearchParams();
  useEffect(() => {
    console.log("location", loc);
    console.log("locationx", searchParams.get("from"))
  }, []);

  return (
    <ProConfigProvider dark>
    <div
    style={{
      backgroundColor: 'white',
      height: '100vh',
    }}
  >
    <LoginFormPage
      logo="https://github.githubassets.com/images/modules/logos_page/Octocat.png"
      backgroundImageUrl="https://mdn.alipayobjects.com/huamei_gcee1x/afts/img/A*y0ZTS6WLwvgAAAAAAAAAAAAADml6AQ/fmt.webp"
      title="OneWay"
      onFinish={async () => {
        console.log("finish");
      }}
      containerStyle={{
        backdropFilter: 'blur(4px)',
      }}
      subTitle="All in one"
      activityConfig={
        {style: 
          {

          }}
      }
    >
    </LoginFormPage>
    </div>
    </ProConfigProvider>
  );
};