import { Hint, Title } from 'types/configuration';

export function internationalize(language: string, internationalizable: Hint | Title): string {
  switch (language) {
    case 'fr':
      return internationalizable.Fr;
    case 'de':
      return internationalizable.De;
    default:
      return internationalizable.En;
  }
}
