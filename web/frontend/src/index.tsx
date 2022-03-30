import React, { FC, ReactElement, createContext, useEffect, useRef, useState } from 'react';
import ReactDOM from 'react-dom';
import { ENDPOINT_PERSONNAL_INFO } from 'components/utils/Endpoints';

import 'index.css';
import App from 'layout/App';
import reportWebVitals from 'reportWebVitals';
import Flash from 'layout/Flash';

const flashTimeout = 4000;

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

export interface FlashState {
  getMessages(): FlashMessage[];
  addMessage(msg: string, level: number): void;
}

// FlashMessage defines the structure of a flash.
class FlashMessage {
  text: string;

  // Level defines the type of flash: info, warn, error
  level: number;

  constructor(text: string, level: number) {
    this.text = text;
    this.level = level;
  }

  getText(): string {
    return this.text;
  }

  getLevel(): number {
    return this.level;
  }
}

// the flash context handles flash messages across the app
export const FlashContext = createContext<FlashState>(undefined);

// A small elements to display that the page is loading, should be something
// more elegant in the future and be its own component.
const Loading: FC = () => <p>App is loading...</p>;

// AppContainer wraps the App with the "auth" and "flash" contexts. It makes
// sure that the App is displayed only when the AuthContext has been updated,
// and displays flash messages.
const AppContainer = () => {
  const [content, setContent] = useState<ReactElement>(<Loading />);
  const [auth, setAuth] = useState<AuthState>(undefined);

  const [flashes, setFlashes] = useState<FlashMessage[]>([]);

  // subtle react thing so the flashes can be used in setTimeout:
  // https://github.com/facebook/react/issues/14010
  const flashesRef = useRef(flashes);
  flashesRef.current = flashes;

  // flashState implements FlashStates. It wraps the necessary functions to be
  // passed to the FlashContext.
  const flashState = {
    getMessages: (): FlashMessage[] => {
      return flashes;
    },

    // add a flash to the list and set a timeout on it
    addMessage: (message: string, level: number) => {
      const newFlashes = [...flashes, new FlashMessage(message, level)];
      setFlashes(newFlashes);

      // remove the flash after some timeout
      setTimeout(() => {
        const removedFlashes = [...flashesRef.current];
        removedFlashes.shift();
        setFlashes(removedFlashes);
      }, flashTimeout);
    },
  };

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
      })
      .catch((e) => {
        flashState.addMessage('failed to get personal info: ' + e, 1);
      });
  }, []);

  return (
    <FlashContext.Provider value={flashState}>
      <Flash />
      <AuthContext.Provider value={auth}>{content}</AuthContext.Provider>
    </FlashContext.Provider>
  );
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
