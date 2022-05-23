import { MinusIcon } from '@heroicons/react/outline';
import { useTranslation } from 'react-i18next';
import { NodeStatus } from 'types/node';

// Custom hook that can display the status of an election and enable changes
// of status (closing, cancelling,...)
const useChangeDKGStatus = (status: NodeStatus) => {
  const { t } = useTranslation();

  const getDKGStatus = () => {
    switch (status) {
      case NodeStatus.NotInitialized:
        return (
          <div className="flex items-center">
            <div>
              <MinusIcon className="ml-2 mr-2 h-5 w-5 text-gray-600" aria-hidden="true" />
            </div>
          </div>
        );
      case NodeStatus.Initialized:
        return (
          <div className="flex items-center">
            <div className="block h-4 w-4 bg-green-500 rounded-full mr-2"></div>
            <div>{t('initialized')}</div>
          </div>
        );
      case NodeStatus.Setup:
        return (
          <div className="flex items-center">
            <div className="block h-4 w-4 bg-green-500 rounded-full mr-2"></div>
            <div>{t('statusSetup')}</div>
          </div>
        );
      case NodeStatus.Failed:
        return (
          <div className="flex items-center">
            <div className="block h-4 w-4 bg-red-500 rounded-full mr-2"></div>
            <div>{t('failed')}</div>
          </div>
        );
      default:
        return null;
    }
  };
  return { getDKGStatus };
};

export default useChangeDKGStatus;
