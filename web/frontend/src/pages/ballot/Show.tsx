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
import { useNavigate } from 'react-router';

const Ballot: FC = () => {
  const { t } = useTranslation();
  const navigate = useNavigate();

  const { electionId } = useParams();
  const UserID = sessionStorage.getItem('id');
  const { loading, configObj, electionID, status, pubKey, ballotSize, chunksPerBallot } =
    useElection(electionId);
  const { configuration, answers, setAnswers } = useConfiguration(configObj);
  const [userErrors, setUserErrors] = useState('');
  const edCurve = kyber.curve.newCurve('edwards25519');
  const [postRequest, setPostRequest] = useState(null);
  const [postError, setPostError] = useState('');
  const [showModal, setShowModal] = useState(false);
  const [modalText, setModalText] = useState(t('voteSuccess') as string);
  const [modalTitle, setModalTitle] = useState('');
  const [navigateDest, setNavigateDest] = useState(null);
  const sendFetchRequest = usePostCall(setPostError);

  useEffect(() => {
    if (postRequest !== null) {
      sendFetchRequest(endpoints.newElectionVote(electionID.toString()), postRequest, setShowModal);
    }
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [postRequest]);

  useEffect(() => {
    if (postError !== null) {
      if (postError.includes('ECONNREFUSED')) {
        setModalText(t('errorServerDown'));
      } else {
        setModalText(t('voteFailure'));
      }
      setModalTitle(t('errorTitle'));
    } else {
      setNavigateDest('/');
      setModalText(t('voteSuccess'));
      setModalTitle(t('voteSuccessful'));
    }
  }, [postError, t]);

  const hexToBytes = (hex: string) => {
    const bytes: number[] = [];
    for (let c = 0; c < hex.length; c += 2) {
      bytes.push(parseInt(hex.substr(c, 2), 16));
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
      setPostRequest(newRequest);
    } catch (e) {
      console.log(e);
      setModalText(t('ballotFailure'));
      setModalTitle(t('errorTitle'));

      setShowModal(true);
    }
    navigate(-1);
  };

  const handleClick = () => {
    if (!ballotIsValid(configuration, answers, setAnswers)) {
      setUserErrors(t('incompleteBallot'));
      return;
    }

    setUserErrors('');
    sendBallot();
  };

  const SubjectElementDisplay = (element: types.SubjectElement) => {
    return (
      <div>
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
      <div className="sm:px-8 pl-2" key={subject.ID}>
        {subject.Order.map((id: ID) => (
          <div key={id}>
            {subject.Elements.get(id).Type === SUBJECT ? (
              <div>
                <h3 className="text-lg font-bold text-gray-600">
                  {subject.Elements.get(id).Title}
                </h3>
                {SubjectTree(subject.Elements.get(id) as types.Subject)}
              </div>
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
      <DragDropContext onDragEnd={(dropRes) => handleOnDragEnd(dropRes, answers, setAnswers)}>
        <div className="shadow-lg rounded-md my-0 sm:my-4 py-8 w-full">
          <h3 className="font-bold uppercase py-4 text-2xl text-center text-gray-600">
            {configuration.MainTitle}
          </h3>
          <div>
            {configuration.Scaffold.map((subject: types.Subject) => SubjectTree(subject))}
            <div className="sm:mx-8 mx-4 text-red-600 text-sm pt-3 pb-5">{userErrors}</div>
            <div className="flex sm:mx-8 mx-4">
              <button
                type="button"
                className="inline-flex items-center px-4 py-2 border border-transparent rounded-md shadow-sm text-sm font-medium text-white bg-indigo-500 hover:bg-indigo-600"
                onClick={handleClick}>
                <CloudUploadIcon className="-ml-1 mr-2 h-5 w-5" aria-hidden="true" />
                {t('castVote')}
              </button>
            </div>
          </div>
        </div>
      </DragDropContext>
    );
  };

  return (
    <div>
      <RedirectToModal
        showModal={showModal}
        setShowModal={setShowModal}
        title={modalTitle}
        buttonRightText={t('close')}
        navigateDestination={navigateDest}>
        {modalText}
      </RedirectToModal>
      {loading ? (
        <Loading />
      ) : (
        <div>{status === STATUS.Open ? ballotDisplay() : <ElectionClosed />}</div>
      )}
    </div>
  );
};

Ballot.propTypes = {
  location: PropTypes.any,
};

export default Ballot;
