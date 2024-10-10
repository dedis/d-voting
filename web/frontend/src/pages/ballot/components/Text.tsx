import { FC, useEffect, useState } from 'react';
import { useTranslation } from 'react-i18next';
import { Answers, TextQuestion } from 'types/configuration';
import { answersFrom } from 'types/getObjectType';
import HintButton from 'components/buttons/HintButton';
import { internationalize, urlizeLabel } from './../../utils';

type TextProps = {
  text: TextQuestion;
  answers: Answers;
  setAnswers: (answers: Answers) => void;
  language: string;
};

const Text: FC<TextProps> = ({ text, answers, setAnswers, language }) => {
  const { t } = useTranslation();
  const [charCounts, setCharCounts] = useState(new Array<number>(text.Choices.length).fill(0));
  const requirementsDisplay = () => {
    let requirements;
    const min = text.MinN;
    const max = text.MaxN;

    if (min !== max) {
      requirements =
        min > 1
          ? t('minText', { minText: min, singularPlural: t('pluralAnswers') })
          : t('minText', { minText: min, singularPlural: t('singularAnswer') });
    } else {
      requirements =
        min > 1
          ? t('fillText', { minText: min, singularPlural: t('pluralAnswers') })
          : t('fillText', { minText: min, singularPlural: t('singularAnswer') });
    }
    return <div className="text-sm pl-2 pb-2 sm:pl-4 text-gray-400">{requirements}</div>;
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
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [answers]);

  const choiceDisplay = (choice: string, url: string, choiceIndex: number) => {
    const columns = text.MaxLength > 50 ? 50 : text.MaxLength;
    const prettyChoice = urlizeLabel(choice, url);
    return (
      <div className="flex mb-2  md:flex-row flex-col" key={choice}>
        <label htmlFor={choice} className="text-gray-600 mr-2 w-24 break-words text-md">
          {prettyChoice}:
        </label>

        <textarea
          id={choice}
          className="mx-2 w-[50%] sm:text-md resize-none border rounded-md text-gray-600 "
          rows={charCounts[choiceIndex] > 0 ? Math.ceil(charCounts[choiceIndex] / columns) : 1}
          //rows={text.MaxLength > 0 ? Math.ceil(text.MaxLength / columns) : 1}
          cols={columns}
          onChange={(e) => handleTextInput(e, choiceIndex)}></textarea>
        <div className="self-end">{charCountDisplay(choiceIndex)}</div>
      </div>
    );
  };
  return (
    <div>
      <div className="grid grid-rows-1 grid-flow-col">
        <div>
          <h3 className="text-lg break-words text-gray-600 w-96">
            {urlizeLabel(internationalize(language, text.Title), text.Title.URL)}
          </h3>
        </div>
        <div className="text-right">
          {<HintButton text={internationalize(language, text.Hint)} />}
        </div>
      </div>
      <div className="pt-1">{requirementsDisplay()}</div>
      {text.ChoicesMap.ChoicesMap.has(language) ? (
        <div className="sm:pl-8 mt-2 pl-6">
          {text.ChoicesMap.ChoicesMap.get(language).map((choice, index) =>
            choiceDisplay(choice, text.ChoicesMap.URLs[index], index)
          )}
        </div>
      ) : (
        <div className="sm:pl-8 mt-2 pl-6">
          {text.ChoicesMap.ChoicesMap.get('en').map((choice, index) =>
            choiceDisplay(choice, text.ChoicesMap.URLs[index], index)
          )}
        </div>
      )}
      <div className="text-red-600 text-sm py-2 sm:pl-2 pl-1">{answers.Errors.get(text.ID)}</div>
    </div>
  );
};

export default Text;
