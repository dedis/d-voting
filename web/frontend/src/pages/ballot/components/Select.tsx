import { FC, Fragment } from 'react';
import { useTranslation } from 'react-i18next';
import { Answers, SelectQuestion } from 'types/configuration';
import { answersFrom } from 'types/getObjectType';
import { QuestionMarkCircleIcon } from '@heroicons/react/outline/';
import { Popover, Transition } from '@headlessui/react';

type SelectProps = {
  select: SelectQuestion;
  answers: Answers;
  setAnswers: (answers: Answers) => void;
};

const Select: FC<SelectProps> = ({ select, answers, setAnswers }) => {
  const { t } = useTranslation();

  const handleChecks = (e: React.ChangeEvent<HTMLInputElement>, choiceIndex: number) => {
    const newAnswers = answersFrom(answers);
    let selectAnswers = newAnswers.SelectAnswers.get(select.ID);

    if (select.MaxN === 1) {
      selectAnswers = new Array<boolean>(select.Choices.length).fill(false);
    }

    selectAnswers[choiceIndex] = e.target.checked;
    newAnswers.SelectAnswers.set(select.ID, selectAnswers);
    const numAnswer = selectAnswers.filter((check: boolean) => check === true).length;

    if (numAnswer >= select.MinN && numAnswer < select.MaxN + 1) {
      newAnswers.Errors.set(select.ID, '');
    } else if (numAnswer >= select.MaxN + 1) {
      newAnswers.Errors.set(select.ID, t('maxSelectError', { max: select.MaxN }));
    }

    setAnswers(newAnswers);
  };

  const requirementsDisplay = () => {
    let requirements = '';
    const max = select.MaxN;
    const min = select.MinN;

    if (max === min) {
      requirements =
        max > 1
          ? t('selectMin', { minSelect: min, singularPlural: t('pluralAnswers') })
          : t('selectMin', { minSelect: min, singularPlural: t('singularAnswer') });
    } else if (min === 0) {
      requirements =
        max > 1
          ? t('selectMax', { maxSelect: max, singularPlural: t('pluralAnswers') })
          : t('selectMax', { maxSelect: max, singularPlural: t('singularAnswer') });
    } else {
      requirements = t('selectBetween', { minSelect: min, maxSelect: max });
    }

    return <div className="text-sm pl-2 pb-2 sm:pl-4 text-gray-400">{requirements}</div>;
  };

  const choiceDisplay = (isChecked: boolean, choice: string, choiceIndex: number) => {
    return (
      <div key={choice}>
        <input
          id={choice}
          className="h-4 w-4 mt-1 mr-2 cursor-pointer accent-indigo-500"
          type="checkbox"
          value={choice}
          checked={isChecked}
          onChange={(e) => handleChecks(e, choiceIndex)}
        />
        <label htmlFor={choice} className="pl-2 break-words text-gray-600 cursor-pointer">
          {choice}
        </label>
      </div>
    );
  };

  return (
    <div>
      <h3 className="text-lg break-words text-gray-600">{select.Title}</h3>
      {requirementsDisplay()}
      <div className="sm:pl-8 mt-2 pl-6">
        {Array.from(answers.SelectAnswers.get(select.ID).entries()).map(
          ([choiceIndex, isChecked]) =>
            choiceDisplay(isChecked, select.Choices[choiceIndex], choiceIndex)
        )}
      </div>
      {select.Hint.length !== 0 && (
        <Popover className="relative">
          <Popover.Button>
            <div className="text-gray-600">
              <QuestionMarkCircleIcon className="color-gray-900 mt-2 h-4 w-4" />
            </div>
          </Popover.Button>
          <Transition
            as={Fragment}
            enter="transition ease-out duration-100"
            enterFrom="transform opacity-0 scale-95"
            enterTo="transform opacity-100 scale-100"
            leave="transition ease-in duration-75"
            leaveFrom="transform opacity-100 scale-100"
            leaveTo="transform opacity-0 scale-95">
            <Popover.Panel className="z-30 absolute p-2 max-w-prose mt-1 ml-2 rounded-md bg-white rounded-lg shadow-lg ring-1 ring-black ring-opacity-5">
              {<div className="text-sm">{select.Hint}</div>}
            </Popover.Panel>
          </Transition>
        </Popover>
      )}
      <div className="text-red-600 text-sm py-2 sm:pl-4 pl-2">{answers.Errors.get(select.ID)}</div>
    </div>
  );
};

export default Select;
