import { useTranslation } from 'react-i18next';
import { Link } from 'react-router-dom';
import { ROUTE_FORM_INDEX } from 'Routes';

export default function FormNotAvailable(props) {
  const { t } = useTranslation();

  return (
    <div className="h-[calc(100vh-130px)]">
      <div className="bg-white min-h-full font-sans px-4 py-16 sm:px-6 sm:py-24 md:grid md:place-items-center lg:px-8">
        <div className="max-w-max mx-auto">
          <main className="sm:flex">
            <div className="sm:ml-6">
              <div className=" sm:border-gray-200 sm:pl-6">
                <h1 className="text-4xl font-extrabold text-gray-900 tracking-tight sm:text-5xl">
                  {props.isVoter ? t('voteImpossible') : t('voteNotVoter')}
                </h1>
                <p className="mt-1 text-base text-gray-500">
                  {props.isVoter ? t('voteImpossibleDescription') : t('voteNotVoterDescription')}
                </p>
              </div>
              <div className="mt-10 flex space-x-3 sm:border-l sm:border-transparent sm:pl-6">
                <Link
                  to={ROUTE_FORM_INDEX}
                  className="inline-flex items-center px-4 py-2 border border-transparent text-sm font-medium rounded-md shadow-sm text-white bg-[#ff0000] hover:bg-[#b51f1f] focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-indigo-500">
                  {t('returnToFormTable')}
                </Link>
              </div>
            </div>
          </main>
        </div>
      </div>
    </div>
  );
}
