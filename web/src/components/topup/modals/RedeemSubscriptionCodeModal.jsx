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
import { Modal, Form, Button } from '@douyinfe/semi-ui';
import { API, showError, showSuccess } from '../../../helpers';

const RedeemSubscriptionCodeModal = ({ visible, onClose, onSuccess, t }) => {
  const [loading, setLoading] = useState(false);
  const [formApi, setFormApi] = useState(null);

  const handleSubmit = async (values) => {
    const key = values.key?.trim();
    if (!key) {
      showError(t('请输入激活码'));
      return;
    }

    if (key.length !== 32) {
      showError(t('激活码格式无效'));
      return;
    }

    setLoading(true);
    try {
      const res = await API.post('/api/subscription_code/redeem', {
        code: key,
      });

      if (res.data?.success) {
        showSuccess(t('激活成功！'));
        formApi?.reset();
        onSuccess?.();
      } else {
        showError(res.data?.message || t('激活失败'));
      }
    } catch (error) {
      showError(error.message || t('激活失败'));
    } finally {
      setLoading(false);
    }
  };

  const handleClose = () => {
    formApi?.reset();
    onClose?.();
  };

  return (
    <Modal
      title={t('兑换订阅激活码')}
      visible={visible}
      onCancel={handleClose}
      footer={null}
      style={{ width: 480 }}
    >
      <Form
        getFormApi={(api) => setFormApi(api)}
        onSubmit={handleSubmit}
        labelPosition='left'
        labelAlign='right'
        labelWidth={100}
      >
        <Form.Input
          field='key'
          label={t('激活码')}
          placeholder={t('请输入32位激活码')}
          rules={[
            { required: true, message: t('请输入激活码') },
            {
              validator: (rule, value) => {
                if (value && value.length !== 32) {
                  return t('激活码必须是32位');
                }
                return true;
              },
            },
          ]}
          maxLength={32}
          showClear
        />

        <div className='flex justify-end gap-2 mt-4'>
          <Button onClick={handleClose}>{t('取消')}</Button>
          <Button
            theme='solid'
            type='primary'
            htmlType='submit'
            loading={loading}
          >
            {t('兑换')}
          </Button>
        </div>
      </Form>
    </Modal>
  );
};

export default RedeemSubscriptionCodeModal;