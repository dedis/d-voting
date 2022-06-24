import { FC } from 'react';
import IndigoSpinnerIcon from './IndigoSpinnerIcon';

type LoadingButtonProps = {
  children: string;
};

const LoadingButton: FC<LoadingButtonProps> = ({ children }) => {
  return (
    <div className="whitespace-nowrap inline-flex items-center justify-center px-4 py-1 mr-2 border border-gray-300 text-sm rounded-full font-medium text-gray-700">
      <IndigoSpinnerIcon />
      {children}
    </div>
  );
};

export default LoadingButton;
