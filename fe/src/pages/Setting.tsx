import React, { useEffect, useRef, useState } from 'react';
import styles from './Setting.less';
import { systemConfig, systemConfigUpdate } from '@/services/controller';
import { PageContainer, ProForm, ProFormInstance, ProFormRadio, ProFormSwitch, ProFormText } from '@ant-design/pro-components';
import { Button, Divider, message } from 'antd';

export default function Page() {
  const [cfg, setCfg] = useState<any>(null);
  const [read, setRead] = useState<boolean>(true);
  const formRef = useRef<ProFormInstance>();

  const loadCfg = async () => {
    const rep = await systemConfig();
    console.log(rep.data);
    setCfg(rep.data);
    // formRef?.current?.setFieldsValue(rep.data);
  }
  useEffect(()=> {
    loadCfg();
  }, []);
  return (
    <PageContainer
      header={{
        title: '全局设置',
      }}
    >
    <ProForm
        readonly={read}
        onFinish={async (values) => {
          const hide = message.loading('正在配置');
          try {
            const rep = await systemConfigUpdate(values);
            hide();
            if (rep.code != 0) {
              message.warning(rep.msg);
              return;
            }
            setRead(!read);
            message.success("更新成功!");
          } catch(e) {
            hide();
            message.error(e);
          }
        }}
        layout='horizontal'
        request={async () => {
          const rep = await systemConfig();
          return rep.data;
        }}
        formRef={formRef}
        submitter={{
          render: (props, doms) => {
            if (read) {
              return [
                <Button
                htmlType="button"
                onClick={() => {
                   setRead(!read);
                }}
                key="edit"
              >
                编辑
              </Button>
              ]
            }
            return [
              ...doms,
            ]},
        }}
      >
        <ProForm.Group title="系统设置">
           <ProFormText name={["System","Host"]} label="Host" initialValue={"0.0.0.0"} />
           <ProFormText name={["System","Port"]} label="Port" initialValue={"443"} />
           <ProFormText name={["System", "Domain"]} label="域名"/>
           <ProFormRadio.Group
              name={["System", "Mode"]} label="模式" radioType="button" initialValue={'strict'}
              options={[
                {
                  label: '严格',
                  value: 'strict',
                },
                {
                  label: '宽松',
                  value: 'loose',
                }
              ]}
            />
        </ProForm.Group>
        <ProForm.Group title="服务端">
          <ProFormText name={["Server", "Domain"]} label="域名"/>
          <ProFormRadio.Group
                name={["Server", "ForceHttps"]} label="强制HTTPS" radioType='button' initialValue={true}
                options={[
                  {
                    label: '开启',
                    value: true,
                  },
                  {
                    label: '关闭',
                    value: false,
                  }
                ]}
              />
        </ProForm.Group>
        <ProForm.Group title="Let'sEncrypt">
          <ProFormRadio.Group
                name={["Cloudflare", "Mode"]} label="DNS" radioType='button' initialValue={'cloudflare'}
                options={[
                  {
                    label: '关闭',
                    value: 'close',
                  },
                  {
                    label: 'Cloudflare',
                    value: 'cloudflare',
                  }
                ]}
              />
          <ProFormText  name={["Cloudflare", "Email"]} label="邮箱"  />
          <ProFormText  name={["Cloudflare", "ApiKey"]} label="ApiKey"  />
          <ProFormText  name={["Cloudflare", "DnsApiToken"]} label="DnsApiToken"  />
          <ProFormText  name={["Cloudflare", "ZoneApiToken"]} label="ZoneApiToken"  />
        </ProForm.Group>
        <ProForm.Group title="OAuth">
        <ProFormRadio.Group
                name={["Auth", "Mode"]} label="站点" radioType='button' initialValue={'github'}
                options={[
                  {
                    label: '关闭',
                    value: 'close',
                  },
                  {
                    label: 'Github',
                    value: 'github',
                  },
                  {
                    label: 'Gitee',
                    value: 'gitee'
                  }
                ]}
              />
        <ProFormText name={["Auth", "Token"]} label="Token" />
        <ProFormText name={["Auth", "ClientId"]} label="ClientId" />
        <ProFormText name={["Auth", "Email"]} label="邮箱" />
        </ProForm.Group>
    </ProForm>
    </PageContainer>
  );
}
