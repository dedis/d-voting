import { FC, useEffect, useState } from 'react';

import { Configuration, Rank, Select, Subject, Text } from '../../../components/utils/types';
import { getObjRank, getObjSelect, getObjSubject, getObjText } from './utils/getObjectType';

import AddButton from './AddButton';
import DeleteButton from './DeleteButton';
import QuestionComponent from './QuestionComponent';

type SubjectComponentProps = {
  schema: Configuration;
  updateSchema(
    parentID: string,
    obj: Subject | Rank | Select | Text,
    type: 'ADD' | 'UPDATE' | 'DELETE',
    target: string
  ): void;
  subject: Subject;
};

const SubjectComponent: FC<SubjectComponentProps> = ({ schema, updateSchema, subject }) => {
  const { ID, Title, Subjects, Ranks, Selects, Texts, Order } = subject;
  const [components, setComponents] = useState<any>([]);

  useEffect(() => {
    const orderData = () => {
      setComponents(
        Order.map((id) => {
          if (Subjects.find((subj) => subj.ID === id)) {
            return (
              <SubjectComponent
                key={`subject${id}`}
                schema={schema}
                updateSchema={updateSchema}
                subject={Subjects.find((subj) => subj.ID === id)}
              />
            );
          } else if (Ranks.find((rank) => rank.ID === id)) {
            return (
              <QuestionComponent
                key={`rank${id}`}
                parentID={ID}
                obj={Ranks.find((rank) => rank.ID === id)}
                updateSchema={updateSchema}
                type={'Rank'}
              />
            );
          } else if (Texts.find((text) => text.ID === id)) {
            return (
              <QuestionComponent
                key={`text${id}`}
                parentID={ID}
                obj={Texts.find((text) => text.ID === id)}
                updateSchema={updateSchema}
                type={'Text'}
              />
            );
          } else if (Selects.find((select) => select.ID === id)) {
            return (
              <QuestionComponent
                key={`select${id}`}
                parentID={ID}
                obj={Selects.find((select) => select.ID === id)}
                updateSchema={updateSchema}
                type={'Select'}
              />
            );
          } else return <></>;
        })
      );
    };

    orderData();
  }, [schema, Order, ID, Ranks, Selects, Texts, Subjects, updateSchema]);

  return (
    <div className="ml-4 mb-4 mr-2 shadow-lg rounded-md">
      <DeleteButton
        text="Subject"
        onClick={() => {
          updateSchema(ID, subject, 'DELETE', 'Subject');
        }}
      />
      <input
        value={Title}
        onChange={(e) => updateSchema(ID, { ...subject, Title: e.target.value }, 'UPDATE', 'Title')}
        name="Title"
        type="text"
        placeholder="Enter the Title"
        className="ml-2 mt-2 mb-2 border rounded-md text-md w-60"
      />
      {components.map((component) => component)}
      <div className="flex">
        <AddButton
          text="Subject"
          onClick={() => updateSchema(ID, getObjSubject(), 'ADD', 'Subjects')}
        />
        <AddButton text="Rank" onClick={() => updateSchema(ID, getObjRank(), 'ADD', 'Ranks')} />
        <AddButton
          text="Select"
          onClick={() => updateSchema(ID, getObjSelect(), 'ADD', 'Selects')}
        />
        <AddButton text="Text" onClick={() => updateSchema(ID, getObjText(), 'ADD', 'Texts')} />
      </div>
    </div>
  );
};

export default SubjectComponent;
