import React, { FC, Suspense, useEffect, useState, createContext, useContext } from 'react';
import { BrowserRouter as Router, Routes, Route, Navigate, useLocation } from 'react-router-dom';

import {
  ROUTE_HOME,
  ROUTE_ABOUT,
  ROUTE_ADMIN,
  ROUTE_ELECTION_CREATE,
  ROUTE_ELECTION_INDEX,
  ROUTE_ELECTION_SHOW,
  ROUTE_RESULT_INDEX,
  ROUTE_RESULT_SHOW,
  ROUTE_BALLOT_INDEX,
  ROUTE_BALLOT_SHOW,
  ROUTE_LOGIN,
} from '../Routes';
import Login from '../pages/Login';
import LoginCallback from '../pages/LoginCallback';
import Home from '../pages/Home';
import About from '../pages/About';
import Admin from '../pages/Admin';
import ElectionIndex from '../pages/ElectionIndex';
import ElectionCreate from '../pages/ElectionCreate';
import ElectionShow from '../pages/ElectionShow';
import ResultIndex from '../pages/ResultIndex';
import ResultShow from '../pages/ResultShow';
import BallotIndex from '../pages/BallotIndex';
import BallotShow from '../pages/BallotShow';
import NavBar from './NavBar';
import Footer from './Footer';

import './App.css';
import { AuthContext } from '..';

const NotFound = () => <div>404 not found</div>;

const App: FC = () => {
  const RequireAuth = ({ children }) => {
    let location = useLocation();

    const authCtx = useContext(AuthContext);

    if (!authCtx.isLogged) {
      return <Navigate to="/login" state={{ from: location }} replace />;
    }

    return children;
  };

  return (
    <Suspense fallback="...loading app">
      <Router>
        <div className="App flex flex-col h-screen justify-between">
          <div className="app-nav">
            <NavBar />
          </div>
          <div
            data-testid="content"
            className="app-page mb-auto flex flex-row justify-center items-center w-full">
            <Routes>
              <Route
                path={ROUTE_ELECTION_CREATE}
                element={
                  <RequireAuth>
                    <ElectionCreate />
                  </RequireAuth>
                }
              />
              <Route path={ROUTE_ELECTION_SHOW + '/:electionId'} element={<ElectionShow />} />
              <Route path={ROUTE_RESULT_INDEX} element={<ResultIndex />} />
              <Route path={ROUTE_RESULT_SHOW + '/:electionId'} element={<ResultShow />} />
              <Route path={ROUTE_BALLOT_INDEX} element={<BallotIndex />} />
              <Route path={ROUTE_BALLOT_SHOW + '/:electionId'} element={<BallotShow />} />
              <Route
                path={ROUTE_ADMIN}
                element={
                  <RequireAuth>
                    <Admin />
                  </RequireAuth>
                }
              />
              <Route path={ROUTE_ABOUT} element={<About />} />
              <Route path={ROUTE_ELECTION_INDEX} element={<ElectionIndex />} />
              <Route path={ROUTE_LOGIN} element={<Login />} />
              <Route path="/" element={<Home />} />
              <Route path="*" element={<NotFound />} />
            </Routes>
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
