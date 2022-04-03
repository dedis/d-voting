import { FlashContext } from 'index';
import React, { FC, useContext, useEffect } from 'react';
import { useTranslation } from 'react-i18next';
import { useNavigate } from 'react-router-dom';

type BallotCastProps = {
  postError: string;
  showFlash: boolean;
};

// Warning: Can't perform a React state update on an unmounted component. This is a no-op, but it indicates a memory leak in your application. To fix, cancel all subscriptions and asynchronous tasks in a useEffect cleanup function.
// Ballot@https://localhost:3000/static/js/bundle.js:5481:69

// Component that redirects to the home page with a flash message
// when a ballot is cast.
const BallotCast: FC<BallotCastProps> = ({ postError, showFlash }) => {
  const navigate = useNavigate();
  const { t } = useTranslation();

  const fctx = useContext(FlashContext);

  useEffect(() => {
    if (showFlash) {
      if (postError !== null) {
        if (postError.includes('ECONNREFUSED')) {
          fctx.addMessage(t('errorServerDown'), 3);
        } else {
          fctx.addMessage(t('voteFailure'), 3);
        }
      } else {
        fctx.addMessage(t('voteSuccess'), 1);
      }

      window.scrollTo(0, 0);
      navigate('/');
    }
  });

  return <></>;
};

export default BallotCast;
