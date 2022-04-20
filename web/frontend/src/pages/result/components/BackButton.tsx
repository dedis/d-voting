import { FC } from 'react';
import { useTranslation } from 'react-i18next';

type BackButtonProps = {};

const BackButton: FC<BackButtonProps> = () => {
  const { t } = useTranslation();

  return (
    <button
      type="button"
      className="inline-flex mx-4 my-4 items-center px-3 py-2 border border-gray-300 rounded-md shadow-sm text-sm font-medium text-gray-700 bg-white hover:bg-gray-50">
      {t('back')}
    </button>
  );
};

export default BackButton;
