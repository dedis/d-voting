import React, { ChangeEvent, FC, useCallback, useState } from 'react';
import PropTypes from 'prop-types';
import { useTranslation } from 'react-i18next';
import { CloudUploadIcon } from '@heroicons/react/outline';

import { ENDPOINT_EVOTING_CREATE } from 'components/utils/Endpoints';

import { Configuration, Rank, Select, Subject, Text } from '../../../components/utils/types';

import AddButton from './AddButton';
import SubjectComponent from './SubjectComponent';

import { getObjSubject } from './utils/getObjectType';

type ElectionFormProps = {
  setShowModal(modal: any): void;
  setTextModal(text: string): void;
};

const ElectionForm: FC<ElectionFormProps> = ({ setShowModal, setTextModal }) => {
  const { t } = useTranslation();
  const [schema, setSchema] = useState<Configuration>({ MainTitle: '', Scaffold: [] });
  const { MainTitle, Scaffold } = schema;

  async function createHandler() {
    const data = {
      Title: '',
      AdminID: '',
      Token: '',
      Format: schema,
    };

    const req = {
      method: 'POST',
      type: 'application/json',
      body: JSON.stringify(data),
    };
    try {
      const res = await fetch(ENDPOINT_EVOTING_CREATE, req);
      if (res.status !== 200) {
        setTextModal('Failed to create election');
        setShowModal(true);
      } else {
        let response = await res.json();
        setTextModal(`Success creating an election ! ElectionID : ${response.ElectionID}`);
        setShowModal(true);
        setSchema({ MainTitle: '', Scaffold: [] });
      }
    } catch (error) {
      setTextModal(error.message);
      setShowModal(true);
    }
  }

  const addSubject = () => {
    Scaffold.push(getObjSubject());
    setSchema({ ...schema, Scaffold });
  };

  const updateSchema: (
    parentID: string,
    obj: Subject | Rank | Text | Select,
    type: 'ADD' | 'UPDATE' | 'DELETE',
    target: string
  ) => void = useCallback(
    (parentID, obj, type, target) => {
      const modifiedSchema: Configuration = { ...schema };
      const stack: any = [[modifiedSchema]];
      while (stack.length) {
        const [curr, parent]: any = stack.pop();
        // check for match on ID
        if (curr.ID === parentID) {
          switch (type) {
            case 'ADD':
              curr[target].push(obj);
              curr.Order.push(obj.ID);
              break;
            case 'UPDATE':
              if (target === 'Title') {
                curr[target] = obj.Title;
              } else {
                curr[target] = curr[target].map((object) => {
                  if (object.ID === obj.ID) {
                    return obj;
                  }
                  return object;
                });
              }
              break;
            case 'DELETE':
              // special case when target is Subject
              // because the parent can either be of key Scaffold or Subjects
              if (target === 'Subject') {
                const parentTarget = parent.hasOwnProperty('Scaffold') ? 'Scaffold' : 'Subjects';
                parent[parentTarget] = parent[parentTarget].filter((value) => value.ID !== obj.ID);
              } else {
                curr[target] = curr[target].filter((value) => value.ID !== obj.ID);
                curr.Order = curr.Order.filter((value) => value !== obj.ID);
              }
              break;
            default:
              break;
          }
        }
        if (curr.hasOwnProperty('Scaffold')) {
          curr.Scaffold.forEach((child) => stack.push([child, curr]));
        } else {
          curr.Subjects.forEach((child) => stack.push([child, curr]));
        }
      }
      setSchema(modifiedSchema);
    },
    [schema]
  );

  const onMainTitleChange: (e: ChangeEvent<HTMLInputElement>) => void = (e) => {
    e.persist();
    setSchema({ ...schema, MainTitle: e.target.value });
  };

  return (
    <>
      <div className="flex flex-col shadow-lg rounded-md">
        <input
          value={MainTitle}
          onChange={onMainTitleChange}
          name="MainTitle"
          type="text"
          placeholder="Enter the Main title"
          className="ml-3 mt-4 w-60 mb-2 text-lg border rounded-md"
        />
        {Scaffold.map((subject) => (
          <SubjectComponent
            key={subject.ID}
            schema={schema}
            subject={subject}
            updateSchema={updateSchema}
          />
        ))}
        <div>
          <AddButton text="Subject" onClick={addSubject} />
        </div>
      </div>
      <button
        type="button"
        className="flex inline-flex mt-2 mb-2 ml-2 items-center px-4 py-2 border border-transparent rounded-md shadow-sm text-sm font-medium text-white bg-indigo-500 hover:bg-indigo-600"
        onClick={createHandler}>
        <CloudUploadIcon className="-ml-1 mr-2 h-5 w-5" aria-hidden="true" />
        {t('createElection')}
      </button>
    </>
  );
};

ElectionForm.propTypes = {
  setShowModal: PropTypes.func.isRequired,
  setTextModal: PropTypes.func.isRequired,
};

export default ElectionForm;
