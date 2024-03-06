import { default as i18n } from 'i18next';
import { urlizeLabel } from './../../../utils';

export const prettifyChoice = (choicesMap, index) => {
  return urlizeLabel(
    (choicesMap.ChoicesMap.has(i18n.language)
      ? choicesMap.ChoicesMap.get(i18n.language)
      : choicesMap.ChoicesMap.get('en'))[index],
    choicesMap.URLs[index]
  );
};
