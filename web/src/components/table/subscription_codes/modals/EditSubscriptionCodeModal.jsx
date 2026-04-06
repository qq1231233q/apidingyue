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

import React, { useEffect, useMemo, useRef, useState } from 'react';
import { useTranslation } from 'react-i18next';
import {
  API,
  downloadTextAsFile,
  showError,
  showSuccess,
} from '../../../../helpers';
import { useIsMobile } from '../../../../hooks/common/useIsMobile';
import {
  Button,
  Modal,
  SideSheet,
  Space,
  Spin,
  Typography,
  Card,
  Tag,
  Form,
  Avatar,
  Row,
  Col,
  Select,
} from '@douyinfe/semi-ui';
import {
  IconCreditCard,
  IconSave,
  IconClose,
  IconGift,
} from '@douyinfe/semi-icons';

const { Text, Title } = Typography;

function normalizeGroupOptions(groups, extraGroups = []) {
  const uniqueGroups = new Set();
  [...(groups || []), ...(extraGroups || [])].forEach((group) => {
    const value = typeof group === 'string' ? group.trim() : '';
    if (value) {
      uniqueGroups.add(value);
    }
  });

  return Array.from(uniqueGroups)
    .sort((a, b) => {
      if (a === 'default') return -1;
      if (b === 'default') return 1;
      return a.localeCompare(b);
    })
    .map((group) => ({
      label: group,
      value: group,
    }));
}

const EditSubscriptionCodeModal = (props) => {
  const { t } = useTranslation();
  const isEdit = props.editingCode.id !== undefined;
  const [loading, setLoading] = useState(isEdit);
  const [groupOptions, setGroupOptions] = useState([]);
  const [groupLoading, setGroupLoading] = useState(false);
  const isMobile = useIsMobile();
  const formApiRef = useRef(null);

  const selectGroupOptions = useMemo(
    () =>
      normalizeGroupOptions(groupOptions, [props.editingCode?.available_group]),
    [groupOptions, props.editingCode?.available_group],
  );

  const selectableGroupsText = useMemo(
    () => selectGroupOptions.map((item) => item.value).join(' / '),
    [selectGroupOptions],
  );

  const getInitValues = () => ({
    name: '',
    quota: 1000000,
    count: 1,
    expired_time: null,
    duration_unit: 'month',
    duration_value: 1,
    custom_seconds: 0,
    available_group: '',
  });

  const handleCancel = () => {
    props.handleClose();
  };

  const loadCode = async () => {
    setLoading(true);
    const res = await API.get(`/api/subscription_code/${props.editingCode.id}`);
    const { success, message, data } = res.data;
    if (success) {
      const next = { ...getInitValues(), ...data };
      next.expired_time =
        next.expired_time && next.expired_time !== 0
          ? new Date(next.expired_time * 1000)
          : null;
      formApiRef.current?.setValues(next);
    } else {
      showError(message);
    }
    setLoading(false);
  };

  useEffect(() => {
    if (!formApiRef.current) return;
    if (isEdit) {
      loadCode();
      return;
    }
    formApiRef.current.setValues(getInitValues());
  }, [props.editingCode.id]);

  useEffect(() => {
    if (!props.visiable) return;
    setGroupLoading(true);
    API.get('/api/group')
      .then((res) => {
        if (res.data?.success) {
          setGroupOptions(res.data?.data || []);
        } else {
          setGroupOptions([]);
        }
      })
      .catch(() => setGroupOptions([]))
      .finally(() => setGroupLoading(false));
  }, [props.visiable]);

  const submit = async (values) => {
    let name = values.name;
    if (!isEdit && (!name || name === '')) {
      name = t('订阅激活码');
    }

    setLoading(true);
    const localInputs = {
      ...values,
      count: parseInt(values.count, 10) || 0,
      quota: parseInt(values.quota, 10) || 0,
      name,
      expired_time: values.expired_time
        ? Math.floor(values.expired_time.getTime() / 1000)
        : 0,
    };

    let res;
    if (isEdit) {
      res = await API.put('/api/subscription_code/', {
        ...localInputs,
        id: parseInt(props.editingCode.id, 10),
      });
    } else {
      res = await API.post('/api/subscription_code/', localInputs);
    }

    const { success, message, data } = res.data;
    if (!success) {
      showError(message);
      setLoading(false);
      return;
    }

    if (isEdit) {
      showSuccess(t('激活码更新成功！'));
      props.refresh();
      props.handleClose();
      setLoading(false);
      return;
    }

    showSuccess(t('激活码创建成功！'));
    props.refresh();
    formApiRef.current?.setValues(getInitValues());
    props.handleClose();

    if (data) {
      const text = data.map((item) => `${item}\n`).join('');
      Modal.confirm({
        title: t('激活码创建成功'),
        content: (
          <div>
            <p>{t('激活码创建成功，是否下载激活码？')}</p>
            <p>{t('激活码将以文本文件的形式下载，文件名为激活码的名称。')}</p>
          </div>
        ),
        onOk: () => {
          downloadTextAsFile(text, `${localInputs.name}.txt`);
        },
      });
    }

    setLoading(false);
  };

  return (
    <>
      <SideSheet
        placement={isEdit ? 'right' : 'left'}
        title={
          <Space>
            {isEdit ? (
              <Tag color='blue' shape='circle'>
                {t('更新')}
              </Tag>
            ) : (
              <Tag color='green' shape='circle'>
                {t('新建')}
              </Tag>
            )}
            <Title heading={4} className='m-0'>
              {isEdit ? t('更新激活码信息') : t('创建新的激活码')}
            </Title>
          </Space>
        }
        bodyStyle={{ padding: '0' }}
        visible={props.visiable}
        width={isMobile ? '100%' : 600}
        footer={
          <div className='flex justify-end bg-white'>
            <Space>
              <Button
                theme='solid'
                onClick={() => formApiRef.current?.submitForm()}
                icon={<IconSave />}
                loading={loading}
              >
                {t('提交')}
              </Button>
              <Button
                theme='light'
                type='primary'
                onClick={handleCancel}
                icon={<IconClose />}
              >
                {t('取消')}
              </Button>
            </Space>
          </div>
        }
        closeIcon={null}
        onCancel={handleCancel}
      >
        <Spin spinning={loading}>
          <Form
            initValues={getInitValues()}
            getFormApi={(api) => {
              formApiRef.current = api;
            }}
            onSubmit={submit}
          >
            {({ values }) => (
              <div className='p-2'>
                <Card className='!rounded-2xl shadow-sm border-0 mb-6'>
                  <div className='flex items-center mb-2'>
                    <Avatar
                      size='small'
                      color='blue'
                      className='mr-2 shadow-md'
                    >
                      <IconGift size={16} />
                    </Avatar>
                    <div>
                      <Text className='text-lg font-medium'>{t('基本信息')}</Text>
                      <div className='text-xs text-gray-600'>
                        {t('设置激活码的基本信息')}
                      </div>
                    </div>
                  </div>

                  <Row gutter={12}>
                    <Col span={24}>
                      <Form.Input
                        field='name'
                        label={t('名称')}
                        placeholder={t('请输入名称')}
                        style={{ width: '100%' }}
                        rules={
                          !isEdit
                            ? []
                            : [{ required: true, message: t('请输入名称') }]
                        }
                        showClear
                      />
                    </Col>
                    <Col span={24}>
                      <Form.DatePicker
                        field='expired_time'
                        label={t('激活码有效截止时间')}
                        type='dateTime'
                        placeholder={t('选择激活码有效截止时间（可选，留空为永久）')}
                        style={{ width: '100%' }}
                        showClear
                      />
                      <div className='text-xs text-gray-500 mt-1'>
                        {t('这里只限制激活码最晚可兑换到什么时候，不是订阅结束时间')}
                      </div>
                    </Col>
                  </Row>
                </Card>

                <Card className='!rounded-2xl shadow-sm border-0'>
                  <div className='flex items-center mb-2'>
                    <Avatar
                      size='small'
                      color='green'
                      className='mr-2 shadow-md'
                    >
                      <IconCreditCard size={16} />
                    </Avatar>
                    <div>
                      <Text className='text-lg font-medium'>{t('额度设置')}</Text>
                      <div className='text-xs text-gray-600'>
                        {t('设置激活码的充值额度和数量')}
                      </div>
                    </div>
                  </div>

                  <Row gutter={12}>
                    <Col span={24}>
                      <Form.InputNumber
                        field='quota'
                        label={t('充值额度')}
                        min={1}
                        rules={[
                          { required: true, message: t('请输入充值额度') },
                          {
                            validator: (rule, v) => {
                              const num = parseInt(v, 10);
                              return num > 0
                                ? Promise.resolve()
                                : Promise.reject(t('充值额度必须大于 0'));
                            },
                          },
                        ]}
                        style={{ width: '100%' }}
                        showClear
                      />
                    </Col>
                  </Row>

                  {!isEdit && (
                    <Row gutter={12} style={{ marginTop: '12px' }}>
                      <Col span={24}>
                        <Form.InputNumber
                          field='count'
                          label={t('生成数量')}
                          min={1}
                          rules={[
                            { required: true, message: t('请输入生成数量') },
                            {
                              validator: (rule, v) => {
                                const num = parseInt(v, 10);
                                return num > 0
                                  ? Promise.resolve()
                                  : Promise.reject(t('生成数量必须大于0'));
                              },
                            },
                          ]}
                          style={{ width: '100%' }}
                          showClear
                        />
                      </Col>
                    </Row>
                  )}
                </Card>

                <Card className='!rounded-2xl shadow-sm border-0 mt-6'>
                  <div className='flex items-center mb-2'>
                    <Avatar
                      size='small'
                      color='purple'
                      className='mr-2 shadow-md'
                    >
                      <IconCreditCard size={16} />
                    </Avatar>
                    <div>
                      <Text className='text-lg font-medium'>{t('订阅时长')}</Text>
                      <div className='text-xs text-gray-600'>
                        {t('用户兑换后，从兑换时刻开始按这里计时')}
                      </div>
                    </div>
                  </div>

                  <Row gutter={12}>
                    <Col span={12}>
                      <Form.Select
                        field='duration_unit'
                        label={t('时长单位')}
                        style={{ width: '100%' }}
                        rules={[
                          { required: true, message: t('请选择时长单位') },
                        ]}
                      >
                        <Select.Option value='year'>{t('年')}</Select.Option>
                        <Select.Option value='month'>{t('月')}</Select.Option>
                        <Select.Option value='day'>{t('天')}</Select.Option>
                        <Select.Option value='hour'>{t('小时')}</Select.Option>
                        <Select.Option value='custom'>
                          {t('自定义秒数')}
                        </Select.Option>
                      </Form.Select>
                    </Col>
                    <Col span={12}>
                      {values.duration_unit !== 'custom' ? (
                        <Form.InputNumber
                          field='duration_value'
                          label={t('时长数值')}
                          min={1}
                          rules={[
                            { required: true, message: t('请输入时长数值') },
                            {
                              validator: (rule, v) => {
                                const num = parseInt(v, 10);
                                return num > 0
                                  ? Promise.resolve()
                                  : Promise.reject(t('时长数值必须大于 0'));
                              },
                            },
                          ]}
                          style={{ width: '100%' }}
                          showClear
                        />
                      ) : (
                        <Form.InputNumber
                          field='custom_seconds'
                          label={t('自定义秒数')}
                          min={1}
                          rules={[
                            { required: true, message: t('请输入秒数') },
                            {
                              validator: (rule, v) => {
                                const num = parseInt(v, 10);
                                return num > 0
                                  ? Promise.resolve()
                                  : Promise.reject(t('秒数必须大于0'));
                              },
                            },
                          ]}
                          style={{ width: '100%' }}
                          showClear
                        />
                      )}
                    </Col>
                  </Row>
                </Card>

                <Card className='!rounded-2xl shadow-sm border-0 mt-6'>
                  <div className='flex items-center mb-2'>
                    <Avatar
                      size='small'
                      color='orange'
                      className='mr-2 shadow-md'
                    >
                      <IconGift size={16} />
                    </Avatar>
                    <div>
                      <Text className='text-lg font-medium'>{t('高级设置')}</Text>
                      <div className='text-xs text-gray-600'>
                        {t('可选的高级配置')}
                      </div>
                    </div>
                  </div>

                  <Row gutter={12}>
                    <Col span={24}>
                      <Form.Select
                        field='available_group'
                        label={t('可用分组')}
                        showClear
                        loading={groupLoading}
                        placeholder={t('所有分组')}
                      >
                        <Select.Option value=''>{t('所有分组')}</Select.Option>
                        {selectGroupOptions.map((group) => (
                          <Select.Option key={group.value} value={group.value}>
                            {group.label}
                          </Select.Option>
                        ))}
                      </Form.Select>
                      <div className='text-xs text-gray-500 mt-1'>
                        {t('限制订阅额度只能用于指定分组的模型')}
                      </div>
                      {selectableGroupsText ? (
                        <div className='text-xs text-gray-500 mt-1'>
                          {t('可选分组')}: {selectableGroupsText}
                        </div>
                      ) : null}
                    </Col>
                  </Row>
                </Card>
              </div>
            )}
          </Form>
        </Spin>
      </SideSheet>
    </>
  );
};

export default EditSubscriptionCodeModal;
