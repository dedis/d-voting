import React, { FC, useContext, useState } from 'react';
import { NavLink, useNavigate } from 'react-router-dom';
import { useTranslation } from 'react-i18next';
import { ENDPOINT_LOGOUT } from '../components/utils/Endpoints';

import {
  ROUTE_ABOUT,
  ROUTE_ADMIN,
  ROUTE_BALLOT_INDEX,
  ROUTE_ELECTION_CREATE,
  ROUTE_ELECTION_INDEX,
  ROUTE_HOME,
} from '../Routes';
import logo from '../assets/logo.png';
import { LanguageSelector } from '../language';
import { AuthContext, FlashContext, FlashLevel } from '..';
import handleLogin from 'pages/session/HandleLogin';
import { PlusIcon } from '@heroicons/react/solid';
import Profile from './components/Profile';

// TODO: change mobile menu : put everything in the dropdown

const NavBar: FC = () => {
  const { t } = useTranslation();

  const authCtx = useContext(AuthContext);
  const [loginError, setLoginError] = useState(null);

  const navigate = useNavigate();

  const fctx = useContext(FlashContext);

  // used for the mobile menu button
  const [menuToggle, setMenuToggle] = useState(false);
  const triggerMenuToggle = () => {
    setMenuToggle(!menuToggle);
  };

  const handleLogout = async (e) => {
    e.preventDefault();

    const opts = { method: 'POST' };

    const res = await fetch(ENDPOINT_LOGOUT, opts);
    if (res.status !== 200) {
      fctx.addMessage(t('logOutError', { error: res.statusText }), FlashLevel.Error);
    } else {
      fctx.addMessage(t('logOutSuccessful'), FlashLevel.Info);
    }

    authCtx.isLogged = false;
    navigate('/');
  };

  return (
    <nav className="w-full border-b">
      <div className="max-w-7xl mx-auto px-2 sm:px-6 lg:px-8">
        <div className="relative flex items-center justify-between h-16">
          {/* Mobile icon */}
          <div className="absolute inset-y-0 left-0 flex items-center sm:hidden">
            <button
              onClick={triggerMenuToggle}
              className="inline-flex items-center justify-center p-2 rounded-md text-gray-400 hover:text-white hover:bg-gray-700 focus:outline-none focus:ring-2 focus:ring-inset focus:ring-white"
              aria-controls="mobile-menu"
              aria-expanded="false">
              <span className="sr-only">Open main menu</span>
              <svg
                className={`${menuToggle ? 'hidden' : 'block'} h-6 w-6`}
                xmlns="http://www.w3.org/2000/svg"
                fill="none"
                viewBox="0 0 24 24"
                stroke="currentColor"
                aria-hidden="true">
                <path
                  strokeLinecap="round"
                  strokeLinejoin="round"
                  strokeWidth="2"
                  d="M4 6h16M4 12h16M4 18h16"
                />
              </svg>

              <svg
                className={`${menuToggle ? 'block' : 'hidden'} h-6 w-6`}
                xmlns="http://www.w3.org/2000/svg"
                fill="none"
                viewBox="0 0 24 24"
                stroke="currentColor"
                aria-hidden="true">
                <path
                  strokeLinecap="round"
                  strokeLinejoin="round"
                  strokeWidth="2"
                  d="M6 18L18 6M6 6l12 12"
                />
              </svg>
            </button>
          </div>

          <div className="flex-1 flex items-center justify-center sm:justify-start">
            <div className="flex-shrink-0 flex items-center">
              <NavLink to={ROUTE_HOME}>
                <img className="block lg:hidden h-10 w-auto" src={logo} alt="Workflow" />
                <img className="hidden lg:block h-10 w-auto" src={logo} alt="Workflow" />
              </NavLink>
            </div>
            <div className="hidden sm:block sm:ml-6">
              <div className="flex space-x-6 mt-1">
                <NavLink
                  to={ROUTE_ELECTION_INDEX}
                  title={t('navBarStatus')}
                  className={'text-black text-lg hover:text-indigo-700'}>
                  {t('navBarStatus')}
                </NavLink>

                {authCtx.isLogged && (authCtx.role === 'admin' || authCtx.role === 'operator') && (
                  <NavLink
                    title={t('navBarVote')}
                    to={ROUTE_BALLOT_INDEX}
                    className={'text-black text-lg hover:text-indigo-700'}>
                    {t('navBarVote')}
                  </NavLink>
                )}

                {authCtx.role === 'admin' && authCtx.isLogged && (
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

          <div className="absolute inset-y-0 right-0 flex items-center pr-2 sm:static sm:inset-auto sm:ml-6 sm:pr-0">
            {authCtx.isLogged && (authCtx.role === 'admin' || authCtx.role === 'operator') && (
              <NavLink title={t('navBarCreateElection')} to={ROUTE_ELECTION_CREATE}>
                <div className="whitespace-nowrap inline-flex items-center justify-center px-4 py-2 border border-transparent rounded-md shadow-sm text-base font-medium text-white bg-indigo-600 hover:bg-indigo-700">
                  <PlusIcon className="-ml-1 mr-2 h-5 w-5" aria-hidden="true" />
                  {t('navBarCreateElection')}
                </div>
              </NavLink>
            )}
            <LanguageSelector />
            <Profile
              authCtx={authCtx}
              handleLogout={handleLogout}
              handleLogin={handleLogin}
              loginError={loginError}
              setLoginError={setLoginError}
            />
          </div>
        </div>
      </div>

      <div className="sm:hidden" style={menuToggle ? {} : { display: 'none' }} id="mobile-menu">
        <div className="px-2 pt-2 pb-3 space-y-1">
          <NavLink
            to={ROUTE_ELECTION_CREATE}
            title={t('navBarCreate')}
            className={(isActive) =>
              isActive
                ? 'bg-gray-900 text-white px-3 py-2'
                : 'text-gray-300 hover:bg-gray-700 hover:text-white block px-3 py-2 rounded-md text-base font-medium'
            }>
            {t('navBarCreate')}
          </NavLink>

          <NavLink
            to={ROUTE_ELECTION_INDEX}
            title={t('navBarCreate')}
            className={(isActive) =>
              isActive
                ? 'bg-gray-900 text-white px-3 py-2'
                : 'text-gray-300 hover:bg-gray-700 hover:text-white block px-3 py-2 rounded-md text-base font-medium'
            }>
            {t('navBarStatus')}
          </NavLink>

          <NavLink
            to={ROUTE_ABOUT}
            className={(isActive) =>
              isActive
                ? 'bg-gray-900 text-white px-3 py-2'
                : 'text-gray-300 hover:bg-gray-700 hover:text-white block px-3 py-2 rounded-md text-base font-medium'
            }>
            {t('navBarAbout')}
          </NavLink>
        </div>
      </div>
    </nav>
  );
};

export default NavBar;
