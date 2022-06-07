import { CheckIcon, PencilIcon, RefreshIcon } from '@heroicons/react/outline';
import { FlashContext, FlashLevel, ProxyContext } from 'index';
import { ChangeEvent, FC, createRef, useContext, useEffect, useState } from 'react';
import * as endpoints from './Endpoints';

const proxyKey = 'proxy';

const ProxyInput: FC = () => {
  const fctx = useContext(FlashContext);
  const pctx = useContext(ProxyContext);

  const [proxy, setProxy] = useState<string>(pctx.getProxy());
  const [inputVal, setInputVal] = useState('');
  const [inputChanging, setInputChanging] = useState(false);
  const [inputWidth, setInputWidth] = useState(0);

  const proxyTextRef = createRef<HTMLDivElement>();

  const fetchFromBackend = async () => {
    try {
      const response = await fetch(endpoints.getProxyConfig);
      if (!response.ok) {
        const js = await response.json();
        throw new Error(JSON.stringify(js));
      } else {
        setProxy(await response.text());
      }
    } catch (e) {
      fctx.addMessage(`Failed to get proxy: ${proxy}`, FlashLevel.Error);
    }
  };

  // update the proxy context and sessionStore each time the proxy changes
  useEffect(() => {
    sessionStorage.setItem(proxyKey, proxy);
    pctx.setProxy(proxy);
    setInputVal(proxy);
  }, [proxy]);

  // function called by the "refresh" button
  const getDefault = () => {
    fetchFromBackend();
    fctx.addMessage('Proxy updated to default', FlashLevel.Info);
  };

  const updateProxy = () => {
    try {
      new URL(inputVal);
    } catch {
      fctx.addMessage('invalid URL', FlashLevel.Error);
      return;
    }

    setInputChanging(false);
    setProxy(inputVal);
  };

  const editProxy = () => {
    setInputWidth(proxyTextRef.current.clientWidth);
    setInputChanging(true);
  };

  return (
    <div className="flex flex-row items-center">
      {inputChanging ? (
        <>
          <input
            value={inputVal}
            onChange={(e: ChangeEvent<HTMLInputElement>) => setInputVal(e.target.value)}
            className="mt-1 ml-3 border rounded-md p-2"
            style={{ width: `${inputWidth + 3}px` }}
          />
          <div className="ml-1">
            <button className={`border p-1 rounded-md }`} onClick={updateProxy}>
              <CheckIcon className="h-5 w-5" aria-hidden="true" />
            </button>
          </div>
        </>
      ) : (
        <>
          <div
            ref={proxyTextRef}
            className="mt-1 ml-3 border border-transparent p-2"
            onClick={editProxy}>
            {inputVal}
          </div>
          <div className="">
            <button className="hover:text-indigo-500 p-1 rounded-md" onClick={editProxy}>
              <PencilIcon className="m-1 h-3 w-3" aria-hidden="true" />
            </button>
          </div>
        </>
      )}
      <button
        onClick={getDefault}
        className="flex flex-row items-center hover:text-indigo-500 p-1 rounded-md">
        get default
        <RefreshIcon className="m-1 h-3 w-3" aria-hidden="true" />
      </button>
    </div>
  );
};

export default ProxyInput;
