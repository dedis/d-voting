import React, { FC, Fragment, useState } from 'react';
import { useTranslation } from 'react-i18next';
import { default as i18n } from 'i18next';

import { availableLanguages } from './Configuration';

import { Menu, Transition } from '@headlessui/react';
import { GlobeAltIcon } from '@heroicons/react/outline';

const classNames = (...classes: string[]) => {
  return classes.filter(Boolean).join(' ');
};

const LanguageSelector: FC = () => {
  const { t } = useTranslation();
  const [toggle, setToggle] = useState(false);
  const triggerToggle = () => {
    setToggle(!toggle);
  };

  return (
    <div className="relative inline-block text-left">
      <Menu as="div" className="ml-6">
        <div>
          <Menu.Button
            onClick={triggerToggle}
            className="flex text-sm mr-6 rounded-full text-gray-400 hover:text-white">
            <span className="sr-only">Language</span>
            <GlobeAltIcon className="h-7 w-7 text-neutral-600" aria-hidden="true" />
          </Menu.Button>
        </div>
        <Transition
          as={Fragment}
          enter="transition ease-out duration-100"
          enterFrom="transform opacity-0 scale-95"
          enterTo="transform opacity-100 scale-100"
          leave="transition ease-in duration-75"
          leaveFrom="transform opacity-100 scale-100"
          leaveTo="transform opacity-0 scale-95">
          <Menu.Items className="origin-top-right absolute right-0 mt-2 w-48 rounded-md shadow-lg py-1 bg-white ring-1 ring-black ring-opacity-5 focus:outline-none">
            {availableLanguages.map((lang) => (
              <Menu.Item key={lang}>
                <div
                  onClick={() => {
                    if (i18n.language !== lang) i18n.changeLanguage(lang);
                  }}
                  className={classNames(
                    i18n.language === lang ? 'bg-gray-100' : 'cursor-pointer',
                    ' block px-4 py-2 text-sm text-gray-700'
                  )}>
                  {t(lang)}
                </div>
              </Menu.Item>
            ))}
          </Menu.Items>
        </Transition>
      </Menu>
    </div>
  );
};

export default LanguageSelector;
