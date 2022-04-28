import React from 'react';
import { useTranslation } from 'react-i18next';
import { STATUS } from 'types/electionInfo';

// Custom hook that can display the status of an election and enable changes
// of status (closing, cancelling,...)
const useChangeStatus = (status: STATUS) => {
  const { t } = useTranslation();

  const getStatus = () => {
    switch (status) {
      case STATUS.Initial:
        return (
          <span className="election-status">
            <span className="election-status-on"></span>
            <span className="election-status-text">{t('statusInitial')}</span>
          </span>
        );
      case STATUS.Open:
        return (
          <span className="election-status">
            <span className="election-status-on"></span>
            <span className="election-status-text">{t('statusOpen')}</span>
          </span>
        );
      case STATUS.Closed:
        return (
          <span className="election-status">
            <span className="election-status-closed"></span>
            <span className="election-status-text">{t('statusClose')}</span>
          </span>
        );
      case STATUS.ShuffledBallots:
        return (
          <span className="election-status">
            <span className="election-status-closed"></span>
            <span className="election-status-text">{t('statusShuffle')}</span>
          </span>
        );
      case STATUS.ResultAvailable:
        return (
          <span className="election-status">
            <span className="election-status-closed"></span>
            <span className="election-status-text">{t('resultsAvailable')}</span>
          </span>
        );
      case STATUS.Canceled:
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
