import { FlashContext, FlashLevel } from 'index';
import { useContext } from 'react';
import {
  ExclamationCircleIcon,
  ExclamationIcon,
  InformationCircleIcon,
  XIcon,
} from '@heroicons/react/outline';

const Flash = () => {
  const fctx = useContext(FlashContext);

  const closeFlash = (id: number) => {
    fctx.hideMessage(id);
    document.getElementById(id.toString()).setAttribute('class', 'hidden');
  };

  return (
    <div className="w-full fixed z-50">
      {fctx.getMessages().map((msg, i) => (
        <div key={i.toString()}>
          {msg.isVisible() ? (
            <div
              id={i.toString()}
              className={`flex items-center text-white text-sm font-bold px-4 py-3 
              ${msg.getLevel() === FlashLevel.Info && 'bg-indigo-500'} 
              ${msg.getLevel() === FlashLevel.Warning && 'bg-orange-500'} 
              ${msg.getLevel() === FlashLevel.Error && 'bg-red-500'}`}
              role="alert">
              <div className="px-2">
                {msg.getLevel() === FlashLevel.Info && (
                  <InformationCircleIcon className="h-6 w-6" />
                )}
                {msg.getLevel() === FlashLevel.Warning && <ExclamationIcon className="h-6 w-6" />}
                {msg.getLevel() === FlashLevel.Error && (
                  <ExclamationCircleIcon className="h-6 w-6" />
                )}
              </div>
              <p>{msg.getText()}</p>
              <button
                type="button"
                className={`ml-auto -mx-1.5 -my-1.5 rounded-lg focus:ring-2 p-1.5 inline-flex h-8 w-8 
                ${
                  msg.getLevel() === FlashLevel.Info && 'focus:ring-indigo-400 hover:bg-indigo-600'
                } 
                ${
                  msg.getLevel() === FlashLevel.Warning &&
                  'focus:ring-orange-400 hover:bg-orange-600'
                } 
                ${msg.getLevel() === FlashLevel.Error && 'focus:ring-red-400 hover:bg-red-600'}`}
                onClick={() => closeFlash(i)}
                aria-label="Close">
                <XIcon />
              </button>
            </div>
          ) : null}
        </div>
      ))}
    </div>
  );
};

export default Flash;
