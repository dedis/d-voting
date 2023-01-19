import React, { FC } from 'react';
import { useTranslation } from 'react-i18next';
type LanguageButtonsProps = {
  availableLanguages: string[];
  setLanguage: (lang: string) => void;
};

const LanguageButtons: FC<LanguageButtonsProps> = ({ availableLanguages, setLanguage }) => {
  const { t } = useTranslation();
  return (
    <div className="py-6 px-5 space-y-6">
      <form className="flex gap-y-4 gap-x-8">
        {availableLanguages.map((lang, index) => (
          <label key={index} id={'lang' + lang}>
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
  );
};

export default LanguageButtons;
