import React, { Suspense, useContext } from 'react';
import { Navigate, Route, BrowserRouter as Router, Routes, useLocation } from 'react-router-dom';

import {
  ROUTE_ABOUT,
  ROUTE_ADMIN,
  ROUTE_BALLOT_INDEX,
  ROUTE_BALLOT_SHOW,
  ROUTE_ELECTION_CREATE,
  ROUTE_ELECTION_INDEX,
  ROUTE_LOGGED,
  ROUTE_LOGIN,
} from '../Routes';
import Login from '../pages/session/Login';
import Home from '../pages/Home';
import About from '../pages/About';
import Admin from '../pages/Admin';
import ElectionIndex from '../pages/election/Index';
import ElectionCreate from '../pages/election/New';
import ElectionResult from '../pages/election/Result';
import ElectionShow from '../pages/election/Show';
import BallotIndex from '../pages/ballot/Index';
import BallotShow from '../pages/ballot/Show';
import NavBar from './NavBar';
import Footer from './Footer';

import './App.css';
import { AuthContext } from '..';
import Logged from 'pages/session/Logged';

const NotFound = () => <div>404 not found</div>;

const App = () => {
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
              <Route path={'/elections/:electionId'} element={<ElectionShow />} />
              <Route path={'/elections/:electionId/result'} element={<ElectionResult />} />
              <Route path={ROUTE_BALLOT_INDEX} element={<BallotIndex />} />
              <Route
                path={ROUTE_BALLOT_SHOW + '/:electionId'}
                element={
                  <RequireAuth>
                    <BallotShow />
                  </RequireAuth>
                }
              />
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
              <Route path={ROUTE_LOGGED} element={<Logged />} />
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
