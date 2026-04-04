# distributed-sales

This is the implementation of a **RabbitMQ / Microsservices** assignment for our **Distributed Systems** class.

# Running
The RabbitMQ broker can be deployed by running
```bash
docker compose up --build
```
on the root of this repository.

Alternatively, you may use your own broker.
However, the default ports and credentials are expected.
Furthermore, a virtual host called `assignment1` is also expected.

## Cryptographic keys 

In this project, we're using RSA keys only,
and we're following the PKCS8, PKIX and x509 standards for key generation.

Services assume they already have all of the required public keys to verify any valid message.

All signed messages are JSON encoded, with a designated field for the signature, and another for the payload.
The signature is calculated on the SHA-256 hash of the payload.

### Generating keys

#### Generating private keys
```bash
openssl genrsa -out private_key.pem 2048
```

#### Extracting public key from private
```bash
openssl rsa -in private_key.pem -pubout -out public_key.pem
```
