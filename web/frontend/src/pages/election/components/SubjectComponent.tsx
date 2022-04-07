import { FC, ReactElement, useEffect, useRef, useState } from 'react';
import PropTypes from 'prop-types';

import { ID, Rank, Select, Subject, Text } from '../../../types/configuration';
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

  const { Title, Subjects, Ranks, Selects, Texts, Order } = subject;

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
    const newSubjects = [...Subjects];
    newSubjects[newSubjects.findIndex((s) => s.ID === subj.ID)] = subj;
    setSubject({ ...subject, Subjects: newSubjects, Title });
  };

  const notifySubject = (question: Rank | Select | Text) => {
    switch (question.Type) {
      case 'RANK':
        const r = question as Rank;
        const newRanks = [...Ranks];
        newRanks[newRanks.findIndex((rank) => rank.ID === r.ID)] = r;
        setSubject({ ...subject, Ranks: newRanks });
        break;
      case 'SELECT':
        const s = question as Select;
        const newSelects = [...Selects];
        newSelects[newSelects.findIndex((select) => select.ID === s.ID)] = s;
        setSubject({ ...subject, Selects: newSelects });
        break;
      case 'TEXT':
        const t = question as Text;
        const newTexts = [...Texts];
        newTexts[newTexts.findIndex((text) => text.ID === t.ID)] = t;
        setSubject({ ...subject, Texts: newTexts });
        break;
      default:
        break;
    }
  };

  const addSubject = () => {
    const newSubjects = [...Subjects];
    const newSubj = newSubject();
    newSubjects.push(newSubj);
    setSubject({ ...subject, Subjects: newSubjects, Order: [...Order, newSubj.ID] });
  };

  const addQuestion = (question: Rank | Select | Text) => {
    switch (question.Type) {
      case 'RANK':
        const r = question as Rank;
        const newRanks = [...Ranks];
        newRanks.push(r);
        setSubject({ ...subject, Ranks: newRanks, Order: [...Order, r.ID] });
        break;
      case 'TEXT':
        const t = question as Text;
        const newTexts = [...Texts];
        newTexts.push(t);
        setSubject({ ...subject, Texts: newTexts, Order: [...Order, t.ID] });
        break;
      case 'SELECT':
        const s = question as Select;
        const newSelects = [...Selects];
        newSelects.push(s);
        setSubject({ ...subject, Selects: newSelects, Order: [...Order, s.ID] });
        break;
      default:
        break;
    }
  };

  const localRemoveSubject = (subjID: ID) => () => {
    const newSubjects = [...Subjects];
    setSubject({
      ...subject,
      Subjects: newSubjects.filter((subj) => subj.ID !== subjID),
      Order: Order.filter((id) => id !== subjID),
    });
  };

  const removeChildQuestion = (question: Rank | Select | Text) => () => {
    const targetID = question.ID;
    switch (question.Type) {
      case 'RANK':
        setSubject({
          ...subject,
          Ranks: Ranks.filter((rank) => rank.ID !== targetID),
          Order: Order.filter((id) => id !== targetID),
        });
        break;
      case 'TEXT':
        setSubject({
          ...subject,
          Texts: Texts.filter((text) => text.ID !== targetID),
          Order: Order.filter((id) => id !== targetID),
        });
        break;
      case 'SELECT':
        setSubject({
          ...subject,
          Selects: Selects.filter((select) => select.ID !== targetID),
          Order: Order.filter((id) => id !== targetID),
        });
        break;
      default:
        break;
    }
  };

  // Sorts the questions components & sub-subjects according to their Order into
  // the components state array
  useEffect(() => {
    // findQuestion return the react element based on the question/subject ID.
    // Returns undefined if the question/subject ID is unknown.
    const findQuestion = (id: string): ReactElement => {
      const all: (Subject | Text | Select | Rank)[] = [...Subjects, ...Texts, ...Selects, ...Ranks];
      const found = all.find((el) => el.ID === id);

      if (found === undefined) {
        return undefined;
      }

      switch (found.Type) {
        case 'TEXT':
          const text = found as Text;
          return (
            <Question
              key={`text${text.ID}`}
              question={text}
              notifyParent={notifySubject}
              removeQuestion={removeChildQuestion(text)}
            />
          );
        case 'SUBJECT':
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
        case 'RANK':
          const rank = found as Rank;
          return (
            <Question
              key={`rank${rank.ID}`}
              question={rank}
              notifyParent={notifySubject}
              removeQuestion={removeChildQuestion(rank)}
            />
          );
        case 'SELECT':
          const select = found as Select;
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
  }, [Title, Subjects, Ranks, Selects, Texts, Order, nestedLevel]);

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
