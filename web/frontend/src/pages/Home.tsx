import { FC } from 'react';
import { useTranslation } from 'react-i18next';

const Home: FC = () => {
  const { t } = useTranslation();

  return (
    <div className="flex flex-col">
      <h1>{t('homeTitle')}</h1>
      <div>{t('homeText')}</div>
    </div>
  );
};

export default Home;
