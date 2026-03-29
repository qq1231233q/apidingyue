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
import CardPro from '../../common/ui/CardPro';
import SubscriptionCodesTable from './SubscriptionCodesTable';
import SubscriptionCodesActions from './SubscriptionCodesActions';
import SubscriptionCodesFilters from './SubscriptionCodesFilters';
import SubscriptionCodesDescription from './SubscriptionCodesDescription';
import EditSubscriptionCodeModal from './modals/EditSubscriptionCodeModal';
import { useSubscriptionCodesData } from '../../../hooks/subscription_codes/useSubscriptionCodesData';
import { useIsMobile } from '../../../hooks/common/useIsMobile';
import { createCardProPagination } from '../../../helpers/utils';

const SubscriptionCodesPage = () => {
  const codesData = useSubscriptionCodesData();
  const isMobile = useIsMobile();

  const {
    showEdit,
    editingCode,
    closeEdit,
    refresh,
    selectedKeys,
    setEditingCode,
    setShowEdit,
    batchCopyCodes,
    batchDeleteCodes,
    formInitValues,
    setFormApi,
    searchCodes,
    loading,
    searching,
    compactMode,
    setCompactMode,
    t,
  } = codesData;

  return (
    <>
      <EditSubscriptionCodeModal
        refresh={refresh}
        editingCode={editingCode}
        visiable={showEdit}
        handleClose={closeEdit}
      />

      <CardPro
        type='type1'
        descriptionArea={
          <SubscriptionCodesDescription
            compactMode={compactMode}
            setCompactMode={setCompactMode}
            t={t}
          />
        }
        actionsArea={
          <div className='flex flex-col md:flex-row justify-between items-center gap-2 w-full'>
            <SubscriptionCodesActions
              selectedKeys={selectedKeys}
              setEditingCode={setEditingCode}
              setShowEdit={setShowEdit}
              batchCopyCodes={batchCopyCodes}
              batchDeleteCodes={batchDeleteCodes}
              t={t}
            />

            <div className='w-full md:w-full lg:w-auto order-1 md:order-2'>
              <SubscriptionCodesFilters
                formInitValues={formInitValues}
                setFormApi={setFormApi}
                searchCodes={searchCodes}
                loading={loading}
                searching={searching}
                t={t}
              />
            </div>
          </div>
        }
        paginationArea={createCardProPagination({
          currentPage: codesData.activePage,
          pageSize: codesData.pageSize,
          total: codesData.codeCount,
          onPageChange: codesData.handlePageChange,
          onPageSizeChange: codesData.handlePageSizeChange,
          isMobile: isMobile,
          t: codesData.t,
        })}
        t={codesData.t}
      >
        <SubscriptionCodesTable {...codesData} />
      </CardPro>
    </>
  );
};

export default SubscriptionCodesPage;