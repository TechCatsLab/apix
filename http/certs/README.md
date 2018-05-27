- Private Key:

```sh
openssl genrsa -out server.key 2048
```

- Public Key:

```sh
openssl req -new -x509 -key server.key -out server.crt -days 3650
```

