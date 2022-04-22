import React, { FC } from 'react';
import TextButton from 'components/buttons/TextButton';
import { useTranslation } from 'react-i18next';
import { Link } from 'react-router-dom';
import { ROUTE_BALLOT_SHOW, ROUTE_ELECTION_INDEX } from 'Routes';
import { Configuration, ID } from 'types/configuration';
import { STATUS } from 'types/electionInfo';
import Action from '../../election/components/Action';
import Status from '../../election/components/Status';

type ResultNotAvailableProps = {
  status: STATUS;
  setStatus: (status: STATUS) => void;
  setIsResultAvailable: (isResultAvailable: boolean) => void;
  configuration: Configuration;
  electionID: ID;
};

const ResultNotAvailable: FC<ResultNotAvailableProps> = ({
  status,
  setStatus,
  setIsResultAvailable,
  configuration,
  electionID,
}) => {
  const { t } = useTranslation();

  return (
    <>
      <div className="shadow-lg rounded-md w-full px-4 my-0 sm:my-4">
        <h3 className="py-6 uppercase text-2xl text-center text-gray-700">
          {configuration.MainTitle}
        </h3>
        <div className="px-4">
          {t('status')}: <Status status={status} />
          <span className="mx-4">{t('action')}:</span>
          <Action
            status={status}
            electionID={electionID}
            setStatus={setStatus}
            setResultAvailable={setIsResultAvailable}
          />
        </div>
      </div>
      <div className="flex my-4">
        {status === STATUS.OPEN ? (
          <Link to={ROUTE_BALLOT_SHOW + '/' + electionID}>
            <TextButton>{t('navBarVote')}</TextButton>
          </Link>
        ) : null}
        <Link to={ROUTE_ELECTION_INDEX}>
          <TextButton>{t('back')}</TextButton>
        </Link>
      </div>
    </>
  );
};

export default ResultNotAvailable;
