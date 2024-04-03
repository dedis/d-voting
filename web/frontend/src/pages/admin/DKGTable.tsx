import React, { FC, useEffect, useState } from 'react';
import { useTranslation } from 'react-i18next';
import AddProxyModal from './components/AddProxyModal';
import EditProxyModal from './components/EditProxyModal';
import RemoveProxyModal from './components/RemoveProxyModal';
import ProxyRow from './ProxyRow';

export const NODE_PROXY_PER_PAGE = 5;

type DKGTableProps = {
  nodeProxyAddresses: Map<string, string>;
  setNodeProxyAddresses: (nodeProxyAddress: Map<string, string>) => void;
};

const DKGTable: FC<DKGTableProps> = ({ nodeProxyAddresses, setNodeProxyAddresses }) => {
  const { t } = useTranslation();
  const [nodeProxyToDisplay, setNodeProxyToDisplay] = useState<Array<[string, string]>>([]);
  const [pageIndex, setPageIndex] = useState(0);
  const [showAddProxy, setShowAddProxy] = useState(false);
  const [showEditProxy, setShowEditProxy] = useState(false);
  const [showDeleteProxy, setShowDeleteProxy] = useState(false);
  const [nodeToEdit, setNodeToEdit] = useState(null);

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

  const handleAddProxy = (node: string, proxy: string) => {
    const newNodeProxy = new Map(nodeProxyAddresses);
    newNodeProxy.set(node, proxy);
    setNodeProxyAddresses(newNodeProxy);

    setPageIndex(partitionMap(newNodeProxy, NODE_PROXY_PER_PAGE).length - 1);
  };

  const handleDeleteProxy = () => {
    const newNodeProxy = new Map(nodeProxyAddresses);
    newNodeProxy.delete(nodeToEdit);
    setNodeProxyAddresses(newNodeProxy);

    if (newNodeProxy.size % NODE_PROXY_PER_PAGE === 0) {
      setPageIndex(pageIndex - 1);
    }
  };
  const handleEditProxy = (node: string, proxy: string) => {
    const newNodeProxy = new Map(nodeProxyAddresses);
    newNodeProxy.delete(nodeToEdit);
    newNodeProxy.set(node, proxy);
    setNodeProxyAddresses(newNodeProxy);

    setPageIndex(partitionMap(newNodeProxy, NODE_PROXY_PER_PAGE).length - 1);
  };

  return (
    <div>
      <AddProxyModal
        open={showAddProxy}
        setOpen={setShowAddProxy}
        handleAddProxy={handleAddProxy}
      />

      <EditProxyModal
        open={showEditProxy}
        setOpen={setShowEditProxy}
        nodeProxy={nodeProxyAddresses}
        setNodeProxy={setNodeProxyAddresses}
        node={nodeToEdit}
        handleEditProxy={handleEditProxy}
      />

      <RemoveProxyModal
        open={showDeleteProxy}
        setOpen={setShowDeleteProxy}
        node={nodeToEdit}
        handleDeleteProxy={handleDeleteProxy}
      />

      <div className="flex flex-col sm:flex-row sm:items-center justify-between mb-4 py-6 pl-2">
        <div>
          <div className="font-bold uppercase text-lg text-gray-700">{t('nodes')}</div>

          <div className="mt-1 flex flex-col sm:flex-row sm:flex-wrap sm:mt-0 sm:space-x-6">
            <div className="mt-2 flex items-center text-sm text-gray-500">{t('nodeDetails')}</div>
          </div>
        </div>

        <div className="mt-5 flex lg:mt-0 lg:ml-4">
          <span className="sm:ml-3">
            <button
              type="button"
              onClick={() => setShowAddProxy(true)}
              className="inline-flex items-center px-4 py-2 border border-transparent rounded-md shadow-sm text-sm font-medium text-white bg-[#ff0000] hover:bg-[#b51f1f]">
              {t('addProxy')}
            </button>
          </span>
        </div>
      </div>

      <div className="relative divide-y overflow-x-auto shadow-md sm:rounded-lg mt-2">
        <table className="w-full text-sm text-left text-gray-500">
          <thead className="text-xs text-gray-700 uppercase bg-gray-50">
            <tr>
              <th scope="col" className="px-6 py-3">
                {t('node')}
              </th>
              <th scope="col" className="px-6 py-3">
                {t('proxy')}
              </th>
              <th scope="col" className=" px-2 py-3">
                <span className="sr-only">{t('edit')}</span>
              </th>
              <th scope="col" className=" px-2 py-3">
                <span className="sr-only">{t('delete')}</span>
              </th>
            </tr>
          </thead>
          <tbody>
            {nodeProxyToDisplay !== undefined &&
              nodeProxyToDisplay.map(([node, proxy], index) => (
                <ProxyRow
                  node={node}
                  proxy={proxy}
                  index={index + pageIndex * NODE_PROXY_PER_PAGE}
                  setShowEditProxy={setShowEditProxy}
                  setShowDeleteProxy={setShowDeleteProxy}
                  setNodeToEdit={setNodeToEdit}
                  key={node}
                />
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
              disabled={
                partitionMap(nodeProxyAddresses, NODE_PROXY_PER_PAGE).length <= pageIndex + 1
              }
              onClick={handleNext}
              className="ml-3 relative inline-flex items-center px-4 py-2 border border-gray-300 text-sm font-medium rounded-md text-gray-700 bg-white hover:bg-gray-50">
              {t('next')}
            </button>
          </div>
        </nav>
      </div>
    </div>
  );
};

export default DKGTable;
