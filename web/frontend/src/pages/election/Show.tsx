import React, { FC, useEffect, useState } from 'react';
import { useParams } from 'react-router-dom';
import PropTypes from 'prop-types';
import { useTranslation } from 'react-i18next';

import useElection from 'components/utils/useElection';
import './Show.css';
import useGetResults from 'components/utils/useGetResults';
import { STATUS } from 'types/election';
import Status from './components/Status';
import Action from './components/Action';
import StatusTimeline from './components/StatusTimeline';

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
    <div className="w-full md:w-[60rem] px-4 py-6">
      {!loading ? (
        <>
          <div className="border px-8 border-gray-200 w-full md:px-10 my-0">
            <div className="pt-6 font-bold uppercase text-2xl text-gray-700">
              {configObj.MainTitle}
            </div>
            <h2>Election ID : {electionId}</h2>
            <div className="py-6 pl-2">
              <div className="font-bold uppercase text-lg text-gray-700">Workflow</div>

              <div className="px-2">
                {t('status')}: <Status status={status} />
                <StatusTimeline status={status} />
              </div>
            </div>
            <div className="py-6 pl-2 pb-8">
              <div className="font-bold uppercase text-lg text-gray-700">{t('action')}</div>
              <div className="px-2">
                <Action
                  status={status}
                  electionID={electionID}
                  setStatus={setStatus}
                  setResultAvailable={setIsResultAvailable}
                />
              </div>
            </div>
          </div>
        </>
      ) : (
        <p className="loading">{t('loading')}</p>
      )}
    </div>
  );
};

ElectionShow.propTypes = {
  location: PropTypes.any,
};

export default ElectionShow;
