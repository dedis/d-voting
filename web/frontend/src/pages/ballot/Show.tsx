import React, { FC, useEffect, useState } from 'react';
import { CloudUploadIcon } from '@heroicons/react/outline';
import { useTranslation } from 'react-i18next';
import { Link, useParams } from 'react-router-dom';
import kyber from '@dedis/kyber';
import PropTypes from 'prop-types';
import { Buffer } from 'buffer';
import { DragDropContext, Droppable } from 'react-beautiful-dnd';

import { ROUTE_BALLOT_INDEX } from '../../Routes';
import useElection from 'components/utils/useElection';
import usePostCall from 'components/utils/usePostCall';
import { ENDPOINT_EVOTING_CAST_BALLOT } from 'components/utils/Endpoints';
import Modal from 'components/modal/Modal';
import { OPEN } from 'components/utils/StatusNumber';
import { encryptVote } from './components/VoteEncrypt';
import useConfiguration, {
  Error,
  Question,
  RANK,
  ROOT_ID,
  RankAnswer,
  SELECT,
  SUBJECT,
  SelectAnswer,
  TEXT,
} from 'components/utils/useConfiguration';
import { ballotIsValid } from './HandleAnswers';
import { selectDisplay, selectHintDisplay } from './ShowSelects';
import { handleOnDragEnd, rankDisplay } from './ShowRanks';
import { textDisplay, textHintDisplay } from './ShowTexts';
import { ID, Rank, Select, Text } from 'components/utils/types';
import voteEncode from './components/VoteEncode';

const Ballot: FC = () => {
  const { t } = useTranslation();
  const { electionId } = useParams();
  const token = sessionStorage.getItem('token');
  const { loading, configuration, electionID, status, pubKey } = useElection(electionId, token);
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

  const createBallot = (K: Buffer, C: Buffer) => {
    let vote = JSON.stringify({ K: Array.from(K), C: Array.from(C) });
    return {
      ElectionID: electionID,
      UserId: sessionStorage.getItem('id'),
      Ballot: Buffer.from(vote),
      Token: token,
    };
  };

  const sendBallot = async () => {
    let encodedAnswers = voteEncode(answers);
    const [K, C] = encryptVote(encodedAnswers, Buffer.from(hexToBytes(pubKey).buffer), edCurve);
    //sending the ballot to evoting server
    let ballot = createBallot(K, C);
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
          <div className="px-8">
            <div>
              {question.Type === SUBJECT ? (
                <h3 className="font-bold text-lg text-gray-600">{question.Content.Title}</h3>
              ) : (
                <h3 className="text-lg text-gray-600">{question.Content.Title}</h3>
              )}
            </div>
            <div>
              {question.Type === SELECT ? (
                <div>
                  {selectHintDisplay(question.Content as Select)}
                  <div className="pl-8">
                    {Array.from(
                      answers.SelectAnswers.find(
                        (s: SelectAnswer) => s.ID === question.Content.ID
                      ).Answers.entries()
                    ).map(([choiceIndex, isChecked]) =>
                      selectDisplay(
                        isChecked,
                        (question.Content as Select).Choices[choiceIndex],
                        question.Content as Select,
                        answers,
                        setAnswers
                      )
                    )}
                  </div>
                </div>
              ) : question.Type === RANK ? (
                <div className="mt-5 pl-8">
                  <Droppable droppableId={String(question.Content.ID)}>
                    {(provided) => (
                      <ul
                        className={question.Content.ID}
                        {...provided.droppableProps}
                        ref={provided.innerRef}>
                        {Array.from(
                          answers.RankAnswers.find(
                            (r: RankAnswer) => r.ID === question.Content.ID
                          ).Answers.entries()
                        ).map(([rankIndex, choiceIndex]) =>
                          rankDisplay(
                            rankIndex,
                            (question.Content as Rank).Choices[choiceIndex],
                            question.Content as Rank,
                            answers,
                            setAnswers
                          )
                        )}
                        {provided.placeholder}
                      </ul>
                    )}
                  </Droppable>
                </div>
              ) : question.Type === TEXT ? (
                <div>
                  {textHintDisplay(question.Content as Text)}
                  <div className="pl-8">
                    {(question.Content as Text).Choices.map((choice) =>
                      textDisplay(choice, question.Content as Text, answers, setAnswers)
                    )}
                  </div>
                </div>
              ) : null}
            </div>
            <div>
              {question.Type === SUBJECT ? (
                subjectTree(sorted, question.Content.ID)
              ) : (
                <div className="text-red-600 text-sm py-2 pl-2">
                  {answers.Errors.find((e: Error) => e.ID === question.Content.ID).Message}
                </div>
              )}
            </div>
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
            <div className="mx-8 text-red-600 text-sm py-2">{userErrors}</div>
            <div className="flex mx-8">
              <button
                type="button"
                className="flex inline-flex items-center px-4 py-2 border border-transparent rounded-md shadow-sm text-sm font-medium text-white bg-indigo-500 hover:bg-indigo-600"
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
            className="flex inline-flex mt-2 mb-2 ml-2 items-center px-4 py-2 border border-transparent rounded-md shadow-sm text-sm font-medium text-white bg-indigo-500 hover:bg-indigo-600">
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
/*<div className="mx-4 my-4 px-8 py-4 shadow-lg rounded-md">*/

Ballot.propTypes = {
  location: PropTypes.any,
};

export default Ballot;
