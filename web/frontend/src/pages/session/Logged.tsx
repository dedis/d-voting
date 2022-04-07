import { AuthContext, FlashContext, FlashLevel } from 'index';
import React, { FC, useContext, useEffect } from 'react';
import { useTranslation } from 'react-i18next';
import { useNavigate } from 'react-router-dom';

// This components handles the callback when successfully logged. It only
// redirects to the home page with a flash message.
const Logged: FC = () => {
  const navigate = useNavigate();
  const { t } = useTranslation();

  const authCtx = useContext(AuthContext);
  const fctx = useContext(FlashContext);

  useEffect(() => {
    if (authCtx.isLogged) {
      fctx.addMessage(t('loggedIn'), FlashLevel.Info);
    } else {
      fctx.addMessage(t('notLoggedIn'), FlashLevel.Error);
    }

    navigate('/');
  });

  return <></>;
};

export default Logged;
