import { FC, ReactElement, useEffect, useRef, useState } from 'react';
import PropTypes from 'prop-types';

import {
  ID,
  RANK,
  RankQuestion,
  SELECT,
  SUBJECT,
  SelectQuestion,
  Subject,
  SubjectElement,
  TEXT,
  TextQuestion,
} from '../../../types/configuration';
import { newRank, newSelect, newSubject, newText } from '../../../types/getObjectType';

import AddButton from './AddButton';
import Question from './Question';
import DeleteButton from './DeleteButton';

const MAX_NESTED_SUBJECT = 2;

type SubjectComponentProps = {
  notifyParent: (subject: Subject) => void;
  removeSubject: () => void;
  subjectObject: Subject;
  nestedLevel: number;
};

const SubjectComponent: FC<SubjectComponentProps> = ({
  notifyParent,
  removeSubject,
  subjectObject,
  nestedLevel,
}) => {
  const [subject, setSubject] = useState<Subject>(subjectObject);
  const isSubjectMounted = useRef<Boolean>(false);
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

  const localNotifyParent = (subj: Subject) => {
    const newElements = new Map(Elements);
    newElements.set(subj.ID, subj);
    setSubject({ ...subject, Elements: newElements, Title });
  };

  const notifySubject = (question: SubjectElement) => {
    const newElements = new Map(Elements);

    switch (question.Type) {
      case RANK:
        const r = question as RankQuestion;
        newElements.set(r.ID, r);
        break;
      case SELECT:
        const s = question as SelectQuestion;
        newElements.set(s.ID, s);
        break;
      case TEXT:
        const t = question as TextQuestion;
        newElements.set(t.ID, t);
        break;
      default:
        break;
    }

    setSubject({ ...subject, Elements: newElements });
  };

  const addSubject = () => {
    const newElements = new Map(Elements);
    const newSubj = newSubject();
    newElements.set(newSubj.ID, newSubj);
    setSubject({ ...subject, Elements: newElements, Order: [...Order, newSubj.ID] });
  };

  const addQuestion = (question: SubjectElement) => {
    const newElements = new Map(Elements);
    newElements.set(question.ID, question);
    setSubject({ ...subject, Elements: newElements, Order: [...Order, question.ID] });
  };

  const localRemoveSubject = (subjID: ID) => () => {
    const newElements = new Map(Elements);
    newElements.delete(subjID);
    setSubject({
      ...subject,
      Elements: newElements,
      Order: Order.filter((id) => id !== subjID),
    });
  };

  const removeChildQuestion = (question: SubjectElement) => () => {
    const newElements = new Map(Elements);
    newElements.delete(question.ID);
    setSubject({
      ...subject,
      Elements: newElements,
      Order: Order.filter((id) => id !== question.ID),
    });
  };

  // Sorts the questions components & sub-subjects according to their Order into
  // the components state array
  useEffect(() => {
    // findQuestion return the react element based on the question/subject ID.
    // Returns undefined if the question/subject ID is unknown.
    const findQuestion = (id: ID): ReactElement => {
      if (!Elements.has(id)) {
        return undefined;
      }

      const found = Elements.get(id);

      switch (found.Type) {
        case TEXT:
          const text = found as TextQuestion;
          return (
            <Question
              key={`text${text.ID}`}
              question={text}
              notifyParent={notifySubject}
              removeQuestion={removeChildQuestion(text)}
            />
          );
        case SUBJECT:
          const sub = found as Subject;
          return (
            <SubjectComponent
              notifyParent={localNotifyParent}
              removeSubject={localRemoveSubject(sub.ID)}
              subjectObject={sub}
              nestedLevel={nestedLevel + 1}
              key={sub.ID}
            />
          );
        case RANK:
          const rank = found as RankQuestion;
          return (
            <Question
              key={`rank${rank.ID}`}
              question={rank}
              notifyParent={notifySubject}
              removeQuestion={removeChildQuestion(rank)}
            />
          );
        case SELECT:
          const select = found as SelectQuestion;
          return (
            <Question
              key={`select${select.ID}`}
              question={select}
              notifyParent={notifySubject}
              removeQuestion={removeChildQuestion(select)}
            />
          );
      }
    };

    setComponents(Order.map((id) => findQuestion(id)));

    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [Title, Elements, Order, nestedLevel]);

  return (
    <div className="ml-4 mb-4 mr-2 shadow-lg rounded-md">
      <DeleteButton onClick={removeSubject}>Subject</DeleteButton>
      <input
        value={Title}
        onChange={(e) => setSubject({ ...subject, Title: e.target.value })}
        name="Title"
        type="text"
        placeholder="Enter the Subject Title"
        className="ml-2 mt-2 mb-2 border rounded-md text-md w-60"
      />
      {components.map((component) => component)}

      <div className="flex hidden justify-between overflow-x-auto sm:flex sm:pr-2">
        <div>
          <AddButton onClick={() => addQuestion(newRank())}>Rank</AddButton>
          <AddButton onClick={() => addQuestion(newSelect())}>Select</AddButton>
          <AddButton onClick={() => addQuestion(newText())}>Text</AddButton>
        </div>
        {nestedLevel < MAX_NESTED_SUBJECT && <AddButton onClick={addSubject}>Subject</AddButton>}
      </div>
      <div className="flex visible overflow-x-auto sm:hidden">
        <AddButton onClick={() => addQuestion(newRank())}>Rank</AddButton>
        <AddButton onClick={() => addQuestion(newSelect())}>Select</AddButton>
        <AddButton onClick={() => addQuestion(newText())}>Text</AddButton>
        {nestedLevel < MAX_NESTED_SUBJECT && <AddButton onClick={addSubject}>Subject</AddButton>}
      </div>
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
