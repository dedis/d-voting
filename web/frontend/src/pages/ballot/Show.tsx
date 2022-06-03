import { FC, useEffect, useState } from 'react';
import { useTranslation } from 'react-i18next';
import { useParams } from 'react-router-dom';
import kyber from '@dedis/kyber';
import PropTypes from 'prop-types';
import { Buffer } from 'buffer';

import SpinnerIcon from 'components/utils/SpinnerIcon';
import { MailIcon } from '@heroicons/react/outline';
import { useNavigate } from 'react-router';

import useElection from 'components/utils/useElection';
import usePostCall from 'components/utils/usePostCall';
import * as endpoints from 'components/utils/Endpoints';
import { encryptVote } from './components/VoteEncrypt';
import { voteEncode } from './components/VoteEncode';
import { useConfiguration } from 'components/utils/useConfiguration';
import { Status } from 'types/election';
import { ballotIsValid } from './components/ValidateAnswers';
import BallotDisplay from './components/BallotDisplay';
import ElectionClosed from './components/ElectionClosed';
import Loading from 'pages/Loading';
import RedirectToModal from 'components/modal/RedirectToModal';

const Ballot: FC = () => {
  const { t } = useTranslation();

  const { electionId } = useParams();
  const UserID = sessionStorage.getItem('id');
  const { loading, configObj, electionID, status, pubKey, ballotSize, chunksPerBallot } =
    useElection(electionId);
  const { configuration, answers, setAnswers } = useConfiguration(configObj);

  const [userErrors, setUserErrors] = useState('');
  const edCurve = kyber.curve.newCurve('edwards25519');
  const [postError, setPostError] = useState('');
  const [showModal, setShowModal] = useState(false);
  const [modalText, setModalText] = useState(t('voteSuccess') as string);
  const [modalTitle, setModalTitle] = useState('');
  const [castVoteLoading, setCastVoteLoading] = useState(false);
  const sendFetchRequest = usePostCall(setPostError);

  const navigate = useNavigate();

  useEffect(() => {
    if (postError !== null) {
      if (postError.includes('ECONNREFUSED')) {
        setModalText(t('errorServerDown'));
      } else {
        setModalText(t('voteFailure'));
      }
      setModalTitle(t('errorTitle'));
    } else {
      setModalText(t('voteSuccess'));
      setModalTitle(t('voteSuccessful'));
    }
  }, [postError, t]);

  const hexToBytes = (hex: string) => {
    const bytes: number[] = [];
    for (let c = 0; c < hex.length; c += 2) {
      bytes.push(parseInt(hex.substring(c, c + 2), 16));
    }
    return new Uint8Array(bytes);
  };

  const createBallot = (EGPairs: Array<Buffer[]>) => {
    const vote = [];
    EGPairs.forEach(([K, C]) => vote.push({ K: Array.from(K), C: Array.from(C) }));
    return {
      Ballot: vote,
      UserID,
    };
  };

  const sendBallot = async () => {
    try {
      const ballotChunks = voteEncode(answers, ballotSize, chunksPerBallot);
      const EGPairs = Array<Buffer[]>();
      ballotChunks.forEach((chunk) =>
        EGPairs.push(encryptVote(chunk, Buffer.from(hexToBytes(pubKey).buffer), edCurve))
      );
      //sending the ballot to evoting server
      const ballot = createBallot(EGPairs);
      const newRequest = {
        method: 'POST',
        body: JSON.stringify(ballot),
        headers: {
          'Content-Type': 'Application/json',
        },
      };
      await sendFetchRequest(
        endpoints.newElectionVote(electionID.toString()),
        newRequest,
        setShowModal
      );
    } catch (e) {
      console.log(e);
      setModalText(t('ballotFailure'));
      setModalTitle(t('errorTitle'));

      setShowModal(true);
    }
    setCastVoteLoading(false);
  };

  const handleClick = () => {
    if (!ballotIsValid(configuration, answers, setAnswers)) {
      setUserErrors(t('incompleteBallot'));
      return;
    }
    setCastVoteLoading(true);

    setUserErrors('');
    sendBallot();
  };

  return (
    <>
      <RedirectToModal
        showModal={showModal}
        setShowModal={setShowModal}
        title={modalTitle}
        buttonRightText={t('close')}
        navigateDestination={-1}>
        {modalText}
      </RedirectToModal>
      {loading ? (
        <Loading />
      ) : (
        <>
          {status === Status.Open && (
            <div className="w-[60rem] font-sans px-4 pt-8 pb-4">
              <div className="pb-2">
                <h2 className="text-2xl font-bold leading-7 text-gray-900 sm:text-3xl sm:truncate">
                  {t('vote')}
                </h2>
                <div className="mt-2 text-sm text-gray-500">{t('voteExplanation')}</div>
              </div>

              <BallotDisplay
                configuration={configuration}
                answers={answers}
                setAnswers={setAnswers}
                userErrors={userErrors}
              />

              <div className="flex mb-4 sm:mb-6">
                <button
                  type="button"
                  className="inline-flex items-center mr-2 sm:mr-4 px-4 py-2 border border-transparent rounded-md shadow-sm text-sm font-medium text-white bg-indigo-500 hover:bg-indigo-600"
                  onClick={handleClick}>
                  {castVoteLoading ? (
                    <SpinnerIcon />
                  ) : (
                    <MailIcon className="-ml-1 mr-2 h-5 w-5" aria-hidden="true" />
                  )}
                  {t('castVote')}
                </button>
                <button
                  type="button"
                  onClick={() => navigate(-1)}
                  className=" text-gray-700 mr-2 items-center px-4 py-2 border rounded-md text-sm hover:text-indigo-500">
                  {t('back')}
                </button>
              </div>
            </div>
          )}
          {status !== Status.Open && <ElectionClosed />}
        </>
      )}
    </>
  );
};

Ballot.propTypes = {
  location: PropTypes.any,
};

export default Ballot;
