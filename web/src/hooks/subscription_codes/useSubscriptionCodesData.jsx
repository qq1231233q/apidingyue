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

  // Basic state
  const [codes, setCodes] = useState([]);
  const [loading, setLoading] = useState(true);
  const [searching, setSearching] = useState(false);
  const [activePage, setActivePage] = useState(1);
  const [pageSize, setPageSize] = useState(ITEMS_PER_PAGE);
  const [codeCount, setCodeCount] = useState(0);
  const [selectedKeys, setSelectedKeys] = useState([]);

  // Edit state
  const [editingCode, setEditingCode] = useState({ id: undefined });
  const [showEdit, setShowEdit] = useState(false);

  // Form API
  const [formApi, setFormApi] = useState(null);

  // UI state
  const [compactMode, setCompactMode] = useTableCompactMode('subscription_codes');

  // Form state
  const formInitValues = {
    searchKeyword: '',
  };

  // Get form values
  const getFormValues = () => {
    const formValues = formApi ? formApi.getValues() : {};
    return {
      searchKeyword: formValues.searchKeyword || '',
    };
  };

  // Load subscription codes list
  const loadCodes = async (page = 1, pageSize) => {
    setLoading(true);
    try {
      const res = await API.get(
        `/api/subscription_code/?p=${page}&page_size=${pageSize}`,
      );
      const { success, message, data } = res.data;
      if (success) {
        const newPageData = data.items;
        setActivePage(data.page <= 0 ? 1 : data.page);
        setCodeCount(data.total);
        setCodes(newPageData);
      } else {
        showError(message);
      }
    } catch (error) {
      showError(error.message);
    }
    setLoading(false);
  };

  // Search subscription codes
  const searchCodes = async () => {
    const { searchKeyword } = getFormValues();
    if (searchKeyword === '') {
      await loadCodes(1, pageSize);
      return;
    }

    setSearching(true);
    try {
      const res = await API.get(
        `/api/subscription_code/search?keyword=${searchKeyword}&p=1&page_size=${pageSize}`,
      );
      const { success, message, data } = res.data;
      if (success) {
        const newPageData = data.items;
        setActivePage(data.page || 1);
        setCodeCount(data.total);
        setCodes(newPageData);
      } else {
        showError(message);
      }
    } catch (error) {
      showError(error.message);
    }
    setSearching(false);
  };

  // Manage subscription codes (CRUD operations)
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

      const { success, message } = res.data;
      if (success) {
        showSuccess(t('操作成功完成！'));
        let code = res.data.data;
        let newCodes = [...codes];
        if (action !== SUBSCRIPTION_CODE_ACTIONS.DELETE) {
          record.status = code.status;
        }
        setCodes(newCodes);
      } else {
        showError(message);
      }
    } catch (error) {
      showError(error.message);
    }
    setLoading(false);
  };

  // Refresh data
  const refresh = async (page = activePage) => {
    const { searchKeyword } = getFormValues();
    if (searchKeyword === '') {
      await loadCodes(page, pageSize);
    } else {
      await searchCodes();
    }
  };

  // Handle page change
  const handlePageChange = (page) => {
    setActivePage(page);
    const { searchKeyword } = getFormValues();
    if (searchKeyword === '') {
      loadCodes(page, pageSize);
    } else {
      searchCodes();
    }
  };

  // Handle page size change
  const handlePageSizeChange = (size) => {
    setPageSize(size);
    setActivePage(1);
    const { searchKeyword } = getFormValues();
    if (searchKeyword === '') {
      loadCodes(1, size);
    } else {
      searchCodes();
    }
  };

  // Row selection configuration
  const rowSelection = {
    onSelect: (record, selected) => {},
    onSelectAll: (selected, selectedRows) => {},
    onChange: (selectedRowKeys, selectedRows) => {
      setSelectedKeys(selectedRows);
    },
  };

  // Row style handling
  const handleRow = (record, index) => {
    const isExpired = (rec) => {
      return (
        rec.status === SUBSCRIPTION_CODE_STATUS.UNUSED &&
        rec.expired_time !== 0 &&
        rec.expired_time < Math.floor(Date.now() / 1000)
      );
    };

    if (record.status !== SUBSCRIPTION_CODE_STATUS.UNUSED || isExpired(record)) {
      return {
        style: {
          background: 'var(--semi-color-disabled-border)',
        },
      };
    } else {
      return {};
    }
  };

  // Copy text
  const copyText = async (text) => {
    if (await copy(text)) {
      showSuccess('已复制到剪贴板！');
    } else {
      Modal.error({
        title: '无法复制到剪贴板，请手动复制',
        content: text,
        size: 'large',
      });
    }
  };

  // Batch copy codes
  const batchCopyCodes = async () => {
    if (selectedKeys.length === 0) {
      showError(t('请至少选择一个激活码！'));
      return;
    }

    let keys = '';
    for (let i = 0; i < selectedKeys.length; i++) {
      keys += selectedKeys[i].name + '    ' + selectedKeys[i].key + '\n';
    }
    await copyText(keys);
  };

  // Batch delete codes (clear invalid)
  const batchDeleteCodes = async () => {
    Modal.confirm({
      title: t('确定清除所有失效激活码？'),
      content: t('将删除已使用、已禁用及过期的激活码，此操作不可撤销。'),
      onOk: async () => {
        setLoading(true);
        const res = await API.delete('/api/subscription_code/invalid');
        const { success, message, data } = res.data;
        if (success) {
          showSuccess(t('已删除 {{count}} 条失效激活码', { count: data }));
          await refresh();
        } else {
          showError(message);
        }
        setLoading(false);
      },
    });
  };

  // Close edit modal
  const closeEdit = () => {
    setShowEdit(false);
    setTimeout(() => {
      setEditingCode({ id: undefined });
    }, 500);
  };

  // Remove record
  const removeRecord = (key) => {
    let newDataSource = [...codes];
    if (key != null) {
      let idx = newDataSource.findIndex((data) => data.key === key);
      if (idx > -1) {
        newDataSource.splice(idx, 1);
        setCodes(newDataSource);
      }
    }
  };

  // Initialize data loading
  useEffect(() => {
    loadCodes(1, pageSize)
      .then()
      .catch((reason) => {
        showError(reason);
      });
  }, [pageSize]);

  return {
    // Data state
    codes,
    loading,
    searching,
    activePage,
    pageSize,
    codeCount,
    selectedKeys,

    // Edit state
    editingCode,
    showEdit,

    // Form state
    formApi,
    formInitValues,

    // UI state
    compactMode,
    setCompactMode,

    // Data operations
    loadCodes,
    searchCodes,
    manageCode,
    refresh,
    copyText,
    removeRecord,

    // State updates
    setActivePage,
    setPageSize,
    setSelectedKeys,
    setEditingCode,
    setShowEdit,
    setFormApi,
    setLoading,

    // Event handlers
    handlePageChange,
    handlePageSizeChange,
    rowSelection,
    handleRow,
    closeEdit,
    getFormValues,

    // Batch operations
    batchCopyCodes,
    batchDeleteCodes,

    // Translation function
    t,
  };
};
