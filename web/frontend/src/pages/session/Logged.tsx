import { AuthContext, FlashContext, FlashLevel } from 'index';
import { FC, useContext, useEffect } from 'react';
import { useTranslation } from 'react-i18next';
import { useNavigate } from 'react-router-dom';

// This components handles the callback when successfully logged. It only
// redirects to the home page with a flash message.
const Logged: FC = () => {
  const navigate = useNavigate();
  const { t } = useTranslation();

  const authCtx = useContext(AuthContext);
  const fctx = useContext(FlashContext);

  const getCookie = () => {
    let name = 'redirect' + '=';
    let decodedCookie = decodeURIComponent(document.cookie);
    let ca = decodedCookie.split(';');
    for (let i = 0; i < ca.length; i++) {
      let c = ca[i];
      while (c.charAt(0) == ' ') {
        c = c.substring(1);
      }
      if (c.indexOf(name) == 0) {
        return c.substring(name.length, c.length);
      }
    }
    return '';
  };

  useEffect(() => {
    if (authCtx.isLogged) {
      fctx.addMessage(t('loggedIn'), FlashLevel.Info);
    } else {
      fctx.addMessage(t('notLoggedIn'), FlashLevel.Error);
    }
    const redir = getCookie();
    console.log('Logged: redirecting to', redir);
    navigate(redir);
  });

  return <></>;
};

export default Logged;
