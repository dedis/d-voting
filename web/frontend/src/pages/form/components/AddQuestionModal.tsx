import { FC, Fragment, useRef, useState } from 'react';
import PropTypes from 'prop-types';
import { Dialog, Transition } from '@headlessui/react';
import { useTranslation } from 'react-i18next';
import { CheckIcon, MinusCircleIcon, PlusCircleIcon } from '@heroicons/react/outline';
import {
  RANK,
  RankQuestion,
  SELECT,
  SelectQuestion,
  TEXT,
  TextQuestion,
} from 'types/configuration';
import { ranksSchema, selectsSchema, textsSchema } from '../../../schema/configurationValidation';
import useQuestionForm from './utils/useQuestionForm';
import DisplayTypeIcon from './DisplayTypeIcon';

import { availableLanguages } from 'language/Configuration';

//import TranslatableInput from './TranslatableInput';
type AddQuestionModalProps = {
  question: RankQuestion | SelectQuestion | TextQuestion;
  open: boolean;
  setOpen(opened: boolean): void;
  notifyParent(question: RankQuestion | SelectQuestion | TextQuestion): void;
  handleClose: () => void;
};

const MAX_MINN = 20;

const AddQuestionModal: FC<AddQuestionModalProps> = ({
  question,
  open,
  setOpen,
  handleClose,
  notifyParent,
}) => {
  const { ID, Type } = question;
  const { t } = useTranslation();
  const {
    state: values,
    handleChange,
    addChoice,
    deleteChoice,
    updateChoice,
  } = useQuestionForm(question);
  const [language, setLanguage] = useState('en');
  const { Title, TitleDe, TitleFr, MaxN, MinN, Choices, ChoicesMap, Hint, HintFr, HintDe } = values;
  const [errors, setErrors] = useState([]);
  const handleSave = async () => {
    try {
      switch (Type) {
        case TEXT:
          await textsSchema.validate(values, { abortEarly: false });
          break;
        case RANK:
          await ranksSchema.validate(values, { abortEarly: false });
          break;
        case SELECT:
          await selectsSchema.validate(values, { abortEarly: false });
          break;
        default:
          break;
      }
      setErrors([]);
      notifyParent(values);
      setOpen(false);
    } catch (err: any) {
      console.log('erreur', err);
      setErrors(err.errors);
    }
  };

  const handleAddChoice = (e) => {
    switch (Type) {
      case RANK:
        handleChange('addChoiceRank')(e);
        break;
      default:
        addChoice(language);
        break;
    }
  };

  const handleDeleteChoice = (index: number) => (e) => {
    switch (Type) {
      case RANK:
        handleChange('deleteChoiceRank', index)(e);
        break;
      default:
        deleteChoice(index);
        break;
    }
  };

  const cancelButtonRef = useRef(null);

  const displayExtraFields = () => {
    switch (Type) {
      case TEXT:
        const tq = values as TextQuestion;
        return (
          <>
            <label className="block text-md font-medium text-gray-500">MaxLength</label>
            <input
              value={tq.MaxLength}
              onChange={handleChange('TextMaxLength')}
              name="MaxLength"
              min="1"
              type="number"
              placeholder={t('enterMaxLength')}
              className="my-1 px-1 w-32 ml-1 border rounded-md"
            />
            <label className="block text-md font-medium text-gray-500">Regex</label>
            <input
              value={tq.Regex}
              onChange={handleChange()}
              name="Regex"
              type="text"
              placeholder={t('enterRegex')}
              className="my-1 px-1 w-40 ml-1 border rounded-md"
            />
          </>
        );
      default:
        return;
    }
  };
  /*type renderQuestionModalProps = {
    lg : string;
    nameC: string[];
  };
  const RenderQuestionModal : FC<renderQuestionModalProps> = ({lg ,nameC}) => {
    return(<div className="pb-2">
                        {Choices.get('en').map((choice: string, idx: number) => (
                          <div className="flex w-60" key={`${ID}wrapper${idx}`}>
                            <input
                              key={`${ID}choice${idx}`}
                              value={choice}
                              onChange={updateChoice(idx, language)}
                              name="Choice"
                              type="text"
                              placeholder={
                                Type !== TEXT ? `${'Choice'}` + ` ${idx + 1}` : `Answer ${idx + 1}`
                              }
                              className="my-1 px-1 w-60 ml-2 border rounded-md"
                            />
                            <div className="flex ml-1 mt-1.2">
                              {Choices.get('en').length > 1 && (
                                <button
                                  key={`${ID}deleteChoice${idx}`}
                                  type="button"
                                  className="inline-flex items-center border border-transparent rounded-full font-medium text-gray-300 hover:text-gray-400"
                                  onClick={handleDeleteChoice(idx)}>
                                  <MinusCircleIcon className="h-5 w-5" aria-hidden="true" />
                                </button>
                              )}
                              {idx === Choices.get('en').length - 1 && (
                                <button
                                  key={`${ID}addChoice${idx}`}
                                  type="button"
                                  className="inline-flex items-center border border-transparent rounded-full font-medium text-green-600 hover:text-green-800"
                                  onClick={handleAddChoice}>
                                  <PlusCircleIcon className="h-5 w-5" aria-hidden="true" />
                                </button>
                              )}
                            </div>
                          </div>
                        ))}
                      </div>
                    )}
  
  }*/
  //{language === 'en' && ( <RenderQuestionModal lg="Choice" nameC= {Choices}/>)}
  //{language === 'fr' && ( <RenderQuestionModal lg="Choix" nameC={ChoicesFr}/>)}
  //{language === 'de' && (<RenderQuestionModal lg="Choix" nameC= {ChoicesDe}/>)}
  return (
    <Transition.Root show={open} as={Fragment}>
      <Dialog
        as="div"
        className="fixed z-10 inset-0 px-4 sm:px-0 overflow-y-auto"
        initialFocus={cancelButtonRef}
        onClose={setOpen}>
        <div className="block items-end justify-center min-h-screen text-center">
          <Transition.Child
            as={Fragment}
            enter="ease-out duration-300"
            enterFrom="opacity-0"
            enterTo="opacity-100"
            leave="ease-in duration-200"
            leaveFrom="opacity-100"
            leaveTo="opacity-0">
            <Dialog.Overlay className="fixed inset-0 bg-gray-500 bg-opacity-75 transition-opacity" />
          </Transition.Child>

          {/* This element is to trick the browser into centering the modal contents. */}
          <span className="inline-block align-middle h-screen" aria-hidden="true">
            &#8203;
          </span>
          <Transition.Child
            as={Fragment}
            enter="ease-out duration-300"
            enterFrom="opacity-0 translate-y-4 sm:translate-y-0 sm:scale-95"
            enterTo="opacity-100 translate-y-0 sm:scale-100"
            leave="ease-in duration-200"
            leaveFrom="opacity-100 translate-y-0 sm:scale-100"
            leaveTo="opacity-0 translate-y-4 sm:translate-y-0 sm:scale-95">
            <div className="inline-block bg-white rounded-lg text-left overflow-hidden shadow-xl transform transition-all my-8 align-middle max-w-lg w-full md:min-h-[25rem] md:max-w-2xl md:w-[60rem]">
              <div className="flex bg-gray-50 pt-3 pl-3 pb-2 mb-4">
                <div className="rounded-full bg-gray-100 mr-2 ml-1">
                  <DisplayTypeIcon Type={Type} />
                </div>
                <Dialog.Title
                  as="h3"
                  className="text-lg pt-1.5 leading-6 font-medium text-gray-500">
                  {t(`addQuestion${Type}`)}
                </Dialog.Title>
              </div>
              <div className="pb-6 pr-6 pl-6">
                <div className="flex flex-col sm:flex-row sm:min-h-[18rem] ">
                  <div className="flex flex-col w-[55%]">
                    <div className="py-6 px-5 space-y-6">
                      <form className="flex gap-y-4 gap-x-8">
                        {availableLanguages.map((lang, index) => (
                          <label id={'lang' + lang}>
                            <input
                              className="hidden peer"
                              type="radio"
                              key={index}
                              id={'lang' + lang}
                              name="lang"></input>
                            <div
                              className="peer-checked:bg-gray-300 text-base font-small text-gray-900 hover:text-gray-700"
                              onClick={() => setLanguage(lang)}>
                              {t(lang)}
                            </div>
                          </label>
                        ))}
                      </form>
                    </div>
                    <div className="pb-4">{t('mainProperties')} </div>
                    <div>
                      <label className="block text-md mt font-medium text-gray-500">Title</label>
                      {language === 'en' && (
                        <input
                          value={Title}
                          onChange={handleChange()}
                          name="Title"
                          type="text"
                          placeholder={'Enter your Title'}
                          className="my-1 px-1 w-60 ml-1 border rounded-md"
                        />
                      )}
                      {language === 'fr' && (
                        <input
                          value={TitleFr}
                          onChange={handleChange()}
                          name="TitleFr"
                          type="text"
                          placeholder={'Entrer votre Titre'}
                          className="my-1 px-1 w-60 ml-1 border rounded-md"
                        />
                      )}
                      {language === 'de' && (
                        <input
                          value={TitleDe}
                          onChange={handleChange()}
                          name="TitleDe"
                          type="text"
                          placeholder={'enterTitle'}
                          className="my-1 px-1 w-60 ml-1 border rounded-md"
                        />
                      )}
                    </div>
                    <div className="text-red-600">
                      {errors
                        .filter((err) => err.startsWith('Title'))
                        .map((v, i) => (
                          <div key={i}>{v}</div>
                        ))}
                    </div>
                    <div>
                      <label className="block text-md mt font-medium text-gray-500">Hint</label>
                      {language === 'en' && (
                        <input
                          value={Hint}
                          onChange={handleChange()}
                          name="Hint"
                          type="text"
                          placeholder={t('enterHint')}
                          className="my-1 px-1 w-60 ml-1 border rounded-md"
                        />
                      )}
                      {language === 'fr' && (
                        <input
                          value={HintFr}
                          onChange={handleChange()}
                          name="HintFr"
                          type="text"
                          placeholder={t('enterHint')}
                          className="my-1 px-1 w-60 ml-1 border rounded-md"
                        />
                      )}
                      {language === 'de' && (
                        <input
                          value={HintDe}
                          onChange={handleChange()}
                          name="HintDe"
                          type="text"
                          placeholder={t('enterHint')}
                          className="my-1 px-1 w-60 ml-1 border rounded-md"
                        />
                      )}
                    </div>
                    <label className="flex pt-2 text-md font-medium text-gray-500">
                      {Type !== TEXT ? t('choices') : t('answers')}
                    </label>

                    {language === 'en' && (
                      <div className="pb-2">
                        {ChoicesMap.get('en').map((choice: string, idx: number) => (
                          <div className="flex w-60" key={`${ID}wrapper${idx}`}>
                            <input
                              key={`${ID}choice${idx}`}
                              value={choice}
                              onChange={updateChoice(idx, language)}
                              name="Choice"
                              type="text"
                              placeholder={
                                Type !== TEXT ? `${'Choice'}` + ` ${idx + 1}` : `Answer ${idx + 1}`
                              }
                              className="my-1 px-1 w-60 ml-2 border rounded-md"
                            />
                            <div className="flex ml-1 mt-1.2">
                              {ChoicesMap.get('en').length > 1 && (
                                <button
                                  key={`${ID}deleteChoice${idx}`}
                                  type="button"
                                  className="inline-flex items-center border border-transparent rounded-full font-medium text-gray-300 hover:text-gray-400"
                                  onClick={handleDeleteChoice(idx)}>
                                  <MinusCircleIcon className="h-5 w-5" aria-hidden="true" />
                                </button>
                              )}
                              {idx === ChoicesMap.get('en').length - 1 && (
                                <button
                                  key={`${ID}addChoice${idx}`}
                                  type="button"
                                  className="inline-flex items-center border border-transparent rounded-full font-medium text-green-600 hover:text-green-800"
                                  onClick={handleAddChoice}>
                                  <PlusCircleIcon className="h-5 w-5" aria-hidden="true" />
                                </button>
                              )}
                            </div>
                          </div>
                        ))}
                      </div>
                    )}
                    {language === 'fr' && (
                      <div className="pb-2">
                        {ChoicesMap.get('fr').map((choice: string, idx: number) => (
                          <div className="flex w-60" key={`${ID}wrapper${idx}`}>
                            <input
                              key={`${ID}choice${idx}`}
                              value={choice}
                              onChange={updateChoice(idx, language)}
                              name="Choice"
                              type="text"
                              placeholder={
                                Type !== TEXT ? `${'Choix'}` + ` ${idx + 1}` : `Answer ${idx + 1}`
                              }
                              className="my-1 px-1 w-60 ml-2 border rounded-md"
                            />
                            <div className="flex ml-1 mt-1.2">
                              {ChoicesMap.get('fr').length > 1 && (
                                <button
                                  key={`${ID}deleteChoice${idx}`}
                                  type="button"
                                  className="inline-flex items-center border border-transparent rounded-full font-medium text-gray-300 hover:text-gray-400"
                                  onClick={handleDeleteChoice(idx)}>
                                  <MinusCircleIcon className="h-5 w-5" aria-hidden="true" />
                                </button>
                              )}
                              {idx === ChoicesMap.get('fr').length - 1 && (
                                <button
                                  key={`${ID}addChoice${idx}`}
                                  type="button"
                                  className="inline-flex items-center border border-transparent rounded-full font-medium text-green-600 hover:text-green-800"
                                  onClick={handleAddChoice}>
                                  <PlusCircleIcon className="h-5 w-5" aria-hidden="true" />
                                </button>
                              )}
                            </div>
                          </div>
                        ))}
                      </div>
                    )}
                    {language === 'de' && (
                      <div className="pb-2">
                        {ChoicesMap.get('de').map((choice: string, idx: number) => (
                          <div className="flex w-60" key={`${ID}wrapper${idx}`}>
                            <input
                              key={`${ID}choice${idx}`}
                              value={choice}
                              onChange={updateChoice(idx, language)}
                              name="Choice"
                              type="text"
                              placeholder={
                                Type !== TEXT ? `${'Choix'}` + ` ${idx + 1}` : `Answer ${idx + 1}`
                              }
                              className="my-1 px-1 w-60 ml-2 border rounded-md"
                            />
                            <div className="flex ml-1 mt-1.2">
                              {ChoicesMap.get('de').length > 1 && (
                                <button
                                  key={`${ID}deleteChoice${idx}`}
                                  type="button"
                                  className="inline-flex items-center border border-transparent rounded-full font-medium text-gray-300 hover:text-gray-400"
                                  onClick={handleDeleteChoice(idx)}>
                                  <MinusCircleIcon className="h-5 w-5" aria-hidden="true" />
                                </button>
                              )}
                              {idx === ChoicesMap.get('de').length - 1 && (
                                <button
                                  key={`${ID}addChoice${idx}`}
                                  type="button"
                                  className="inline-flex items-center border border-transparent rounded-full font-medium text-green-600 hover:text-green-800"
                                  onClick={handleAddChoice}>
                                  <PlusCircleIcon className="h-5 w-5" aria-hidden="true" />
                                </button>
                              )}
                            </div>
                          </div>
                        ))}
                      </div>
                    )}
                    <div className="text-red-600">
                      {errors
                        .filter((err) => err.startsWith('Choices'))
                        .map((v, i) => (
                          <div key={i}>{v}</div>
                        ))}
                    </div>
                  </div>
                  <div className="w-[45%]">
                    {Type !== RANK && (
                      <>
                        <div className="pb-4">{t('additionalProperties')} </div>
                        <label className="block text-md font-medium text-gray-500">
                          {t('maxChoices')}
                        </label>
                        <input
                          value={MaxN}
                          onChange={handleChange()}
                          name="MaxN"
                          min="1"
                          type="number"
                          placeholder={t('enterMaxN')}
                          className="my-1 px-1 w-32 ml-1 border rounded-md"
                        />
                        <div className="text-red-600">
                          {errors
                            .filter((err) => err.startsWith('Max'))
                            .map((v, i) => (
                              <div key={i}>{v}</div>
                            ))}
                        </div>
                        <label className="block text-md font-medium text-gray-500">
                          {t('minChoices')}
                        </label>
                        <input
                          value={MinN}
                          onChange={handleChange()}
                          name="MinN"
                          max={MaxN < MAX_MINN ? MaxN : MAX_MINN}
                          min="0"
                          type="number"
                          placeholder={t('enterMinN')}
                          className="my-1 px-1 w-32 ml-1 border rounded-md"
                        />
                        <div className="text-red-600">
                          {errors
                            .filter((err) => err.startsWith('Min'))
                            .map((v, i) => (
                              <div key={i}>{v}</div>
                            ))}
                        </div>
                      </>
                    )}
                    {displayExtraFields()}
                  </div>
                </div>
                <div className="mt-5 sm:mt-6 sm:grid sm:grid-cols-2 sm:gap-3 sm:grid-flow-row-dense">
                  <button
                    type="button"
                    className="w-full inline-flex justify-center rounded-md border border-transparent shadow-sm px-4 py-2 bg-indigo-600 text-base font-medium text-white hover:bg-indigo-700 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-indigo-500 sm:col-start-2 sm:text-sm"
                    onClick={handleSave}>
                    <CheckIcon className="-ml-1 mr-2 h-5 w-5" aria-hidden="true" />
                    {t('saveQuestion')}
                  </button>
                  <button
                    type="button"
                    className="mt-3 w-full inline-flex justify-center rounded-md border border-gray-300 shadow-sm px-4 py-2 bg-white text-base font-medium text-gray-700 hover:bg-gray-50 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-indigo-500 sm:mt-0 sm:col-start-1 sm:text-sm"
                    onClick={handleClose}
                    ref={cancelButtonRef}>
                    {t('cancel')}
                  </button>
                </div>
              </div>
            </div>
          </Transition.Child>
        </div>
      </Dialog>
    </Transition.Root>
  );
};

AddQuestionModal.propTypes = {
  question: PropTypes.any.isRequired,
  open: PropTypes.bool.isRequired,
  setOpen: PropTypes.func.isRequired,
  notifyParent: PropTypes.func.isRequired,
  handleClose: PropTypes.func.isRequired,
};

export default AddQuestionModal;
