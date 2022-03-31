import { AuthContext, FlashContext } from 'index';
import React, { FC, useContext, useEffect } from 'react';
import { useNavigate } from 'react-router-dom';

// This components handles the callback when successfully logged. It only
// redirects to the home page with a flash message.
const Logged: FC = () => {
  const navigate = useNavigate();

  const authCtx = useContext(AuthContext);
  const fctx = useContext(FlashContext);

  useEffect(() => {
    if (authCtx.isLogged) {
      fctx.addMessage('You are logged in.', 1);
    } else {
      fctx.addMessage('You are not logged in.', 3);
    }

    navigate('/');
  });

  return <></>;
};

export default Logged;
