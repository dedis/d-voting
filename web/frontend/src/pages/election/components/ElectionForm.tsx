import { FC, useState } from 'react';
import { useTranslation } from 'react-i18next';
import PropTypes from 'prop-types';
import { ENDPOINT_EVOTING_CREATE } from 'components/utils/Endpoints';

import { CloudUploadIcon, TrashIcon } from '@heroicons/react/solid';

import SubjectComponent from './SubjectComponent';
import AddButton from './AddButton';
import UploadFile from './UploadFile';

import configurationSchema from '../../../schema/configurationValidation';
import { Configuration, ID, Subject } from '../../../types/configuration';
import { newSubject } from './utils/getObjectType';

// notifyParent must be used by the child to tell the parent if the subject's
// schema changed.
// RemoveSubject is used by the subject child to notify the parent when the "removeSubject" button
// has been clicked.

type ElectionFormProps = {
  setShowModal(modal: any): void;
  setTextModal(text: string): void;
};

const ElectionForm: FC<ElectionFormProps> = ({ setShowModal, setTextModal }) => {
  // conf is the configuration object containing Maintitle and Scaffold which
  // contains an array of subject.
  const { t } = useTranslation();
  const emptyConf: Configuration = { MainTitle: '', Scaffold: [] };
  const [conf, setConf] = useState<Configuration>(emptyConf);
  const { MainTitle, Scaffold } = conf;

  async function createHandler() {
    const data = {
      Format: conf,
    };
    const req = {
      method: 'POST',
      type: 'application/json',
      body: JSON.stringify(data),
    };

    try {
      await configurationSchema.validate(conf);
      try {
        const res = await fetch(ENDPOINT_EVOTING_CREATE, req);
        const response = await res.json();
        if (res.status !== 200) {
          setShowModal(true);
          setTextModal(`Error HTTP ${res.status} : ${res.statusText}`);
        } else {
          setTextModal(`Success creating an election ! ElectionID : ${response.ElectionID}`);
          setShowModal(true);
          setConf(emptyConf);
        }
      } catch (error) {
        setTextModal(error.message);
        setShowModal(true);
      }
    } catch (err) {
      setTextModal(
        'Incorrect election configuration, please fill it completely: ' + err.errors.join(',')
      );
      setShowModal(true);
    }
  }

  // Called by any of our subject child when they update their schema.
  const notifyParent = (targetID: ID, targetObject: Subject) => {
    const newSubjects = [...Scaffold];
    newSubjects[newSubjects.findIndex((subject) => subject.ID === targetID)] = targetObject;
    setConf({ ...conf, Scaffold: newSubjects });
  };

  const addSubject = () => {
    const newSubjects = [...Scaffold];
    newSubjects.push(newSubject());
    setConf({ ...conf, Scaffold: newSubjects });
  };

  const removeSubject = (subjectID: ID) => () => {
    setConf({
      ...conf,
      Scaffold: Scaffold.filter((subject) => subject.ID !== subjectID),
    });
  };

  return (
    <div className="w-screen px-4 md:px-0 md:w-auto">
      <div className="flex flex-col shadow-lg rounded-md">
        <UploadFile setConf={setConf} setShowModal={setShowModal} setTextModal={setTextModal} />
        <div className="hidden sm:block">
          <div className="py-3 px-4">
            <div className="border-t border-gray-200" />
          </div>
        </div>
        <input
          value={MainTitle}
          onChange={(e) => setConf({ ...conf, MainTitle: e.target.value })}
          name="MainTitle"
          type="text"
          placeholder="Enter the Main title"
          className="ml-3 mt-4 w-60 mb-2 text-lg border rounded-md"
        />
        {Scaffold.map((subject) => (
          <SubjectComponent
            notifyParent={notifyParent}
            subjectObject={subject}
            removeSubject={removeSubject(subject.ID)}
            nestedLevel={0}
            key={subject.ID}
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
        {t('createElec')}
      </button>
      <button
        type="button"
        className="flex inline-flex mt-2 mb-2 ml-2 items-center px-4 py-2 border border-transparent rounded-md shadow-sm text-sm font-medium text-white bg-red-600 hover:bg-red-700"
        onClick={() => {
          if (MainTitle.length || Scaffold.length) setConf(emptyConf);
        }}>
        <TrashIcon className="-ml-1 mr-2 h-5 w-5" aria-hidden="true" />
        {t('clearElec')}
      </button>
    </div>
  );
};

ElectionForm.propTypes = {
  setShowModal: PropTypes.func.isRequired,
  setTextModal: PropTypes.func.isRequired,
};

export default ElectionForm;
