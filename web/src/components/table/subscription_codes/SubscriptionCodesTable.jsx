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

import React, { useMemo, useState } from 'react';
import { Empty } from '@douyinfe/semi-ui';
import CardTable from '../../common/ui/CardTable';
import {
  IllustrationNoResult,
  IllustrationNoResultDark,
} from '@douyinfe/semi-illustrations';
import { getSubscriptionCodesColumns } from './SubscriptionCodesColumnDefs';
import DeleteSubscriptionCodeModal from './modals/DeleteSubscriptionCodeModal';

const SubscriptionCodesTable = (codesData) => {
  const {
    codes,
    loading,
    activePage,
    pageSize,
    codeCount,
    compactMode,
    handlePageChange,
    handlePageSizeChange,
    rowSelection,
    handleRow,
    manageCode,
    copyText,
    setEditingCode,
    setShowEdit,
    refresh,
    t,
  } = codesData;

  const [showDeleteModal, setShowDeleteModal] = useState(false);
  const [deletingRecord, setDeletingRecord] = useState(null);

  const showDeleteCodeModal = (record) => {
    setDeletingRecord(record);
    setShowDeleteModal(true);
  };

  const columns = useMemo(
    () =>
      getSubscriptionCodesColumns({
        t,
        manageCode,
        copyText,
        setEditingCode,
        setShowEdit,
        showDeleteCodeModal,
      }),
    [t, manageCode, copyText, setEditingCode, setShowEdit],
  );

  const tableColumns = useMemo(() => {
    return compactMode
      ? columns.map((col) => {
          if (col.dataIndex === 'operate') {
            const { fixed, ...rest } = col;
            return rest;
          }
          return col;
        })
      : columns;
  }, [compactMode, columns]);

  return (
    <>
      <CardTable
        columns={tableColumns}
        dataSource={codes}
        scroll={compactMode ? undefined : { x: 'max-content' }}
        pagination={{
          currentPage: activePage,
          pageSize,
          total: codeCount,
          showSizeChanger: true,
          pageSizeOptions: [10, 20, 50, 100],
          onPageSizeChange: handlePageSizeChange,
          onPageChange: handlePageChange,
        }}
        hidePagination={true}
        loading={loading}
        rowSelection={rowSelection}
        onRow={handleRow}
        empty={
          <Empty
            image={<IllustrationNoResult style={{ width: 150, height: 150 }} />}
            darkModeImage={
              <IllustrationNoResultDark style={{ width: 150, height: 150 }} />
            }
            description={t('暂无激活码记录')}
            style={{ padding: 30 }}
          />
        }
        className='rounded-xl overflow-hidden'
        size='middle'
      />

      <DeleteSubscriptionCodeModal
        visible={showDeleteModal}
        onCancel={() => setShowDeleteModal(false)}
        record={deletingRecord}
        manageCode={manageCode}
        refresh={refresh}
        codes={codes}
        activePage={activePage}
        t={t}
      />
    </>
  );
};

export default SubscriptionCodesTable;
