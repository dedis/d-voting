import React, { FC, useState } from 'react';
import { useTranslation } from 'react-i18next';
import PropTypes from 'prop-types';

import { ENDPOINT_GET_TEQ_KEY } from 'components/utils/Endpoints';
import './Login.css';
import { useLocation } from 'react-router-dom';

const Login: FC = () => {
  const { t } = useTranslation();

  const [loginError] = useState();

  const [content, setContent] = useState('');

  // The backend will provide the client the URL to make a Tequila
  // authentication. We therefore redirect to this address.
  const handleClick = async () => {
    fetch(ENDPOINT_GET_TEQ_KEY)
      .then((resp) => {
        const jsonData = resp.json();
        jsonData.then((result) => {
          window.location = result.url;
        });
      })
      .catch((error) => {
        console.log(error);
      });

    return <div>{loginError === null ? <div></div> : t('errorServerDown')}</div>;
  };

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

  return (
    <div className="login-container">
      <div className="login-txt">{content}</div>
      <button id="login-button" className="login-btn" onClick={handleClick}>
        {t('login')}
      </button>
    </div>
  );
};

Login.propTypes = {
  setToken: PropTypes.func,
};

export default Login;
