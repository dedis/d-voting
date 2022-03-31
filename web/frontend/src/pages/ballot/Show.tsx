import { FC, useEffect, useState } from 'react';
import { CloudUploadIcon } from '@heroicons/react/outline';
import { useTranslation } from 'react-i18next';
import { Link, useParams } from 'react-router-dom';
import kyber from '@dedis/kyber';
import PropTypes from 'prop-types';
import { Buffer } from 'buffer';

import { ROUTE_BALLOT_INDEX } from '../../Routes';
import useElection from 'components/utils/useElection';
import usePostCall from 'components/utils/usePostCall';
import { ENDPOINT_EVOTING_CAST_BALLOT } from 'components/utils/Endpoints';
import Modal from 'components/modal/Modal';
import { OPEN } from 'components/utils/StatusNumber';
import { encryptVote } from './components/VoteEncrypt';
import { ballotIsValid } from './components/HandleAnswers';
import { handleOnDragEnd } from './components/RankDisplay';
import { voteEncode } from './components/VoteEncode';
import useConfiguration from 'components/utils/useConfiguration';
import { ID, Question, RANK, ROOT_ID, SELECT, SUBJECT, TEXT } from 'types/configuration';
import { DragDropContext } from 'react-beautiful-dnd';

const Ballot: FC = () => {
  const { t } = useTranslation();
  const { electionId } = useParams();
  const token = sessionStorage.getItem('token');
  const { loading, configuration, electionID, status, pubKey, ballotSize } = useElection(
    electionId,
    token
  );
  const { sortedQuestions, answers, setAnswers } = useConfiguration(configuration);
  const [userErrors, setUserErrors] = useState('');
  const edCurve = kyber.curve.newCurve('edwards25519');
  const [postRequest, setPostRequest] = useState(null);
  const [postError, setPostError] = useState('');
  const { postData } = usePostCall(setPostError);
  const [showModal, setShowModal] = useState(false);
  const [modalText, setModalText] = useState(t('voteSuccess') as string);

  useEffect(() => {
    if (postRequest !== null) {
      setPostError(null);
      postData(ENDPOINT_EVOTING_CAST_BALLOT, postRequest, setShowModal);
      setPostRequest(null);
    }
  }, [postData, postRequest]);

  useEffect(() => {
    if (postError !== null) {
      if (postError.includes('ECONNREFUSED')) {
        setModalText(t('errorServerDown'));
      } else {
        setModalText(t('voteFailure'));
      }
    } else {
      setModalText(t('voteSuccess'));
    }
  }, [postError, t]);

  const hexToBytes = (hex: string) => {
    let bytes: number[] = [];
    for (let c = 0; c < hex.length; c += 2) {
      bytes.push(parseInt(hex.substr(c, 2), 16));
    }
    return new Uint8Array(bytes);
  };

  const createBallot = (EGPairs: Array<Buffer[]>) => {
    let vote = [];
    EGPairs.forEach(([K, C]) => vote.push({ K: Array.from(K), C: Array.from(C) }));
    return {
      ElectionID: electionID,
      UserId: sessionStorage.getItem('id'),
      Ballot: vote,
      Token: token,
    };
  };

  const sendBallot = async () => {
    let ballotChunks = voteEncode(answers, ballotSize);
    let EGPairs = Array<Buffer[]>();
    ballotChunks.forEach((chunk) =>
      EGPairs.push(encryptVote(chunk, Buffer.from(hexToBytes(pubKey).buffer), edCurve))
    );
    //sending the ballot to evoting server
    let ballot = createBallot(EGPairs);
    let newRequest = {
      method: 'POST',
      body: JSON.stringify(ballot),
    };
    setPostRequest(newRequest);
  };

  const handleClick = () => {
    if (ballotIsValid(sortedQuestions, answers, setAnswers)) {
      setUserErrors('');
      sendBallot();
    } else {
      setUserErrors(t('incompleteBallot'));
    }
  };

  const subjectTree = (sorted: Question[], parentId: ID) => {
    const questions = sorted.filter((question) => question.ParentID === parentId);
    if (!questions.length) {
      return null;
    }

    return (
      <div>
        {questions.map((question) => (
          <div key={question.Content.ID} className="px-8">
            {question.Type === SUBJECT ? question.render(question) : null}
            {question.Type === RANK ? question.render(question, answers) : null}
            {question.Type === TEXT ? question.render(question, answers, setAnswers) : null}
            {question.Type === SELECT ? question.render(question, answers, setAnswers) : null}
            <div>{question.Type === SUBJECT ? subjectTree(sorted, question.Content.ID) : null}</div>
          </div>
        ))}
      </div>
    );
  };

  const ballotDisplay = () => {
    return (
      <DragDropContext onDragEnd={(e) => handleOnDragEnd(e, answers, setAnswers)}>
        <div className="shadow-lg rounded-md my-4 py-8">
          <h3 className="font-bold uppercase py-4 text-2xl text-center text-gray-600">
            {configuration.MainTitle}
          </h3>
          <div>
            {subjectTree(sortedQuestions, ROOT_ID)}
            <div className="mx-8 text-red-600 text-sm pt-3 pb-5">{userErrors}</div>
            <div className="flex mx-8">
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

  const electionClosedDisplay = () => {
    return (
      <div>
        <div> {t('voteImpossible')}</div>
        <Link to={ROUTE_BALLOT_INDEX}>
          <button
            type="button"
            className="inline-flex mt-2 mb-2 ml-2 items-center px-4 py-2 border border-transparent rounded-md shadow-sm text-sm font-medium text-white bg-indigo-500 hover:bg-indigo-600">
            {t('back')}
          </button>
        </Link>
      </div>
    );
  };

  return (
    <div>
      <Modal
        showModal={showModal}
        setShowModal={setShowModal}
        textModal={modalText}
        buttonRightText={t('close')}
      />
      {loading ? (
        <p className="loading">{t('loading')}</p>
      ) : (
        <div>{status === OPEN ? ballotDisplay() : electionClosedDisplay()}</div>
      )}
    </div>
  );
};

Ballot.propTypes = {
  location: PropTypes.any,
};

export default Ballot;
