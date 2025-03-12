import { FlashContext, FlashLevel } from 'index';
import { useContext } from 'react';
import {
  ExclamationCircleIcon,
  ExclamationIcon,
  InformationCircleIcon,
  XIcon,
} from '@heroicons/react/outline';
import styles from './Flash.module.css';

const Flash = () => {
  const fctx = useContext(FlashContext);

  return (
    <div className="w-full z-50">
      {fctx.getMessages().map((msg) => (
        <div
          key={msg.id}
          className={`relative
                      ${msg.getLevel() === FlashLevel.Info && 'bg-[#ff0000]'} 
                      ${msg.getLevel() === FlashLevel.Warning && 'bg-orange-500'} 
                      ${msg.getLevel() === FlashLevel.Error && 'bg-red-500'}`}>
          <div
            id={msg.id}
            className={`flex items-center text-white text-sm font-bold py-3 max-w-7xl mx-auto px-2 md:px-6 lg:px-8`}
            role="alert">
            <div className="px-2">
              {msg.getLevel() === FlashLevel.Info && <InformationCircleIcon className="h-6 w-6" />}
              {msg.getLevel() === FlashLevel.Warning && <ExclamationIcon className="h-6 w-6" />}
              {msg.getLevel() === FlashLevel.Error && <ExclamationCircleIcon className="h-6 w-6" />}
            </div>
            <p>{msg.getText()}</p>
            <button
              type="button"
              className={`ml-auto -mx-1.5 -my-1.5 rounded-lg focus:ring-2 p-1.5 inline-flex h-8 w-8 
                ${msg.getLevel() === FlashLevel.Info && 'focus:ring-[#ff0000] hover:bg-[#ff0000]'} 
                ${
                  msg.getLevel() === FlashLevel.Warning &&
                  'focus:ring-orange-400 hover:bg-orange-600'
                } 
                ${msg.getLevel() === FlashLevel.Error && 'focus:ring-red-400 hover:bg-red-600'}`}
              onClick={() => fctx.hideMessage(msg.id)}
              aria-label="Close">
              <XIcon />
            </button>
          </div>
          <div className={styles.progress} />
        </div>
      ))}
    </div>
  );
};

export default Flash;
