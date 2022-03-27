import { buildAnswer, getIndices } from './HandleAnswers';
import { Draggable, DropResult } from 'react-beautiful-dnd';
import { Answers, Error, RANK, RankAnswer } from 'components/utils/useConfiguration';
import { t } from 'i18next';
import { Rank } from 'components/utils/types';

const reorderRankAnswers = (
  sourceIndex: number,
  destinationIndex: number,
  questionIndex: number,
  answers: Answers,
  setAnswers: React.Dispatch<React.SetStateAction<Answers>>
) => {
  const [reorderedItem] = answers.RankAnswers[questionIndex].Answers.splice(sourceIndex, 1);
  answers.RankAnswers[questionIndex].Answers.splice(destinationIndex, 0, reorderedItem);
  setAnswers(answers);
};

const handleOnDragEnd = (
  result: DropResult,
  answers: Answers,
  setAnswers: React.Dispatch<React.SetStateAction<Answers>>
) => {
  if (!result.destination) {
    return;
  }
  let questionIndex = answers.RankAnswers.findIndex(
    (r: RankAnswer) => r.ID === result.destination.droppableId
  );
  let errorIndex = answers.Errors.findIndex((e: Error) => e.ID === result.destination.droppableId);
  let newAnswers = buildAnswer(answers);

  newAnswers.Errors[errorIndex].Message = '';
  reorderRankAnswers(
    result.source.index,
    result.destination.index,
    questionIndex,
    newAnswers,
    setAnswers
  );
};

const handleRankInput = (
  e: React.ChangeEvent<HTMLInputElement>,
  question: Rank,
  choice: string,
  rankIndex: number,
  answers: Answers,
  setAnswers: React.Dispatch<React.SetStateAction<Answers>>
) => {
  let { questionIndex, errorIndex, newAnswers } = getIndices(question, choice, answers, RANK);

  newAnswers.Errors[errorIndex].Message = '';
  let destinationIndex = +e.target.value;
  if (e.target.value !== '') {
    if (destinationIndex > question.MaxN || destinationIndex <= 0) {
      newAnswers.Errors[errorIndex].Message = t('rankRange') + question.MaxN;
      setAnswers(newAnswers);
    } else {
      reorderRankAnswers(rankIndex, destinationIndex - 1, questionIndex, newAnswers, setAnswers);
    }
  }
  e.target.value = '';
};

const RankList = () => {
  return (
    <svg
      className="flex-none mx-3"
      width="20"
      height="16"
      viewBox="0 0 20 16"
      fill="none"
      xmlns="http://www.w3.org/2000/svg">
      <rect y="4" width="20" height="2" rx="1" fill="#C4C4C4" />
      <rect y="10" width="20" height="2" rx="1" fill="#C4C4C4" />
      <rect y="7" width="20" height="2" rx="1" fill="#C4C4C4" />
      <path
        d="M9.82259 0.0877967C9.93439 0.0324678 10.0656 0.0324677 10.1774 0.0877967L14.5294 2.2415C14.9086 2.4291 14.775 3 14.352 3H5.64796C5.22498 3 5.09145 2.4291 5.47055 2.2415L9.82259 0.0877967Z"
        fill="#C4C4C4"
      />
      <path
        d="M10.1774 15.9122C10.0656 15.9675 9.93439 15.9675 9.82259 15.9122L5.47055 13.7585C5.09145 13.5709 5.22498 13 5.64796 13H14.352C14.775 13 14.9086 13.5709 14.5295 13.7585L10.1774 15.9122Z"
        fill="#C4C4C4"
      />
    </svg>
  );
};

const rankDisplay = (
  rankIndex: number,
  choice: string,
  question: Rank,
  answers: Answers,
  setAnswers: React.Dispatch<React.SetStateAction<Answers>>
) => {
  return (
    <Draggable key={choice} draggableId={choice} index={rankIndex}>
      {(provided) => (
        <li
          id={choice}
          ref={provided.innerRef}
          {...provided.draggableProps}
          {...provided.dragHandleProps}
          className="block items-center mb-2 w-2/3 bg-white rounded-lg border border-gray-200 shadow-md hover:bg-gray-100">
          <div className="flex py-4 justify-between items-center text-sm text-gray-900">
            <input
              type="text"
              placeholder={(rankIndex + 1).toString()}
              size={question.MaxN / 10 + 1}
              className="flex-none mx-3 rounded text-center"
              onChange={(e) => handleRankInput(e, question, choice, rankIndex, answers, setAnswers)}
            />
            <div className="flex-auto text-gray-600">{choice}</div>
            <RankList />
          </div>
        </li>
      )}
    </Draggable>
  );
};

export { handleOnDragEnd, rankDisplay };
