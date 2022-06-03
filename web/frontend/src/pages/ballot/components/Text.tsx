import { FC, useEffect, useState } from 'react';
import { useTranslation } from 'react-i18next';
import { Answers, TextQuestion } from 'types/configuration';
import { answersFrom } from 'types/getObjectType';

type TextProps = {
  text: TextQuestion;
  answers: Answers;
  setAnswers: (answers: Answers) => void;
};

const Text: FC<TextProps> = ({ text, answers, setAnswers }) => {
  const { t } = useTranslation();
  const [charCounts, setCharCounts] = useState(new Array<number>(text.Choices.length).fill(0));

  const hintDisplay = () => {
    let hint = '';
    const min = text.MinN;
    const max = text.MaxN;

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

  const handleTextInput = (e: React.ChangeEvent<HTMLTextAreaElement>, choiceIndex: number) => {
    const newAnswers = answersFrom(answers);

    const textAnswers = newAnswers.TextAnswers.get(text.ID);
    textAnswers[choiceIndex] = e.target.value.trim();
    newAnswers.TextAnswers.set(text.ID, textAnswers);

    newAnswers.Errors.set(text.ID, '');

    if (text.Regex !== '') {
      const regexp = new RegExp(text.Regex);
      for (const answer of textAnswers) {
        if (!regexp.test(answer) && answer !== '') {
          newAnswers.Errors.set(text.ID, t('regexpCheck', { regexp: text.Regex }));
        }
      }
    }

    setAnswers(newAnswers);
  };

  const charCountDisplay = (choiceIndex: number) => {
    return (
      <div className="justify-center text-sm">
        {charCounts[choiceIndex] > text.MaxLength ? (
          <p className="text-red-500">
            {charCounts[choiceIndex]} / {text.MaxLength}
          </p>
        ) : (
          <p className="text-gray-400">
            {charCounts[choiceIndex]} / {text.MaxLength}
          </p>
        )}
      </div>
    );
  };

  useEffect(() => {
    const newCount = new Array<number>();
    answers.TextAnswers.get(text.ID).map((answer) => newCount.push(answer.length));
    setCharCounts(newCount);
  }, [answers]);

  const choiceDisplay = (choice: string, choiceIndex: number) => {
    const columns = text.MaxLength > 50 ? 50 : text.MaxLength;
    return (
      <div className="flex mb-2 " key={choice}>
        <label htmlFor={choice} className="text-gray-600 text-md">
          {choice == '' ? '' : choice + ': '}
        </label>

        <textarea
          id={choice}
          className="mx-2 sm:text-md border rounded-md text-gray-600 "
          rows={charCounts[choiceIndex] > 0 ? Math.ceil(charCounts[choiceIndex] / columns) : 1}
          //rows={text.MaxLength > 0 ? Math.ceil(text.MaxLength / columns) : 1}
          cols={columns}
          style={{ resize: 'none' }}
          onChange={(e) => handleTextInput(e, choiceIndex)}></textarea>
        <div className="self-end">{charCountDisplay(choiceIndex)}</div>
      </div>
    );
  };

  return (
    <div>
      <h3 className="text-lg text-gray-600">{text.Title}</h3>
      {hintDisplay()}
      <div className="sm:pl-8 mt-2 pl-6">
        {text.Choices.map((choice, index) => choiceDisplay(choice, index))}
      </div>
      <div className="text-red-600 text-sm py-2 sm:pl-2 pl-1">{answers.Errors.get(text.ID)}</div>
    </div>
  );
};

export default Text;
