import { useTranslation } from 'react-i18next';
import { Link } from 'react-router-dom';
import { ROUTE_ELECTION_INDEX } from 'Routes';

export default function ElectionClosed() {
  const { t } = useTranslation();

  return (
    <div className="h-[calc(100vh-130px)]">
      <div className="bg-white min-h-full font-sans px-4 py-16 sm:px-6 sm:py-24 md:grid md:place-items-center lg:px-8">
        <div className="max-w-max mx-auto">
          <main className="sm:flex">
            <div className="sm:ml-6">
              <div className=" sm:border-gray-200 sm:pl-6">
                <h1 className="text-4xl font-extrabold text-gray-900 tracking-tight sm:text-5xl">
                  {t('voteImpossible')}
                </h1>
                <p className="mt-1 text-base text-gray-500">{t('voteImpossibleDescription')}</p>
              </div>
              <div className="mt-10 flex space-x-3 sm:border-l sm:border-transparent sm:pl-6">
                <Link
                  to={ROUTE_ELECTION_INDEX}
                  className="inline-flex items-center px-4 py-2 border border-transparent text-sm font-medium rounded-md shadow-sm text-white bg-indigo-600 hover:bg-indigo-700 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-indigo-500">
                  {t('notFoundVoteImpossible')}
                </Link>
              </div>
            </div>
          </main>
        </div>
      </div>
    </div>
  );
}
