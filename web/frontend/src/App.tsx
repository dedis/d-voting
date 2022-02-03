import React, { FC, Suspense, useState } from "react";
import { BrowserRouter as Router, Routes, Route } from "react-router-dom";

import "./App.css";
import "./language/Lang";
import { GET_PERSONNAL_INFOS } from "./components/utils/ExpressEndoints";
import CreateElection from "./components/election/CreateElection";
import Election from "./components/election/ElectionStatus";
import NavBar from "./components/NavBar";
import Home from "./components/Home";
import BallotsTable from "./components/election/BallotsTable";
import Ballot from "./components/election/Ballot";
import ResultTable from "./components/election/ResultTable";
import ResultPage from "./components/election/ResultPage";
import About from "./components/About";
import Admin from "./components/Admin";
import Footer from "./components/Footer";
import ElectionDetails from "./components/election/ElectionDetails";
import Login from "./components/Login";

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
      <div className="App flex flex-col h-screen justify-between">
        <Router>
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
                <Login id="login-id" />
              </div>
            ) : (
              <div className="p-10 w-full max-w-screen-xl">
                <Routes>
                  <Route path="/" component={Home} />
                  <Route path="/create-election" component={CreateElection} />
                  <Route path="/elections" component={Election} />
                  <Route path="/elections/:id" component={ElectionDetails} />
                  <Route path="/vote" component={BallotsTable} />
                  <Route path="/results" component={ResultTable} />
                  <Route path="/results/:id" component={ResultPage} />
                  <Route path="/vote/:id" component={Ballot} />
                  <Route path="/about" component={About} />
                  <Route path="/admin" component={Admin} />
                  <Route path="/404" component={NotFoundPage} />
                  <Redirect to="/404" />
                </Routes>
              </div>
            )}
          </div>
          <Footer />
        </Router>
      </div>
    </Suspense>
  );
};

export default App;
