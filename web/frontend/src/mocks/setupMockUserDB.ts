import ShortUniqueId from 'short-unique-id';
import { ROLE } from 'types/userRole';

const uid = new ShortUniqueId({ length: 8 });

const mockUser1 = {
  id: uid(),
  sciper: '123456',
  role: ROLE.Admin,
};

const mockUser2 = {
  id: uid(),
  sciper: '234567',
  role: ROLE.Operator,
};

const mockUser3 = {
  id: uid(),
  sciper: '345678',
  role: ROLE.Voter,
};

const user = {
  id: uid(),
  sciper: '561934',
  role: ROLE.Admin,
};

const setupMockUserDB = () => {
  const userDB = [];
  userDB.push(mockUser1);
  userDB.push(mockUser2);
  userDB.push(mockUser3);
  userDB.push(user);
  return userDB;
};

export default setupMockUserDB;
