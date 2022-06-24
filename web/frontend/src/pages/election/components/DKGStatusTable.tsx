import { FC, useEffect, useState } from 'react';
import { useTranslation } from 'react-i18next';
import { ID } from 'types/configuration';
import { NodeStatus } from 'types/node';
import DKGStatusRow from './DKGStatusRow';

const NODES_PER_PAGE = 5;

type DKGStatusTableProps = {
  roster: string[];
  electionId: ID;
  loading: Map<string, boolean>;
  setLoading: (loading: Map<string, boolean>) => void;
  nodeProxyAddresses: Map<string, string>;
  setNodeProxyAddresses: (nodeProxy: Map<string, string>) => void;
  DKGStatuses: Map<string, NodeStatus>;
  setDKGStatuses: (DKFStatuses: Map<string, NodeStatus>) => void;
  setTextModalError: (error: string) => void;
  setShowModalError: (show: boolean) => void;
};

const DKGStatusTable: FC<DKGStatusTableProps> = ({
  roster,
  electionId,
  loading,
  setLoading,
  nodeProxyAddresses,
  setNodeProxyAddresses,
  DKGStatuses,
  setDKGStatuses,
  setTextModalError,
  setShowModalError,
}) => {
  const { t } = useTranslation();

  const [nodesToDisplay, setNodesToDisplay] = useState([]);
  const [pageIndex, setPageIndex] = useState(0);

  const partitionArray = (array: string[], size: number) =>
    array.map((_v, i) => (i % size === 0 ? array.slice(i, i + size) : null)).filter((v) => v);

  useEffect(() => {
    if (roster.length) {
      setNodesToDisplay(partitionArray(roster, NODES_PER_PAGE)[pageIndex]);
    }
  }, [roster, pageIndex]);

  const handlePrevious = (): void => {
    if (pageIndex > 0) {
      setPageIndex(pageIndex - 1);
    }
  };

  const handleNext = (): void => {
    if (partitionArray(roster, NODES_PER_PAGE).length > pageIndex + 1) {
      setPageIndex(pageIndex + 1);
    }
  };

  return (
    <div>
      <div className="relative divide-y overflow-x-auto shadow-md sm:rounded-lg mt-2">
        <table className="w-full text-sm text-left text-gray-500">
          <thead className="text-xs text-gray-700 uppercase bg-gray-50">
            <tr>
              <th scope="col" className="px-6 py-3">
                {t('node')}
              </th>
              <th scope="col" className="px-6 py-3">
                {t('status')}
              </th>
            </tr>
          </thead>
          <tbody>
            {nodesToDisplay !== undefined &&
              nodesToDisplay.map((node, index) => (
                <DKGStatusRow
                  key={index}
                  electionId={electionId}
                  node={node}
                  index={index}
                  loading={loading}
                  setLoading={setLoading}
                  nodeProxyAddresses={nodeProxyAddresses}
                  setNodeProxyAddresses={setNodeProxyAddresses}
                  DKGStatuses={DKGStatuses}
                  setDKGStatuses={setDKGStatuses}
                  setTextModalError={setTextModalError}
                  setShowModalError={setShowModalError}
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
              m: partitionArray(roster, NODES_PER_PAGE).length,
              x: roster.length,
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
              disabled={partitionArray(roster, NODES_PER_PAGE).length <= pageIndex + 1}
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

export default DKGStatusTable;
