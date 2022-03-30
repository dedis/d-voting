import {
  Answers,
  Question,
  RankAnswer,
  RankQuestion,
  SelectAnswer,
  SelectQuestion,
  Subject,
  TextQuestion,
} from 'types/configuration';
import { RankDisplay } from './components/RankDisplay';
import { Droppable } from 'react-beautiful-dnd';
import { TextDisplay, TextHintDisplay } from './components/TextDisplay';
import { SelectDisplay, SelectHintDisplay } from './components/SelectDisplay';

export type HintDisplayProps = {
  questionContent: RankQuestion | TextQuestion | SelectQuestion | Subject;
};

export function renderRank(
  question: Question,
  answers: Answers,
  setAnswers: React.Dispatch<React.SetStateAction<Answers>>
) {
  let rank = question.Content as RankQuestion;

  return (
    <div>
      <h3 className="text-lg text-gray-600">{rank.Title}</h3>
      <div className="mt-5 pl-8">
        <Droppable droppableId={String(rank.ID)}>
          {(provided) => (
            <ul
              className={question.Content.ID}
              {...provided.droppableProps}
              ref={provided.innerRef}>
              {Array.from(
                answers.RankAnswers.find((r: RankAnswer) => r.ID === rank.ID).Answers.entries()
              ).map(([rankIndex, choiceIndex]) => (
                <RankDisplay
                  rankIndex={rankIndex}
                  choice={rank.Choices[choiceIndex]}
                  question={rank}
                  answers={answers}
                  setAnswers={setAnswers}
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
      <div className="pl-8">
        {text.Choices.map((choice) => (
          <TextDisplay choice={choice} question={text} answers={answers} setAnswers={setAnswers} />
        ))}
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
      <div className="pl-8">
        {Array.from(
          answers.SelectAnswers.find((s: SelectAnswer) => s.ID === select.ID).Answers.entries()
        ).map(([choiceIndex, isChecked]) => (
          <SelectDisplay
            isChecked={isChecked}
            choice={select.Choices[choiceIndex]}
            question={select}
            answers={answers}
            setAnswers={setAnswers}
          />
        ))}
      </div>
    </div>
  );
}

export function renderSubject(question: Question) {
  return <h3 className="font-bold text-lg text-gray-600">{question.Content.Title}</h3>;
}
