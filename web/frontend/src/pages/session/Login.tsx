import React, { FC, useContext, useEffect, useState } from 'react';
import { useTranslation } from 'react-i18next';
import PropTypes from 'prop-types';

import { useLocation } from 'react-router-dom';
import { FlashContext } from 'index';
import handleLogin from './HandleLogin';

const Login: FC = () => {
  const { t } = useTranslation();
  const [content, setContent] = useState('');
  const location = useLocation();

  const fctx = useContext(FlashContext);

  type LocationState = {
    from: Location;
  };

  useEffect(() => {
    const state = location.state as LocationState;
    if (state !== null) {
      setContent(t('loginText', { from: state.from.pathname }));
    }
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [location]);

  return (
    <div>
      <div className="flex py-8">{content}</div>
      <button id="login-button" className="flex" onClick={() => handleLogin(fctx)}>
        {t('login')}
      </button>
    </div>
  );
};

Login.propTypes = {
  setToken: PropTypes.func,
};

export default Login;
