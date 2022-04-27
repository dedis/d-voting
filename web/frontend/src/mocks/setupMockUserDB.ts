import ShortUniqueId from 'short-unique-id';

const uid = new ShortUniqueId({ length: 8 });

const mockUser1 = {
  id: uid(),
  sciper: '123456',
  role: 'admin',
};

const mockUser2 = {
  id: uid(),
  sciper: '234567',
  role: 'operator',
};

const mockUser3 = {
  id: uid(),
  sciper: '345678',
  role: 'voter',
};

const setupMockUserDB = () => {
  const userDB = [];
  userDB.push(mockUser1);
  userDB.push(mockUser2);
  userDB.push(mockUser3);

  return userDB;
};

export default setupMockUserDB;
