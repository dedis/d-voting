import { Menu, Transition } from '@headlessui/react';
import { DotsVerticalIcon } from '@heroicons/react/outline';
import { Fragment } from 'react';
import { useTranslation } from 'react-i18next';

const SubjectDropdown = ({ dropdownContent }) => {
  const { t } = useTranslation();

  return (
    <div className="flex my-2">
      <Menu as="div" className="relative inline-block text-left">
        <div>
          <Menu.Button className="inline-flex w-full justify-center bg-opacity-20 px-2 py-1 text-sm font-medium">
            <DotsVerticalIcon className="h-5 w-5 text-black" />
          </Menu.Button>
        </div>
        <Transition
          as={Fragment}
          enter="transition ease-out duration-100"
          enterFrom="transform opacity-0 scale-95"
          enterTo="transform opacity-100 scale-100"
          leave="transition ease-in duration-75"
          leaveFrom="transform opacity-100 scale-100"
          leaveTo="transform opacity-0 scale-95">
          <Menu.Items className="z-30 absolute right-0 mt-2 w-56 origin-top-right divide-y divide-gray-100 rounded-md bg-white shadow-lg ring-1 ring-black ring-opacity-5 focus:outline-none">
            <div className="px-1 py-1 ">
              {dropdownContent.map(({ name, icon, onClick }) => (
                <Menu.Item key={name}>
                  {({ active }) => (
                    <button
                      key={name}
                      onClick={onClick}
                      className={`${
                        active ? 'bg-violet-500 text-white' : 'text-gray-900'
                      } group flex w-full items-center rounded-md px-2 py-2 text-sm`}>
                      {icon}
                      {t(name)}
                    </button>
                  )}
                </Menu.Item>
              ))}
            </div>
          </Menu.Items>
        </Transition>
      </Menu>
    </div>
  );
};

export default SubjectDropdown;
