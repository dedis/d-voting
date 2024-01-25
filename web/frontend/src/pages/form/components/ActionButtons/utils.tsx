import { ID } from './../../../../types/configuration';

export function isManager(formID: ID, authorization: Map<String, String[]>, isLogged: boolean) {
  return (
    isLogged && // must be logged in
    authorization.has('election') &&
    authorization.get('election').includes('create') && // must be able to create elections
    authorization.has(formID) &&
    authorization.get(formID).includes('own') // must own the election
  );
}
