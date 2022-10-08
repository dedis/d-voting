import { FC } from 'react';
import { useTranslation } from 'react-i18next';
import { ChevronRightIcon } from '@heroicons/react/solid';
import { ROUTE_FORM_INDEX } from 'Routes';
import { Link } from 'react-router-dom';
import introductionImage from '../assets/introductionImage.png';

const Home: FC = () => {
  const { t } = useTranslation();
  return (
    <div className="bg-white pb-8 sm:pb-6 lg:pb-6">
      <div className="pt-8 overflow-hidden sm:pt-12 lg:relative lg:py-44">
        <div className="mx-auto max-w-md px-4 sm:max-w-3xl sm:px-6 lg:px-8 lg:max-w-7xl lg:grid lg:grid-cols-2 lg:gap-24">
          <div>
            <div className="mt-20">
              <div>
                <a href="https://github.com/dedis/d-voting" className="inline-flex space-x-4">
                  <span className="rounded bg-indigo-50 px-2.5 py-1 text-xs font-semibold text-indigo-600 tracking-wide uppercase">
                    {t('homeWhatsNew')}
                  </span>
                  <span className="inline-flex items-center text-sm font-medium text-indigo-600 space-x-1">
                    <span>{t('homeJustShippedVersion')} 1.0.0</span>
                    <ChevronRightIcon className="h-5 w-5" aria-hidden="true" />
                  </span>
                </a>
              </div>
              <div className="mt-6 sm:max-w-xl">
                <h1 className="text-4xl font-extrabold text-gray-900 tracking-tight sm:text-5xl">
                  {t('homeTitle')}
                </h1>
                <p className="mt-6 text-xl text-gray-500">{t('homeText')}</p>
              </div>
              <div className="mt-12 sm:max-w-lg sm:w-full sm:flex">
                <div className="mt-4 sm:mt-0 sm:ml-3">
                  <Link to={ROUTE_FORM_INDEX}>
                    <button className="block w-full rounded-md border border-transparent px-5 py-3 bg-indigo-600 text-base font-medium text-white shadow hover:bg-indigo-700 focus:outline-none focus:ring-2 focus:ring-indigo-500 focus:ring-offset-2 sm:px-10">
                      Get Started
                    </button>
                  </Link>
                </div>
              </div>
            </div>
          </div>
        </div>

        <div className="sm:mx-auto sm:max-w-3xl sm:px-6">
          <div className="py-12 sm:relative sm:mt-12 sm:py-16 lg:absolute lg:inset-y-0 lg:right-0 lg:w-1/2">
            <div className="hidden sm:block">
              <svg
                className="absolute top-8 right-1/2 -mr-3 lg:m-0 lg:left-0"
                width={404}
                height={392}
                fill="none"
                viewBox="0 0 404 392">
                <defs>
                  <pattern
                    id="837c3e70-6c3a-44e6-8854-cc48c737b659"
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
                <rect width={404} height={392} fill="url(#837c3e70-6c3a-44e6-8854-cc48c737b659)" />
              </svg>
            </div>
            <div className="relative pl-4 -mr-40 sm:mx-auto sm:max-w-3xl sm:px-0 lg:max-w-none lg:h-full lg:pl-12">
              <img
                className="w-full rounded-md shadow-xl ring-1 ring-black ring-opacity-5 lg:h-full lg:w-auto lg:max-w-none"
                src={introductionImage}
                alt=""
              />
            </div>
          </div>
        </div>
      </div>
    </div>
  );
};

export default Home;
