import React, { FC } from 'react';
import { Dialog } from '@headlessui/react';
import { useTranslation } from 'react-i18next';

type WarningModalProps = {
  message: string;
  isShown: boolean;
  setIsShown: (isShown: boolean) => void;
  action: () => void;
};

const WarningModal: FC<WarningModalProps> = ({ message, isShown, setIsShown, action }) => {
  const { t } = useTranslation();
  if (isShown) {
    return (
      <Dialog open={isShown} onClose={() => {}}>
        <Dialog.Overlay className="fixed inset-0 bg-black opacity-30" />
        <div className="fixed inset-0 flex items-center justify-center">
          <div className="bg-white content-center rounded-lg shadow-lg p-3 w-80">
            <Dialog.Title as="h3" className="text-lg font-medium leading-6 text-gray-900">
              Warning
            </Dialog.Title>
            <Dialog.Description className="mt-2 mx-auto text-sm text-center text-gray-500">
              {message}
            </Dialog.Description>
            <div className="mt-4 sm:mt-6 sm:grid sm:grid-cols-2 sm:gap-3 sm:grid-flow-row-dense">
              <button
                type="button"
                className="inline-flex justify-center px-4 py-2 text-sm font-medium text-white bg-[#ff0000] border border-transparent rounded-md hover:bg-gray-300 focus:outline-none focus-visible:ring-2 focus-visible:ring-offset-2 focus-visible:ring-[#ff0000]"
                onClick={() => setIsShown(false)}>
                {t('cancel')}
              </button>
              <button
                type="button"
                className="inline-flex justify-center px-4 py-2 text-sm font-medium text-white bg-red-600 border border-transparent rounded-md hover:bg-red-700 focus:outline-none focus-visible:ring-2 focus-visible:ring-offset-2 focus-visible:ring-red-500"
                onClick={() => {
                  setIsShown(false);
                  action();
                }}>
                {t('continue')}
              </button>
            </div>
          </div>
        </div>
      </Dialog>
    );
  } else {
    return <></>;
  }
};

export default WarningModal;
