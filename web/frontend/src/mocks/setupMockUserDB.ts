import ShortUniqueId from 'short-unique-id';
import { UserRole } from 'types/userRole';

const uid = new ShortUniqueId({ length: 8 });

const mockUser1 = {
  id: uid(),
  sciper: '123456',
  role: UserRole.Admin,
};

const mockUser2 = {
  id: uid(),
  sciper: '234567',
  role: UserRole.Operator,
};

const mockUser3 = {
  id: uid(),
  sciper: '345678',
  role: UserRole.Voter,
};

const setupMockUserDB = () => {
  const userDB = [];
  userDB.push(mockUser1);
  userDB.push(mockUser2);
  userDB.push(mockUser3);
  return userDB;
};

export default setupMockUserDB;
