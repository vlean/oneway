import {
  ModalForm,
  ProFormDateTimePicker,
  ProFormRadio,
  ProFormSelect,
  ProFormText,
  ProFormTextArea,
  StepsForm,
} from '@ant-design/pro-components';
import { Modal } from 'antd';
import { Button, Form, message } from 'antd';
import React, {PropsWithChildren} from 'react';

export interface FormValueType extends Partial<any> {
  target?: string;
  template?: string;
  type?: string;
  time?: string;
  frequency?: string;
}

export interface UpdateFormProps {
  onCancel: (flag?: boolean, formVals?: FormValueType) => void;
  onSubmit: (values: FormValueType) => Promise<void>;
  updateModalVisible: boolean;
  values: Partial<any>;
}


const UpdateForm: React.FC<PropsWithChildren<any>> = (props) => {
  const { updateModalVisible, onCancel, onSubmit, values } = props;
  const [form] = Form.useForm<any>();

  return (
    <ModalForm<any>
      title="修改规则"
      open={updateModalVisible}
      initialValues={values}
      layout='horizontal'
      form={form}
      width={420}
      labelCol={{span: 8}}
      wrapperCol={{span: 14}}
      autoFocusFirstInput
      modalProps={{
        destroyOnClose: true,
        onCancel: () => onCancel(),
      }}
      onFinish={(v) => onSubmit(v)}
    >
        <ProFormText
          // width="md"
          name="id"
          label="ID"
          disabled
        />
        <ProFormRadio.Group
          name="schema"
          label="协议"
          radioType='button'
          options={[
            {
              label: 'HTTPS',
              value: 'https',
            },
            {
              label: 'HTTP',
              value: 'http',
            },
          ]}
        />
         <ProFormText 
          name="from"
          label="来源域名"
        />
        <ProFormText 
          name="to"
          label="转发域名"
        />
         <ProFormText 
          name="client"
          label="客户端"
        />
    </ModalForm>
  );
};

export default UpdateForm;