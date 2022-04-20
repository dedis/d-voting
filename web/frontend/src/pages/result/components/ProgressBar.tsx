import { FC } from 'react';

type ProgressBarProps = {
  candidate: string;
  isBest: boolean;
  children: string;
};

const ProgressBar: FC<ProgressBarProps> = ({ candidate, isBest, children }) => {
  return (
    <div className="px-4 flex justify-between items-center">
      <div className="flex-initial w-2/5 max-w-xs break-words pr-2">
        <span className={`${isBest && 'font-bold'}`}>{candidate}</span>:
      </div>
      <div className="flex-auto w-4/5 h-min bg-gray-400 rounded-full">
        {children === '0.00' ? (
          <div
            className="bg-gray-400 px-1 text-xs font-medium text-white p-0.5 leading-none rounded-full"
            style={{ width: `100%` }}>
            {children}
          </div>
        ) : (
          <div
            className={`${!isBest && 'bg-indigo-400'} ${
              isBest && 'bg-indigo-600'
            } text-xs font-medium text-white text-center p-0.5 leading-none rounded-full`}
            style={{ width: `${children}%` }}>
            {children}
          </div>
        )}
      </div>
    </div>
  );
};

export default ProgressBar;
