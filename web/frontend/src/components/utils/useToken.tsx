import { useState } from "react";

/*Custom hook that models a token that the user gets when signin
  TODO: this looks like a good candidate for application state management,
        as we're mixing state management and storage (sessionStorage).
        We could use an application state management framework instead.
*/
const useToken = () => {
  const getToken = () => {
    let tok = sessionStorage.getItem("token");
    if (tok) {
      return tok;
    }
    return null;
  };

  const [token, setToken] = useState(getToken());

  const saveToken = (tok) => {
    sessionStorage.setItem("token", tok);
    setToken(tok);
  };

  return { token, saveToken };
};

export default useToken;
