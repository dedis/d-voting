import ProxyInput from 'components/utils/proxy';

const Footer = () => (
  <div className="flex flex-row border-t justify-center bg-white items-center w-full p-4 text-gray-300 text-xs">
    <footer>
      <div className="hidden sm:flex flex-row items-center max-w-7xl mx-auto py-2 px-4 overflow-hidden sm:px-6 lg:px-8">
        <span className="text-gray-400"> &copy; 2022 DEDIS LAB - </span>
        <a className="text-gray-600" href="https://github.com/dedis/dela">
          https://github.com/dedis/dela
        </a>
        <div className="px-10">
          <ProxyInput />
        </div>
      </div>
      <div className="flex sm:hidden flex-col items-center max-w-7xl mx-auto py-2 px-4 overflow-hidden sm:px-6 lg:px-8">
        <span className="text-gray-400"> &copy; 2022 DEDIS LAB - </span>
        <a className="text-gray-600" href="https://github.com/dedis/dela">
          https://github.com/dedis/dela
        </a>
        <div className="pt-2">
          <ProxyInput />
        </div>
      </div>
    </footer>
  </div>
);

export default Footer;
