import React, { FC } from 'react';
import { useTranslation } from 'react-i18next';
import dVotingSystem from '../assets/d-votingSystem.png';

const About: FC = () => {
  const { t } = useTranslation();
  return (
    <div className="py-14 overflow-hidden">
      <div className="max-w-7xl mx-auto px-4 space-y-8 sm:px-6 lg:px-8">
        <div className="text-base max-w-prose mx-auto lg:max-w-none">
          <h2 className="text-base text-indigo-600 font-semibold tracking-wide uppercase">
            {t('aboutPlatform')}
          </h2>
          <p className="mt-2 text-3xl leading-8 font-extrabold tracking-tight text-gray-900 sm:text-4xl">
            {t('whatMakesUsDifferent')}
          </p>
        </div>
        <div className="relative z-10 text-base max-w-prose mx-auto lg:max-w-5xl lg:mx-0 lg:pr-72">
          <p className="text-lg text-gray-500">{t('about0')}</p>
        </div>
        <div className="lg:grid lg:grid-cols-2 lg:gap-8 lg:items-start">
          <div className="relative z-10">
            <div className="prose prose-indigo text-gray-500 mx-auto lg:max-w-none">
              <p>{t('about1')}</p>
              <p>{t('about2')}</p>
              <p>{t('about3')}</p>
              <p>{t('about4')}</p>
            </div>
            <div className="mt-10 flex text-base max-w-prose mx-auto lg:max-w-none">
              <div className="rounded-md shadow ml-4">
                <a
                  href="https://dedis.github.io/d-voting/#/"
                  className="w-full flex items-center justify-center px-5 py-3 border border-transparent text-base font-medium rounded-md text-indigo-600 bg-white hover:bg-gray-50">
                  {t('learnMore')}
                </a>
              </div>
            </div>
          </div>
          <div className="mt-12 relative text-base max-w-prose mx-auto lg:mt-0 lg:max-w-none">
            <svg
              className="absolute top-0 right-0 -mt-20 -mr-20 lg:top-auto lg:right-auto lg:bottom-1/2 lg:left-1/2 lg:mt-0 lg:mr-0 xl:top-0 xl:right-0 xl:-mt-20 xl:-mr-20"
              width={404}
              height={384}
              fill="none"
              viewBox="0 0 404 384"
              aria-hidden="true">
              <defs>
                <pattern
                  id="bedc54bc-7371-44a2-a2bc-dc68d819ae60"
                  x={0}
                  y={0}
                  width={20}
                  height={20}
                  patternUnits="userSpaceOnUse">
                  <rect
                    x={0}
                    y={0}
                    width={4}
                    height={4}
                    className="text-gray-200"
                    fill="currentColor"
                  />
                </pattern>
              </defs>
              <rect width={404} height={384} fill="url(#bedc54bc-7371-44a2-a2bc-dc68d819ae60)" />
            </svg>
            <div className="relative bg-white rounded-lg shadow-lg">
              <img src={dVotingSystem} alt="" />
            </div>
          </div>
        </div>
      </div>
    </div>
  );
};

export default About;
