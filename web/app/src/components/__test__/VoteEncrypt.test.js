/*global require, describe, Buffer, expect, test, global*/
/*eslint no-undef: "error"*/

import kyber from "@dedis/kyber";
global.TextEncoder = require("util").TextEncoder; //https://github.com/heineiuo/rippledb/issues/148w
const encrypt = require('../voting/VoteEncrypt');

describe("Encryption tests", () => {

    
    const edCurve = kyber.curve.newCurve("edwards25519")
    /*Generate random private/public key pair to test the encryption and decryption
    In reality the public dkg key is retrieved by the app but the private key
    is not accessible and is actually a distributed private key */
    const privateKey = edCurve.scalar().pick();
    const publicKey = edCurve.point().mul(privateKey,null).marshalBinary();

    const vote = 'Tom';

    const decryptVote = (ephemeralKey, encryptedVote, privateKey) => {
        const ephKeyUnmarsh = edCurve.point();
        ephKeyUnmarsh.unmarshalBinary(ephemeralKey)
        const encryptedVoteUnmarsh = edCurve.point();
        encryptedVoteUnmarsh.unmarshalBinary(encryptedVote);
        const S = edCurve.point().mul(privateKey, ephKeyUnmarsh);
        const M = edCurve.point().sub(encryptedVoteUnmarsh,S);
        return M.data().toString();
    }
    

    test('correctly decrypt string', () => {
        const [ephemeralKey, encryptedVote] = encrypt.encryptVote(vote, publicKey, edCurve);
        expect(decryptVote(ephemeralKey, encryptedVote, privateKey)).toBe(vote);
    });

    test('embed plain point is not equal of encrypted cipher point', () => {
        const enc = new TextEncoder();
        const voteByte = enc.encode(vote); //vote as []byte  
        const voteBuff = Buffer.from(voteByte.buffer);
        const M = edCurve.point().embed(voteBuff); 
        const [, encryptedVote] = encrypt.encryptVote(vote, publicKey, edCurve);
        const encryptedVoteUnmarsh = edCurve.point();
        encryptedVoteUnmarsh.unmarshalBinary(encryptedVote);
        expect(M.equals(encryptedVoteUnmarsh)).toBe(false);
    })
})