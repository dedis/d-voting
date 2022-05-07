import { useTranslation } from 'react-i18next';
import { STATUS } from 'types/election';

// Custom hook that can display the status of an election and enable changes
// of status (closing, cancelling,...)
const useChangeStatus = (status: STATUS) => {
  const { t } = useTranslation();

  const getStatus = () => {
    console.log('status: ' + status);
    switch (status) {
      case STATUS.Initial:
        return (
          <div className="flex">
            <div className="block h-4 w-4 bg-green-500 rounded-full mr-2"></div>
            <div>{t('statusInitial')}</div>
          </div>
        );
      case STATUS.Initialized:
        return (
          <div className="flex">
            <div className="block h-4 w-4 bg-green-500 rounded-full mr-2"></div>
            <div>{t('statusInitializedNodes')}</div>
          </div>
        );
      case STATUS.OnGoingSetup:
        return (
          <div className="flex">
            {/*<span className="election-status">
            <span className="election-status-on"></span>
            <span className="election-status-text">{t('statusInitializedNodes')}</span>
        </span>*/}

            <div className="block h-4 w-4 bg-green-500 rounded-full mr-2"></div>
            <div>{t('statusOnGoingSetup')}</div>
          </div>
        );
      case STATUS.Setup:
        return (
          <div className="flex">
            <div className="block h-4 w-4 bg-green-500 rounded-full mr-2"></div>
            <div>{t('statusSetup')}</div>
          </div>
        );
      case STATUS.Open:
        return (
          <div className="flex">
            <div className="block h-4 w-4 bg-green-500 rounded-full mr-2"></div>
            <div>{t('statusOpen')}</div>
          </div>
        );
      case STATUS.Closed:
        return (
          <div className="flex">
            <div className="block h-4 w-4 bg-gray-400 rounded-full mr-2"></div>
            <div>{t('statusClose')}</div>
          </div>
        );
      case STATUS.OnGoingShuffle:
        return (
          <div className="flex">
            {/*<span className="election-status">
            <span className="election-status-closed"></span>
            <span className="election-status-text">{t('statusClose')}</span>
        </span>*/}

            <div className="block h-4 w-4 bg-gray-400 rounded-full mr-2"></div>
            <div>{t('statusOnGoingShuffle')}</div>
          </div>
        );
      case STATUS.ShuffledBallots:
        return (
          <div className="flex">
            <div className="block h-4 w-4 bg-gray-400 rounded-full mr-2"></div>
            <div>{t('statusShuffle')}</div>
          </div>
        );
      case STATUS.OnGoingDecryption:
        return (
          <div className="flex">
            {/*<span className="election-status">
            <span className="election-status-closed"></span>
            <span className="election-status-text">{t('statusShuffle')}</span>
        </span>*/}

            <div className="block h-4 w-4 bg-gray-400 rounded-full mr-2"></div>
            <div>{t('statusOnGoingDecryption')}</div>
          </div>
        );
      case STATUS.PubSharesSubmitted:
        return (
          <div className="flex">
            <div className="block h-4 w-4 bg-gray-400 rounded-full mr-2"></div>
            <div>{t('statusDecrypted')}</div>
          </div>
        );
      case STATUS.ResultAvailable:
        return (
          <div className="flex">
            <div className="block h-4 w-4 bg-gray-400 rounded-full mr-2"></div>
            <div>{t('statusResultAvailable')}</div>
          </div>
        );
      case STATUS.Canceled:
        return (
          <div className="flex">
            <div className="block h-4 w-4 bg-red-500 rounded-full mr-2"></div>
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
