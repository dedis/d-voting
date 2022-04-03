import { FC, useEffect, useRef } from 'react';

import { MinusSmIcon, PlusSmIcon, XIcon } from '@heroicons/react/outline';
import { Disclosure } from '@headlessui/react';
import { ChevronUpIcon } from '@heroicons/react/solid';

import { newRank, newSelect, newText } from './utils/getObjectType';
import useQuestionForm from './utils/useQuestionForm';

import PropTypes from 'prop-types';
import { Rank, Select, Text } from 'types/configuration';

type QuestionComponentProps = {
  obj: Rank | Select | Text;
  notifyParent(obj: Rank | Select | Text): void;
  removeQuestion: () => void;
  type: 'RANK' | 'TEXT' | 'SELECT';
};

const QuestionComponent: FC<QuestionComponentProps> = ({
  obj,
  notifyParent,
  removeQuestion,
  type,
}) => {
  const { ID } = obj;
  const isQuestionMounted = useRef<Boolean>(false);

  const initTypeObject: { RANK: Rank; SELECT: Select; TEXT: Text } = {
    RANK: newRank(ID),
    SELECT: newSelect(ID),
    TEXT: newText(ID),
  };
  const [values, [handleChange, addChoice, deleteChoice, clearChoices, updateChoice]] =
    useQuestionForm(obj.Title === '' ? initTypeObject[type] : obj);

  const { Title, MaxN, MinN, Choices, Regex, MaxLength } = values;

  useEffect(() => {
    if (!isQuestionMounted.current) {
      isQuestionMounted.current = true;
      return;
    }
    notifyParent(values);
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [values]);

  return (
    <Disclosure>
      {({ open }) => (
        <div className="py-2 px-5">
          <div className="flex">
            <Disclosure.Button className="flex justify-between w-full px-4 py-2 text-sm font-semibold text-left text-gray-500 bg-gray-100 rounded-lg hover:bg-gray-200 focus:outline-none focus-visible:ring focus-visible:ring-grey-500 focus-visible:ring-opacity-75">
              <span className="uppercase">{type}</span>
              <ChevronUpIcon
                className={`${open ? '' : 'transform rotate-180'} w-5 h-5 text-gray-500`}
              />
            </Disclosure.Button>
            <button
              className="flex justify-between w-15 px-4 py-2 text-sm font-semibold text-left text-gray-500 bg-gray-100 rounded-lg hover:bg-gray-200 focus:outline-none focus-visible:ring focus-visible:ring-grey-500 focus-visible:ring-opacity-75"
              onClick={removeQuestion}>
              <XIcon className="w-5 h-5 text-red-500" aria-hidden="true" />
            </button>
          </div>
          <Disclosure.Panel className="px-3 pt-4 pb-2 text-sm text-gray-500">
            <div className="flex flex-col">
              <label className="block text-md mt font-medium text-gray-500">ID: {ID}</label>
              <label className="block text-md mt font-medium text-gray-500">Title</label>
              <input
                value={Title}
                onChange={handleChange}
                name="Title"
                type="text"
                placeholder="Enter your Title"
                className="my-1 w-60 ml-1 border rounded-md"
              />
              <label className="block text-md font-medium text-gray-500">MaxN</label>
              <input
                value={MaxN}
                onChange={handleChange}
                name="MaxN"
                min={MinN}
                type="number"
                placeholder="Enter the MaxN"
                className="my-1 w-60 ml-1 border rounded-md"
              />
              <label className="block text-md font-medium text-gray-500">MinN</label>
              <input
                value={MinN}
                onChange={handleChange}
                name="MinN"
                max={MaxN}
                min="0"
                type="number"
                placeholder="Enter the MinN"
                className="my-1 w-60 ml-1 border rounded-md"
              />
              {type === 'TEXT' && (
                <>
                  <label className="block text-md font-medium text-gray-500">MaxLength</label>
                  <input
                    value={MaxLength}
                    onChange={handleChange}
                    name="MaxLength"
                    min="0"
                    type="number"
                    placeholder="Enter the MaxLength"
                    className="my-1 w-60 ml-1 border rounded-md"
                  />
                </>
              )}
              {type === 'TEXT' && (
                <>
                  <label className="block text-md font-medium text-gray-500">Regex</label>
                  <input
                    value={Regex}
                    onChange={handleChange}
                    name="Regex"
                    type="text"
                    placeholder="Enter your Regex"
                    className="my-1 w-60 ml-1 border rounded-md"
                  />
                </>
              )}
              <label className="block text-md font-medium text-gray-500">Choices</label>
              {Choices.map((choice: string, idx: number) => (
                <div key={`${ID}wrapper${idx}`}>
                  <input
                    key={`${ID}choice${idx}`}
                    value={choice}
                    onChange={updateChoice(idx)}
                    name="Choice"
                    type="text"
                    placeholder="Enter your choice"
                    className="my-1 w-60 ml-1 border rounded-md"
                  />
                  <button
                    key={`${ID}deleteChoice${idx}`}
                    type="button"
                    className="inline-flex ml-2 items-center border border-transparent rounded-md shadow-sm text-sm font-medium text-white bg-indigo-600 hover:bg-indigo-700"
                    onClick={deleteChoice(idx)}>
                    <MinusSmIcon className="h-4 w-4" aria-hidden="true" />
                  </button>
                </div>
              ))}
              <button
                type="button"
                className="flex p-2 h-8 w-8 mb-2 rounded-md bg-green-600 hover:bg-green-800 sm:-mr-2"
                onClick={addChoice}>
                <PlusSmIcon className="h-5 w-5 text-white" aria-hidden="true" />
              </button>
              <button
                onClick={clearChoices}
                className="h-8 w-40 rounded-md bg-red-600 hover:bg-red-700 text-white text-md">
                Clear all choices
              </button>
            </div>
          </Disclosure.Panel>
        </div>
      )}
    </Disclosure>
  );
};

QuestionComponent.propTypes = {
  obj: PropTypes.any.isRequired,
  notifyParent: PropTypes.func.isRequired,
  removeQuestion: PropTypes.func.isRequired,
  type: PropTypes.oneOf(['RANK', 'TEXT', 'SELECT']),
};
export default QuestionComponent;
