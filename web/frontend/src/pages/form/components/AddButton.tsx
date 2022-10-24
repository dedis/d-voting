import { PlusSmIcon } from '@heroicons/react/outline';
import PropTypes from 'prop-types';
import { FC } from 'react';

type AddButtonProps = {
  onClick(): void;
  children: string;
};

const AddButton: FC<AddButtonProps> = ({ onClick, children }) => {
  return (
    <button
      type="button"
      className={`inline-flex mb-2 ml-2 items-center px-3 py-2 border border-gray-300 rounded-md shadow-sm text-sm font-medium text-gray-700 bg-white hover:bg-gray-50 ${
        children === 'Subject' && 'bg-gray-200'
      }`}
      onClick={onClick}>
      <PlusSmIcon className="-ml-1 mr-2 h-5 w-5" aria-hidden="true" />
      {children}
    </button>
  );
};

AddButton.propTypes = {
  onClick: PropTypes.func.isRequired,
  children: PropTypes.string.isRequired,
};

export default AddButton;
