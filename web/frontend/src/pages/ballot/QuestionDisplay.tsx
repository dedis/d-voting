import {
  Answers,
  Error,
  Question,
  RankAnswer,
  RankQuestion,
  SelectAnswer,
  SelectQuestion,
  Subject,
  TextQuestion,
} from 'types/configuration';
import { RankDisplay } from './components/RankDisplay';
import { TextDisplay, TextHintDisplay } from './components/TextDisplay';
import { SelectDisplay, SelectHintDisplay } from './components/SelectDisplay';
import { Droppable } from 'react-beautiful-dnd';

export type HintDisplayProps = {
  questionContent: RankQuestion | TextQuestion | SelectQuestion | Subject;
};

export function renderRank(question: Question, answers: Answers) {
  let rank = question.Content as RankQuestion;

  return (
    <div className="mb-6">
      <h3 className="text-lg text-gray-600">{rank.Title}</h3>
      <div className="mt-5 sm:pl-8 w-3/4">
        <Droppable droppableId={String(rank.ID)}>
          {(provided) => (
            <ul className={rank.ID} {...provided.droppableProps} ref={provided.innerRef}>
              {Array.from(
                answers.RankAnswers.find((r: RankAnswer) => r.ID === rank.ID).Answers.entries()
              ).map(([rankIndex, choiceIndex]) => (
                <RankDisplay
                  rankIndex={rankIndex}
                  choice={rank.Choices[choiceIndex]}
                  key={rank.Choices[choiceIndex]}
                />
              ))}
              {provided.placeholder}
            </ul>
          )}
        </Droppable>
      </div>
    </div>
  );
}

export function renderText(
  question: Question,
  answers: Answers,
  setAnswers: React.Dispatch<React.SetStateAction<Answers>>
) {
  let text = question.Content as TextQuestion;

  return (
    <div>
      <h3 className="text-lg text-gray-600">{text.Title}</h3>
      <TextHintDisplay questionContent={text} />
      <div className="sm:pl-8 mt-2 pl-6">
        {text.Choices.map((choice) => (
          <TextDisplay
            choice={choice}
            question={text}
            answers={answers}
            setAnswers={setAnswers}
            key={choice}
          />
        ))}
      </div>
      <div className="text-red-600 text-sm py-2 sm:pl-2 pl-1">
        {answers.Errors.find((e: Error) => e.ID === text.ID).Message}
      </div>
    </div>
  );
}

export function renderSelect(
  question: Question,
  answers: Answers,
  setAnswers: React.Dispatch<React.SetStateAction<Answers>>
) {
  let select = question.Content as SelectQuestion;

  return (
    <div>
      <h3 className="text-lg text-gray-600">{select.Title}</h3>
      <SelectHintDisplay questionContent={select} />
      <div className="sm:pl-8 pl-6">
        {Array.from(
          answers.SelectAnswers.find((s: SelectAnswer) => s.ID === select.ID).Answers.entries()
        ).map(([choiceIndex, isChecked]) => (
          <SelectDisplay
            isChecked={isChecked}
            choice={select.Choices[choiceIndex]}
            question={select}
            answers={answers}
            setAnswers={setAnswers}
            key={select.Choices[choiceIndex]}
          />
        ))}
      </div>
      <div className="text-red-600 text-sm py-2 sm:pl-2 pl-1">
        {answers.Errors.find((e: Error) => e.ID === select.ID).Message}
      </div>
    </div>
  );
}

export function renderSubject(question: Question) {
  return <h3 className="font-bold text-lg text-gray-600">{question.Content.Title}</h3>;
}
