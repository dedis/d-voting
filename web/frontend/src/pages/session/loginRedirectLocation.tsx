let value: string = '/';

export const getRedirectToLogin = () => {
  return value;
};
export const setRedirectToLogin = (v: string) => {
  console.log('setRedirectToLogin', v);
  value = v;
};
