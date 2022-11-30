import React, { FC, ReactElement, createContext, useEffect, useRef, useState } from 'react';
import ReactDOM from 'react-dom';
import { ENDPOINT_PERSONAL_INFO } from 'components/utils/Endpoints';

import 'index.css';
import App from 'layout/App';
import reportWebVitals from 'reportWebVitals';
import ShortUniqueId from 'short-unique-id';

import * as endpoints from 'components/utils/Endpoints';

const flashTimeout = 4000;

// By default we load the mock messages when not in production. This is handy
// because it removes the need to have a backend server.
if (process.env.NODE_ENV !== 'production' && process.env.REACT_APP_NOMOCK !== 'on') {
  const { dvotingserver } = require('./mocks/dvotingserver');
  dvotingserver.start();
}
const arr = new Map<String, Array<String>>();
const defaultAuth = { isLogged: false, firstname: '', lastname: '', role: '', authorization: arr };

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
  authorization: Map<String, Array<String>>;
}

export interface FlashState {
  getMessages(): FlashMessage[];
  addMessage(msg: string, level: number): void;
  hideMessage(index: string): void;
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
  level: FlashLevel;

  // A uniq string identifier
  id: string;

  constructor(text: string, level: FlashLevel) {
    this.text = text;
    this.level = level;
    this.id = new ShortUniqueId({ length: 8 })();
  }

  getText(): string {
    return this.text;
  }

  getLevel(): FlashLevel {
    return this.level;
  }
}
const flashM = new FlashMessage('', 1);
const defaultFlashState = {
  getMessages: function (): FlashMessage[] {
    return [flashM];
  },
  addMessage: function (msg: string, level: FlashLevel): void {},
  hideMessage: function (index: string): void {},
};
// the flash context handles flash messages across the app
export const FlashContext = createContext<FlashState>(defaultFlashState);

// the proxy state provides the proxy address across all the app
export interface ProxyState {
  getProxy(): string;
  setProxy(p: string): void;
}

export class ProxyHolder implements ProxyState {
  proxy!: string;

  getProxy(): string {
    return this.proxy;
  }

  setProxy(p: string) {
    this.proxy = p;
  }
}

const defaultProxyState = new ProxyHolder();

export const ProxyContext = createContext<ProxyState>(defaultProxyState);

// A small elements to display that the page is loading, should be something
// more elegant in the future and be its own component.
const Loading: FC = () => (
  <div className="flex h-screen">
    <div className="m-auto">
      <div className="text-center pb-2">
        <svg
          role="status"
          className="inline w-16 h-16 mr-2 text-gray-200 animate-spin dark:text-gray-400 fill-indigo-600"
          viewBox="0 0 100 101"
          fill="none"
          xmlns="http://www.w3.org/2000/svg">
          <path
            d="M100 50.5908C100 78.2051 77.6142 100.591 50 100.591C22.3858 100.591 0 78.2051 0 50.5908C0 22.9766 22.3858 0.59082 50 0.59082C77.6142 0.59082 100 22.9766 100 50.5908ZM9.08144 50.5908C9.08144 73.1895 27.4013 91.5094 50 91.5094C72.5987 91.5094 90.9186 73.1895 90.9186 50.5908C90.9186 27.9921 72.5987 9.67226 50 9.67226C27.4013 9.67226 9.08144 27.9921 9.08144 50.5908Z"
            fill="currentColor"
          />
          <path
            d="M93.9676 39.0409C96.393 38.4038 97.8624 35.9116 97.0079 33.5539C95.2932 28.8227 92.871 24.3692 89.8167 20.348C85.8452 15.1192 80.8826 10.7238 75.2124 7.41289C69.5422 4.10194 63.2754 1.94025 56.7698 1.05124C51.7666 0.367541 46.6976 0.446843 41.7345 1.27873C39.2613 1.69328 37.813 4.19778 38.4501 6.62326C39.0873 9.04874 41.5694 10.4717 44.0505 10.1071C47.8511 9.54855 51.7191 9.52689 55.5402 10.0491C60.8642 10.7766 65.9928 12.5457 70.6331 15.2552C75.2735 17.9648 79.3347 21.5619 82.5849 25.841C84.9175 28.9121 86.7997 32.2913 88.1811 35.8758C89.083 38.2158 91.5421 39.6781 93.9676 39.0409Z"
            fill="currentFill"
          />
        </svg>
      </div>
      <p>App is loading...</p>
    </div>
  </div>
);

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
  const [auth, setAuth] = useState<AuthState>(defaultAuth);

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
      const flash = new FlashMessage(message, level);
      const newFlashes = [...flashes, flash];
      setFlashes(newFlashes);

      // remove the flash after some timeout
      setTimeout(() => {
        let removedFlashes = [...flashesRef.current];
        removedFlashes = removedFlashes.filter((f) => f.id !== flash.id);
        setFlashes(removedFlashes);
      }, flashTimeout);
    },

    // Set the visibility of flashMessage to false
    hideMessage: (id: string) => {
      let removedFlashes = [...flashesRef.current];
      removedFlashes = removedFlashes.filter((f) => f.id !== id);
      setFlashes(removedFlashes);
    },
  };

  const setDefaultProxy = async () => {
    let proxy = sessionStorage.getItem('proxy');

    if (proxy === null) {
      const response = await fetch(endpoints.getProxyConfig);
      if (!response.ok) {
        const js = await response.json();
        throw new Error(`Failed to get the default proxy: ${JSON.stringify(js)}`);
      }

      proxy = await response.text();
    }

    defaultProxyState.setProxy(proxy);
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
          authorization: new Map(Object.entries(result.authorization)),
        });

        // wait for the default proxy to be set
        await setDefaultProxy();

        setContent(<App />);
      } catch (e: any) {
        setContent(<Failed>{e.toString()}</Failed>);
        console.log('error:', e);
      }
    }

    fetchData();
  }, []);

  return (
    <FlashContext.Provider value={flashState}>
      <AuthContext.Provider value={auth}>
        <ProxyContext.Provider value={defaultProxyState}>{content}</ProxyContext.Provider>
      </AuthContext.Provider>
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
