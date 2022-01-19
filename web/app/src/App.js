import React from 'react';
import { useState } from 'react';
import './App.css';
import CreateElection from './components/election-creation/CreateElection';
import Election from './components/election-status/Election';
import NavBar from './components/navigation/NavBar';
import Home from './components/homepage/Home';
import BallotsTable from './components/voting/BallotsTable';
import Ballot from './components/voting/Ballot';
import ResultTable from './components/result-page/ResultTable';
import ResultPage from './components/result-page/ResultPage';
import About from './components/about/About';
import Admin from './components/admin/Admin';
import Footer from './components/footer/Footer';
import { BrowserRouter as Router, Switch, Route } from 'react-router-dom';
import ElectionDetails from './components/election-status/ElectionDetails';
import { LanguageContext } from './components/language/LanguageContext';
import Login from './components/login/Login';
import {GET_PERSONNAL_INFOS} from './components/utils/ExpressEndoints';


const App = () => {

  const getBrowserLanguage = () => {
    var userLang = navigator.userLanguage || navigator.language;
    if (userLang.substring(0, 2) === 'fr') {
      return 'fr'
    }
    return 'en';
  }

  //language state
  const [lanContext, setLanContext] =  useState(getBrowserLanguage());

  const [isLogged, setIsLogged] = useState(undefined);
  const [name, setName] = useState('');
  const [firstname, setFirstname] = useState('');
  const [sciper, setSciper] = useState(0);
  const [role, setRole] = useState('')

  fetch(GET_PERSONNAL_INFOS)
      .then(res => res.json())
      .then((result) => {
        setIsLogged(result.islogged);
        setName(result.name);
        setFirstname(result.firstname);
        setSciper(result.sciper);
        setRole(result.role);
      });
  return (
    <div className="App flex flex-col h-screen justify-between">

      <Router>
        <LanguageContext.Provider value={[lanContext, setLanContext]}>
          <div className='app-nav'>
            <NavBar firstname={firstname} name={name} sciper={sciper} role={role}/>
          </div>
          <div data-testid="content" className='app-page mb-auto flex flex-row justify-center items-center w-full'>
          {!isLogged? (<div className='login-container'><Login id='login-id'/></div>): (<div className="p-10 w-full max-w-screen-xl">
            <Switch>
              <Route path="/" exact component={Home}/>
              <Route path="/create-election" component={CreateElection}/>
              <Route path="/elections" exact component={Election}/>
              <Route path="/elections/:id" component={ElectionDetails}/>
              <Route path="/vote" exact component={BallotsTable}/>
              <Route path="/results" exact component={ResultTable}/>
              <Route path="/results/:id" exact component={ResultPage}/>
              <Route path = "/vote/:id" component = {Ballot}/>
              <Route path="/about" component={About}/>
              <Route path="/admin" component={Admin}/>
            </Switch>
            </div>)}
          </div>
          <Footer />
        </LanguageContext.Provider>
      </Router>
    </div>
  );
}

export default App;
