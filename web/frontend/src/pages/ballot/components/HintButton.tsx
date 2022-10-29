import { FC, Fragment } from 'react';
import { QuestionMarkCircleIcon } from '@heroicons/react/outline/';
import { Popover, Transition } from '@headlessui/react';

type HintButtonProps = {
  text: string;
};

const HintButton: FC<HintButtonProps> = ({ text }) => {
  return (
    text.length !== 0 && (
      <Popover className="relative ">
        <Popover.Button>
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
          <Popover.Panel className="z-30 absolute p-2 max-w-prose mt-1 ml-2 rounded-md bg-white rounded-lg shadow-lg ring-1 ring-black ring-opacity-5">
            {<div className="text-sm">{text}</div>}
          </Popover.Panel>
        </Transition>
      </Popover>
    )
  );
};

export default HintButton;
