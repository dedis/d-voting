import { FC, Fragment, useEffect, useState } from 'react';
import { useTranslation } from 'react-i18next';
import { newForm } from 'components/utils/Endpoints';

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
import RemoveElementModal from './RemoveElementModal';
import { useConfiguration } from 'components/utils/useConfiguration';
import BallotDisplay from 'pages/ballot/components/BallotDisplay';
import usePostCall from 'components/utils/usePostCall';
import * as endpoints from 'components/utils/Endpoints';

// notifyParent must be used by the child to tell the parent if the subject's
// schema changed.

// removeSubject is used by the subject child to notify the
// parent when the "removeSubject" button has been clicked.

type FormFormProps = {};

const FormForm: FC<FormFormProps> = () => {
  // conf is the configuration object containing MainTitle and Scaffold which
  // contains an array of subject.
  const { t } = useTranslation();
  const emptyConf: Configuration = emptyConfiguration();
  const [conf, setConf] = useState<Configuration>(emptyConf);
  const [loading, setLoading] = useState<boolean>(false);
  const [showModal, setShowModal] = useState<boolean>(false);
  const [showRemoveSubjectModal, setShowRemoveSubjectModal] = useState<boolean>(false);
  const [textModal, setTextModal] = useState<string>('');
  const [currentTab, setCurrentTab] = useState<string>('formForm');
  const [subjectIdToRemove, setSubjectIdToRemove] = useState<ID>('');
  const [titleChanging, setTitleChanging] = useState<boolean>(true);
  const [navigateDestination, setNavigateDestination] = useState(null);
  const [marshalledConf, setMarshalledConf] = useState<any>(marshalConfig(conf));
  const { configuration: previewConf, answers, setAnswers } = useConfiguration(marshalledConf);

  const { MainTitle, Scaffold } = conf;
  const regexPattern = /[^a-zA-Z0-9]/g;

  const [postError, setPostError] = useState(null);
  const [, setIsPosting] = useState(false);
  const sendFetchRequest = usePostCall(setPostError);
  const AuthorizationUpdate = (FormID: string): Promise<boolean> => {
    const req = {
      method: 'PUT',
      body: JSON.stringify({
        FormID,
      }),
      headers: {
        'Content-Type': 'application/json',
      },
    };

    return sendFetchRequest(endpoints.addFormAuthorization, req, setIsPosting);
  };

  useEffect(() => {
    setMarshalledConf(marshalConfig(conf));
  }, [conf]);

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
      const res = await fetch(newForm, req);
      if (res.status !== 200) {
        const response = await res.text();
        setTextModal(`Error HTTP ${res.status} (${res.statusText}) : ${response}`);
        setShowModal(true);
      } else {
        const response = await res.json();
        AuthorizationUpdate(response.FormID);
        setNavigateDestination('/forms/' + response.FormID);
        setTextModal(`${t('successCreateForm')} ${response.FormID}`);
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
    const title = MainTitle.replace(regexPattern, '_').slice(0, 99); // replace spaces with underscores
    link.download = title + '.json';
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

  const handleConfirmRemoveSubject = () => {
    setConf({
      ...conf,
      Scaffold: Scaffold.filter((subject) => subject.ID !== subjectIdToRemove),
    });
    setSubjectIdToRemove('');
  };

  const displayFormForm = () => {
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
                  className="ml-3 px-1 w-60 text-lg border rounded-md"
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
                <div
                  className="mt-1 ml-3 w-[90%] break-words"
                  onClick={() => setTitleChanging(true)}>
                  {MainTitle}
                </div>
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
              removeSubject={() => {
                setSubjectIdToRemove(subject.ID);
                setShowRemoveSubjectModal(true);
              }}
              nestedLevel={0}
              key={subject.ID}
            />
          ))}
          <button
            onClick={addSubject}
            className="flex w-full h-12  border-t  px-4 py-3 text-left text-sm font-medium hover:bg-gray-50">
            <PlusSmIcon className="mr-2 h-5 w-5" aria-hidden="true" />
            {t('addSubject')}
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
            {t('createForm')}
          </button>
          <button
            type="button"
            className="inline-flex my-2 ml-2 items-center px-4 py-2 border border-transparent rounded-md text-sm font-medium text-white bg-red-600 hover:bg-red-700"
            onClick={() => {
              setTitleChanging(true);
              setConf(emptyConf);
            }}>
            <TrashIcon className="-ml-1 mr-2 h-5 w-5" aria-hidden="true" />
            {t('clearForm')}
          </button>
          <DownloadButton exportData={exportData}>{t('exportJSON')}</DownloadButton>
        </div>
      </div>
    );
  };

  const displayPreviewForm = () => {
    return (
      <div className="w-screen px-4 md:px-0 mb-4 md:w-auto">
        <div className="border rounded-md">
          <div className="ml-2 w-[95%]">
            <BallotDisplay
              configuration={previewConf}
              answers={answers}
              setAnswers={setAnswers}
              userErrors=""
            />
          </div>
        </div>
      </div>
    );
  };

  const switchTabs = () => {
    switch (currentTab) {
      case 'formForm':
        return displayFormForm();
      case 'previewForm':
        return displayPreviewForm();
    }
  };

  return (
    <>
      <RemoveElementModal
        showModal={showRemoveSubjectModal}
        setShowModal={setShowRemoveSubjectModal}
        textModal={t('confirmRemovesubject')}
        handleConfirm={handleConfirmRemoveSubject}
      />
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
        {displayFormForm()}
        {displayPreviewForm()}
      </div>
      <div className="flex flex-col md:hidden">
        <Tabs currentTab={currentTab} setCurrentTab={setCurrentTab} />
        {switchTabs()}
      </div>
    </>
  );
};

FormForm.propTypes = {};

export default FormForm;
