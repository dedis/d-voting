import React, { FC, useContext, useEffect, useState } from 'react';
import { LightElectionInfo, Status } from 'types/election';
import * as endpoints from 'components/utils/Endpoints';
import { NodeStatus } from 'types/node';
import { FlashContext, FlashLevel } from 'index';
import { Link } from 'react-router-dom';
import ElectionStatus from './ElectionStatus';
import QuickAction from './QuickAction';
import ElectionStatusLoading from './ElectionStatusLoading';
import { ID } from 'types/configuration';

type ElectionRowProps = {
  election: LightElectionInfo;
  electionStatuses: Map<ID, Status>;
  setElectionsStatuses: (status: Map<ID, Status>) => void;
  initialElectionStatuses: Map<ID, Status>;
  setInitialElectionStatuses: (status: Map<ID, Status>) => void;
};

const ElectionRow: FC<ElectionRowProps> = ({
  election,
  electionStatuses,
  setElectionsStatuses,
  initialElectionStatuses,
  setInitialElectionStatuses,
}) => {
  const fctx = useContext(FlashContext);

  const [DKGStatuses, setDKGStatuses] = useState<Map<string, NodeStatus>>(null);
  const [DKGLoading, setDKGLoading] = useState(true);
  const [electionStatus, setElectionStatus] = useState<Status>(null);
  const abortController = new AbortController();
  const signal = abortController.signal;

  const notifyParent = (newStatus: Status) => {
    const newElectionStatuses = new Map(electionStatuses);
    newElectionStatuses.set(election.ElectionID, newStatus);
    // notify parent to update the filter
    setElectionsStatuses(newElectionStatuses);
    setInitialElectionStatuses(newElectionStatuses);
    // notify parent table that the row has updated at least once
    setElectionsStatuses(newElectionStatuses);
    setInitialElectionStatuses(newElectionStatuses);
  };

  // Fetch the NodeStatus of each node
  useEffect(() => {
    if (election !== null) {
      // Only fetch if the nodes haven't been setup yet and the row
      // hasn't updated its status yet
      if (
        election.Status === Status.Initial &&
        initialElectionStatuses.get(election.ElectionID) === undefined
      ) {
        const request = {
          method: 'GET',
          signal: signal,
        };

        const fetchDKGStatus = async (node: string, proxyAddress: string) => {
          try {
            const response = await fetch(
              endpoints.getDKGActors(proxyAddress, election.ElectionID),
              request
            );

            // The node is not initialized
            if (response.status === 404) {
              return { Node: node, Status: NodeStatus.NotInitialized };
            }

            if (!response.ok) {
              const js = await response.json();
              throw new Error(JSON.stringify(js));
            }

            let dkgStatus = await response.json();
            return { Node: node, Status: dkgStatus.Status as NodeStatus };
          } catch (e) {
            if (!(e instanceof DOMException)) {
              fctx.addMessage(e.message, FlashLevel.Error);
            }
            return Promise.reject();
          }
        };

        const fetchProxies = async () => {
          try {
            const response = await fetch(
              endpoints.getProxiesAddresses(election.ElectionID),
              request
            );

            if (!response.ok) {
              const js = await response.json();
              throw new Error(JSON.stringify(js));
            }

            let dataReceived = await response.json();
            const newNodeProxyAddresses = new Map<string, string>();

            dataReceived.Proxies.forEach((value) => {
              Object.entries(value).forEach(([node, proxy]) => {
                newNodeProxyAddresses.set(node, proxy as string);
              });
            });

            return newNodeProxyAddresses;
          } catch (e) {
            if (!(e instanceof DOMException)) {
              fctx.addMessage(e.message, FlashLevel.Error);
            }
            return Promise.reject();
          }
        };

        fetchProxies().then(
          (nodeProxy) => {
            const promises = Array.from(nodeProxy).map(([node, proxy]) => {
              return fetchDKGStatus(node, proxy);
            });

            Promise.all(promises)
              .then(
                (value) => {
                  const newDKGStatuses: Map<string, NodeStatus> = new Map();
                  value.forEach((v) => {
                    newDKGStatuses.set(v.Node, v.Status);
                  });
                  setDKGStatuses(newDKGStatuses);
                },
                () => ({})
              )
              .finally(() => setDKGLoading(false));
          },
          () => ({})
        );
      } else {
        setDKGLoading(false);
      }
    }

    return () => {
      abortController.abort();
    };
  }, [election]);

  // Update de status
  useEffect(() => {
    if (election !== null) {
      var newStatus = election.Status;

      if (election.Status === Status.Initial) {
        if (initialElectionStatuses.get(election.ElectionID) !== undefined) {
          newStatus = initialElectionStatuses.get(election.ElectionID);
        }

        if (DKGStatuses !== null) {
          const dkgStatuses = Array.from(DKGStatuses.values());

          if (!dkgStatuses.includes(NodeStatus.NotInitialized)) {
            newStatus = Status.Initialized;
          }

          if (dkgStatuses.includes(NodeStatus.Setup)) {
            newStatus = Status.Setup;
          }
        }
      }

      setElectionStatus(newStatus);
      notifyParent(newStatus);
    }
  }, [election, DKGStatuses]);

  return (
    <tr className="bg-white border-b hover:bg-gray-50 ">
      <td scope="row" className="px-6 py-4 font-medium text-gray-900 whitespace-nowrap">
        <Link
          className="election-link text-gray-700 hover:text-indigo-500"
          to={`/elections/${election.ElectionID}`}>
          {election.Title}
        </Link>
      </td>
      <td className="px-6 py-4">
        {DKGLoading ? <ElectionStatusLoading /> : <ElectionStatus status={electionStatus} />}
      </td>
      <td className="px-6 py-4 text-right">
        <QuickAction status={election.Status} electionID={election.ElectionID} />
      </td>
    </tr>
  );
};

export default ElectionRow;
