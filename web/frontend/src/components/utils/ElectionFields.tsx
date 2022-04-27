import { LightElectionInfo } from 'types/electionInfo';
import { useFillLightElectionInfo } from './FillElectionInfo';

/**
 *
 * @param {*} electionData a json object of an election
 * @returns the fields of an election and a function to change the status field
 */
const ElectionFields = (electionData: LightElectionInfo) => {
  const { title, id, status, pubKey, setStatus } = useFillLightElectionInfo(electionData);
  return { title, id, status, pubKey, setStatus };
};

export default ElectionFields;
