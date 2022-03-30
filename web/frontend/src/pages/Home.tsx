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
      <button
        className="bg-blue-500 hover:bg-blue-700 text-white font-bold py-2 px-4 rounded-full"
        onClick={() => fctx.addMessage(`Hello from Home !`, 1)}>
        Add flash
      </button>
    </div>
  );
};

export default Home;
