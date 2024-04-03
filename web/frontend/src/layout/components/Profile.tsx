import React, { FC, Fragment } from 'react';
import { useTranslation } from 'react-i18next';
import { AuthState, FlashState } from 'index';
import PropTypes from 'prop-types';

import { Menu, Transition } from '@headlessui/react';
import { UserCircleIcon } from '@heroicons/react/outline';

type ProfileProps = {
  authCtx: AuthState;
  handleLogout: (e: any) => Promise<void>;
  handleLogin: (arg0: FlashState) => Promise<void>;
  fctx: FlashState;
};

const Profile: FC<ProfileProps> = ({ authCtx, handleLogout, handleLogin, fctx }) => {
  const { t } = useTranslation();

  return authCtx.isLogged ? (
    <Menu as="div">
      <div>
        <Menu.Button className="flex text-sm rounded-full text-gray-400 hover:text-white">
          <span className="sr-only">Profile</span>
          <UserCircleIcon className="h-7 w-7 text-neutral-600" aria-hidden="true" />
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
        <Menu.Items className="z-40 origin-top-right absolute right-0 mt-2 w-48 rounded-md shadow-lg py-1 bg-white ring-1 ring-black ring-opacity-5 focus:outline-none">
          <Menu.Item>
            <p className="block px-4 py-2 text-sm text-gray-400">
              Logged as {authCtx.firstName} {authCtx.lastName}
            </p>
          </Menu.Item>
          <Menu.Item>
            <div
              onClick={handleLogout}
              className={'cursor-pointer block px-4 py-2 text-sm text-gray-700'}>
              {t('logout')}
            </div>
          </Menu.Item>
        </Menu.Items>
      </Transition>
    </Menu>
  ) : (
    <button
      onClick={() => handleLogin(fctx)}
      className="whitespace-nowrap inline-flex items-center justify-center px-4 py-2 border border-transparent rounded-md shadow-sm text-base font-medium text-white bg-[#ff0000] hover:bg-[#b51f1f]">
      {t('login')}
    </button>
  );
};

Profile.propTypes = {
  authCtx: PropTypes.any.isRequired,
  handleLogout: PropTypes.func.isRequired,
  handleLogin: PropTypes.func.isRequired,
  fctx: PropTypes.any.isRequired,
};

export default Profile;
