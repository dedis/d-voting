import pollTransaction from 'pages/form/components/utils/TransactionPoll';
import { checkTransaction } from './Endpoints';

// Custom hook that post a request to an endpoint
const usePostCall = (setError) => {
  return async (endpoint, request, setIsPosting) => {
    let success = true;
    const response = await fetch(endpoint, request);
    const result = await response.json();
    console.log('result:', result);

    if (!response.ok) {
      const txt = await response.text();
      setError(new Error(txt));
      success = false;
      return success;
    }

    if (result.Token) {
      pollTransaction(checkTransaction, result.Token, 1000, 30).then(
        () => {
          setIsPosting((prev) => !prev);
          console.log('Transaction included');
        },
        (err) => {
          console.log('Transaction rejected');
          setError(err.message);
          success = false;
        }
      );
    }

    if (success) setError(null);
    setIsPosting((prev) => !prev);
    return success;
  };
};

export default usePostCall;
