/*
Copyright (C) 2025 QuantumNous

This program is free software: you can redistribute it and/or modify
it under the terms of the GNU Affero General Public License as
published by the Free Software Foundation, either version 3 of the
License, or (at your option) any later version.

This program is distributed in the hope that it will be useful,
but WITHOUT ANY WARRANTY; without even the implied warranty of
MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
GNU Affero General Public License for more details.

You should have received a copy of the GNU Affero General Public License
along with this program. If not, see <https://www.gnu.org/licenses/>.

For commercial licensing, please contact support@quantumnous.com
*/

import React, { useState } from 'react';
import {
  Modal,
  Form,
  Button,
  Space,
  Banner,
} from '@douyinfe/semi-ui';
import { IconClose, IconSave } from '@douyinfe/semi-icons';
import { API, showError, showSuccess } from '../../../../helpers';

const RedeemSubscriptionCodeModal = ({
  visible,
  handleClose,
  planRecord,
  refresh,
  t,
}) => {
  const [loading, setLoading] = useState(false);
  const [formApi, setFormApi] = useState(null);

  const handleSubmit = async (values) => {
    if (!values.code || values.code.trim() === '') {
      showError(t('请输入激活码'));
      return;
    }

    setLoading(true);
    try {
      const res = await API.post('/api/subscription_code/redeem', {
        code: values.code.trim(),
      });

      if (res.data?.success) {
        showSuccess(t('激活成功'));
        handleClose();
        refresh?.();
      } else {
        showError(res.data?.message || t('激活失败'));
      }
    } catch (e) {
      showError(t('请求失败'));
    } finally {
      setLoading(false);
    }
  };

  return (
    <Modal
      title={t('兑换激活码')}
      visible={visible}
      onCancel={handleClose}
      footer={
        <Space>
          <Button
            theme='solid'
            onClick={() => formApi?.submitForm()}
            icon={<IconSave />}
            loading={loading}
          >
            {t('兑换')}
          </Button>
          <Button
            theme='light'
            type='primary'
            onClick={handleClose}
            icon={<IconClose />}
          >
            {t('取消')}
          </Button>
        </Space>
      }
      centered
    >
      <div className='space-y-4'>
        {planRecord?.plan?.title && (
          <Banner
            type='info'
            description={`${t('套餐')}：${planRecord.plan.title}`}
            closeIcon={null}
          />
        )}

        <Form
          initValues={{ code: '' }}
          getFormApi={(api) => setFormApi(api)}
          onSubmit={handleSubmit}
        >
          <Form.Input
            field='code'
            label={t('激活码')}
            placeholder={t('请输入激活码')}
            required
            rules={[{ required: true, message: t('请输入激活码') }]}
            showClear
          />
        </Form>
      </div>
    </Modal>
  );
};

export default RedeemSubscriptionCodeModal;