import en from './../src/language/en.json';
import fr from './../src/language/fr.json';
import de from './../src/language/de.json';

export function initI18n () {
  i18n.init({
    resources: { en, fr, de },
    fallbackLng: ['en', 'fr', 'de'],
  });
}
