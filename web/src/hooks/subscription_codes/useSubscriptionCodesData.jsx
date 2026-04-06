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

import { useState, useEffect } from 'react';
import { API, showError, showSuccess, copy } from '../../helpers';
import { ITEMS_PER_PAGE } from '../../constants';
import {
  SUBSCRIPTION_CODE_ACTIONS,
  SUBSCRIPTION_CODE_STATUS,
} from '../../constants/subscription_code.constants';
import { Modal } from '@douyinfe/semi-ui';
import { useTranslation } from 'react-i18next';
import { useTableCompactMode } from '../common/useTableCompactMode';

export const useSubscriptionCodesData = () => {
  const { t } = useTranslation();

  const [codes, setCodes] = useState([]);
  const [loading, setLoading] = useState(true);
  const [searching, setSearching] = useState(false);
  const [activePage, setActivePage] = useState(1);
  const [pageSize, setPageSize] = useState(ITEMS_PER_PAGE);
  const [codeCount, setCodeCount] = useState(0);
  const [selectedKeys, setSelectedKeys] = useState([]);

  const [editingCode, setEditingCode] = useState({ id: undefined });
  const [showEdit, setShowEdit] = useState(false);
  const [formApi, setFormApi] = useState(null);

  const [compactMode, setCompactMode] = useTableCompactMode('subscription_codes');

  const formInitValues = {
    searchKeyword: '',
  };

  const getFormValues = () => {
    const formValues = formApi ? formApi.getValues() : {};
    return {
      searchKeyword: formValues.searchKeyword || '',
    };
  };

  const loadCodes = async (page = 1, size) => {
    setLoading(true);
    try {
      const res = await API.get(
        `/api/subscription_code/?p=${page}&page_size=${size}`,
      );
      const { success, message, data } = res.data;
      if (success) {
        setActivePage(data.page <= 0 ? 1 : data.page);
        setCodeCount(data.total);
        setCodes(data.items || []);
      } else {
        showError(message);
      }
    } catch (error) {
      showError(error.message);
    } finally {
      setLoading(false);
    }
  };

  const searchCodes = async (page = 1, size = pageSize) => {
    const { searchKeyword } = getFormValues();
    if (searchKeyword === '') {
      await loadCodes(page, size);
      return;
    }

    setSearching(true);
    try {
      const res = await API.get(
        `/api/subscription_code/search?keyword=${searchKeyword}&p=${page}&page_size=${size}`,
      );
      const { success, message, data } = res.data;
      if (success) {
        setActivePage(data.page || 1);
        setCodeCount(data.total);
        setCodes(data.items || []);
      } else {
        showError(message);
      }
    } catch (error) {
      showError(error.message);
    } finally {
      setSearching(false);
    }
  };

  const manageCode = async (id, action, record) => {
    setLoading(true);
    let data = { id };
    let res;

    try {
      switch (action) {
        case SUBSCRIPTION_CODE_ACTIONS.DELETE:
          res = await API.delete(`/api/subscription_code/${id}/`);
          break;
        case SUBSCRIPTION_CODE_ACTIONS.ENABLE:
          data.status = SUBSCRIPTION_CODE_STATUS.UNUSED;
          res = await API.put('/api/subscription_code/?status_only=true', data);
          break;
        case SUBSCRIPTION_CODE_ACTIONS.DISABLE:
          data.status = SUBSCRIPTION_CODE_STATUS.DISABLED;
          res = await API.put('/api/subscription_code/?status_only=true', data);
          break;
        default:
          throw new Error('Unknown operation type');
      }

      const { success, message, data: responseData } = res.data;
      if (success) {
        showSuccess(t('操作成功'));
        if (action !== SUBSCRIPTION_CODE_ACTIONS.DELETE) {
          setCodes((prev) =>
            prev.map((item) =>
              item.id === id
                ? {
                    ...item,
                    status: responseData?.status ?? record?.status,
                  }
                : item,
            ),
          );
        }
      } else {
        showError(message);
      }
    } catch (error) {
      showError(error.message);
    } finally {
      setLoading(false);
    }
  };

  const refresh = async (page = activePage) => {
    const { searchKeyword } = getFormValues();
    if (searchKeyword === '') {
      await loadCodes(page, pageSize);
    } else {
      await searchCodes(page, pageSize);
    }
  };

  const handlePageChange = (page) => {
    setActivePage(page);
    const { searchKeyword } = getFormValues();
    if (searchKeyword === '') {
      loadCodes(page, pageSize);
    } else {
      searchCodes(page, pageSize);
    }
  };

  const handlePageSizeChange = (size) => {
    setPageSize(size);
    setActivePage(1);
    const { searchKeyword } = getFormValues();
    if (searchKeyword === '') {
      loadCodes(1, size);
    } else {
      searchCodes(1, size);
    }
  };

  const rowSelection = {
    onSelect: () => {},
    onSelectAll: () => {},
    onChange: (_, selectedRows) => {
      setSelectedKeys(selectedRows);
    },
  };

  const handleRow = (record) => {
    const expired =
      record.status === SUBSCRIPTION_CODE_STATUS.UNUSED &&
      record.expired_time !== 0 &&
      record.expired_time < Math.floor(Date.now() / 1000);

    if (record.status !== SUBSCRIPTION_CODE_STATUS.UNUSED || expired) {
      return {
        style: {
          background: 'var(--semi-color-disabled-border)',
        },
      };
    }
    return {};
  };

  const copyText = async (text) => {
    if (await copy(text)) {
      showSuccess(t('已复制到剪贴板'));
    } else {
      Modal.error({
        title: t('无法复制到剪贴板，请手动复制'),
        content: text,
        size: 'large',
      });
    }
  };

  const batchCopyCodes = async () => {
    if (selectedKeys.length === 0) {
      showError(t('请至少选择一个激活码'));
      return;
    }

    const keys = selectedKeys
      .map((item) => `${item.name}    ${item.key}`)
      .join('\n');
    await copyText(keys);
  };

  const batchDeleteCodes = async () => {
    Modal.confirm({
      title: t('确定清除所有失效激活码？'),
      content: t('将删除已使用、已禁用及过期的激活码，此操作不可撤销。'),
      onOk: async () => {
        setLoading(true);
        try {
          const res = await API.delete('/api/subscription_code/invalid');
          const { success, message, data } = res.data;
          if (success) {
            showSuccess(t('已删除 {{count}} 条失效激活码', { count: data }));
            await refresh();
          } else {
            showError(message);
          }
        } finally {
          setLoading(false);
        }
      },
    });
  };

  const closeEdit = () => {
    setShowEdit(false);
    setTimeout(() => {
      setEditingCode({ id: undefined });
    }, 500);
  };

  const removeRecord = (key) => {
    setCodes((prev) => prev.filter((item) => item.key !== key));
  };

  useEffect(() => {
    loadCodes(1, pageSize)
      .then()
      .catch((reason) => {
        showError(reason);
      });
  }, [pageSize]);

  return {
    codes,
    loading,
    searching,
    activePage,
    pageSize,
    codeCount,
    selectedKeys,

    editingCode,
    showEdit,

    formApi,
    formInitValues,

    compactMode,
    setCompactMode,

    loadCodes,
    searchCodes,
    manageCode,
    refresh,
    copyText,
    removeRecord,

    setActivePage,
    setPageSize,
    setSelectedKeys,
    setEditingCode,
    setShowEdit,
    setFormApi,
    setLoading,

    handlePageChange,
    handlePageSizeChange,
    rowSelection,
    handleRow,
    closeEdit,
    getFormValues,

    batchCopyCodes,
    batchDeleteCodes,

    t,
  };
};
