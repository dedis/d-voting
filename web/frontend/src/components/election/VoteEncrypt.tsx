export function encryptVote(vote, dkgKey, edCurve){
     
    //embed the vote into a curve point
    const enc = new TextEncoder();
    const voteByte = enc.encode(vote); //vote as []byte  
    const voteBuff = Buffer.from(voteByte.buffer);
    const M = edCurve.point().embed(voteBuff); 

   
    //dkg public key as a point on the EC 
    const keyBuff = dkgKey;
    const p = edCurve.point();
    p.unmarshalBinary(keyBuff); //unmarshall dkg public key
    const pubKeyPoint = p.clone(); //get the point corresponding to the dkg public key

    const k = edCurve.scalar().pick();  //ephemeral private key
    const K = edCurve.point().mul(k, null); // ephemeral DH public key
    
    const S = edCurve.point().mul(k, pubKeyPoint); //ephemeral DH shared secret
    const C = S.add(S,M); //message blinded with secret
    
    //(K,C) are what we'll send to the backend
    return [K.marshalBinary(),C.marshalBinary()];
}