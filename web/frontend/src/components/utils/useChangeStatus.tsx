import React from 'react';
import { useTranslation } from 'react-i18next';
import { STATUS } from 'types/electionInfo';

// Custom hook that can display the status of an election and enable changes
// of status (closing, cancelling,...)
const useChangeStatus = (status: STATUS) => {
  const { t } = useTranslation();

  const getStatus = () => {
    switch (status) {
      case STATUS.INITIAL:
        return (
          <span className="election-status">
            <span className="election-status-on"></span>
            <span className="election-status-text">{t('statusInitial')}</span>
          </span>
        );
      case STATUS.OPEN:
        return (
          <span className="election-status">
            <span className="election-status-on"></span>
            <span className="election-status-text">{t('statusOpen')}</span>
          </span>
        );
      case STATUS.CLOSED:
        return (
          <span className="election-status">
            <span className="election-status-closed"></span>
            <span className="election-status-text">{t('statusClose')}</span>
          </span>
        );
      case STATUS.SHUFFLED_BALLOTS:
        return (
          <span className="election-status">
            <span className="election-status-closed"></span>
            <span className="election-status-text">{t('statusShuffle')}</span>
          </span>
        );
      case STATUS.RESULT_AVAILABLE:
        return (
          <span className="election-status">
            <span className="election-status-closed"></span>
            <span className="election-status-text">{t('resultsAvailable')}</span>
          </span>
        );
      case STATUS.CANCELED:
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
