import React, { Fragment, useState } from 'react';
import { Menu, Transition } from '@headlessui/react';
import { UserCircleIcon } from '@heroicons/react/outline';
import { useTranslation } from 'react-i18next';

// TODO: proptype & interface validation

const Profile = ({ authCtx, handleLogout, handleLogin, loginError, setLoginError }) => {
  const { t } = useTranslation();

  const [toggle, setToggle] = useState(false);
  const triggerToggle = () => {
    setToggle(!toggle);
  };

  return (
    <Menu as="div">
      <div>
        <Menu.Button
          onClick={triggerToggle}
          className="flex text-sm rounded-full text-gray-400 hover:text-white">
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
        <Menu.Items className="origin-top-right absolute right-0 mt-2 w-48 rounded-md shadow-lg py-1 bg-white ring-1 ring-black ring-opacity-5 focus:outline-none">
          {authCtx.isLogged ? (
            <>
              <Menu.Item>
                {/* <div className={'cursor-pointer block px-4 py-2 text-sm text-gray-700'}>
                    Logged as {authCtx.firstname} {authCtx.lastname}
                  </div> */}
                <p className="block px-4 py-2 text-sm text-gray-400">
                  Logged as {authCtx.firstname} {authCtx.lastname}
                </p>
              </Menu.Item>
              <Menu.Item onClick={handleLogout}>
                <div className={'cursor-pointer block px-4 py-2 text-sm text-gray-700'}>
                  {t('logout')}
                </div>
              </Menu.Item>
            </>
          ) : (
            <Menu.Item>
              <div
                onClick={() => handleLogin(loginError, setLoginError)}
                className={'cursor-pointer block px-4 py-2 text-sm text-gray-700'}>
                {t('login')}
              </div>
            </Menu.Item>
          )}
        </Menu.Items>
      </Transition>
    </Menu>
  );
};

export default Profile;
