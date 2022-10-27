import { XIcon } from '@heroicons/react/outline';
import PropTypes from 'prop-types';
import { FC } from 'react';

type DeleteButtonProps = {
  onClick(): void;
  children: string;
};

const DeleteButton: FC<DeleteButtonProps> = ({ onClick, children }) => {
  return (
    <div className="relative">
      <button
        type="button"
        className="-mr-1 flex p-2 absolute top-1 right-3 rounded-md bg-red-600 hover:bg-red-700 sm:-mr-2"
        onClick={onClick}>
        <span className="sr-only">Delete {children}</span>
        <XIcon className="h-3 w-3 text-white" aria-hidden="true" />
      </button>
    </div>
  );
};

DeleteButton.propTypes = {
  onClick: PropTypes.func.isRequired,
  children: PropTypes.string.isRequired,
};

export default DeleteButton;
