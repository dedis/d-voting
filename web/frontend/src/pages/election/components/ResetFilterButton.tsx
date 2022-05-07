import React, { FC } from 'react';
import { useTranslation } from 'react-i18next';
import { Status } from 'types/election';

type ResetFilterButtonProps = {
  setStatusToKeep: (status: Status) => void;
};

const ResetFilterButton: FC<ResetFilterButtonProps> = ({ setStatusToKeep }) => {
  const { t } = useTranslation();

  return (
    <button
      type="button"
      onClick={() => setStatusToKeep(null)}
      className="text-gray-700 my-2 mx-2 items-center px-4 py-2 border border-transparent rounded-md shadow-sm text-sm shadow-md font-medium hover:text-white hover:bg-indigo-500">
      {t('resetFilter')}
    </button>
  );
};

export default ResetFilterButton;
