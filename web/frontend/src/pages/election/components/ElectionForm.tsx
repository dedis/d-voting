import { FC, Fragment, useState } from 'react';
import { useTranslation } from 'react-i18next';
import { newElection } from 'components/utils/Endpoints';

import { CloudUploadIcon, PencilIcon, TrashIcon } from '@heroicons/react/solid';

import SubjectComponent from './SubjectComponent';
import UploadFile from './UploadFile';

import configurationSchema from '../../../schema/configurationValidation';
import { Configuration, ID, Subject } from '../../../types/configuration';
import { emptyConfiguration, newSubject } from '../../../types/getObjectType';
import { marshalConfig } from '../../../types/JSONparser';
import DownloadButton from 'components/buttons/DownloadButton';
import SpinnerIcon from 'components/utils/SpinnerIcon';
import RedirectToModal from 'components/modal/RedirectToModal';
import { CheckIcon, PlusSmIcon } from '@heroicons/react/outline';
import Tabs from './Tabs';

// notifyParent must be used by the child to tell the parent if the subject's
// schema changed.

// removeSubject is used by the subject child to notify the
// parent when the "removeSubject" button has been clicked.

type ElectionFormProps = {};

const ElectionForm: FC<ElectionFormProps> = () => {
  // conf is the configuration object containing MainTitle and Scaffold which
  // contains an array of subject.
  const { t } = useTranslation();
  const emptyConf: Configuration = emptyConfiguration();
  const [conf, setConf] = useState<Configuration>(emptyConf);
  const [loading, setLoading] = useState<boolean>(false);
  const [showModal, setShowModal] = useState<boolean>(false);
  const [textModal, setTextModal] = useState<string>('');
  const [currentTab, setCurrentTab] = useState<string>('electionForm');
  const [titleChanging, setTitleChanging] = useState<boolean>(true);
  const [navigateDestination, setNavigateDestination] = useState(null);
  const { MainTitle, Scaffold } = conf;

  async function createHandler() {
    const data = {
      Configuration: marshalConfig(conf),
    };
    const req = {
      method: 'POST',
      body: JSON.stringify(data),
      headers: { 'Content-Type': 'application/json' },
    };

    try {
      await configurationSchema.validate(data.Configuration);
    } catch (err) {
      setTextModal(t('errorIncorrectConfSchema') + err.errors.join(','));
      setShowModal(true);
      return;
    }

    try {
      setLoading(true);
      const res = await fetch(newElection, req);
      if (res.status !== 200) {
        const response = await res.text();
        setTextModal(`Error HTTP ${res.status} (${res.statusText}) : ${response}`);
        setShowModal(true);
      } else {
        const response = await res.json();
        setNavigateDestination('/elections/' + response.ElectionID);
        setTextModal(`${t('successCreateElection')} ${response.ElectionID}`);
        setShowModal(true);
        setConf(emptyConf);
      }
      setLoading(false);
    } catch (error) {
      setTextModal(error.message);
      setShowModal(true);
      setLoading(false);
    }
  }

  // exports the data to a JSON file, marshall the configuration state object
  // before exporting it
  const exportData = async () => {
    const data = marshalConfig(conf);
    try {
      await configurationSchema.validate(data);
    } catch (err) {
      setTextModal(t('errorIncorrectConfSchema') + err.errors.join(','));
      setShowModal(true);
      return;
    }
    const jsonString = `data:text/json;chatset=utf-8,${encodeURIComponent(JSON.stringify(data))}`;
    const link = document.createElement('a');
    link.href = jsonString;
    link.download = 'election_configuration.json';
    link.click();
  };

  // Called by any of our subject child when they update their schema.
  const notifyParent = (subject: Subject) => {
    const newSubjects = [...Scaffold];
    newSubjects[newSubjects.findIndex((s) => s.ID === subject.ID)] = subject;
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

  const DisplayElectionForm = () => {
    return (
      <div className="w-screen px-4 md:px-0 md:w-auto">
        <div className="flex flex-col border rounded-md">
          <div className="flex mt-3 mb-2">
            {titleChanging ? (
              <>
                <input
                  value={MainTitle}
                  onChange={(e) => setConf({ ...conf, MainTitle: e.target.value })}
                  name="MainTitle"
                  type="text"
                  placeholder={t('enterMainTitle')}
                  className="ml-3 w-60 text-lg border rounded-md"
                />
                <div className="ml-1">
                  <button
                    className={`border p-1 rounded-md ${
                      MainTitle.length === 0 ? 'bg-gray-100' : ' '
                    }`}
                    disabled={MainTitle.length === 0}
                    onClick={() => setTitleChanging(false)}>
                    <CheckIcon className="h-5 w-5" aria-hidden="true" />
                  </button>
                </div>
              </>
            ) : (
              <>
                <div className="mt-1 ml-3">{MainTitle}</div>
                <div className="ml-1">
                  <button
                    className="hover:text-indigo-500 p-1 rounded-md"
                    onClick={() => setTitleChanging(true)}>
                    <PencilIcon className="m-1 h-3 w-3" aria-hidden="true" />
                  </button>
                </div>
              </>
            )}
          </div>

          {Scaffold.map((subject) => (
            <SubjectComponent
              notifyParent={notifyParent}
              subjectObject={subject}
              removeSubject={removeSubject(subject.ID)}
              nestedLevel={0}
              key={subject.ID}
            />
          ))}
          <button
            onClick={addSubject}
            className="flex w-full h-12  border-t  px-4 py-3 text-left text-sm font-medium hover:bg-gray-50">
            <PlusSmIcon className="mr-2 h-5 w-5" aria-hidden="true" />
            Add Subject
          </button>
        </div>
        <div className="my-2">
          <button
            type="button"
            className="inline-flex my-2 ml-2 items-center px-4 py-2 border border-transparent rounded-md  text-sm font-medium text-white bg-indigo-500 hover:bg-indigo-600"
            onClick={createHandler}>
            {loading ? (
              <SpinnerIcon />
            ) : (
              <CloudUploadIcon className="-ml-1 mr-2 h-5 w-5" aria-hidden="true" />
            )}
            {t('createElec')}
          </button>
          <button
            type="button"
            className="inline-flex my-2 ml-2 items-center px-4 py-2 border border-transparent rounded-md text-sm font-medium text-white bg-red-600 hover:bg-red-700"
            onClick={() => {
              setTitleChanging(true);
              setConf(emptyConf);
            }}>
            <TrashIcon className="-ml-1 mr-2 h-5 w-5" aria-hidden="true" />
            {t('clearElec')}
          </button>
          <DownloadButton exportData={exportData}>{t('exportElecJSON')}</DownloadButton>
        </div>
      </div>
    );
  };

  const DisplayPreviewElection = () => {
    return (
      <div className="border mx-4 sm:mx-0 mb-4 rounded-md">
        <div className="h-[calc(100vh-265px)]  ml-2">preview</div>
      </div>
    );
  };

  const SwitchTabs = () => {
    switch (currentTab) {
      case 'electionForm':
        return <DisplayElectionForm />;
      case 'previewForm':
        return <DisplayPreviewElection />;
        break;
    }
  };

  return (
    <>
      <RedirectToModal
        showModal={showModal}
        setShowModal={setShowModal}
        title={t('notification')}
        buttonRightText={t('close')}
        navigateDestination={navigateDestination}>
        {textModal}
      </RedirectToModal>

      <UploadFile
        updateForm={(config: Configuration) => {
          setTitleChanging(false);
          setConf(config);
        }}
        setShowModal={setShowModal}
        setTextModal={setTextModal}
      />

      <div className="hidden md:grid grid-cols-2 gap-2">
        <DisplayElectionForm />
        <DisplayPreviewElection />
      </div>
      <div className="flex flex-col md:hidden">
        <Tabs currentTab={currentTab} setCurrentTab={setCurrentTab} />
        <SwitchTabs />
      </div>
    </>
  );
};

ElectionForm.propTypes = {};

export default ElectionForm;
