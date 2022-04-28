import ShortUniqueId from 'short-unique-id';
import { Role } from 'types/userRole';

const uid = new ShortUniqueId({ length: 8 });

const mockUser1 = {
  id: uid(),
  sciper: '123456',
  role: Role.Admin,
};

const mockUser2 = {
  id: uid(),
  sciper: '234567',
  role: Role.Operator,
};

const mockUser3 = {
  id: uid(),
  sciper: '345678',
  role: Role.Voter,
};

const setupMockUserDB = () => {
  const userDB = [];
  userDB.push(mockUser1);
  userDB.push(mockUser2);
  userDB.push(mockUser3);

  return userDB;
};

export default setupMockUserDB;
