import { FC } from 'react';

type ProgressBarProps = {
  isBest: boolean;
  children: string;
};

const ProgressBar: FC<ProgressBarProps> = ({ isBest, children }) => {
  return (
    <div className="h-min bg-gray-300 rounded-full">
      {children === '0.00' ? (
        <div
          className="bg-gray-300 px-1 text-xs font-medium text-white p-0.5 leading-none rounded-full"
          style={{ width: `100%` }}>
          {`${children}%`}
        </div>
      ) : (
        <div
          className={`${!isBest && 'bg-indigo-300'} ${
            isBest && 'bg-indigo-500'
          } text-xs font-medium text-white text-center p-0.5 leading-none rounded-full`}
          style={{ width: `${children}%` }}>
          {`${children}%`}
        </div>
      )}
    </div>
  );
};

export default ProgressBar;
