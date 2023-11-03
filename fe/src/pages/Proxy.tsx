
import styles from './Proxy.less';
import {
  ActionType,
  FooterToolbar,
  PageContainer,
  ProDescriptions,
  ProDescriptionsItemProps,
  ProTable,
} from '@ant-design/pro-components';
import { Button, Divider, Drawer, message, Switch, Tag } from 'antd';
import React, { useRef, useState } from 'react';
import CreateForm from '@/components/Proxy/CreateForm';
import UpdateForm, { FormValueType } from '@/components/Proxy/UpdateForm';
import { forwardDel, forwardList, forwardSave } from '@/services/controller';
import { Link } from '@umijs/max';

/**
 * 添加节点
 * @param fields
 */
const handleAdd = async (fields: any) => {
  const hide = message.loading('正在添加');
  try {
    await forwardSave({ ...fields });
    hide();
    message.success('添加成功');
    return true;
  } catch (error) {
    hide();
    message.error('添加失败请重试！');
    return false;
  }
};

/**
 * 更新节点
 * @param fields
 */
const handleUpdate = async (fields: any) => {
  const hide = message.loading('正在配置');
  try {
    await forwardSave(
      {
        id: fields.id,
        from: fields.from || '',
        to: fields.to || '',
        client: fields.client || 'default',
        schema: fields.schema || 'https',
        status: fields.status || 1,
      }
    );
    hide();

    message.success('配置成功');
    return true;
  } catch (error) {
    hide();
    message.error('配置失败请重试！');
    return false;
  }
};

/**
 *  删除节点
 * @param selectedRows
 */
const handleRemove = async (selectedRows: any[]) => {
  const hide = message.loading('正在删除');
  if (!selectedRows) return true;
  try {
    await forwardDel({
      ids: selectedRows.map((row) => row.id) || [],
    });
    hide();
    message.success('删除成功，即将刷新');
    return true;
  } catch (error) {
    hide();
    message.error('删除失败，请重试');
    return false;
  }
};

const Proxy: React.FC<unknown> = () => {
  const [createModalVisible, handleModalVisible] = useState<boolean>(false);
  const [updateModalVisible, handleUpdateModalVisible] =
    useState<boolean>(false);
  const [formValues, setFormValues] = useState({});
  const actionRef = useRef<ActionType>();
  const [row, setRow] = useState<any>();
  const columns: ProDescriptionsItemProps<any>[] = [
    {
      title: 'ID',
      dataIndex: 'id',
      tip: '',
      hideInForm: true,
      hideInSearch: true,
    },
    {
      title: '来源域名',
      dataIndex: 'from',
      valueType: 'text',
      render: (_, record) => (
        <>
           <Link to={'https://'+record.from} target='_blank'> <Tag color="success">https://{record.from}</Tag> </Link>
        </>
      )
    },
    {
      title: '协议',
      dataIndex: 'schema',
      // valueType: 'text',
      hideInTable: true,
      hideInSearch: true,
      valueEnum: {
        https: { text: 'HTTPS', status: 'Success' },
        http: { text: 'HTTP', status: 'Info' },
      },
    },
    {
      title: '转发域名',
      dataIndex: 'to',
      valueType: 'text',
      render: (_, record) => (
        <>

           <Tag color={record.schema == 'https'? 'success' :'default'}>{record.schema}://{record.to}</Tag>
        </>
      )
    },
    {
      title: '客户端',
      dataIndex: 'client',
      valueType: 'text',
      hideInSearch: true,
    },
    {
      title: '状态',
      dataIndex: 'status', 
      valueType: 'option',
      hideInSearch: true,
      render: (_, record) => (
        <>
          <Switch checkedChildren="开启" unCheckedChildren="关闭" checked={record.status == 1}
             onChange={(c: boolean) => {
              console.log({...record, status: c ? 1: 2})
              handleUpdate({
                ...record,
                status: c ? 1 : 2
              })
              if (actionRef.current) {
                actionRef.current.reload();
              }

          }}  />
        </>
      )
    },
    {
      title: '操作',
      dataIndex: 'option',
      valueType: 'option',
      render: (_, record) => (
        <>
          <a
            onClick={() => {
              setFormValues(record);
              handleUpdateModalVisible(true);
            }}
          >
            修改
          </a>
          <Divider type="vertical" />
          <a href='#' onClick={()=> {
            handleRemove([record])
            if (actionRef.current) {
              actionRef.current.reload();
            }
          }}>删除</a>
        </>
      ),
    },
  ];

  return (
    <PageContainer
      header={{
        title: '转发规则配置',
      }}
    >
      <ProTable<any>
        headerTitle="转发配置"
        actionRef={actionRef}
        rowKey="id"
        search={{
          labelWidth: 120,
        }}
        toolBarRender={() => [
          <Button
            key="1"
            type="primary"
            onClick={() => handleModalVisible(true)}
          >
            新建
          </Button>,
        ]}
        request={async (params, sorter, filter) => {
          const { data, code } = await forwardList({
            ...params,
            // FIXME: remove @ts-ignore
            // @ts-ignore
            sorter,
            filter,
          });
          return {
            data: data || [],
            success: code == 0,
          };
        }}
        columns={columns}
      />
      <CreateForm
        onCancel={() => handleModalVisible(false)}
        modalVisible={createModalVisible}
      >
        <ProTable<any, any>
          onSubmit={async (value) => {
            const success = await handleAdd(value);
            if (success) {
              handleModalVisible(false);
              if (actionRef.current) {
                actionRef.current.reload();
              }
            }
          }}
          rowKey="id"
          type="form"
          columns={columns}
        />
      </CreateForm>
      {formValues && Object.keys(formValues).length ? (
        <UpdateForm
          onSubmit={async (value) => {
            const success = await handleUpdate(value);
            if (success) {
              handleUpdateModalVisible(false);
              setFormValues({});
              if (actionRef.current) {
                actionRef.current.reload();
              }
            }
          }}
          onCancel={() => {
            handleUpdateModalVisible(false);
            setFormValues({});
          }}
          updateModalVisible={updateModalVisible}
          values={formValues}
        >
        </UpdateForm>
      ) : null}

      <Drawer
        width={600}
        open={!!row}
        onClose={() => {
          setRow(undefined);
        }}
        closable={false}
      >
        {row?.name && (
          <ProDescriptions<any>
            column={2}
            title={row?.name}
            request={async () => ({
              data: row || {},
            })}
            params={{
              id: row?.name,
            }}
            columns={columns}
          />
        )}
      </Drawer>
    </PageContainer>
  );
};

export default Proxy;
