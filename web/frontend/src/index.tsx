import React, { FC, ReactElement, createContext, useEffect, useRef, useState } from 'react';
import ReactDOM from 'react-dom';
import { ENDPOINT_PERSONAL_INFO } from 'components/utils/Endpoints';

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

export const enum FlashLevel {
  Info = 1,
  Warning,
  Error,
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

const Failed: FC = ({ children }) => (
  <div className="flex items-center justify-center w-screen h-screen bg-gradient-to-r from-red-600 to-red-700">
    <div className="px-5 py-3 bg-white rounded-md shadow-xl">
      <div className="flex flex-col items-center">
        <div className="p-4">
          <h1 className="text-2xl font-medium text-slate-600 pb-2">Failed to get personal info.</h1>
          <p className="text-sm tracking-tight font-light text-slate-400 leading-6">
            Is the backend running ?
          </p>
          <code className="text-sm tracking-tight font-light text-slate-400 leading-6">
            {children}
          </code>
        </div>
      </div>
    </div>
  </div>
);

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

    async function fetchData() {
      try {
        const res = await fetch(ENDPOINT_PERSONAL_INFO, req);

        if (res.status !== 200) {
          const txt = await res.text();
          throw new Error(`unexpected status: ${res.status} - ${txt}`);
        }

        const result = await res.json();

        setAuth({
          isLogged: result.islogged,
          firstname: result.firstname,
          lastname: result.lastname,
          role: result.role,
        });

        setContent(<App />);
      } catch (e) {
        setContent(<Failed>{e.toString()}</Failed>);
        console.log('error:', e);
      }
    }

    fetchData();
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
