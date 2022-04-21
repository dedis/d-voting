// information accessed through the middleware
export const ENDPOINT_GET_TEQ_KEY = '/api/get_teq_key';
export const ENDPOINT_PERSONAL_INFO = '/api/personal_info';
export const ENDPOINT_LOGOUT = '/api/logout';
export const ENDPOINT_USER_RIGHTS = '/api/user_rights';
export const ENDPOINT_ADD_ROLE = '/api/add_role';
export const ENDPOINT_REMOVE_ROLE = '/api/remove_role';

export const newElection = '/api/evoting/elections';
export const editElection = (ElectionID: string) => `/api/evoting/elections/${ElectionID}`;
export const newElectionVote = (ElectionID: string) => `/api/evoting/elections/${ElectionID}/vote`;
export const editShuffle = (ElectionID: string) => `/evoting/services/shuffle/${ElectionID}`;
// Decrypt
export const editDKGActors = (ElectionID: string) => `/evoting/services/dkg/actors/${ElectionID}`;

// public information can be directly fetched from dela nodes
export const election = (ElectionID: string) => `/evoting/elections/${ElectionID}`;
export const elections = '/evoting/elections';

// To remove
export const ENDPOINT_EVOTING_RESULT = '/api/evoting/result';
