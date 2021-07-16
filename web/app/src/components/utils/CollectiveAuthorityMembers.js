/*addresses and public keys of each node*/

const address1 = 'RjEyNy4wLjAuMToyMDAx'; //address of a collective authority member
const PK1 = 'wtxsvHDkV9ahgEf1Nn6jgURmKyCVYWwc58xDSme4kxA='; //key of a collective authority memeber
const address2 = 'RjEyNy4wLjAuMToyMDAy';
const PK2 = 'ww6wDduiJhH+xiCHpqIC+0D0vgrBLVRHBQq0Zjt9hWI=';
const address3 = 'RjEyNy4wLjAuMToyMDAz';
const PK3 = 'aS9Na4kQGah07l2cU9fCGzv8RImJDH+kPxO9Ge00BqY=';
export const SHUFFLE_THRESHOLD = 3; //always needs to be at least 2/3 of the number of nodes
export const COLLECTIVE_AUTHORITY_MEMBERS = [{'Address' : address1,'PublicKey':PK1}, {'Address' : address2,'PublicKey':PK2}, {'Address' : address3,'PublicKey':PK3}];
