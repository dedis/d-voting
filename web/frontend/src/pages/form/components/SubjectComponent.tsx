import { FC, ReactElement, useEffect, useRef, useState } from 'react';
import PropTypes from 'prop-types';

import * as types from '../../../types/configuration';
import { RANK, SELECT, SUBJECT, TEXT } from '../../../types/configuration';
import { newRank, newSelect, newSubject, newText } from '../../../types/getObjectType';

import Question from './Question';
import SubjectDropdown from './SubjectDropdown';
import {
  CheckIcon,
  ChevronUpIcon,
  CursorClickIcon,
  FolderIcon,
  MenuAlt1Icon,
  SwitchVerticalIcon,
  XIcon,
} from '@heroicons/react/outline';
import { PencilIcon } from '@heroicons/react/solid';
import AddQuestionModal from './AddQuestionModal';
import { useTranslation } from 'react-i18next';
import RemoveElementModal from './RemoveElementModal';
import { internationalize } from './../../utils';
const MAX_NESTED_SUBJECT = 1;

type SubjectComponentProps = {
  notifyParent: (subject: types.Subject) => void;
  removeSubject: () => void;
  subjectObject: types.Subject;
  nestedLevel: number;
  language: string;
};
const SubjectComponent: FC<SubjectComponentProps> = ({
  notifyParent,
  removeSubject,
  subjectObject,
  nestedLevel,
  language,
}) => {
  const { t } = useTranslation();
  const emptyElementToRemove = { ID: '', Type: '' };

  const [subject, setSubject] = useState<types.Subject>(subjectObject);
  const [currentQuestion, setCurrentQuestion] = useState<
    types.RankQuestion | types.SelectQuestion | types.TextQuestion | null
  >();
  const isSubjectMounted = useRef<boolean>(false);
  const [isOpen, setIsOpen] = useState<boolean>(false);
  const [titleChanging, setTitleChanging] = useState<boolean>(
    subjectObject.Title.En.length ? false : true
  );
  const [openModal, setOpenModal] = useState<boolean>(false);
  const [showRemoveElementModal, setShowRemoveElementModal] = useState<boolean>(false);
  const [textRemoveElementModal, setTextRemoveElementModal] = useState<string>('');
  const [elementToRemove, setElementToRemove] = useState(emptyElementToRemove);
  const [components, setComponents] = useState<ReactElement[]>([]);

  const { Title, Order, Elements } = subject;
  // When a property changes, we notify the parent with the new subject object
  useEffect(() => {
    // We only notify the parent when the subject is mounted
    if (!isSubjectMounted.current) {
      isSubjectMounted.current = true;
      return;
    }
    notifyParent(subject);
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [subject]);

  const localNotifyParent = (subj: types.Subject) => {
    const newElements = new Map(Elements);
    newElements.set(subj.ID, subj);
    setSubject({ ...subject, Elements: newElements, Title });
  };

  const notifySubject = (question: types.SubjectElement) => {
    if (question.Type !== SUBJECT) {
      const newElements = new Map(Elements);
      newElements.set(question.ID, question);
      setSubject({ ...subject, Elements: newElements });
    }
  };

  const addSubject = () => {
    const newElements = new Map(Elements);
    const newSubj = newSubject();
    newElements.set(newSubj.ID, newSubj);
    setSubject({ ...subject, Elements: newElements, Order: [...Order, newSubj.ID] });
  };

  const addQuestion = (question: types.SubjectElement) => {
    const newElements = new Map(Elements);
    newElements.set(question.ID, question);
    setSubject({ ...subject, Elements: newElements, Order: [...Order, question.ID] });
  };

  const localRemoveSubject = (subjID: types.ID) => () => {
    const newElements = new Map(Elements);
    newElements.delete(subjID);
    setSubject({
      ...subject,
      Elements: newElements,
      Order: Order.filter((id) => id !== subjID),
    });
  };

  const removeChildQuestion = (questionID: types.ID) => () => {
    const newElements = new Map(Elements);
    newElements.delete(questionID);
    setSubject({
      ...subject,
      Elements: newElements,
      Order: Order.filter((id) => id !== questionID),
    });
  };

  const handleConfirmRemoveElement = () => {
    switch (elementToRemove.Type) {
      case SUBJECT:
        localRemoveSubject(elementToRemove.ID)();
        break;
      default:
        removeChildQuestion(elementToRemove.ID)();
        break;
    }
    setElementToRemove(emptyElementToRemove);
  };

  // Sorts the questions components & sub-subjects according to their Order into
  // the components state array
  useEffect(() => {
    // findQuestion return the react element based on the question/subject ID.
    // Returns undefined if the question/subject ID is unknown.
    const findQuestion = (id: types.ID): ReactElement => {
      if (!Elements.has(id)) {
        return undefined;
      }

      const found = Elements.get(id);

      switch (found.Type) {
        case TEXT:
          const text = found as types.TextQuestion;
          return (
            <Question
              key={`text${text.ID}`}
              question={text}
              notifyParent={notifySubject}
              removeQuestion={() => {
                setElementToRemove({ ID: text.ID, Type: TEXT });
                setTextRemoveElementModal(t(`confirmRemove${found.Type}`));
                setShowRemoveElementModal(true);
              }}
              language={language}
            />
          );
        case SUBJECT:
          const sub = found as types.Subject;
          return (
            <SubjectComponent
              notifyParent={localNotifyParent}
              removeSubject={() => {
                setElementToRemove({ ID: sub.ID, Type: SUBJECT });
                setTextRemoveElementModal(t(`confirmRemove${found.Type}`));
                setShowRemoveElementModal(true);
              }}
              subjectObject={sub}
              nestedLevel={nestedLevel + 1}
              key={sub.ID}
              language={language}
            />
          );
        case RANK:
          const rank = found as types.RankQuestion;
          return (
            <Question
              key={`rank${rank.ID}`}
              question={rank}
              notifyParent={notifySubject}
              removeQuestion={() => {
                setElementToRemove({ ID: rank.ID, Type: RANK });
                setTextRemoveElementModal(t(`confirmRemove${found.Type}`));
                setShowRemoveElementModal(true);
              }}
              language={language}
            />
          );
        case SELECT:
          const select = found as types.SelectQuestion;
          return (
            <Question
              key={`select${select.ID}`}
              question={select}
              notifyParent={notifySubject}
              removeQuestion={() => {
                setElementToRemove({ ID: select.ID, Type: SELECT });
                setTextRemoveElementModal(t(`confirmRemove${found.Type}`));
                setShowRemoveElementModal(true);
              }}
              language={language}
            />
          );
      }
    };

    setComponents(Order.map((id) => findQuestion(id)));

    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [Title, Elements, Order, nestedLevel]);

  useEffect(() => {
    if (components.length === 0) {
      // resets the icon state if there are no questions
      setIsOpen(false);
    }
  }, [components]);

  useEffect(() => {
    // When the modal is closed we reset the current question sent to the
    // question modal
    if (isOpen === false) {
      setCurrentQuestion(null);
    }
  }, [isOpen]);

  const handleAddQuestion = (
    question: types.TextQuestion | types.RankQuestion | types.SelectQuestion
  ) => {
    setIsOpen(true);
    setOpenModal(true);
    setCurrentQuestion(question);
  };

  const dropdownContent: {
    name: string;
    icon: JSX.Element;
    onClick: () => void;
  }[] = [
    {
      name: 'addRank',
      icon: <SwitchVerticalIcon className="mr-2 h-5 w-5" aria-hidden="true" />,
      onClick: () => {
        handleAddQuestion(newRank());
      },
    },
    {
      name: 'addSelect',
      icon: <CursorClickIcon className="mr-2 h-5 w-5" aria-hidden="true" />,
      onClick: () => {
        handleAddQuestion(newSelect());
      },
    },
    {
      name: 'addText',
      icon: <MenuAlt1Icon className="mr-2 h-5 w-5" aria-hidden="true" />,
      onClick: () => {
        handleAddQuestion(newText());
      },
    },
    {
      name: 'removeSubject',
      icon: <XIcon className="mr-2 h-5 w-5" aria-hidden="true" />,
      onClick: removeSubject,
    },
  ];
  if (nestedLevel < MAX_NESTED_SUBJECT) {
    dropdownContent.splice(3, 0, {
      name: 'addSubject',
      icon: <FolderIcon className="mr-2 h-5 w-5" aria-hidden="true" />,
      onClick: () => {
        setIsOpen(true);
        addSubject();
      },
    });
  }

  const QuestionModal = () => {
    return currentQuestion ? (
      <AddQuestionModal
        open={openModal}
        setOpen={setOpenModal}
        notifyParent={addQuestion}
        question={currentQuestion}
        handleClose={() => {
          setIsOpen(false);
          setOpenModal(false);
        }}
      />
    ) : null;
  };

  return (
    <div className={`${nestedLevel === 0 ? 'border-t' : 'pl-3'} `}>
      <RemoveElementModal
        showModal={showRemoveElementModal}
        setShowModal={setShowRemoveElementModal}
        textModal={textRemoveElementModal}
        handleConfirm={handleConfirmRemoveElement}
      />

      <QuestionModal />
      <div className="flex flex-row justify-between w-full h-30 ">
        <div className="flex flex-col max-w-full pl-2">
          <div className="mt-3 flex">
            <div className="h-9 w-9 rounded-full bg-gray-100 mr-2 ml-1">
              <FolderIcon className="m-2 h-5 w-5 text-gray-400" aria-hidden="true" />
            </div>
            {titleChanging ? (
              <div className="flex flex-col mt-3  mb-2">
                {language === 'en' && (
                  <input
                    value={Title.En}
                    onChange={(e) =>
                      setSubject({ ...subject, Title: { ...Title, En: e.target.value } })
                    }
                    name="Title"
                    type="text"
                    placeholder={t('enterSubjectTitleLg')}
                    className={`m-3 px-1 w-120 border rounded-md ${
                      nestedLevel === 0 ? 'text-lg' : 'text-md'
                    } `}
                  />
                )}
                {language === 'fr' && (
                  <input
                    value={Title.Fr}
                    onChange={(e) =>
                      setSubject({ ...subject, Title: { ...Title, Fr: e.target.value } })
                    }
                    name="Title"
                    type="text"
                    placeholder={t('enterSubjectTitleLg1')}
                    className={`m-3 px-1 w-120 border rounded-md ${
                      nestedLevel === 0 ? 'text-lg' : 'text-md'
                    } `}
                  />
                )}
                {language === 'de' && (
                  <input
                    value={Title.De}
                    onChange={(e) =>
                      setSubject({ ...subject, Title: { ...Title, De: e.target.value } })
                    }
                    name="Title"
                    type="text"
                    placeholder={t('enterSubjectTitleLg2')}
                    className={`m-3 px-1 w-120 border rounded-md ${
                      nestedLevel === 0 ? 'text-lg' : 'text-md'
                    } `}
                  />
                )}
                <div className="ml-1">
                  <button
                    className={`border p-1 rounded-md ${Title.En.length === 0 && 'bg-gray-100'}`}
                    disabled={Title.En.length === 0}
                    onClick={() => setTitleChanging(false)}>
                    <CheckIcon className="h-5 w-5" aria-hidden="true" />
                  </button>
                </div>
              </div>
            ) : (
              <div className="flex mb-2 max-w-md truncate">
                <div className="pt-1.5 truncate" onClick={() => setTitleChanging(true)}>
                  {internationalize(language, Title)}
                </div>
                <div className="ml-1 pr-10">
                  <button
                    className="hover:text-indigo-500 p-1 rounded-md"
                    onClick={() => setTitleChanging(true)}>
                    <PencilIcon className="m-1 h-3 w-3" aria-hidden="true" />
                  </button>
                </div>
              </div>
            )}
          </div>

          <div className="flex mt-2 ml-2">
            <button
              disabled={!(components.length > 0)}
              onClick={() => setIsOpen(!isOpen)}
              className="text-left text-sm font-medium rounded-full text-gray-900">
              <ChevronUpIcon
                className={`${!isOpen ? 'rotate-180 transform' : ''} h-5 w-5 ${
                  components.length > 0 ? 'text-gray-600' : 'text-gray-300'
                } `}
              />
            </button>
            <div className="ml-2">{t('subject')}</div>
          </div>
        </div>
        <div className="relative">
          <div className="-mr-2 flex absolute right-3">
            <SubjectDropdown dropdownContent={dropdownContent} />
          </div>
        </div>
      </div>
      {components.length > 0 && isOpen && (
        <div className="text-sm bg-gray-50">{components.map((component) => component)}</div>
      )}
    </div>
  );
};

SubjectComponent.propTypes = {
  notifyParent: PropTypes.func.isRequired,
  removeSubject: PropTypes.func.isRequired,
  subjectObject: PropTypes.any.isRequired,
  nestedLevel: PropTypes.number.isRequired,
};

export default SubjectComponent;
