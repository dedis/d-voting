import { FC, useEffect, useState } from 'react';
import { useTranslation } from 'react-i18next';
import { useParams } from 'react-router-dom';
import kyber from '@dedis/kyber';
import PropTypes from 'prop-types';
import { Buffer } from 'buffer';

import useElection from 'components/utils/useElection';
import usePostCall from 'components/utils/usePostCall';
import * as endpoints from 'components/utils/Endpoints';
import { encryptVote } from './components/VoteEncrypt';
import { voteEncode } from './components/VoteEncode';
import { useConfiguration } from 'components/utils/useConfiguration';
import * as types from 'types/configuration';
import { ID, RANK, SELECT, SUBJECT, TEXT } from 'types/configuration';
import { DragDropContext } from 'react-beautiful-dnd';
import RedirectToModal from 'components/modal/RedirectToModal';
import Select from './components/Select';
import Rank, { handleOnDragEnd } from './components/Rank';
import Text from './components/Text';
import { ballotIsValid } from './components/ValidateAnswers';
import { STATUS } from 'types/election';
import ElectionClosed from './components/ElectionClosed';
import Loading from 'pages/Loading';
import { CloudUploadIcon } from '@heroicons/react/solid';
import SpinnerIcon from 'components/utils/SpinnerIcon';

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

  const SubjectElementDisplay = (element: types.SubjectElement) => {
    return (
      <div className="pl-4">
        {element.Type === RANK && <Rank rank={element as types.RankQuestion} answers={answers} />}
        {element.Type === SELECT && (
          <Select
            select={element as types.SelectQuestion}
            answers={answers}
            setAnswers={setAnswers}
          />
        )}
        {element.Type === TEXT && (
          <Text text={element as types.TextQuestion} answers={answers} setAnswers={setAnswers} />
        )}
      </div>
    );
  };

  const SubjectTree = (subject: types.Subject) => {
    return (
      <div className="" key={subject.ID}>
        <h3 className="text-xl font-bold text-gray-600">{subject.Title}</h3>
        {subject.Order.map((id: ID) => (
          <div key={id}>
            {subject.Elements.get(id).Type === SUBJECT ? (
              <div className="pl-4">{SubjectTree(subject.Elements.get(id) as types.Subject)}</div>
            ) : (
              SubjectElementDisplay(subject.Elements.get(id))
            )}
          </div>
        ))}
      </div>
    );
  };

  const ballotDisplay = () => {
    return (
      <div className="w-[60rem] font-sans px-4 pt-8 pb-4">
        <DragDropContext onDragEnd={(dropRes) => handleOnDragEnd(dropRes, answers, setAnswers)}>
          <div className="flex items-center">
            <h2 className="text-2xl font-bold leading-7 text-gray-900 sm:text-3xl sm:truncate">
              {t('vote')}
            </h2>
          </div>

          <div className="w-full pb-4 my-0 sm:my-4">
            <h3 className="py-6 text-2xl text-center text-gray-700">{configuration.MainTitle}</h3>
            <div className="flex flex-col">
              {configuration.Scaffold.map((subject: types.Subject) => SubjectTree(subject))}
              <div className="text-red-600 text-sm pt-3 pb-1">{userErrors}</div>
            </div>
            <div className="flex mt-4">
              <button
                type="button"
                className="inline-flex items-center px-4 py-2 border border-transparent rounded-md shadow-sm text-sm font-medium text-white bg-indigo-500 hover:bg-indigo-600"
                onClick={handleClick}>
                {castVoteLoading ? (
                  <SpinnerIcon />
                ) : (
                  <CloudUploadIcon className="-ml-1 mr-2 h-5 w-5" aria-hidden="true" />
                )}
                {t('castVote')}
              </button>
            </div>
          </div>
        </DragDropContext>
      </div>
    );
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
      {loading ? <Loading /> : <>{status === STATUS.Open ? ballotDisplay() : <ElectionClosed />}</>}
    </>
  );
};

Ballot.propTypes = {
  location: PropTypes.any,
};

export default Ballot;
