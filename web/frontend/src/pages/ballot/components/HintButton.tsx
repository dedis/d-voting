import { FC, Fragment, useRef } from 'react';
import { QuestionMarkCircleIcon } from '@heroicons/react/outline/';
import { Popover, Transition } from '@headlessui/react';

type HintButtonProps = {
  text: string;
};

const HintButton: FC<HintButtonProps> = ({ text }) => {
  const buttonRef = useRef(null);
  const timeoutDuration = 200;
  let timeout;

  const closePopover = () => {
    return buttonRef.current?.dispatchEvent(
      new KeyboardEvent('keydown', {
        key: 'Escape',
        bubbles: true,
        cancelable: true,
      })
    );
  };

  const onMouseEnter = (open) => {
    clearTimeout(timeout);
    if (open) return;
    return buttonRef.current?.click();
  };

  const onMouseLeave = (open) => {
    if (!open) return;
    timeout = setTimeout(() => closePopover(), timeoutDuration);
  };

  return (
    text.length !== 0 && (
      <Popover className="relative ">
        {({ open }) => {
          return (
            <>
              <div onMouseLeave={onMouseLeave.bind(null, open)}>
                <Popover.Button
                  ref={buttonRef}
                  onMouseEnter={onMouseEnter.bind(null, open)}
                  onMouseLeave={onMouseLeave.bind(null, open)}>
                  <div className="text-gray-600">
                    <QuestionMarkCircleIcon className="color-gray-900 mt-2 h-4 w-4" />
                  </div>
                </Popover.Button>
                <Transition
                  as={Fragment}
                  enter="transition ease-out duration-100"
                  enterFrom="transform opacity-0 scale-95"
                  enterTo="transform opacity-100 scale-100"
                  leave="transition ease-in duration-75"
                  leaveFrom="transform opacity-100 scale-100"
                  leaveTo="transform opacity-0 scale-95">
                  <Popover.Panel
                    onMouseEnter={onMouseEnter.bind(null, open)}
                    onMouseLeave={onMouseLeave.bind(null, open)}
                    className="z-30 absolute right-0 p-2 w-96 mt-1 ml-2 rounded-md bg-white rounded-lg shadow-lg ring-1 ring-black ring-opacity-5">
                    {<div className="text-sm">{text}</div>}
                  </Popover.Panel>
                </Transition>
              </div>
            </>
          );
        }}
      </Popover>
    )
  );
};

export default HintButton;
