import React, { Suspense, useContext } from 'react';
import { Navigate, Route, BrowserRouter as Router, Routes, useLocation } from 'react-router-dom';

import {
  ROUTE_ABOUT,
  ROUTE_ADMIN,
  ROUTE_BALLOT_SHOW,
  ROUTE_FORM_CREATE,
  ROUTE_FORM_INDEX,
  ROUTE_LOGGED,
  ROUTE_LOGIN,
  ROUTE_UNAUTHORIZED,
} from '../Routes';
import Login from '../pages/session/Login';
import Home from '../pages/Home';
import About from '../pages/About';
import Admin from 'pages/admin/Admin';
import FormIndex from '../pages/form/Index';
import FormCreate from '../pages/form/New';
import FormResult from '../pages/form/Result';
import FormShow from '../pages/form/Show';
import BallotShow from '../pages/ballot/Show';
import NavBar from './NavBar';
import Footer from './Footer';

import './App.css';
import { AuthContext } from '..';
import Logged from 'pages/session/Logged';
import Flash from './Flash';
import ClientError from './ClientError';
import { UserRole } from 'types/userRole';

const App = () => {
  const RequireAuth = ({
    children,
    roles,
  }: {
    children: JSX.Element;
    roles?: string[];
  }): JSX.Element => {
    let location = useLocation();

    const authCtx = useContext(AuthContext);

    if (!authCtx.isLogged) {
      return <Navigate to={ROUTE_LOGIN} state={{ from: location }} replace />;
    } else {
      if (roles && !roles.includes(authCtx.role)) {
        return <Navigate to={ROUTE_UNAUTHORIZED} state={{ from: location }} replace />;
      }
    }

    return children;
  };

  return (
    <Suspense fallback="...loading app">
      <Router>
        <div className="App flex flex-col h-screen justify-between">
          <div>
            <NavBar />
          </div>
          <div
            data-testid="content"
            className=" mb-auto max-w-[80rem] mx-auto flex flex-row justify-center items-center w-full">
            <Routes>
              <Route
                path={ROUTE_FORM_CREATE}
                element={
                  <RequireAuth roles={[UserRole.Admin, UserRole.Operator]}>
                    <FormCreate />
                  </RequireAuth>
                }
              />
              <Route path={'/forms/:formId'} element={<FormShow />} />
              <Route path={'/forms/:formId/result'} element={<FormResult />} />
              <Route
                path={ROUTE_BALLOT_SHOW + '/:formId'}
                element={
                  <RequireAuth roles={null}>
                    <BallotShow />
                  </RequireAuth>
                }
              />
              <Route
                path={ROUTE_ADMIN}
                element={
                  <RequireAuth roles={[UserRole.Admin]}>
                    <Admin />
                  </RequireAuth>
                }
              />
              <Route path={ROUTE_ABOUT} element={<About />} />
              <Route path={ROUTE_FORM_INDEX} element={<FormIndex />} />
              <Route path={ROUTE_LOGIN} element={<Login />} />
              <Route path={ROUTE_LOGGED} element={<Logged />} />
              <Route path={ROUTE_UNAUTHORIZED} element={<ClientError statusCode={403} />} />
              <Route path="/" element={<Home />} />
              <Route path="*" element={<ClientError statusCode={404} />} />
            </Routes>
          </div>
          <div>
            <Flash />
          </div>
          <div>
            <Footer />
          </div>
        </div>
      </Router>
    </Suspense>
  );
};

export default App;
