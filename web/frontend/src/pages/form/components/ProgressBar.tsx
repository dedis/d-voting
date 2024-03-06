import { FC } from 'react';

type ProgressBarProps = {
  isBest: boolean;
  children: string;
};

type SelectProgressBarProps = {
  percent: string;
  totalCount: number;
  numberOfBallots: number;
  isBest: boolean;
};

export const ProgressBar: FC<ProgressBarProps> = ({ isBest, children }) => {
  return (
    <div className="sm:ml-1 md:ml-2 w-3/5 sm:w-4/5">
      <div className="h-min bg-white rounded-full mr-1 md:mr-2 w-full flex items-center">
        <div
          className={`${!isBest && children !== '0.00' && 'bg-[#ff0000]'} ${
            !isBest && children === '0.00' && 'bg-[#ff0000]'
          } ${isBest && 'bg-[#ff0000]'}  flex-none mr-2 text-white h-2 sm:h-3 p-0.5 rounded-full`}
          style={{ width: `${children}%` }}></div>

        <div className="text-gray-700 text-sm">{`${children}%`}</div>
      </div>
    </div>
  );
};

export const SelectProgressBar: FC<SelectProgressBarProps> = ({
  percent,
  totalCount,
  numberOfBallots,
  isBest,
}) => {
  return (
    <div className="sm:ml-1 md:ml-2 w-3/5 sm:w-4/5">
      <div className="h-min bg-white rounded-full mr-1 md:mr-2 w-full flex items-center">
        <div
          className={`${!isBest && totalCount !== 0 && 'bg-[#ff0000]'} ${
            !isBest && totalCount === 0 && 'bg-[#ff0000]'
          } ${isBest && 'bg-[#ff0000]'}  flex-none mr-2 text-white h-2 sm:h-3 p-0.5 rounded-full`}
          style={{ width: `${percent}%` }}></div>

        <div className="text-gray-700 text-sm">{`${totalCount}/${numberOfBallots}`}</div>
      </div>
    </div>
  );
};
