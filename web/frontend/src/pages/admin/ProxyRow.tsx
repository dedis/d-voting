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
      <td className="px-6 py-4 font-medium text-gray-900 whitespace-nowrap break-all">{node}</td>
      <td className="px-6 py-4 break-all">{proxy}</td>

      <td className="px-2 py-4 text-right">
        <div className="block sm:hidden">
          <button onClick={handleEdit} className="font-medium text-[#ff0000] hover:underline ">
            {t('edit')}
          </button>
        </div>
      </td>
      <td className="sm:flex px-2 py-4 sm:flex-row-reverse">
        <button onClick={handleDelete} className="font-medium text-red-600 hover:underline mr-2">
          {t('delete')}
        </button>
        <div className="hidden sm:block">
          <button onClick={handleEdit} className="font-medium text-[#ff0000] hover:underline mr-4">
            {t('edit')}
          </button>
        </div>
      </td>
    </tr>
  );
};

export default ProxyRow;
