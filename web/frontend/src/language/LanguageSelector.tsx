import React, { useState } from 'react';
import { default as i18n } from 'i18next';
import { availableLanguages } from './Configuration';

const LanguageSelector = () => {
  const [toggle, setToggle] = useState(false);
  const triggerToggle = () => {
    setToggle(!toggle);
  };

  return (
    <div className="relative inline-block text-left">
      <div>
        <button
          onClick={triggerToggle}
          type="button"
          className="inline-flex justify-center w-full text-gray-300 px-2 py-0.5 text-sm font-medium focus:outline-none"
          id="menu-button"
          aria-expanded="true"
          aria-haspopup="true">
          {availableLanguages}
          <svg
            className="-mr-1 ml-2 h-5 w-5"
            xmlns="http://www.w3.org/2000/svg"
            viewBox="0 0 20 20"
            fill="currentColor"
            aria-hidden="true">
            <path
              fillRule="evenodd"
              d="M5.293 7.293a1 1 0 011.414 0L10 10.586l3.293-3.293a1 1 0 111.414 1.414l-4 4a1 1 0 01-1.414 0l-4-4a1 1 0 010-1.414z"
              clipRule="evenodd"
            />
          </svg>
        </button>
      </div>

      <div
        className={`${
          toggle
            ? 'ease-out duration-100 transform opacity-100 scale-100'
            : 'ease-in duration-75 transform opacity-0 scale-95'
        } transition origin-top-right absolute right-0 mt-2 w-56 rounded-md shadow-lg bg-white ring-1 ring-black ring-opacity-5 focus:outline-none`}
        role="menu"
        aria-orientation="vertical"
        aria-labelledby="menu-button"
        tabIndex={-1}>
        <div className="py-1" role="none">
          <select
            className="text-gray-700 block px-4 py-2 text-sm"
            defaultValue={i18n.language}
            onChange={(e) => i18n.changeLanguage(e.target.value)}>
            {availableLanguages.map((lang) => (
              <option key={lang}>{lang}</option>
            ))}
          </select>
        </div>
      </div>
    </div>
  );
};

export default LanguageSelector;
