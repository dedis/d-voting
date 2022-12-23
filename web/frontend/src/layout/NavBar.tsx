import React, { FC, Fragment, useContext, useState } from 'react';
import { NavLink, useNavigate } from 'react-router-dom';
import { useTranslation } from 'react-i18next';
import { default as i18n } from 'i18next';

import { ENDPOINT_LOGOUT } from '../components/utils/Endpoints';
import {
  ROUTE_ABOUT,
  ROUTE_ADMIN,
  ROUTE_FORM_CREATE,
  ROUTE_FORM_INDEX,
  ROUTE_HOME,
} from '../Routes';

import WarningModal from './components/WarningModal';
import { AuthContext, FlashContext, FlashLevel } from '..';
import handleLogin from 'pages/session/HandleLogin';
import Profile from './components/Profile';

import { availableLanguages } from 'language/Configuration';
import { LanguageSelector } from '../language';

import logo from '../assets/logo.png';
import { Popover, Transition } from '@headlessui/react';
import { LoginIcon, LogoutIcon, MenuIcon, XIcon } from '@heroicons/react/outline';
import { PlusIcon } from '@heroicons/react/solid';

const SUBJECT_ELECTION = 'election';
const ACTION_CREATE = 'create';
const SUBJECT_ROLES = 'roles';
const ACTION_ADD = 'add';
const ACTION_LIST = 'list';

function hasAuthorization(authCtx, subject: string, action: string): boolean {
  return (
    authCtx.authorization.has(subject) && authCtx.authorization.get(subject).indexOf(action) !== -1
  );
}
const MobileMenu = ({ authCtx, handleLogout, fctx, t }) => (
  <Popover>
    <div className="-mr-2 -my-2 md:hidden">
      <Popover.Button className="bg-white rounded-md p-2 inline-flex items-center justify-center text-gray-400 hover:text-gray-500 hover:bg-gray-100 focus:outline-none focus:ring-2 focus:ring-inset focus:ring-indigo-500">
        <span className="sr-only">Open menu</span>
        <MenuIcon className="h-6 w-6" aria-hidden="true" />
      </Popover.Button>
    </div>
    <Transition
      as={Fragment}
      enter="duration-200 ease-out"
      enterFrom="opacity-0 scale-100"
      enterTo="opacity-100 scale-100"
      leave="duration-100 ease-in"
      leaveFrom="opacity-100 scale-100"
      leaveTo="opacity-0 scale-100">
      <Popover.Panel
        focus
        className="absolute top-0 z-40 inset-x-0 p-2 transition transform origin-top-right md:hidden">
        <div className="rounded-lg shadow-lg ring-1 ring-black ring-opacity-5 bg-white divide-y-2 divide-gray-50">
          <div className="pt-5 pb-6 px-5">
            <div className="flex items-center justify-between">
              <div className="-mr-2">
                <Popover.Button className="bg-white rounded-md p-2 inline-flex items-center justify-center text-gray-400 hover:text-gray-500 hover:bg-gray-100 focus:outline-none focus:ring-2 focus:ring-inset focus:ring-indigo-500">
                  <span className="sr-only">Close menu</span>
                  <XIcon className="h-6 w-6" aria-hidden="true" />
                </Popover.Button>
              </div>
              <div>
                <img className="h-10 w-auto" src={logo} alt="Workflow" />
              </div>
            </div>
            <div className="mt-6">
              <nav className="grid gap-y-8">
                {
                  <NavLink to={ROUTE_FORM_INDEX}>
                    <Popover.Button className=" w-full -m-3 p-3 flex items-center rounded-md hover:bg-gray-50">
                      <span className="ml-3 text-base font-medium text-gray-900">
                        {t('navBarStatus')}
                      </span>
                    </Popover.Button>
                  </NavLink>
                }
                {authCtx.isLogged && hasAuthorization(authCtx, SUBJECT_ROLES, ACTION_ADD) && (
                  <NavLink to={ROUTE_ADMIN}>
                    <Popover.Button className=" w-full -m-3 p-3 flex items-center rounded-md hover:bg-gray-50">
                      <span className="ml-3 text-base font-medium text-gray-900">
                        {t('navBarAdmin')}
                      </span>
                    </Popover.Button>
                  </NavLink>
                )}
                {!authCtx.isLogged && (
                  <NavLink to={ROUTE_ABOUT}>
                    <Popover.Button className=" w-full -m-3 p-3 flex items-center rounded-md hover:bg-gray-50">
                      <span className="ml-3 text-base font-medium text-gray-900">
                        {t('navBarAbout')}
                      </span>
                    </Popover.Button>
                  </NavLink>
                )}
              </nav>
            </div>
            <div className="pt-4">
              {authCtx.isLogged && hasAuthorization(authCtx, SUBJECT_ELECTION, ACTION_CREATE) && (
                <NavLink to={ROUTE_FORM_CREATE}>
                  <Popover.Button className="w-full flex items-center justify-center px-4 py-2 border border-transparent rounded-md shadow-sm text-base font-medium text-white bg-indigo-600 hover:bg-indigo-700">
                    <PlusIcon className="-ml-1 mr-2 h-5 w-5" aria-hidden="true" />
                    {t('navBarCreateForm')}
                  </Popover.Button>
                </NavLink>
              )}
            </div>
          </div>

          <div className="py-6 px-5 space-y-6">
            <div className="grid grid-cols-2 gap-y-4 gap-x-8">
              {availableLanguages.map(
                (lang) =>
                  i18n.language !== lang && (
                    <Popover.Button key={lang}>
                      <div
                        className="text-base font-medium text-gray-900 hover:text-gray-700"
                        onClick={() => i18n.changeLanguage(lang)}>
                        {t(lang)}
                      </div>
                    </Popover.Button>
                  )
              )}
            </div>
          </div>

          <div className="py-6 px-5 space-y-6">
            {authCtx.isLogged && (
              <div>
                Logged as {authCtx.firstname} {authCtx.lastname}
              </div>
            )}
            <div>
              {authCtx.isLogged ? (
                <div onClick={handleLogout}>
                  <Popover.Button className="w-full flex items-center justify-center px-4 py-2 border  rounded-md shadow-sm text-base font-medium">
                    <LogoutIcon className="-ml-1 mr-2 h-5 w-5" aria-hidden="true" />
                    {t('logout')}
                  </Popover.Button>
                </div>
              ) : (
                <div onClick={() => handleLogin(fctx)}>
                  <Popover.Button className="w-full flex items-center justify-center px-4 py-2 border border-transparent rounded-md shadow-sm text-base font-medium text-white bg-indigo-600 hover:bg-indigo-700">
                    <LoginIcon className="-ml-1 mr-2 h-5 w-5" aria-hidden="true" />
                    {t('login')}
                  </Popover.Button>
                </div>
              )}
            </div>
          </div>
        </div>
      </Popover.Panel>
    </Transition>
  </Popover>
);

const RightSideNavBar = ({ authCtx, handleLogout, fctx, t }) => (
  <div className="absolute hidden inset-y-0 right-0 flex items-center pr-2 md:static md:inset-auto md:flex md:ml-6 md:pr-0">
    {authCtx.isLogged && hasAuthorization(authCtx, SUBJECT_ELECTION, ACTION_CREATE) && (
      <NavLink title={t('navBarCreateForm')} to={ROUTE_FORM_CREATE}>
        <div className="whitespace-nowrap inline-flex items-center justify-center px-4 py-2 border-2 border-indigo-500 rounded-md shadow-sm text-base font-medium text-indigo-500 bg-white hover:bg-indigo-500 hover:text-white">
          <PlusIcon className="-ml-1 mr-2 h-5 w-5" aria-hidden="true" />
          {t('navBarCreateForm')}
        </div>
      </NavLink>
    )}
    <LanguageSelector />
    <Profile authCtx={authCtx} handleLogout={handleLogout} handleLogin={handleLogin} fctx={fctx} />
  </div>
);

const LeftSideNavBar = ({ authCtx, t }) => (
  <div className="flex-1 flex items-center justify-center md:justify-start">
    <div className="flex-shrink-0 flex items-center">
      <NavLink to={ROUTE_HOME}>
        <img className="block lg:hidden h-10 w-auto" src={logo} alt="Workflow" />
        <img className="hidden lg:block h-10 w-auto" src={logo} alt="Workflow" />
      </NavLink>
    </div>
    <div className="hidden md:block md:ml-6">
      <div className="flex space-x-6 mt-1">
        <NavLink
          to={ROUTE_FORM_INDEX}
          title={t('navBarStatus')}
          className={'text-black text-lg hover:text-indigo-700'}>
          {t('navBarStatus')}
        </NavLink>
        {authCtx.isLogged && hasAuthorization(authCtx, SUBJECT_ROLES, ACTION_LIST) && (
          <NavLink to={ROUTE_ADMIN} className={'text-black text-lg hover:text-indigo-700'}>
            Admin
          </NavLink>
        )}
        <NavLink to={ROUTE_ABOUT} className={'text-black text-lg hover:text-indigo-700'}>
          {t('navBarAbout')}
        </NavLink>
      </div>
    </div>
  </div>
);

const NavBar: FC = () => {
  const { t } = useTranslation();

  const authCtx = useContext(AuthContext);

  const navigate = useNavigate();

  const fctx = useContext(FlashContext);
  const [isShown, setIsShown] = useState(false);

  const logout = async () => {
    const opts = { method: 'POST' };

    const res = await fetch(ENDPOINT_LOGOUT, opts);
    if (res.status !== 200) {
      fctx.addMessage(t('logOutError', { error: res.statusText }), FlashLevel.Error);
    } else {
      fctx.addMessage(t('logOutSuccessful'), FlashLevel.Info);
    }
    // TODO: should be a setAuth function passed to AuthContext rather than
    // changing the state directly
    authCtx.isLogged = false;
    authCtx.firstname = undefined;
    authCtx.role = undefined;
    authCtx.lastname = undefined;
    navigate('/');
  };
  const handleLogout = async (e) => {
    e.preventDefault();
    setIsShown(true);
  };

  return (
    <nav className="w-full border-b">
      <div className="max-w-7xl mx-auto px-2 md:px-6 lg:px-8">
        <div className="relative flex items-center justify-between h-16">
          <MobileMenu authCtx={authCtx} handleLogout={handleLogout} fctx={fctx} t={t} />
          <LeftSideNavBar authCtx={authCtx} t={t} />
          <RightSideNavBar authCtx={authCtx} handleLogout={handleLogout} fctx={fctx} t={t} />
          <WarningModal
            isShown={isShown}
            setIsShown={setIsShown}
            action={async () => logout()}
            message={t('logoutWarning')}
          />
        </div>
      </div>
    </nav>
  );
};

export default NavBar;
