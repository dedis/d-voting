import React, { FC } from 'react';
import { useTranslation } from 'react-i18next';

import Login from './Login';
import './Home.css';

type HomeProps = {
  isLogged: boolean;
};

const Home: FC<HomeProps> = ({ isLogged }) => {
  const { t } = useTranslation();

  return isLogged ? (
    <div className="home">
      <h1>{t('homeTitle')}</h1>
      <div className="home-txt">{t('homeText')}</div>
    </div>
  ) : (
    <div className="login-container">
      <Login />
    </div>
  );
};

export default Home;
