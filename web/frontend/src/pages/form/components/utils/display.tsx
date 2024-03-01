import { default as i18n } from 'i18next';

export const prettifyChoice = (choicesMap, index) => {
  const choice = (
    choicesMap.ChoicesMap.has(i18n.language)
      ? choicesMap.ChoicesMap.get(i18n.language)
      : choicesMap.ChoicesMap.get('en')
  )[index];
  const url = choicesMap.URLs[index];
  return url ? (
    <a href={url} style={{ color: 'blue', textDecoration: 'underline' }}>
      {choice}
    </a>
  ) : (
    choice
  );
};
