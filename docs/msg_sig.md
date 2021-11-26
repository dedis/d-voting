# Message signature

API messages must be signed by the proxy server, which every Dela node must
trust. We assume the proxy server owns a key pair `secret_key`/`public_key`.
Dela nodes have the `public_key`. We are using the common "[Magic
signature](https://web.archive.org/web/20210418211626/https://www.abstractioneer.org/2010/01/magic-signatures-for-salmon.html)"
scheme to sign messages.

Let's say the proxy wants to send a message to one of the Dela node:

```json
json := {
    "foo": "bar",
    ...
}
```

First, the message is going to be `base64url` encoded:

```
encoded := base64url_encode(json)
```

Then, a signature on the encoded hash is to be created:

```
signature := sign(secret_key, sha256(encoded))
``` 

Finally, a json message with the encoded original message and the signature can
be sent to the Dela node:

```json
message := {
    "payload": encoded,
    "signature": signature
}
```

Upon receiving the message, a Dela node is going to verify the signature:

```
ok := verify_signature(public_key, message.signature, sha256(message.payload))
```

Lastly, the Dela node can decode the original json message, which has been
authenticated, and process it:

```
json := base64url_decode(message.payload)
=> {
    "foof": "bar",
    ...
}
```

In order to prevent replay attack a secure channel such as TLS over HTTP must be
used to exchange messages between the proxy and the Dela nodes.