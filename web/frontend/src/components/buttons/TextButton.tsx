import { FC } from 'react';

type TextButtonProps = {
  children: string;
};

// A button with text
const TextButton: FC<TextButtonProps> = ({ children }) => {
  return (
    <button
      type="button"
      className="text-gray-700 my-2 mr-2 items-center px-4 py-2 border rounded-md text-sm hover:text-white hover:bg-indigo-500">
      {children}
    </button>
  );
};

export default TextButton;
