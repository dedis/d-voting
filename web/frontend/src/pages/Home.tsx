import { FlashContext, FlashLevel } from 'index';
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
      <div className="flex">
        <button
          className="flex inline-flex my-2 ml-2 items-center px-4 py-2 border border-transparent rounded-md shadow-sm text-sm font-medium text-white bg-indigo-500 hover:bg-indigo-600"
          onClick={() => {
            fctx.addMessage('Hello world!', FlashLevel.Info);
          }}>
          Add flash info
        </button>
        <button
          className="flex inline-flex my-2 ml-2 items-center px-4 py-2 border border-transparent rounded-md shadow-sm text-sm font-medium text-white bg-indigo-500 hover:bg-indigo-600"
          onClick={() => {
            fctx.addMessage('Hello world!', FlashLevel.Warning);
          }}>
          Add flash warning
        </button>
        <button
          className="flex inline-flex my-2 ml-2 items-center px-4 py-2 border border-transparent rounded-md shadow-sm text-sm font-medium text-white bg-indigo-500 hover:bg-indigo-600"
          onClick={() => {
            fctx.addMessage('Hello world!', FlashLevel.Error);
          }}>
          Add flash error
        </button>
      </div>
    </div>
  );
};

export default Home;
