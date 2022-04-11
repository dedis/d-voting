import useFillElectionFields from './useFillElectionFields';

/**
 *
 * @param {*} electionData a json object of an election
 * @returns the fields of an election and a function to change the status field
 */
const ElectionFields = (electionData) => {
  const { electionTitle, id, status, pubKey, result, setStatus } =
    useFillElectionFields(electionData);
  return { electionTitle, id, status, pubKey, result, setStatus };
};

export default ElectionFields;
