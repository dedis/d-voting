import { CloudDownloadIcon } from '@heroicons/react/solid';
import { FC } from 'react';

type DownloadButtonProps = {
  exportData: (() => Promise<void>) | (() => void);
  children: string;
};

const DownloadButton: FC<DownloadButtonProps> = ({ exportData, children }) => {
  return (
    <button
      type="button"
      className="flex my-2 mx-2 items-center px-4 py-2 border rounded-md text-sm font-medium hover:text-[#ff0000]"
      onClick={exportData}>
      <CloudDownloadIcon className="-ml-1 mr-2 h-5 w-5" aria-hidden="true" />
      {children}
    </button>
  );
};

export default DownloadButton;
