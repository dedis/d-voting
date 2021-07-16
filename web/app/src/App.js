import React from 'react';
import {useState} from 'react';
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
import Footer from './components/footer/Footer';
import {BrowserRouter as Router, Switch, Route} from 'react-router-dom';
import ElectionDetails from './components/election-status/ElectionDetails';
import {LanguageContext} from './components/language/LanguageContext';
import Login from './components/login/Login';
import useToken from './components/utils/useToken';


const App = () => {
  
  const getBrowserLanguage = () => {
    var userLang = navigator.userLanguage || navigator.language; 
    if(userLang.substring(0,2) === 'fr'){
      return 'fr'
    }
    return 'en';
  }

  //language state
  const [lanContext, setLanContext] =  useState(getBrowserLanguage());
  const {token, saveToken} = useToken();

  return (
    <div className="App">
      
     <Router>
        <LanguageContext.Provider value={[lanContext, setLanContext]}>         
          <div className='app-nav'>
            <Route path='/:page' component={NavBar} />        
            <Route exact path='/' component={NavBar}/>
          </div>
          <div data-testid="content" className='app-page'>
          {!token? (<div className='login-container'><Login id='login-id' setToken={saveToken}/></div>): (<div>
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
            </Switch>
            </div>)}
          </div>
          <div className='footer-container'>
            <Footer/>
          </div>
        </LanguageContext.Provider>
    </Router>
  </div>
  );
}

export default App;
