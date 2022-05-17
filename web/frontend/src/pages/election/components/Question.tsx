import { FC, useEffect, useRef, useState } from 'react';

import {
  CursorClickIcon,
  MenuAlt1Icon,
  MinusSmIcon,
  PencilAltIcon,
  PlusSmIcon,
  SwitchVerticalIcon,
  XIcon,
} from '@heroicons/react/outline';
import { ChevronUpIcon } from '@heroicons/react/solid';

import useQuestionForm from './utils/useQuestionForm';

import PropTypes from 'prop-types';
import {
  RANK,
  RankQuestion,
  SELECT,
  SelectQuestion,
  SubjectElement,
  TEXT,
  TextQuestion,
} from 'types/configuration';
import SubjectDropdown from './SubjectDropdown';

type QuestionProps = {
  question: RankQuestion | SelectQuestion | TextQuestion;
  notifyParent(question: RankQuestion | SelectQuestion | TextQuestion): void;
  removeQuestion: () => void;
};

const MAX_MINN = 20;

const Question: FC<QuestionProps> = ({ question, notifyParent, removeQuestion }) => {
  const { ID, Type } = question;
  const isQuestionMounted = useRef<Boolean>(false);
  const [isOpen, setIsOpen] = useState<boolean>(false);

  const {
    state: values,
    handleChange,
    addChoice,
    deleteChoice,
    updateChoice,
  } = useQuestionForm(question);

  const { Title, MaxN, MinN, Choices } = values;

  useEffect(() => {
    // We only notify the parent when the question is mounted
    if (!isQuestionMounted.current) {
      isQuestionMounted.current = true;
      return;
    }
    notifyParent(values);
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [values]);

  const showExtraFields = (quest: SubjectElement) => {
    switch (quest.Type) {
      case TEXT:
        const t = question as TextQuestion;
        return (
          <>
            <label className="block text-md font-medium text-gray-500">MaxLength</label>
            <input
              value={t.MaxLength}
              onChange={handleChange}
              name="MaxLength"
              min="0"
              type="number"
              placeholder="Enter the MaxLength"
              className="my-1 w-60 ml-1 border rounded-md"
            />
            <label className="block text-md font-medium text-gray-500">Regex</label>
            <input
              value={t.Regex}
              onChange={handleChange}
              name="Regex"
              type="text"
              placeholder="Enter your Regex"
              className="my-1 w-60 ml-1 border rounded-md"
            />
          </>
        );
      default:
        return null;
    }
  };

  const dropdownContent = [
    {
      name: `Edit ${Type}`,
      icon: <PencilAltIcon className="mr-2 h-5 w-5" aria-hidden="true" />,
      onClick: () => {
        console.log('should have popover');
      },
    },
    {
      name: `Remove ${Type}`,
      icon: <XIcon className="mr-2 h-5 w-5" aria-hidden="true" />,
      onClick: removeQuestion,
    },
  ];

  const DisplayTypeIcon = () => {
    switch (Type) {
      case RANK:
        return <SwitchVerticalIcon className="m-2 h-5 w-5 text-gray-400" aria-hidden="true" />;
      case SELECT:
        return <CursorClickIcon className="m-2 h-5 w-5 text-gray-400" aria-hidden="true" />;
      case TEXT:
        return <MenuAlt1Icon className="m-2 h-5 w-5 text-gray-400" aria-hidden="true" />;
      default:
        return null;
    }
  };

  return (
    <div className="pl-3">
      <div className="flex flex-row justify-between w-full h-24 ">
        <div className="flex flex-col pl-2">
          <div className="mt-3 flex">
            <div className="h-9 w-9 rounded-full bg-gray-100 mr-2 ml-1">
              <DisplayTypeIcon />
            </div>
            <div className="pt-1.5">{Title.length ? Title : `Enter ${Type} title`}</div>
          </div>

          <div className="flex mt-2 ml-2">
            <button
              onClick={() => setIsOpen(!isOpen)}
              className="text-left text-sm font-medium rounded-full text-gray-900">
              <ChevronUpIcon
                className={`${!isOpen ? 'rotate-180 transform' : ''} h-5 w-5 text-gray-600 `}
              />
            </button>
            <div className="ml-2">{Type.charAt(0).toUpperCase() + Type.slice(1)}</div>
          </div>
        </div>
        <div className="relative">
          <div className="-mr-2 flex absolute right-3">
            <SubjectDropdown dropdownContent={dropdownContent} />
          </div>
        </div>
      </div>
      <div className="px-10">
        {isOpen && (
          <div className="flex flex-col">
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
              max={MaxN < MAX_MINN ? MaxN : MAX_MINN}
              min="0"
              type="number"
              placeholder="Enter the MinN"
              className="my-1 w-60 ml-1 border rounded-md"
            />
            {showExtraFields(question)}
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
          </div>
        )}
      </div>
    </div>
  );
};

Question.propTypes = {
  question: PropTypes.any.isRequired,
  notifyParent: PropTypes.func.isRequired,
  removeQuestion: PropTypes.func.isRequired,
};
export default Question;
