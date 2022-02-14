import i18n from "i18next";
import {initReactI18next} from "react-i18next";
import LanguageDetector from 'i18next-browser-languagedetector';

import de from "./de.json"
import en from "./en.json";
import fr from "./fr.json";

const resources = {
  de,
  en,
  fr
}

export const availableLanguages = Object.keys(resources)

i18n.use(initReactI18next)
  .use(LanguageDetector)
  .init({
    resources,
    defaultNS: "common",
    fallbackLng: "en",
    debug: true,
    interpolation: {
        escapeValue: false // react already safes from xss => https://www.i18next.com/translation-function/interpolation#unescape
    }
  });

export { i18n };