import { FlashContext } from 'index';
import React, { FC, useContext } from 'react';
import { useTranslation } from 'react-i18next';

import './Home.css';

const Home: FC = () => {
  const { t } = useTranslation();
  const fctx = useContext(FlashContext);

  return (
    <div className="home">
      <h1>{t('homeTitle')}</h1>
      <div className="home-txt">{t('homeText')}</div>
    </div>
  );
};

export default Home;
