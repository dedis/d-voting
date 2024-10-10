import { use } from 'i18next';
import { initReactI18next } from 'react-i18next';
import LanguageDetector from 'i18next-browser-languagedetector';

import de from './de.json';
import en from './en.json';
import fr from './fr.json';

const resources = {
  de,
  en,
  fr,
};

export const availableLanguages = Object.keys(resources);

use(initReactI18next)
  .use(LanguageDetector)
  .init({
    resources,
    fallbackLng: ['en', 'fr', 'de'],
    debug: true,
    interpolation: {
      escapeValue: false, // react already safe from xss => https://www.i18next.com/translation-function/interpolation#unescape
    },
  });
