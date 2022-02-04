import { useState } from 'react';

/*Custom hook that models a token that the user gets when signin */
const useToken = () => {
    const getToken = () => {
        let tok = sessionStorage.getItem('token');
        if(tok){
            return tok;
        }
        return null;
    }
    const [token,setToken] = useState(getToken());
    
    const saveToken = (token) => {
        sessionStorage.setItem('token', token);
        setToken(token);
    }

    return {token, saveToken};
}

export default useToken;