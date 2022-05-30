import React, { FC, useEffect, useState } from 'react';
import { useTranslation } from 'react-i18next';
import ProxyRow from './ProxyRow';

const NODE_PROXY_PER_PAGE = 5;

type DKGTableProps = {
  nodeProxyAddresses: Map<string, string>;
};

const DKGTable: FC<DKGTableProps> = ({ nodeProxyAddresses }) => {
  const { t } = useTranslation();
  const [nodeProxyToDisplay, setNodeProxyToDisplay] = useState<Array<[string, string]>>([]);
  const [pageIndex, setPageIndex] = useState(0);

  const partitionMap = (nodeProxy: Map<string, string>, size: number): [string, string][][] => {
    const array: [string, string][] = Array.from(nodeProxy);
    return array
      .map((_value, index) => (index % size === 0 ? array.slice(index, index + size) : null))
      .filter((v) => v);
  };

  useEffect(() => {
    if (nodeProxyAddresses.size) {
      setNodeProxyToDisplay(partitionMap(nodeProxyAddresses, NODE_PROXY_PER_PAGE)[pageIndex]);
    }
  }, [nodeProxyAddresses, pageIndex]);

  const handlePrevious = (): void => {
    if (pageIndex > 0) {
      setPageIndex(pageIndex - 1);
    }
  };

  const handleNext = (): void => {
    if (partitionMap(nodeProxyAddresses, NODE_PROXY_PER_PAGE).length > pageIndex + 1) {
      setPageIndex(pageIndex + 1);
    }
  };

  return (
    <div className="relative divide-y overflow-x-auto shadow-md sm:rounded-lg mt-2">
      <table className="w-full text-sm text-left text-gray-500">
        <thead className="text-xs text-gray-700 uppercase bg-gray-50">
          <tr>
            <th scope="col" className="px-6 py-3">
              {t('nodes')}
            </th>
            <th scope="col" className="px-6 py-3">
              {t('proxies')}
            </th>
            <th scope="col" className="px-6 py-3">
              <span className="sr-only">{t('edit')}</span>
            </th>
          </tr>
        </thead>
        <tbody>
          {nodeProxyToDisplay !== undefined &&
            nodeProxyToDisplay.map(([node, proxy], index) => (
              <ProxyRow node={node} proxy={proxy} index={index} />
            ))}
        </tbody>
      </table>

      <nav
        className="bg-white px-4 py-3 flex items-center justify-between border-t border-gray-200 sm:px-6"
        aria-label="Pagination">
        <div className="hidden sm:block text-sm text-gray-700">
          {t('showingNOverMOfXResults', {
            n: pageIndex + 1,
            m: partitionMap(nodeProxyAddresses, NODE_PROXY_PER_PAGE).length,
            x: `${nodeProxyAddresses !== null ? nodeProxyAddresses.size : 0}`,
          })}
        </div>
        <div className="flex-1 flex justify-between sm:justify-end">
          <button
            disabled={pageIndex === 0}
            onClick={handlePrevious}
            className="relative inline-flex items-center px-4 py-2 border border-gray-300 text-sm font-medium rounded-md text-gray-700 bg-white hover:bg-gray-50">
            {t('previous')}
          </button>
          <button
            disabled={partitionMap(nodeProxyAddresses, NODE_PROXY_PER_PAGE).length <= pageIndex + 1}
            onClick={handleNext}
            className="ml-3 relative inline-flex items-center px-4 py-2 border border-gray-300 text-sm font-medium rounded-md text-gray-700 bg-white hover:bg-gray-50">
            {t('next')}
          </button>
        </div>
      </nav>
    </div>
  );
};

export default DKGTable;
