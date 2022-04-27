// Custom hook that post a request to an endpoint
const usePostCall = (setError) => {
  return async (endpoint, request, setIsPosting) => {
    let success = true;
    try {
      const response = await fetch(endpoint, request);
      if (!response.ok) {
        const txt = await response.text();
        throw new Error(txt);
      }
      setError(null);
    } catch (error) {
      setError(error.message);
      success = false;
    }
    setIsPosting((prev) => !prev);
    return success;
  };
};

export default usePostCall;
