import { ENDPOINT_DEV_LOGIN, ENDPOINT_GET_TEQ_KEY } from 'components/utils/Endpoints';
import { FlashLevel, FlashState } from 'index';

// The backend will provide the client the URL to make a Tequila authentication.
// We therefore redirect to this address.
// If REACT_APP_DEV_LOGIN === "true", we allow an automatic login with SCIPER 100100.
const handleLogin = async (fctx: FlashState) => {
  try {
    let res;
    if (process.env.REACT_APP_DEV_LOGIN === 'true') {
      await fetch(ENDPOINT_DEV_LOGIN);
      window.location.reload();
      return;
    } else {
      res = await fetch(ENDPOINT_GET_TEQ_KEY);
    }

    const d = new Date();
    d.setTime(d.getTime() + 120000);
    let expires = d.toUTCString();
    document.cookie = `redirect=${window.location.pathname}; expires=${expires}; path=/`;

    if (res.status !== 200) {
      const txt = await res.text();
      throw new Error(`unexpected status: ${res.status} - ${txt}`);
    }

    const json = await res.json();
    window.location = json.url;
  } catch (error: any) {
    fctx.addMessage(error.toString(), FlashLevel.Error);
  }
};

export default handleLogin;
