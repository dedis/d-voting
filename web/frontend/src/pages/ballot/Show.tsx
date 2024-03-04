import { FC, useContext, useState } from 'react';
import { AuthContext } from 'index';
import { isVoter } from './../../utils/auth';
import { useTranslation } from 'react-i18next';
import { useParams } from 'react-router-dom';
import kyber from '@dedis/kyber';
import PropTypes from 'prop-types';
import { Buffer } from 'buffer';

import SpinnerIcon from 'components/utils/SpinnerIcon';
import { MailIcon } from '@heroicons/react/outline';
import { useNavigate } from 'react-router';

import useForm from 'components/utils/useForm';
import * as endpoints from 'components/utils/Endpoints';
import { encryptVote } from './components/VoteEncrypt';
import { voteEncode } from './components/VoteEncode';
import { useConfiguration } from 'components/utils/useConfiguration';
import { Status } from 'types/form';
import { ballotIsValid } from './components/ValidateAnswers';
import BallotDisplay from './components/BallotDisplay';
import FormNotAvailable from './components/FormNotAvailable';
import Loading from 'pages/Loading';
import RedirectToModal from 'components/modal/RedirectToModal';
import { default as i18n } from 'i18next';

const Ballot: FC = () => {
  const { t } = useTranslation();

  const { formId } = useParams();
  const UserID = sessionStorage.getItem('id');
  const { loading, configObj, formID, status, pubKey, ballotSize, chunksPerBallot } =
    useForm(formId);
  const { configuration, answers, setAnswers } = useConfiguration(configObj);

  const [userErrors, setUserErrors] = useState('');
  const edCurve = kyber.curve.newCurve('edwards25519');
  const [showModal, setShowModal] = useState(false);
  const [modalText, setModalText] = useState(t('voteSuccess') as string);
  const [modalTitle, setModalTitle] = useState('');
  const [castVoteLoading, setCastVoteLoading] = useState(false);

  const navigate = useNavigate();
  const { authorization, isLogged } = useContext(AuthContext);

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
      try {
        const response = await fetch(endpoints.newFormVote(formID.toString()), newRequest);
        if (!response.ok) {
          const txt = await response.text();
          throw new Error(txt);
        }
        setModalText(t('voteSuccess'));
        setModalTitle(t('voteSuccessful'));
      } catch (error) {
        if (error.message.includes('ECONNREFUSED')) {
          setModalText(t('errorServerDown'));
        } else {
          setModalText(t('voteFailure'));
        }
        setModalTitle(t('errorTitle'));
      }

      setShowModal((prev) => !prev);
    } catch (e) {
      console.log(e);
      setModalText(t('ballotFailure'));
      setModalTitle(t('errorTitle'));

      setShowModal(true);
    }
    setCastVoteLoading(false);
  };

  const handleClick = (event) => {
    if (!ballotIsValid(configuration, answers, setAnswers)) {
      setUserErrors(t('incompleteBallot'));
      return;
    }
    setCastVoteLoading(true);

    setUserErrors('');
    sendBallot();
    event.currentTarget.disabled = true;
  };

  const userIsVoter = isVoter(formID, authorization, isLogged);

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
          {status === Status.Open && userIsVoter && (
            <div className="w-[60rem] font-sans px-4 pt-8 pb-4">
              <div className="pb-2">
                <h2 className="text-2xl font-bold leading-7 text-gray-900 sm:text-3xl sm:truncate">
                  {t('vote')}
                </h2>
                <div className="mt-2 text-sm text-gray-500">{t('voteExplanation')}</div>
              </div>
              <div className="border-t mt-3" />
              <BallotDisplay
                configuration={configuration}
                answers={answers}
                setAnswers={setAnswers}
                userErrors={userErrors}
                language={i18n.language}
              />

              <div className="flex mb-4 sm:mb-6">
                <button
                  type="button"
                  className="inline-flex items-center mr-2 sm:mr-4 px-4 py-2 border border-transparent rounded-md shadow-sm text-sm font-medium text-white bg-[#ff0000] hover:bg-[#ff0000]"
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
                  className=" text-gray-700 mr-2 items-center px-4 py-2 border rounded-md text-sm hover:text-[#ff0000]">
                  {t('back')}
                </button>
              </div>
            </div>
          )}
          {!userIsVoter && <FormNotAvailable isVoter={false} />}
          {status !== Status.Open && <FormNotAvailable isVoter={true} />}
        </>
      )}
    </>
  );
};

Ballot.propTypes = {
  location: PropTypes.any,
};

export default Ballot;
