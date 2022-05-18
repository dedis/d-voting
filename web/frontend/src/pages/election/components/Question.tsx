import { FC, useState } from 'react';

import {
  CursorClickIcon,
  MenuAlt1Icon,
  PencilAltIcon,
  SwitchVerticalIcon,
  XIcon,
} from '@heroicons/react/outline';

import PropTypes from 'prop-types';
import {
  RANK,
  RankQuestion,
  SELECT,
  SelectQuestion,
  TEXT,
  TextQuestion,
} from 'types/configuration';
import SubjectDropdown from './SubjectDropdown';
import AddQuestionModal from './AddQuestionModal';

type QuestionProps = {
  question: RankQuestion | SelectQuestion | TextQuestion;
  notifyParent(question: RankQuestion | SelectQuestion | TextQuestion): void;
  removeQuestion: () => void;
};

const Question: FC<QuestionProps> = ({ question, notifyParent, removeQuestion }) => {
  const { Title, Type } = question;
  const [openModal, setOpenModal] = useState<boolean>(false);

  const dropdownContent = [
    {
      name: `edit${Type}`,
      icon: <PencilAltIcon className="mr-2 h-5 w-5" aria-hidden="true" />,
      onClick: () => {
        setOpenModal(true);
      },
    },
    {
      name: `remove${Type}`,
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
      <AddQuestionModal
        open={openModal}
        setOpen={setOpenModal}
        notifyParent={notifyParent}
        question={question}
      />
      <div className="flex flex-row justify-between w-full h-24 ">
        <div className="flex flex-col max-w-full pl-2">
          <div className="mt-3 flex">
            <div className="h-9 w-9 rounded-full bg-gray-100 mr-2 ml-1">
              <DisplayTypeIcon />
            </div>
            <div className="pt-1.5 max-w-md pr-8 truncate">
              {Title.length ? Title : `Enter ${Type} title`}
            </div>
          </div>

          <div className="flex mt-2 ml-2">
            <div className="ml-8">{Type.charAt(0).toUpperCase() + Type.slice(1)}</div>
          </div>
        </div>
        <div className="relative">
          <div className="-mr-2 flex absolute right-3">
            <SubjectDropdown dropdownContent={dropdownContent} />
          </div>
        </div>
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
