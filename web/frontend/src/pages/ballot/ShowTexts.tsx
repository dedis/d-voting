import { Text } from 'components/utils/types';
import { Answers, TEXT } from 'components/utils/useConfiguration';
import { t } from 'i18next';
import { getIndices } from './HandleAnswers';

const textHintDisplay = (content: Text) => {
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

const handleTextInput = (
  e: React.ChangeEvent<HTMLInputElement>,
  question: Text,
  choice: string,
  answers: Answers,
  setAnswers: React.Dispatch<React.SetStateAction<Answers>>
) => {
  let { questionIndex, choiceIndex, errorIndex, newAnswers } = getIndices(
    question,
    choice,
    answers,
    TEXT
  );

  newAnswers.TextAnswers[questionIndex].Answers[choiceIndex] = e.target.value.trim();
  newAnswers.Errors[errorIndex].Message = '';

  if (question.Regex) {
    let regexp = new RegExp(question.Regex);
    for (const answer of newAnswers.TextAnswers[questionIndex].Answers) {
      if (!regexp.test(answer) && answer !== '') {
        newAnswers.Errors[errorIndex].Message = t('regexpCheck', { regexp: question.Regex });
      }
    }
  }
  setAnswers(newAnswers);
};

const textDisplay = (
  choice: string,
  question: Text,
  answers: Answers,
  setAnswers: React.Dispatch<React.SetStateAction<Answers>>
) => {
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
        onChange={(e) => handleTextInput(e, question, choice, answers, setAnswers)}
      />
    </div>
  );
};

export { textDisplay, textHintDisplay };
