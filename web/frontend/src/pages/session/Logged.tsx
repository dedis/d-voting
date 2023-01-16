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
    let decodedCookies = decodeURIComponent(document.cookie);
    const redirectCookie = decodedCookies.split('; ').find((row) => row.startsWith('redirect'));
    return redirectCookie ? redirectCookie.split('=')[1] : '/'; // default value is the home page
  };

  useEffect(() => {
    if (authCtx.isLogged) {
      fctx.addMessage(t('loggedIn'), FlashLevel.Info);
    } else {
      fctx.addMessage(t('notLoggedIn'), FlashLevel.Error);
    }
    const redir = getCookie();
    navigate(redir);
  });

  return <></>;
};

export default Logged;
