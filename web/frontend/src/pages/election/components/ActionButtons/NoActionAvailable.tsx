import { MinusIcon } from '@heroicons/react/outline';
import { useTranslation } from 'react-i18next';

const NoActionAvailable = () => {
  const { t } = useTranslation();

  return (
    <div className="whitespace-nowrap inline-flex items-center justify-center px-4 py-1 mr-2 border border-gray-300 text-sm rounded-full font-medium text-gray-700">
      <MinusIcon className="-ml-1 mr-2 h-5 w-5" aria-hidden="true" />
      {t('actionNotAvailable')}
    </div>
  );
};

export default NoActionAvailable;
