import { FC, useEffect, useRef, useState } from 'react';

import { ID, Subject } from '../../../types/configuration';
import { newRank, newSelect, newSubject, newText } from './utils/getObjectType';

import AddButton from './AddButton';
import QuestionComponent from './QuestionComponent';
import DeleteButton from './DeleteButton';

const MAX_NESTED_SUBJECT = 2;

type SubjectComponentProps = {
  notifyParent: (targetID: ID, targetObject: Subject) => void;
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
  const [components, setComponents] = useState([]);

  const { Title, Subjects, Ranks, Selects, Texts, Order } = subject;

  /** 
    When a property changes, we notify the parent with the new subject object
  **/
  useEffect(() => {
    if (!isSubjectMounted.current) {
      isSubjectMounted.current = true;
      return;
    }
    notifyParent(subject.ID, subject);
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [subject]);

  const localNotifyParent = (targetID: ID, targetObject: any) => {
    const newSubjects = [...Subjects];
    newSubjects[newSubjects.findIndex((subj) => subj.ID === targetID)] = targetObject;
    setSubject({ ...subject, Subjects: newSubjects, Title });
  };

  const notifySubject = (targetID: ID, type: 'RANK' | 'TEXT' | 'SELECT') => (targetObject: any) => {
    switch (type) {
      case 'RANK':
        const newRanks = [...Ranks];
        newRanks[newRanks.findIndex((rank) => rank.ID === targetID)] = targetObject;
        setSubject({ ...subject, Ranks: newRanks });
        break;
      case 'SELECT':
        const newSelects = [...Selects];
        newSelects[newSelects.findIndex((select) => select.ID === targetID)] = targetObject;
        setSubject({ ...subject, Selects: newSelects });
        break;
      case 'TEXT':
        const newTexts = [...Texts];
        newTexts[newTexts.findIndex((text) => text.ID === targetID)] = targetObject;
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

  const addQuestion = (type: 'RANK' | 'TEXT' | 'SELECT') => {
    switch (type) {
      case 'RANK':
        const newRanks = [...Ranks];
        const rank = newRank();
        newRanks.push(rank);
        setSubject({ ...subject, Ranks: newRanks, Order: [...Order, rank.ID] });
        break;
      case 'TEXT':
        const newTexts = [...Texts];
        const text = newText();
        newTexts.push(text);
        setSubject({ ...subject, Texts: newTexts, Order: [...Order, text.ID] });
        break;
      case 'SELECT':
        const newSelects = [...Selects];
        const select = newSelect();
        newSelects.push(select);
        setSubject({ ...subject, Selects: newSelects, Order: [...Order, select.ID] });
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

  const removeChildQuestion = (targetID: ID, type: 'RANK' | 'TEXT' | 'SELECT') => () => {
    switch (type) {
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
    const findSubject = (id: string) => {
      if (Subjects) {
        if (Subjects.find((subj) => subj.ID === id)) {
          const subjectFound = Subjects.find((subj) => subj.ID === id);
          return (
            <SubjectComponent
              notifyParent={localNotifyParent}
              removeSubject={localRemoveSubject(subjectFound.ID)}
              subjectObject={subjectFound}
              nestedLevel={nestedLevel + 1}
              key={subjectFound.ID}
            />
          );
        }
        return false;
      }
      return false;
    };

    const findRanks = (id: string) => {
      if (Ranks) {
        if (Ranks.find((rank) => rank.ID === id)) {
          const rankFound = Ranks.find((rank) => rank.ID === id);
          return (
            <QuestionComponent
              key={`rank${rankFound.ID}`}
              obj={rankFound}
              notifyParent={notifySubject(rankFound.ID, 'RANK')}
              removeQuestion={removeChildQuestion(rankFound.ID, 'RANK')}
              type={'RANK'}
            />
          );
        }
        return false;
      }
      return false;
    };

    const findTexts = (id: string) => {
      if (Texts) {
        if (Texts.find((text) => text.ID === id)) {
          const textFound = Texts.find((rank) => rank.ID === id);
          return (
            <QuestionComponent
              key={`text${textFound.ID}`}
              obj={textFound}
              notifyParent={notifySubject(textFound.ID, 'TEXT')}
              removeQuestion={removeChildQuestion(textFound.ID, 'TEXT')}
              type={'TEXT'}
            />
          );
        }
        return false;
      }
      return false;
    };

    const findSelects = (id: string) => {
      if (Selects) {
        if (Selects.find((select) => select.ID === id)) {
          const selectFound = Selects.find((rank) => rank.ID === id);
          return (
            <QuestionComponent
              key={`select${selectFound.ID}`}
              obj={selectFound}
              notifyParent={notifySubject(selectFound.ID, 'SELECT')}
              removeQuestion={removeChildQuestion(selectFound.ID, 'SELECT')}
              type={'SELECT'}
            />
          );
        }
        return false;
      }
      return false;
    };

    const orderData = () => {
      setComponents(
        Order.map((id) => {
          if (findSubject(id)) {
            return findSubject(id);
          } else if (findRanks(id)) {
            return findRanks(id);
          } else if (findTexts(id)) {
            return findTexts(id);
          } else if (findSelects(id)) {
            return findSelects(id);
          } else return null;
        })
      );
    };

    orderData();
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [Title, Subjects, Ranks, Selects, Texts, Order, nestedLevel]);

  return (
    <div className="ml-4 mb-4 mr-2 shadow-lg rounded-md">
      <DeleteButton text="Subject" onClick={removeSubject} />
      <input
        value={Title}
        onChange={(e) => setSubject({ ...subject, Title: e.target.value })}
        name="Title"
        type="text"
        placeholder="Enter the Subject Title"
        className="ml-2 mt-2 mb-2 border rounded-md text-md w-60"
      />
      <div className="ml-2">ID: {subject.ID}</div>
      {components.map((component) => component)}

      <div className="flex hidden justify-between overflow-x-auto sm:flex sm:pr-2">
        <div>
          <AddButton text="Rank" onClick={() => addQuestion('RANK')} />
          <AddButton text="Select" onClick={() => addQuestion('SELECT')} />
          <AddButton text="Text" onClick={() => addQuestion('TEXT')} />
        </div>
        {nestedLevel < MAX_NESTED_SUBJECT && <AddButton text="Subject" onClick={addSubject} />}
      </div>
      <div className="flex visible overflow-x-auto sm:hidden">
        <AddButton text="Rank" onClick={() => addQuestion('RANK')} />
        <AddButton text="Select" onClick={() => addQuestion('SELECT')} />
        <AddButton text="Text" onClick={() => addQuestion('TEXT')} />
        {nestedLevel < MAX_NESTED_SUBJECT && <AddButton text="Subject" onClick={addSubject} />}
      </div>
    </div>
  );
};

export default SubjectComponent;
