import { getIndexes } from './HandleAnswers';
import { Draggable, DropResult } from 'react-beautiful-dnd';
import { Error, RankAnswer } from 'components/utils/useConfiguration';
import { t } from 'i18next';

const reorderRank = (
  sourceIndex: number,
  destinationIndex: number,
  questionIndex: number,
  rankStates: RankAnswer[],
  setRankStates: React.Dispatch<React.SetStateAction<RankAnswer[]>>
) => {
  const items: Array<RankAnswer> = Array.from(rankStates);
  const [reorderedItem] = items[questionIndex].Answers.splice(sourceIndex, 1);
  items[questionIndex].Answers.splice(destinationIndex, 0, reorderedItem);
  setRankStates(items);
};

const handleOnDragEnd = (
  result: DropResult,
  rankStates: RankAnswer[],
  setRankStates: React.Dispatch<React.SetStateAction<RankAnswer[]>>,
  answerErrors: Error[],
  setAnswerErrors: React.Dispatch<React.SetStateAction<Error[]>>
) => {
  if (!result.destination) {
    return;
  }
  let questionIndex = rankStates.findIndex((s) => s.ID === result.destination.droppableId);
  let errorIndex = answerErrors.findIndex((e) => e.ID === result.destination.droppableId);
  let error = Array.from(answerErrors);
  error[errorIndex].Message = '';
  setAnswerErrors(error);
  reorderRank(
    result.source.index,
    result.destination.index,
    questionIndex,
    rankStates,
    setRankStates
  );
  console.log('rankState: ' + JSON.stringify(rankStates));
};

const handleRankInput = (
  e: React.ChangeEvent<HTMLInputElement>,
  question: any,
  choice: string,
  rankIndex: number,
  rankStates: RankAnswer[],
  setRankStates: React.Dispatch<React.SetStateAction<RankAnswer[]>>,
  answerErrors: Error[],
  setAnswerErrors: React.Dispatch<React.SetStateAction<Error[]>>
) => {
  let { questionIndex, errorIndex } = getIndexes(question, choice, rankStates, answerErrors);
  let error = Array.from(answerErrors);
  error[errorIndex].Message = '';
  let destinationIndex = +e.target.value;
  if (e.target.value !== '') {
    if (destinationIndex > question.MaxN || destinationIndex <= 0) {
      error[errorIndex].Message = t('rankRange') + question.MaxN;
    } else {
      reorderRank(rankIndex, destinationIndex - 1, questionIndex, rankStates, setRankStates);
    }
  }
  e.target.value = '';
  setAnswerErrors(error);

  console.log('rankStates: ' + JSON.stringify(rankStates));
  setAnswerErrors(error);
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
  question: any,
  rankStates: RankAnswer[],
  setRankStates: React.Dispatch<React.SetStateAction<RankAnswer[]>>,
  answerErrors: Error[],
  setAnswerErrors: React.Dispatch<React.SetStateAction<Error[]>>
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
              onChange={(e) =>
                handleRankInput(
                  e,
                  question,
                  choice,
                  rankIndex,
                  rankStates,
                  setRankStates,
                  answerErrors,
                  setAnswerErrors
                )
              }
            />
            <div className="flex-auto">{choice}</div>
            <RankList />
          </div>
        </li>
      )}
    </Draggable>
  );
};

export { handleOnDragEnd, rankDisplay };
