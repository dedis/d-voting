/*addresses and public keys of each node*/

const address1 = 'RjEyNy4wLjAuMToyMDAx'; //address of a collective authority member
const PK1 = 'fbrhQNEluLFqzUOoxE7ZAvKF0C49aeltqtaiVc5iESc='; //key of a collective authority memeber
const address2 = 'RjEyNy4wLjAuMToyMDAy';
const PK2 = 'amYlvJ0+AN/ygE8xhHb8oH6Zax6IcwNhSz6gs/OY/Ow=';
const address3 = 'RjEyNy4wLjAuMToyMDAz';
const PK3 = 'dz9BnoRqT13yavdeAEy2xeO+I+u/COINMGP3N7Nlb9g=';
export const SHUFFLE_THRESHOLD = 3; //always needs to be at least 2/3 of the number of nodes
export const COLLECTIVE_AUTHORITY_MEMBERS = [
  { Address: address1, PublicKey: PK1 },
  { Address: address2, PublicKey: PK2 },
  { Address: address3, PublicKey: PK3 },
];
