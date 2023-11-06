import React, { useEffect, useState } from 'react';
import styles from './Dashboard.less';
import { stat } from '@/services/controller';
import { PageContainer, StatisticCard } from '@ant-design/pro-components';
import { Divider } from 'antd';

export default function Page() {
  const [info, setInfo] = useState<any>(null);
  useEffect(() => {
    stat().then(v => {
      console.log(v);
      setInfo(v.data);
    })
  }, [])
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
      {(info?.client||[]).map((v) => {
        return <StatisticCard.Group direction={'row' }>
            <StatisticCard
          statistic={{
            title: v.name,
          }}
        />
        <Divider type={'vertical'} />
        <StatisticCard
          statistic={{
            title: '连接数',
            value: v.size,
          }}
        />
        <Divider type={'vertical'} />
        <StatisticCard
          statistic={{
            title: '连接使用数',
            value: v.use,
          }}
        />
        </StatisticCard.Group>
      })}
    </PageContainer>
  );
}
