import { FC } from 'react';
import { useTranslation } from 'react-i18next';
import { Answers, TEXT, TextQuestion } from 'types/configuration';
import { getIndices } from './HandleAnswers';
import { HintDisplayProps } from '../QuestionDisplay';

const TextHintDisplay: FC<HintDisplayProps> = ({ questionContent }) => {
  const content = questionContent as TextQuestion;
  const { t } = useTranslation();

  let hint = '';
  let min = content.MinN;
  let max = content.MaxN;

  if (min !== max) {
    hint =
      min > 1
        ? t('minText', { minText: min, singularPlural: t('pluralAnswers') })
        : t('minText', { minText: min, singularPlural: t('singularAnswer') });
  } else {
    hint =
      min > 1
        ? t('fillText', { minText: min, singularPlural: t('pluralAnswers') })
        : t('fillText', { minText: min, singularPlural: t('singularAnswer') });
  }
  return <div className="text-sm pl-2 pb-2 text-gray-400">{hint}</div>;
};

type TextDisplayProps = {
  choice: string;
  question: TextQuestion;
  answers: Answers;
  setAnswers: React.Dispatch<React.SetStateAction<Answers>>;
};

const TextDisplay: FC<TextDisplayProps> = ({ choice, question, answers, setAnswers }) => {
  const { t } = useTranslation();
  const handleTextInput = (e: React.ChangeEvent<HTMLInputElement>) => {
    const { questionIndex, choiceIndex, errorIndex, newAnswers } = getIndices(
      question,
      choice,
      answers,
      TEXT
    );

    newAnswers.TextAnswers[questionIndex].Answers[choiceIndex] = e.target.value.trim();
    newAnswers.Errors[errorIndex].Message = '';

    if (question.Regex) {
      const regexp = new RegExp(question.Regex);
      for (const answer of newAnswers.TextAnswers[questionIndex].Answers) {
        if (!regexp.test(answer) && answer !== '') {
          newAnswers.Errors[errorIndex].Message = t('regexpCheck', { regexp: question.Regex });
        }
      }
    }
    setAnswers(newAnswers);
  };

  return (
    <div>
      <label htmlFor={choice} className="text-gray-600">
        {choice + ': '}
      </label>
      <input
        id={choice}
        type="text"
        key={choice}
        className="mt-1 sm:text-sm border rounded-md w-1/2"
        onChange={handleTextInput}
      />
    </div>
  );
};

export { TextDisplay, TextHintDisplay };