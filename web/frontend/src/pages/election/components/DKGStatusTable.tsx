import { FC } from 'react';
import { useTranslation } from 'react-i18next';
import { ID } from 'types/configuration';
import { OngoingAction } from 'types/election';
import { InternalDKGInfo, NodeStatus } from 'types/node';
import DKGStatusRow from './DKGStatusRow';

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
  // notify to start initialization
  ongoingAction: OngoingAction;
  // notify the parent of the new state
  notifyDKGState: (node: string, info: InternalDKGInfo) => void;
  nodeToSetup: [string, string];
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
  ongoingAction,
  notifyDKGState,
  nodeToSetup,
}) => {
  const { t } = useTranslation();

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
            {roster !== null &&
              roster.map((node, index) => (
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
                  ongoingAction={ongoingAction}
                  notifyDKGState={notifyDKGState}
                  nodeToSetup={nodeToSetup}
                />
              ))}
          </tbody>
        </table>
      </div>
    </div>
  );
};

export default DKGStatusTable;
