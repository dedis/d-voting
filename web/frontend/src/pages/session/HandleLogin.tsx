import { ENDPOINT_GET_TEQ_KEY } from 'components/utils/Endpoints';
import { FlashLevel, FlashState } from 'index';

// The backend will provide the client the URL to make a Tequila authentication.
// We therefore redirect to this address.
const handleLogin = async (fctx: FlashState) => {
  try {
    const res = await fetch(ENDPOINT_GET_TEQ_KEY);

    if (res.status !== 200) {
      const txt = await res.text();
      throw new Error(`unexpected status: ${res.status} - ${txt}`);
    }

    const json = await res.json();
    window.location = json.url;
  } catch (error) {
    fctx.addMessage(error.toString(), FlashLevel.Error);
  }
};

export default handleLogin;
