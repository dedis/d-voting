export async function fetchCall(endpoint, request, setData, setLoading) {
  let err;
  try {
    const response = await fetch(endpoint, request);
    if (!response.ok) {
      err = new Error(await response.text());
    } else {
      let dataReceived = await response.json();
      setData(dataReceived);
      setLoading(false);
    }
  } catch (e) {
    err = e;
  }
  if (err !== undefined) {
    throw err;
  }
}
