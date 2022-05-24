interface User {
  id: string;
  sciper: string;
  role: UserRole;
}

export const enum UserRole {
  Admin = 'admin',
  Operator = 'operator',
  Voter = 'voter',
}

export type { User };
