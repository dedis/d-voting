import React from 'react';
import { useTranslation } from 'react-i18next';

import { OPEN, CLOSED, SHUFFLED_BALLOT, RESULT_AVAILABLE, CANCELED } from './StatusNumber';

/*Custom hook that can display the status of an election and enable changes of status (closing, cancelling,...)*/
const useChangeStatus = (status: number) => {
  const { t } = useTranslation();

  const getStatus = () => {
    switch (status) {
      case OPEN:
        return (
          <span className="election-status">
            <span className="election-status-on"></span>
            <span className="election-status-text">{t('statusOpen')}</span>
          </span>
        );
      case CLOSED:
        return (
          <span className="election-status">
            <span className="election-status-closed"></span>
            <span className="election-status-text">{t('statusClose')}</span>
          </span>
        );
      case SHUFFLED_BALLOT:
        return (
          <span className="election-status">
            <span className="election-status-closed"></span>
            <span className="election-status-text">{t('statusShuffle')}</span>
          </span>
        );
      case RESULT_AVAILABLE:
        return (
          <span className="election-status">
            <span className="election-status-closed"></span>
            <span className="election-status-text">{t('resultsAvailable')}</span>
          </span>
        );
      case CANCELED:
        return (
          <span className="election-status">
            <span className="election-status-cancelled"></span>
            <span className="election-status-text">{t('statusCancel')}</span>
          </span>
        );
      default:
        return null;
    }
  };
  return { getStatus };
};

export default useChangeStatus;
