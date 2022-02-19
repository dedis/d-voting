import React, { FC, Suspense, useEffect, useState } from 'react';
import { BrowserRouter as Router, Routes, Route } from 'react-router-dom';

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
} from './components/Routes';
import About from './pages/About';
import Admin from './pages/Admin';
import ElectionStatus from './pages/ElectionIndex';
import { GET_PERSONNAL_INFOS } from './components/utils/ExpressEndoints';
import CreateElection from './components/CreateElection';
import NavBar from './components/NavBar';
import Home from './pages/Home';
import BallotsTable from './components/BallotsTable';
import Ballot from './components/Ballot';
import ResultTable from './components/ResultTable';
import ResultPage from './components/ResultPage';
import Footer from './components/Footer';
import ElectionDetails from './components/ElectionDetails';
import './App.css';

const App: FC = () => {
  const [isLogged, setIsLogged] = useState(false);
  const [lastname, setLastName] = useState('');
  const [firstname, setFirstname] = useState('');
  const [role, setRole] = useState('');

  useEffect(() => {
    fetch(GET_PERSONNAL_INFOS)
      .then((res) => res.json())
      .then((result) => {
        setIsLogged(result.islogged);
        setLastName(result.lastname);
        setFirstname(result.firstname);
        setRole(result.role);
      });
  }, []);

  return (
    <Suspense fallback="...loading app">
      <Router>
        <div className="App flex flex-col h-screen justify-between">
          <div className="app-nav">
            <NavBar firstname={firstname} lastname={lastname} role={role} isLogged={isLogged} />
          </div>
          <div
            data-testid="content"
            className="app-page mb-auto flex flex-row justify-center items-center w-full">
            <div>
              <Routes>
                <Route path={ROUTE_ADMIN} element={<Admin />} />
                <Route path={ROUTE_ABOUT} element={<About />} />
                <Route path={ROUTE_HOME} element={<Home isLogged={isLogged} />} />
                <Route path={ROUTE_ELECTION_INDEX} element={<ElectionStatus />} />
                <Route path={ROUTE_ELECTION_CREATE} element={<CreateElection />} />
                <Route path={ROUTE_ELECTION_SHOW + '/:electionId'} element={<ElectionDetails />} />
                <Route path={ROUTE_RESULT_INDEX} element={<ResultTable />} />
                <Route path={ROUTE_RESULT_SHOW + '/:electionId'} element={<ResultPage />} />
                <Route path={ROUTE_BALLOT_INDEX} element={<BallotsTable />} />
                <Route path={ROUTE_BALLOT_SHOW + '/:electionId'} element={<Ballot />} />
              </Routes>
            </div>
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
