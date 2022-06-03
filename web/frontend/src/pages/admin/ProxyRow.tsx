import React, { FC } from 'react';
import { useTranslation } from 'react-i18next';

type ProxyRowProps = {
  node: string;
  proxy: string;
  index: number;
  setShowEditProxy: (show: boolean) => void;
  setShowDeleteProxy: (show: boolean) => void;
  setNodeToEdit: (node: string) => void;
};

const ProxyRow: FC<ProxyRowProps> = ({
  node,
  proxy,
  index,
  setShowEditProxy,
  setShowDeleteProxy,
  setNodeToEdit,
}) => {
  const { t } = useTranslation();

  const handleEdit = () => {
    setShowEditProxy(true);
    setNodeToEdit(node);
  };

  const handleDelete = () => {
    setShowDeleteProxy(true);
    setNodeToEdit(node);
  };

  return (
    <tr className="bg-white border-b hover:bg-gray-50">
      <td scope="row" className="px-6 py-4 font-medium text-gray-900 whitespace-nowrap">
        {node}
      </td>
      <td className="px-6 py-4">{proxy}</td>
      <td className="px-6 py-4 text-right">
        <button
          onClick={() => handleEdit()}
          className="font-medium text-indigo-600 hover:underline mr-6">
          {t('edit')}
        </button>
        <button onClick={() => handleDelete()} className="font-medium text-red-600 hover:underline">
          {t('delete')}
        </button>
      </td>
    </tr>
  );
};

export default ProxyRow;
