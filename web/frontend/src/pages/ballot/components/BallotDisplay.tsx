import React, { FC } from 'react';
import { DragDropContext } from 'react-beautiful-dnd';
import { Answers, Configuration, ID, RANK, SELECT, SUBJECT, TEXT } from 'types/configuration';
import * as types from 'types/configuration';
import Rank, { handleOnDragEnd } from './Rank';
import { useTranslation } from 'react-i18next';
import Select from './Select';
import Text from './Text';
import SpinnerIcon from 'components/utils/SpinnerIcon';
import { MailIcon } from '@heroicons/react/outline';
import { useNavigate } from 'react-router';

type BallotDisplayProps = {
  configuration: Configuration;
  answers: Answers;
  setAnswers: (answers: Answers) => void;
  userErrors: string;
  handleClick: () => void;
  castVoteLoading: boolean;
};

const BallotDisplay: FC<BallotDisplayProps> = ({
  configuration,
  answers,
  setAnswers,
  userErrors,
  handleClick,
  castVoteLoading,
}) => {
  const { t } = useTranslation();
  const navigate = useNavigate();

  const SubjectElementDisplay = (element: types.SubjectElement) => {
    return (
      <div className="pl-4 sm:pl-6">
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
      <div key={subject.ID}>
        <h3 className="text-xl pt-1 pb-1 sm:pt-2 sm:pb-2 border-t font-bold text-gray-600">
          {subject.Title}
        </h3>
        {subject.Order.map((id: ID) => (
          <div key={id}>
            {subject.Elements.get(id).Type === SUBJECT ? (
              <div className="pl-4 sm:pl-6">
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

  return (
    <div className="w-[60rem] font-sans px-4 pt-8 pb-4">
      <DragDropContext onDragEnd={(dropRes) => handleOnDragEnd(dropRes, answers, setAnswers)}>
        <div className="pb-2">
          <h2 className="text-2xl font-bold leading-7 text-gray-900 sm:text-3xl sm:truncate">
            {t('vote')}
          </h2>
          <div className="mt-2 text-sm text-gray-500">{t('voteExplanation')}</div>
        </div>

        <div className="w-full pb-4 mb-0 sm:mb-4 mt-4 sm:mt-6">
          <h3 className="py-6 border-t-2 text-2xl text-center text-gray-700">
            {configuration.MainTitle}
          </h3>
          <div className="flex flex-col">
            {configuration.Scaffold.map((subject: types.Subject) => SubjectTree(subject))}
            <div className="text-red-600 text-sm pt-3 pb-1">{userErrors}</div>
          </div>

          <div className="flex mt-4">
            <button
              type="button"
              className="inline-flex flex-none items-center mr-2 sm:mr-4 px-4 py-2 border border-transparent rounded-md shadow-sm text-sm font-medium text-white bg-indigo-500 hover:bg-indigo-600"
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
              className="flex-none text-gray-700 mr-2 items-center px-4 py-2 border rounded-md text-sm hover:text-indigo-500">
              {t('back')}
            </button>
          </div>
        </div>
      </DragDropContext>
    </div>
  );
};

export default BallotDisplay;
