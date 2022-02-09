import React, { FC, Suspense, useState } from "react";
import { BrowserRouter as Router, Routes, Route } from "react-router-dom";

import { GET_PERSONNAL_INFOS } from "./components/utils/ExpressEndoints";
import CreateElection from "./components/election/CreateElection";
import Election from "./components/election/ElectionStatus";
import NavBar from "./components/NavBar";
import Home from "./pages/Home";
import BallotsTable from "./components/election/BallotsTable";
import Ballot from "./components/election/Ballot";
import ResultTable from "./components/election/ResultTable";
import ResultPage from "./components/election/ResultPage";
import About from "./components/About";
import Admin from "./components/Admin";
import Footer from "./components/Footer";
import ElectionDetails from "./components/election/ElectionDetails";
import Login from "./pages/Login";
import "./styles/App.css";

const App: FC = () => {
  const [isLogged, setIsLogged] = useState(undefined);
  const [lastname, setLastName] = useState("");
  const [firstname, setFirstname] = useState("");
  const [sciper, setSciper] = useState(0);
  const [role, setRole] = useState("");

  fetch(GET_PERSONNAL_INFOS)
    .then((res) => res.json())
    .then((result) => {
      setIsLogged(result.islogged);
      setLastName(result.lastname);
      setFirstname(result.firstname);
      setSciper(result.sciper);
      setRole(result.role);
    });

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
            />
          </div>
          <div
            data-testid="content"
            className="app-page mb-auto flex flex-row justify-center items-center w-full"
          >
            {!isLogged ? (
              <div className="login-container">
                <Login />
              </div>
            ) : (
              <div className="p-10 w-full max-w-screen-xl">
                <Routes>
                  <Route path="/" element={<Home />} />
                  <Route path="/create-election" element={<CreateElection />} />
                  <Route path="/elections" element={<Election />} />
                  <Route
                    path="/elections/:electionId"
                    element={<ElectionDetails />}
                  />
                  <Route path="/results" element={<ResultTable />} />
                  <Route path="/results/:electionId" element={<ResultPage />} />
                  <Route path="/vote" element={<BallotsTable />} />
                  <Route path="/vote/:electionId" element={<Ballot />} />
                  <Route path="/about" element={<About />} />
                  <Route path="/admin" element={<Admin />} />
                </Routes>
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
