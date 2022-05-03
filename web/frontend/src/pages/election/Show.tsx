import React, { FC, useEffect, useState } from 'react';
import { useParams } from 'react-router-dom';
import PropTypes from 'prop-types';
import { useTranslation } from 'react-i18next';

import useElection from 'components/utils/useElection';
import useGetResults from 'components/utils/useGetResults';
import { STATUS } from 'types/election';
import Action from './components/Action';
import StatusTimeline from './components/StatusTimeline';
import Loading from 'pages/Loading';

const ElectionShow: FC = () => {
  const { t } = useTranslation();
  const { electionId } = useParams();

  const { loading, electionID, status, setStatus, setResult, configObj, setIsResultSet } =
    useElection(electionId);

  const [, setError] = useState(null);
  const [isResultAvailable, setIsResultAvailable] = useState(false);
  const { getResults } = useGetResults();

  //Fetch result when available after a status change
  useEffect(() => {
    if (status === STATUS.ResultAvailable && isResultAvailable) {
      getResults(electionID, setError, setResult, setIsResultSet);
    }
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [isResultAvailable, status]);

  return (
    <div className="w-[60rem] font-sans px-4 py-4">
      {!loading ? (
        <>
          <h2 className="text-2xl font-bold leading-7 text-gray-900 sm:text-3xl sm:truncate">
            {configObj.MainTitle}
          </h2>

          <h2>Election ID : {electionId}</h2>
          <div className="py-6 pl-2">
            <div className="font-bold uppercase text-lg text-gray-700">{t('status')}</div>

            <div className="px-2 pt-6 flex justify-center">
              <StatusTimeline status={status} />
            </div>
          </div>
          <div className="py-4 pl-2 pb-8">
            <div className="font-bold uppercase text-lg text-gray-700 pb-2">{t('action')}</div>
            <div className="px-2">
              <Action
                status={status}
                electionID={electionID}
                setStatus={setStatus}
                setResultAvailable={setIsResultAvailable}
              />
            </div>
          </div>
        </>
      ) : (
        <Loading />
      )}
    </div>
  );
};

ElectionShow.propTypes = {
  location: PropTypes.any,
};

export default ElectionShow;
