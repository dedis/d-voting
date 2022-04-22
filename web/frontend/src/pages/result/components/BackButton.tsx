import { FC } from 'react';
import { useTranslation } from 'react-i18next';

type BackButtonProps = {};

const BackButton: FC<BackButtonProps> = () => {
  const { t } = useTranslation();

  return (
    <button
      type="button"
      className="text-gray-700 my-2 mx-2 items-center px-4 py-2 border border-transparent rounded-md shadow-sm text-sm shadow-md font-medium hover:text-white hover:bg-indigo-500">
      {t('back')}
    </button>
  );
};

export default BackButton;
