import { FC, useEffect, useState } from 'react';
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

// Component that renders and checks the input of a Choice for a TextQuestion
const TextDisplay: FC<TextDisplayProps> = ({ choice, question, answers, setAnswers }) => {
  const { t } = useTranslation();
  const [charCount, setCharCount] = useState(0);

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

  useEffect(() => {
    const { questionIndex, choiceIndex } = getIndices(question, choice, answers, TEXT);
    setCharCount(answers.TextAnswers[questionIndex].Answers[choiceIndex].length);
  }, [answers]);

  const charCountDisplay = () => {
    return (
      <div className="justify-center text-sm">
        {charCount > question.MaxLength ? (
          <p className="text-red-500">
            {charCount} / {question.MaxLength}
          </p>
        ) : (
          <p className="text-gray-400">
            {charCount} / {question.MaxLength}
          </p>
        )}
      </div>
    );
  };

  return (
    <div className="flex mb-2 items-center">
      <label htmlFor={choice} className="text-gray-600 text-md">
        {choice + ': '}
      </label>
      <input
        id={choice}
        type="text"
        key={choice}
        className="mx-2 sm:text-md border rounded-md w-3/5 text-gray-600"
        onChange={handleTextInput}
      />
      {charCountDisplay()}
    </div>
  );
};

export { TextDisplay, TextHintDisplay };
