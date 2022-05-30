import { FC, Fragment, useState } from 'react';
import { Menu, Transition } from '@headlessui/react';
import { ChevronDownIcon } from '@heroicons/react/solid';
import { useTranslation } from 'react-i18next';
import { Status } from 'types/election';

function classNames(...classes) {
  return classes.filter(Boolean).join(' ');
}

type ElectionTableFilterProps = {
  setStatusToKeep: (status: Status) => void;
};

const ElectionTableFilter: FC<ElectionTableFilterProps> = ({ setStatusToKeep }) => {
  const { t } = useTranslation();
  const [filterByText, setFilterByText] = useState(t('filterByStatus') as string);

  const handleClick = (statusToKeep: Status, filterText) => {
    setStatusToKeep(statusToKeep);
    setFilterByText(filterText);
  };

  return (
    <>
      <Menu as="div" className="relative z-50 inline-block text-left py-6">
        <div>
          <Menu.Button className="inline-flex justify-center w-full rounded-md border border-gray-300 shadow-sm px-4 py-2 bg-white text-sm font-medium text-gray-700 hover:bg-gray-50">
            {filterByText}
            <ChevronDownIcon className="-mr-1 ml-2 h-5 w-5" aria-hidden="true" />
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
          <Menu.Items className="origin-top-left absolute left-0 mt-2 w-56 rounded-md shadow-lg bg-white ring-1 ring-black ring-opacity-5 focus:outline-none">
            <div className="py-1">
              <Menu.Item>
                {({ active }) => (
                  <button
                    onClick={() => handleClick(Status.Initial, t('statusInitial'))}
                    className={classNames(
                      active ? 'bg-gray-100 text-gray-900' : 'text-gray-700',
                      'block px-4 py-2 text-sm w-full text-left'
                    )}>
                    {t('statusInitial')}
                  </button>
                )}
              </Menu.Item>
              <Menu.Item>
                {({ active }) => (
                  <button
                    onClick={() => handleClick(Status.Open, t('statusOpen'))}
                    className={classNames(
                      active ? 'bg-gray-100 text-gray-900' : 'text-gray-700',
                      'block px-4 py-2 text-sm w-full text-left'
                    )}>
                    {t('statusOpen')}
                  </button>
                )}
              </Menu.Item>
              <Menu.Item>
                {({ active }) => (
                  <button
                    onClick={() => handleClick(Status.Closed, t('statusClose'))}
                    className={classNames(
                      active ? 'bg-gray-100 text-gray-900' : 'text-gray-700',
                      'block px-4 py-2 text-sm w-full text-left'
                    )}>
                    {t('statusClose')}
                  </button>
                )}
              </Menu.Item>
              <Menu.Item>
                {({ active }) => (
                  <button
                    onClick={() => handleClick(Status.ShuffledBallots, t('statusShuffle'))}
                    className={classNames(
                      active ? 'bg-gray-100 text-gray-900' : 'text-gray-700',
                      'block px-4 py-2 text-sm w-full text-left'
                    )}>
                    {t('statusShuffle')}
                  </button>
                )}
              </Menu.Item>
              <Menu.Item>
                {({ active }) => (
                  <button
                    onClick={() => handleClick(Status.PubSharesSubmitted, t('statusDecrypted'))}
                    className={classNames(
                      active ? 'bg-gray-100 text-gray-900' : 'text-gray-700',
                      'block px-4 py-2 text-sm w-full text-left'
                    )}>
                    {t('statusDecrypted')}
                  </button>
                )}
              </Menu.Item>
              <Menu.Item>
                {({ active }) => (
                  <button
                    onClick={() => handleClick(Status.ResultAvailable, t('statusResultAvailable'))}
                    className={classNames(
                      active ? 'bg-gray-100 text-gray-900' : 'text-gray-700',
                      'block px-4 py-2 text-sm w-full text-left'
                    )}>
                    {t('statusResultAvailable')}
                  </button>
                )}
              </Menu.Item>
              <Menu.Item>
                {({ active }) => (
                  <button
                    onClick={() => handleClick(Status.Canceled, t('statusCancel'))}
                    className={classNames(
                      active ? 'bg-gray-100 text-gray-900' : 'text-gray-700',
                      'block px-4 py-2 text-sm w-full text-left'
                    )}>
                    {t('statusCancel')}
                  </button>
                )}
              </Menu.Item>
              <Menu.Item>
                {({ active }) => (
                  <button
                    onClick={() => handleClick(null, t('filterByStatus'))}
                    className={classNames(
                      active ? 'bg-gray-100 text-gray-900' : 'text-gray-700',
                      'block px-4 py-2 text-sm w-full text-left'
                    )}>
                    {t('all')}
                  </button>
                )}
              </Menu.Item>
            </div>
          </Menu.Items>
        </Transition>
      </Menu>
      <button
        type="button"
        onClick={() => handleClick(null, t('filterByStatus'))}
        className="text-gray-700 my-2 mx-2 items-center px-4 py-2 border border-gray-300 rounded-md text-sm font-medium hover:text-indigo-500">
        {t('resetFilter')}
      </button>
    </>
  );
};

export default ElectionTableFilter;
