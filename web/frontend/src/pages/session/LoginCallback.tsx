import React, { FC, useEffect } from 'react';
import { useTranslation } from 'react-i18next';

import { useSearchParams } from 'react-router-dom';

const LoginCallback: FC = () => {
  const { t } = useTranslation();

  const [searchParams, setSearchParams] = useSearchParams();

  useEffect(() => {
    console.log('searchParams:', searchParams);
    const key = searchParams.get('key');
    if (key !== undefined && key !== '') {
      console.log('key:', key);
      const req = {
        method: 'GET',
      };
      fetch(`/api/control_key?key=${key}`)
        .then((resp) => {
          console.log('resp:', resp);
          window.location.href = '/';
        })
        .catch((e) => {
          console.error('error:', e);
        });
    }
  }, [searchParams]);

  return (
    <div>
      <p>{t('loginCallback')}</p>
    </div>
  );
};

export default LoginCallback;
