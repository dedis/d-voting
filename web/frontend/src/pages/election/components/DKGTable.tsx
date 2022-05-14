import usePostCall from 'components/utils/usePostCall';
import React, { FC, useEffect, useState } from 'react';
import { useTranslation } from 'react-i18next';
import { NodeStatus } from 'types/node';
import DKGStatus from './DKGStatus';
import * as endpoints from '../../../components/utils/Endpoints';
import { ID } from 'types/configuration';

type DKGTableProps = {
  nodeProxyAddresses: Map<string, string>;
  setNodeProxyAddresses: React.Dispatch<React.SetStateAction<Map<string, string>>>;
  DKGStatuses: Map<string, NodeStatus>;
  electionID: ID;
  setTextModalError: (error: string) => void;
  setShowModalError: (show: boolean) => void;
};

const DKGTable: FC<DKGTableProps> = ({
  nodeProxyAddresses,
  setNodeProxyAddresses,
  DKGStatuses,
  electionID,
  setTextModalError,
  setShowModalError,
}) => {
  const { t } = useTranslation();
  const [proxyAddresses, setProxyAddresses] = useState<Map<string, string>>(null);
  const [prevProxyAddress, setPrevProxyAddress] = useState<Map<string, string>>(null);
  const [isEditMode, setIsEditMode] = useState<Map<string, boolean>>(null);
  const [errors, setErrors] = useState<Map<string, string>>(null);
  const [isPosting, setIsPosting] = useState(false);
  const [postError, setPostError] = useState(null);
  const sendFetchRequest = usePostCall(setPostError);

  const proxyAddressUpdate = async () => {
    const newAddresses = [];
    proxyAddresses.forEach((proxy, node) => newAddresses.push({ [node]: proxy }));

    console.log(newAddresses);
    const req = {
      method: 'PUT',
      body: JSON.stringify({
        Proxies: newAddresses,
      }),
      headers: {
        'Content-Type': 'application/json',
      },
    };
    return sendFetchRequest(endpoints.editProxiesAddresses(electionID), req, setIsPosting);
  };

  useEffect(() => {
    if (postError !== null) {
      setTextModalError(postError);
      setShowModalError(true);
      setPostError(null);
    }
  }, [postError]);

  useEffect(() => {
    if (nodeProxyAddresses !== null) {
      setProxyAddresses(nodeProxyAddresses);
      setPrevProxyAddress(nodeProxyAddresses);
      const newIsEdit = new Map();
      const newErrors = new Map();
      nodeProxyAddresses.forEach((proxy, node) => {
        newIsEdit.set(node, false);
        newErrors.set(node, null);
      });
      setIsEditMode(newIsEdit);
      setErrors(newErrors);
    }
  }, [nodeProxyAddresses]);

  const handleTextInput = (e: React.ChangeEvent<HTMLInputElement>, node: string) => {
    const newAddresses = new Map(proxyAddresses);
    newAddresses.set(node, e.target.value);
    setProxyAddresses(newAddresses);

    const newError = new Map(errors);
    newError.set(node, null);
    setErrors(newError);
  };

  const handleEdit = (node: string) => {
    const editNode = new Map(isEditMode);
    editNode.set(node, true);
    setIsEditMode(editNode);
  };

  const handleSave = (node: string) => {
    const newError = new Map(errors);
    if (proxyAddresses && proxyAddresses.get(node) !== '') {
      const editNode = new Map(isEditMode);
      editNode.set(node, false);
      setIsEditMode(editNode);
      newError.set(node, null);
      setPrevProxyAddress(proxyAddresses);
      setNodeProxyAddresses(proxyAddresses);

      proxyAddressUpdate();
    } else {
      newError.set(node, t('inputProxyAddressError'));
    }
    setErrors(newError);
  };

  const handleCancel = (node: string) => {
    const cancel = new Map(prevProxyAddress);
    setProxyAddresses(cancel);
    const editNode = new Map(isEditMode);
    editNode.set(node, false);
    setIsEditMode(editNode);

    const newError = new Map(errors);
    newError.set(node, null);
    setErrors(newError);
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
              {t('status')}
            </th>
            <th scope="col" className="px-6 py-3">
              {t('proxies')}
            </th>
            <th scope="col" className="px-6 py-3">
              <span className="sr-only">Edit</span>
            </th>
          </tr>
        </thead>
        <tbody>
          <>
            {proxyAddresses !== null && DKGStatuses !== null
              ? Array.from(proxyAddresses).map((node, index) => (
                  <tr key={node[0]} className="bg-white border-b">
                    <td
                      scope="row"
                      className="px-6 py-4 font-medium text-gray-900 whitespace-nowrap">
                      {t('node')} {index}:
                    </td>
                    <td className="px-6 py-4">{<DKGStatus status={DKGStatuses.get(node[0])} />}</td>
                    <td className="px-6 py-4">
                      {isEditMode.get(node[0]) ? (
                        <>
                          <input
                            type="text"
                            className="flex-auto sm:text-md border rounded-md text-gray-600"
                            onChange={(e) => handleTextInput(e, node[0])}
                            placeholder={proxyAddresses.get(node[0]) === '' ? 'https:// ...' : ''}
                            value={proxyAddresses.get(node[0])}
                          />
                          <>
                            {errors.get(node[0]) !== null ? (
                              <div className="text-red-600 text-sm py-2">{errors.get(node[0])}</div>
                            ) : null}
                          </>
                        </>
                      ) : (
                        <>{proxyAddresses.get(node[0])}</>
                      )}
                    </td>
                    <td className="px-6 py-4 text-right">
                      {isEditMode.get(node[0]) ? (
                        <div>
                          <button
                            onClick={() => handleSave(node[0])}
                            className="font-medium text-indigo-600 hover:underline mx-2">
                            {t('save')}
                          </button>
                          <button
                            onClick={() => handleCancel(node[0])}
                            className="font-medium text-indigo-600 hover:underline">
                            {t('cancel')}
                          </button>
                        </div>
                      ) : (
                        <button
                          onClick={() => handleEdit(node[0])}
                          className="font-medium text-indigo-600 hover:underline">
                          {t('edit')}
                        </button>
                      )}
                    </td>
                  </tr>
                ))
              : null}
          </>
        </tbody>
      </table>
    </div>
  );
};

export default DKGTable;
