import { CursorClickIcon, MenuAlt1Icon, SwitchVerticalIcon } from '@heroicons/react/outline';
import { RANK, SELECT, TEXT } from 'types/configuration';

const DisplayTypeIcon = ({ Type }) => {
  switch (Type) {
    case RANK:
      return <SwitchVerticalIcon className="m-2 h-5 w-5 text-gray-400" aria-hidden="true" />;
    case SELECT:
      return <CursorClickIcon className="m-2 h-5 w-5 text-gray-400" aria-hidden="true" />;
    case TEXT:
      return <MenuAlt1Icon className="m-2 h-5 w-5 text-gray-400" aria-hidden="true" />;
    default:
      return null;
  }
};

export default DisplayTypeIcon;
