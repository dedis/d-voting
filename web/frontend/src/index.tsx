import React, { FC, ReactElement, createContext, useEffect, useState } from 'react';
import ReactDOM from 'react-dom';
import { ENDPOINT_PERSONNAL_INFO } from 'components/utils/Endpoints';

import 'index.css';
import App from 'layout/App';
import reportWebVitals from 'reportWebVitals';

// By default we load the mock messages when not in production. This is handy
// because it removes the need to have a backend server.
if (process.env.NODE_ENV !== 'production' && process.env.REACT_APP_NOMOCK !== 'on') {
  const { dvotingserver } = require('./mocks/dvotingserver');
  dvotingserver.start();
}

const defaultAuth = { isLogged: false, firstname: '', lastname: '', role: '' };

// AuthContext is a global state containing the authentication state. React
// Context is a way to manage state globally, without having to pass props to
// components. This context is set when loading the app by asking the backend if
// the user is logged.
export const AuthContext = createContext<AuthState>(defaultAuth);

export interface AuthState {
  isLogged: boolean;
  firstname: string;
  lastname: string;
  role: string;
}

// A small elements to display that the page is loading, should be something
// more elegant in the future and be its own component.
const Loading: FC = () => <p>App is loading...</p>;

// AppContainer wraps the App with the context. It makes sure that the App is
// displayed only when the AuthContext has been updated.
const AppContainer = () => {
  const [content, setContent] = useState<ReactElement>(<Loading />);
  const [auth, setAuth] = useState<AuthState>(undefined);

  useEffect(() => {
    const req = {
      method: 'GET',
    };
    fetch(ENDPOINT_PERSONNAL_INFO, req)
      .then((res) => res.json())
      .then((result) => {
        setAuth({
          isLogged: result.islogged,
          firstname: result.firstname,
          lastname: result.lastname,
          role: result.role,
        });

        setContent(<App />);
      });
  }, []);

  return <AuthContext.Provider value={auth}>{content}</AuthContext.Provider>;
};

// Main entry point
ReactDOM.render(
  <React.StrictMode>
    <AppContainer />
  </React.StrictMode>,
  document.getElementById('root')
);

// If you want to start measuring performance in your app, pass a function
// to log results (for example: reportWebVitals(console.log))
// or send to an analytics endpoint. Learn more: https://bit.ly/CRA-vitals
reportWebVitals();
