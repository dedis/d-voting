import ShortUniqueId from 'short-unique-id';
import { User, UserRole } from 'types/userRole';

const uid = new ShortUniqueId({ length: 8 });

const mockUser1: User = {
  id: uid(),
  sciper: '123456',
  role: UserRole.Admin,
};

const mockUser2: User = {
  id: uid(),
  sciper: '234567',
  role: UserRole.Operator,
};

const mockUser3: User = {
  id: uid(),
  sciper: '345678',
  role: UserRole.Voter,
};

const user: User = {
  id: uid(),
  sciper: '561934',
  role: UserRole.Admin,
};

const setupMockUserDB = (): User[] => {
  const userDB: User[] = [];
  userDB.push(mockUser1);
  userDB.push(mockUser2);
  userDB.push(mockUser3);
  userDB.push(user);
  return userDB;
};

export default setupMockUserDB;
