import React, { useEffect, useState } from 'react';
import styles from './Dashboard.less';
import { stat } from '@/services/controller';
import { PageContainer, ProList, StatisticCard } from '@ant-design/pro-components';
import { Divider, Space, Tag } from 'antd';
import { ApiOutlined } from '@ant-design/icons';

export default function Page() {
  const [info, setInfo] = useState<any>(null);
  useEffect(() => {
    stat().then(v => {
      console.log(v);
      setInfo(v.data);
    })
  }, [])
  const formatApproximateTime = (seconds: any) => {
    if (seconds < 60) {
      return `${seconds}秒`;
    } else if (seconds < 3600) {
      const minutes = Math.floor(seconds / 60);
      return `${minutes}分钟`;
    } else if (seconds < 86400) {
      const hours = Math.floor(seconds / 3600);
      return `${hours}小时`;
    } else {
      const days = Math.floor(seconds / 86400);
      return `${days}天`;
    }
  }
  
  return (
    <PageContainer title="监控">
      <StatisticCard.Group direction={'row' }>
        <StatisticCard
          statistic={{
            title: '请求次数',
            value: info?.http?.request || 0,
          }}
        />
        <Divider type={'vertical'} />
        <StatisticCard
          statistic={{
            title: '认证失败',
            value: info?.http?.auth_fail || 0,
          }}
        />
        <Divider type={'vertical'} />
        <StatisticCard
          statistic={{
            title: '流量',
            value: info?.http?.body_size || 0,
          }}
        />
      </StatisticCard.Group>
      <Divider/>
      <ProList<any>
      rowKey="name"
      headerTitle="客户端列表"
      dataSource={info?.client||[]}
      metas={{
        title: {
          dataIndex: 'name',
        },
        description: {
          render: (_: any, v: any) => {
            return (
              <StatisticCard.Group direction={'row' }>
            <StatisticCard
              statistic={{
                title: '连接数',
                value: v.size,
              }}
            />
          <Divider type={'vertical'} />
          <StatisticCard
            statistic={{
              title: '使用中',
              value: v.size -v.use,
            }}
          />
          </StatisticCard.Group>
          )
          }
        },
        subTitle: {
          render: (_, v: any) => {
            const run = v?.run_time || -1
            return (
              <Space size={0}>
                {run > 0 && <Tag color="blue" title='运行时间' >{formatApproximateTime(run) }</Tag>}
              </Space>
            );
          },
        },
      }}
    />
    </PageContainer>
  );
}
