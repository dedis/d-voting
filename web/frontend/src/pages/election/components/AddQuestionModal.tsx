import { FC, Fragment, useRef } from 'react';
import PropTypes from 'prop-types';
import { Dialog, Transition } from '@headlessui/react';

import { useTranslation } from 'react-i18next';
import { MinusSmIcon, PlusSmIcon, UserAddIcon } from '@heroicons/react/outline';
import { RankQuestion, SelectQuestion, TEXT, TextQuestion } from 'types/configuration';
import useQuestionForm from './utils/useQuestionForm';

type AddQuestionModalProps = {
  question: RankQuestion | SelectQuestion | TextQuestion;
  open: boolean;
  setOpen(opened: boolean): void;
  notifyParent(question: RankQuestion | SelectQuestion | TextQuestion): void;
  // removeQuestion: () => void;
};

const MAX_MINN = 20;

const AddQuestionModal: FC<AddQuestionModalProps> = ({ question, open, setOpen, notifyParent }) => {
  const { ID, Type } = question;
  const { t } = useTranslation();
  const {
    state: values,
    handleChange,
    addChoice,
    deleteChoice,
    updateChoice,
  } = useQuestionForm(question);

  const { Title, MaxN, MinN, Choices } = values;

  const handleClose = () => setOpen(false);

  const handleSave = async () => {
    notifyParent(values);
    setOpen(false);
  };
  const cancelButtonRef = useRef(null);

  const DisplayExtraFields = () => {
    switch (Type) {
      case TEXT:
        const tq = question as TextQuestion;
        return (
          <>
            <label className="block text-md font-medium text-gray-500">MaxLength</label>
            <input
              value={tq.MaxLength}
              onChange={handleChange}
              name="MaxLength"
              min="0"
              type="number"
              placeholder="Enter the MaxLength"
              className="my-1 w-60 ml-1 border rounded-md"
            />
            <label className="block text-md font-medium text-gray-500">Regex</label>
            <input
              value={tq.Regex}
              onChange={handleChange}
              name="Regex"
              type="text"
              placeholder="Enter your Regex"
              className="my-1 w-60 ml-1 border rounded-md"
            />
          </>
        );
      default:
        return null;
    }
  };

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
            <div className="inline-block bg-white rounded-lg text-left overflow-hidden shadow-xl transform transition-all my-8 align-middle max-w-lg w-full p-6">
              <div>
                <div className="text-center">
                  <Dialog.Title as="h3" className="text-lg leading-6 font-medium text-gray-900">
                    {t(`addQuestion${Type}`)}
                  </Dialog.Title>
                  <div className="flex flex-col pt-4">
                    <div>
                      <label className="block text-md mt font-medium text-gray-500">Title</label>
                      <input
                        value={Title}
                        onChange={handleChange}
                        name="Title"
                        type="text"
                        placeholder="Enter your Title"
                        className="my-1 w-60 ml-1 border rounded-md"
                      />
                      <label className="block text-md font-medium text-gray-500">MaxN</label>
                      <input
                        value={MaxN}
                        onChange={handleChange}
                        name="MaxN"
                        min={MinN}
                        type="number"
                        placeholder="Enter the MaxN"
                        className="my-1 w-60 ml-1 border rounded-md"
                      />
                      <label className="block text-md font-medium text-gray-500">MinN</label>

                      <input
                        value={MinN}
                        onChange={handleChange}
                        name="MinN"
                        max={MaxN < MAX_MINN ? MaxN : MAX_MINN}
                        min="0"
                        type="number"
                        placeholder="Enter the MinN"
                        className="my-1 w-60 ml-1 border rounded-md"
                      />
                    </div>
                    <DisplayExtraFields />
                    <label className="block text-md font-medium text-gray-500">Choices</label>
                    {Choices.map((choice: string, idx: number) => (
                      <div key={`${ID}wrapper${idx}`}>
                        <input
                          key={`${ID}choice${idx}`}
                          value={choice}
                          onChange={updateChoice(idx)}
                          name="Choice"
                          type="text"
                          placeholder="Enter your choice"
                          className="my-1 w-60 ml-1 border rounded-md"
                        />
                        <button
                          key={`${ID}deleteChoice${idx}`}
                          type="button"
                          className="inline-flex ml-2 items-center border border-transparent rounded-md shadow-sm text-sm font-medium text-white bg-indigo-600 hover:bg-indigo-700"
                          onClick={deleteChoice(idx)}>
                          <MinusSmIcon className="h-4 w-4" aria-hidden="true" />
                        </button>
                      </div>
                    ))}
                    <button
                      type="button"
                      className="flex p-2 h-8 w-8 mb-2 rounded-md bg-green-600 hover:bg-green-800 sm:-mr-2"
                      onClick={addChoice}>
                      <PlusSmIcon className="h-5 w-5 text-white" aria-hidden="true" />
                    </button>
                  </div>
                </div>
              </div>
              <div className="mt-5 sm:mt-6 sm:grid sm:grid-cols-2 sm:gap-3 sm:grid-flow-row-dense">
                <button
                  type="button"
                  className="w-full inline-flex justify-center rounded-md border border-transparent shadow-sm px-4 py-2 bg-indigo-600 text-base font-medium text-white hover:bg-indigo-700 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-indigo-500 sm:col-start-2 sm:text-sm"
                  onClick={handleSave}>
                  <UserAddIcon className="-ml-1 mr-2 h-5 w-5" aria-hidden="true" />
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
          </Transition.Child>
        </div>
      </Dialog>
    </Transition.Root>
  );
};

AddQuestionModal.propTypes = {
  open: PropTypes.bool.isRequired,
  setOpen: PropTypes.func.isRequired,
  notifyParent: PropTypes.func.isRequired,
};

export default AddQuestionModal;
