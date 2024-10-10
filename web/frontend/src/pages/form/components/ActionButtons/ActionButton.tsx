import { FC } from 'react';
import IndigoSpinnerIcon from '../IndigoSpinnerIcon';

type ActionButtonProps = {
  handleClick: () => any;
  ongoing: boolean;
  ongoingText: string;
  children: JSX.Element;
};

const ActionButton: FC<ActionButtonProps> = ({ handleClick, ongoing, ongoingText, children }) => {
  return !ongoing ? (
    <button onClick={handleClick}>
      <div className="whitespace-nowrap inline-flex items-center justify-center px-4 py-1 mr-2 border border-gray-300 text-sm rounded-full font-medium text-gray-700 hover:text-[#ff0000]">
        {children}
      </div>
    </button>
  ) : (
    <div className="whitespace-nowrap inline-flex items-center justify-center px-4 py-1 mr-2 border border-gray-300 text-sm rounded-full font-medium text-gray-700">
      <IndigoSpinnerIcon />
      {ongoingText}
    </div>
  );
};

export default ActionButton;
