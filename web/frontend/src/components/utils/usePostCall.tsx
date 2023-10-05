import pollTransaction from 'pages/form/components/utils/TransactionPoll';
import { checkTransaction } from './Endpoints';

// Custom hook that posts a request to an endpoint
const usePostCall = (setError) => {
  return async (endpoint, request, setIsPosting) => {
    let success = true;
    const response = await fetch(endpoint, request);

    if (!response.ok) {
      const txt = await response.text();
      setError(new Error(txt));
      success = false;
      return success;
    }

    try {
      const result = await response.json();
      if (result.Token) {
        pollTransaction(checkTransaction, result.Token, 1000, 30).then(
          () => {
            setIsPosting((prev) => !prev);
          },
          (err) => {
            setError(err.message);
            success = false;
          }
        );
      }
    } catch (e) {
      // Too bad, didn't find a token.
    }

    if (success) setError(null);
    setIsPosting((prev) => !prev);
    return success;
  };
};

export default usePostCall;
