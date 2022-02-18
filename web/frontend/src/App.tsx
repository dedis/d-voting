import React, { FC, Suspense, useEffect, useState } from 'react';
import { BrowserRouter as Router, Routes, Route } from 'react-router-dom';

import About from './pages/About';
import Admin from './pages/Admin';
import ElectionStatus from './pages/ElectionIndex';
import Login from './pages/Login';

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
  const [sciper, setSciper] = useState(0);
  const [role, setRole] = useState('');

  useEffect(() => {
    fetch(GET_PERSONNAL_INFOS)
      .then((res) => res.json())
      .then((result) => {
        setIsLogged(result.islogged);
        setLastName(result.lastname);
        setFirstname(result.firstname);
        setSciper(result.sciper);
        setRole(result.role);
      });
  }, []);

  return (
    <Suspense fallback="...loading app">
      <Router>
        <div className="App flex flex-col h-screen justify-between">
          <div className="app-nav">
            <NavBar
              firstname={firstname}
              lastname={lastname}
              sciper={sciper}
              role={role}
              isLogged={isLogged}
            />
          </div>
          <div
            data-testid="content"
            className="app-page mb-auto flex flex-row justify-center items-center w-full">
            {isLogged ? (
              <div className="p-10 w-full max-w-screen-xl">
                <Routes>
                  <Route path="/" element={<Home />} />
                  {/* <Route path="/elections/:electionId/ballots" element={<Ballots />} /> */}
                  <Route path="/create-election" element={<CreateElection />} />
                  <Route path="/elections" element={<ElectionStatus />} />
                  <Route path="/elections/:electionId" element={<ElectionDetails />} />
                  <Route path="/results" element={<ResultTable />} />
                  <Route path="/results/:electionId" element={<ResultPage />} />
                  <Route path="/vote" element={<BallotsTable />} />
                  <Route path="/vote/:electionId" element={<Ballot />} />
                  <Route path="/about" element={<About />} />
                  <Route path="/admin" element={<Admin />} />
                </Routes>
              </div>
            ) : (
              <div className="login-container">
                <Login />
              </div>
            )}
          </div>
          <Footer />
        </div>
      </Router>
    </Suspense>
  );
};

export default App;
