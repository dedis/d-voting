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
export const editShuffle = (ElectionID: string) => `/api/evoting/services/shuffle/${ElectionID}`;
// setup and decrypt
export const editDKGActors = (ElectionID: string) =>
  `/api/evoting/services/dkg/actors/${ElectionID}`;
// initialize the nodes
export const dkgActors = `/api/evoting/services/dkg/actors`;

export const getDKGActors = (Proxy: string, ElectionID: string) =>
  encodeURI(`${Proxy}/evoting/services/dkg/actors/${ElectionID}`);

export const newProxyAddress = '/api/proxies/';
export const editProxyAddress = (NodeAddr: string) => encodeURI(`/api/proxies/${NodeAddr}`);
export const getProxyAddress = (NodeAddr: string) => encodeURI(`/api/proxies/${NodeAddr}`);
export const getProxiesAddresses = '/api/proxies';

// public information can be directly fetched from dela nodes
export const election = (ElectionID: string) =>
  `${process.env.REACT_APP_PROXY}/evoting/elections/${ElectionID}`;
export const elections = `${process.env.REACT_APP_PROXY}/evoting/elections`;

// To remove
export const ENDPOINT_EVOTING_RESULT = '/api/evoting/result';
