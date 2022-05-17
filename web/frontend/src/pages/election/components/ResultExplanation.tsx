import { Popover, Transition } from '@headlessui/react';
import { Fragment } from 'react';
import { useTranslation } from 'react-i18next';
import { QuestionMarkCircleIcon } from '@heroicons/react/outline';

const ResultExplanation = () => {
  const { t } = useTranslation();

  return (
    <div className="top-16 w-full max-w-sm">
      <Popover className="relative">
        {({ open }) => (
          <>
            {
              <Popover.Button
                className={`
                ${open ? '' : 'text-opacity-90'}
                group inline-flex items-center rounded-md px-3 py-2 text-base font-medium text-white hover:text-opacity-100 focus:outline-none focus-visible:ring-2 focus-visible:ring-white focus-visible:ring-opacity-75`}>
                <QuestionMarkCircleIcon
                  className="-ml-1 mr-2 h-5 w-5 text-gray-500 hover:text-indigo-500"
                  aria-hidden="true"
                />
              </Popover.Button>
            }
            <Transition
              as={Fragment}
              enter="transition ease-out duration-200"
              enterFrom="opacity-0 translate-y-1"
              enterTo="opacity-100 translate-y-0"
              leave="transition ease-in duration-150"
              leaveFrom="opacity-100 translate-y-0"
              leaveTo="opacity-0 translate-y-1">
              <Popover.Panel className="absolute z-10 mt-3 ml-28 w-screen max-w-sm -translate-x-1/2 transform px-4 sm:px-0">
                <div className="overflow-hidden rounded-lg shadow-lg ring-1 ring-black ring-opacity-5">
                  <div className="relative bg-white p-7 text-sm text-justify">
                    {t('resultExplanation')}
                  </div>
                </div>
              </Popover.Panel>
            </Transition>
          </>
        )}
      </Popover>
    </div>
  );
};

export default ResultExplanation;
