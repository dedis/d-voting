import { FC, useEffect, useState } from 'react';
import { useTranslation } from 'react-i18next';
import PropTypes from 'prop-types';

import { useLocation } from 'react-router-dom';
import ClientError from 'layout/ClientError';

const Login: FC = () => {
  const { t } = useTranslation();
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
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [location]);

  return <ClientError statusCode={401} description={content} />;
};

Login.propTypes = {
  setToken: PropTypes.func,
};

export default Login;
