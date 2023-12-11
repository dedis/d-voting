import { useTranslation } from 'react-i18next';

const Footer = () => {
  const { t } = useTranslation();
  return (
    <div className="flex flex-row border-t justify-center bg-white items-center w-full p-4 text-gray-300 text-xs">
      <footer data-testid="footer">
        <div className="hidden sm:flex flex-row items-center max-w-7xl mx-auto py-2 px-4 overflow-hidden sm:px-6 lg:px-8">
          <span data-testid="footerCopyright" className="text-gray-400">
            &copy; {new Date().getFullYear()} {t('footerCopyright')}&nbsp;
            <a className="text-gray-600" href="https://github.com/c4dt/d-voting">
              https://github.com/c4dt/d-voting
            </a>
          </span>
        </div>
        <div className="text-center">
          version:
          {process.env.REACT_APP_VERSION || 'unknown'} - build{' '}
          {process.env.REACT_APP_BUILD || 'unknown'} - on{' '}
          {process.env.REACT_APP_BUILD_TIME || 'unknown'}
        </div>
      </footer>
    </div>
  );
};

export default Footer;
