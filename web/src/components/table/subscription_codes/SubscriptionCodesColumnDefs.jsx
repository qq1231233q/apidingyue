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

import React from 'react';
import { Tag, Button, Space, Popover, Dropdown } from '@douyinfe/semi-ui';
import { IconMore } from '@douyinfe/semi-icons';
import { timestamp2string } from '../../../helpers';
import { formatSubscriptionDuration } from '../../../helpers/subscriptionFormat';
import {
  SUBSCRIPTION_CODE_STATUS,
  SUBSCRIPTION_CODE_STATUS_MAP,
  SUBSCRIPTION_CODE_ACTIONS,
} from '../../../constants/subscription_code.constants';

export const isExpired = (record) => {
  return (
    record.status === SUBSCRIPTION_CODE_STATUS.UNUSED &&
    record.expired_time !== 0 &&
    record.expired_time < Math.floor(Date.now() / 1000)
  );
};

const renderTimestamp = (timestamp) => {
  return <>{timestamp2string(timestamp)}</>;
};

const renderStatus = (status, record, t) => {
  if (isExpired(record)) {
    return (
      <Tag color='orange' shape='circle'>
        {t('已过期')}
      </Tag>
    );
  }

  const statusConfig = SUBSCRIPTION_CODE_STATUS_MAP[status];
  if (statusConfig) {
    return (
      <Tag color={statusConfig.color} shape='circle'>
        {t(statusConfig.text)}
      </Tag>
    );
  }

  return (
    <Tag color='black' shape='circle'>
      {t('未知状态')}
    </Tag>
  );
};

const renderAvailableGroup = (group, t) => {
  return (
    <Tag color='blue' shape='circle'>
      {group || t('所有分组')}
    </Tag>
  );
};

export const getSubscriptionCodesColumns = ({
  t,
  manageCode,
  copyText,
  setEditingCode,
  setShowEdit,
  showDeleteCodeModal,
}) => {
  return [
    {
      title: t('ID'),
      dataIndex: 'id',
      width: 80,
    },
    {
      title: t('名称'),
      dataIndex: 'name',
      width: 180,
    },
    {
      title: t('状态'),
      dataIndex: 'status',
      key: 'status',
      width: 110,
      render: (text, record) => <div>{renderStatus(text, record, t)}</div>,
    },
    {
      title: t('充值额度'),
      dataIndex: 'quota',
      width: 130,
      render: (text) => (
        <Tag color='grey' shape='circle'>
          {text}
        </Tag>
      ),
    },
    {
      title: t('订阅时长'),
      dataIndex: 'duration_value',
      width: 160,
      render: (_, record) => <div>{formatSubscriptionDuration(record, t)}</div>,
    },
    {
      title: t('可用分组'),
      dataIndex: 'available_group',
      width: 140,
      render: (text) => renderAvailableGroup(text, t),
    },
    {
      title: t('创建时间'),
      dataIndex: 'created_time',
      width: 180,
      render: (text) => <div>{renderTimestamp(text)}</div>,
    },
    {
      title: t('激活码有效期'),
      dataIndex: 'expired_time',
      width: 180,
      render: (text) => (
        <div>{text === 0 ? t('永久有效') : renderTimestamp(text)}</div>
      ),
    },
    {
      title: t('使用用户ID'),
      dataIndex: 'used_user_id',
      width: 120,
      render: (text) => <div>{text === 0 ? t('未使用') : text}</div>,
    },
    {
      title: '',
      dataIndex: 'operate',
      fixed: 'right',
      width: 205,
      render: (_, record) => {
        const moreMenuItems = [
          {
            node: 'item',
            name: t('删除'),
            type: 'danger',
            onClick: () => {
              showDeleteCodeModal(record);
            },
          },
        ];

        if (
          record.status === SUBSCRIPTION_CODE_STATUS.UNUSED &&
          !isExpired(record)
        ) {
          moreMenuItems.push({
            node: 'item',
            name: t('禁用'),
            type: 'warning',
            onClick: () => {
              manageCode(record.id, SUBSCRIPTION_CODE_ACTIONS.DISABLE, record);
            },
          });
        } else if (!isExpired(record)) {
          moreMenuItems.push({
            node: 'item',
            name: t('启用'),
            type: 'secondary',
            onClick: () => {
              manageCode(record.id, SUBSCRIPTION_CODE_ACTIONS.ENABLE, record);
            },
            disabled: record.status === SUBSCRIPTION_CODE_STATUS.USED,
          });
        }

        return (
          <Space>
            <Popover content={record.key} style={{ padding: 20 }} position='top'>
              <Button type='tertiary' size='small'>
                {t('查看')}
              </Button>
            </Popover>
            <Button
              size='small'
              onClick={async () => {
                await copyText(record.key);
              }}
            >
              {t('复制')}
            </Button>
            <Button
              type='tertiary'
              size='small'
              onClick={() => {
                setEditingCode(record);
                setShowEdit(true);
              }}
              disabled={record.status !== SUBSCRIPTION_CODE_STATUS.UNUSED}
            >
              {t('编辑')}
            </Button>
            <Dropdown
              trigger='click'
              position='bottomRight'
              menu={moreMenuItems}
            >
              <Button type='tertiary' size='small' icon={<IconMore />} />
            </Dropdown>
          </Space>
        );
      },
    },
  ];
};
