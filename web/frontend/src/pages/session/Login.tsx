import React, { FC, useEffect, useState } from 'react';
import { useTranslation } from 'react-i18next';
import PropTypes from 'prop-types';

import { useLocation } from 'react-router-dom';
import { ENDPOINT_GET_TEQ_KEY } from 'components/utils/Endpoints';

const Login: FC = () => {
  const { t } = useTranslation();
  const [loginError, setLoginError] = useState(null);
  const [content, setContent] = useState('');
  const location = useLocation();

  type LocationState = {
    from: Location;
  };

  useEffect(() => {
    const state = location.state as LocationState;
    if (state !== null) {
      setContent(t('loginText', { from: state.from.pathname }));
    }
  }, [location]);

  const handleLogin = async () => {
    fetch(ENDPOINT_GET_TEQ_KEY)
      .then((resp) => {
        const jsonData = resp.json();
        jsonData.then((result) => {
          window.location = result.url;
        });
      })
      .catch((error) => {
        setLoginError(error);
        console.log(error);
      });
  };

  return (
    <div>
      <div className="flex py-8">{content}</div>
      <button id="login-button" className="flex" onClick={handleLogin}>
        {t('login')}
      </button>
    </div>
  );
};

Login.propTypes = {
  setToken: PropTypes.func,
};

export default Login;
