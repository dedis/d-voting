import { FC } from 'react';
import { Draggable, DropResult, Droppable } from 'react-beautiful-dnd';
import { Answers, ID, RankQuestion } from 'types/configuration';
import { answersFrom } from 'types/getObjectType';
import HintButton from 'components/buttons/HintButton';

export const handleOnDragEnd = (
  result: DropResult,
  answers: Answers,
  setAnswers: (answers: Answers) => void
) => {
  if (!result.destination) {
    return;
  }

  const rankID = result.destination.droppableId as ID;
  const newAnswers = answersFrom(answers);
  const rankAnswer = newAnswers.RankAnswers.get(rankID);

  const [reorderedItem] = rankAnswer.splice(result.source.index, 1);
  rankAnswer.splice(result.destination.index, 0, reorderedItem);
  newAnswers.RankAnswers.set(rankID, rankAnswer);

  setAnswers(newAnswers);
};

type RankProps = {
  rank: RankQuestion;
  answers: Answers;
};

const Rank: FC<RankProps> = ({ rank, answers }) => {
  const RankListIcon = () => {
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

  const choiceDisplay = (choice: string, rankIndex: number) => {
    return (
      <Draggable key={choice} draggableId={choice} index={rankIndex}>
        {(provided) => (
          <li
            id={choice}
            ref={provided.innerRef}
            {...provided.draggableProps}
            {...provided.dragHandleProps}
            className="mb-2 rounded-lg border bg-white border-gray-200 shadow-md hover:bg-gray-100">
            <div className="flex justify-between py-4 text-sm text-gray-900">
              <p className="flex-none mx-5 rounded text-center text-gray-400">{rankIndex + 1}</p>
              <div className="flex-auto w-80 overflow-y-auto break-words pr-4 text-gray-600">
                {choice}
              </div>
              <RankListIcon />
            </div>
          </li>
        )}
      </Draggable>
    );
  };

  return (
    <div className="mb-6">
      <div className="grid grid-rows-1 grid-flow-col">
        <div>
          <h3 className="text-lg break-words text-gray-600 w-96">{rank.Title}</h3>
        </div>
        <div>
          <HintButton text={rank.Hint} />
        </div>
      </div>
      <div className="mt-5 px-4 max-w-[300px] sm:pl-8 sm:max-w-md">
        <>
          <Droppable droppableId={String(rank.ID)}>
            {(provided) => (
              <ul className={rank.ID} {...provided.droppableProps} ref={provided.innerRef}>
                {Array.from(answers.RankAnswers.get(rank.ID).entries()).map(
                  ([rankIndex, choiceIndex]) => choiceDisplay(rank.Choices[choiceIndex], rankIndex)
                )}
                {provided.placeholder}
              </ul>
            )}
          </Droppable>
        </>
      </div>
    </div>
  );
};

export default Rank;
