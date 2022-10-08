import { useTranslation } from 'react-i18next';
import { Status } from 'types/form';

// Custom hook that can display the status of an form and enable changes
// of status (closing, cancelling,...)
const useChangeStatus = (status: Status) => {
  const { t } = useTranslation();

  const getStatus = () => {
    switch (status) {
      case Status.Initial:
        return (
          <div className="flex">
            <div className="h-4 px-2 bg-green-500 rounded-full mr-2" />
            <div>{t('statusInitial')}</div>
          </div>
        );
      case Status.Initialized:
        return (
          <div className="flex">
            <div className="h-4 px-2 bg-green-500 rounded-full mr-2" />
            <div>{t('statusInitializedNodes')}</div>
          </div>
        );
      case Status.Setup:
        return (
          <div className="flex">
            <div className="h-4 px-2 bg-green-500 rounded-full mr-2" />
            <div>{t('statusSetup')}</div>
          </div>
        );
      case Status.Open:
        return (
          <div className="flex">
            <div className="h-4 px-2 bg-green-500 rounded-full mr-2" />
            <div>{t('statusOpen')}</div>
          </div>
        );
      case Status.Closed:
        return (
          <div className="flex">
            <div className="h-4 px-2 bg-gray-400 rounded-full mr-2" />
            <div>{t('statusClose')}</div>
          </div>
        );
      case Status.ShuffledBallots:
        return (
          <div className="flex">
            <div className="h-4 px-2 bg-gray-400 rounded-full mr-2" />
            <div>{t('statusShuffle')}</div>
          </div>
        );
      case Status.PubSharesSubmitted:
        return (
          <div className="flex">
            <div className="h-4 px-2 bg-gray-400 rounded-full mr-2" />
            <div>{t('statusDecrypted')}</div>
          </div>
        );
      case Status.ResultAvailable:
        return (
          <div className="flex">
            <div className="h-4 px-2 bg-gray-400 rounded-full mr-2" />
            <div>{t('statusResultAvailable')}</div>
          </div>
        );
      case Status.Canceled:
        return (
          <div className="flex">
            <div className="h-4 px-2 bg-red-500 rounded-full mr-2" />
            <div>{t('statusCancel')}</div>
          </div>
        );
      default:
        return null;
    }
  };
  return { getStatus };
};

export default useChangeStatus;
