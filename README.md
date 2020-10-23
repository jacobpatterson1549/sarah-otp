# sarah-otp

[![Go Report Card](https://goreportcard.com/badge/github.com/jacobpatterson1549/sarah-otp)](https://goreportcard.com/report/github.com/jacobpatterson1549/sahah-otp)


## A One-Time-Pad messaging app

Sarah-OTP allows users to exchange messages confidentially. After users share a key, a message can be passed between the users with "perfect" secrecy. This "perfect" secrecy is accomplished using a [One-Time-Pad](https://en.m.wikipedia.org/wiki/One-time_pad) (OTP). The OTP combines a key with a message to create a cipher that can only be decrypted by combining it with the key. The key must be at least as long as the cipher to ensure all of the message can be encrypted. 

### Example

A substition cipher can be used with the OTP, If a message of `CAT` and a cipher of `APPLE` are encrypted together, the encrypted cipher would be `CPCLE`. Assuming only characters can be passed, the message (`CAT`), would be mapped to [2, 0, 19, 0, 0] and the cipher (`APPLE`), would be mapped to [0, 15, 15, 11, 4]. Note that the message is padded with zeroes to make it as long as the key. The the letters are added together and truncated to be between 0 and 25.  For example, `T` + `P` = `19` + `15` = `34`, which has a remainder of `8` when divided by 26, so the letter `I` is used.
```
  2 0  19  0 0 ( C A T A A )
+ 0 15 15 11 4 ( A P P L E )
--------------
= 2 15  8 11 4 ( C P I L E )
```
When the cipher is decrypted, the same procedure is used, but with subtraction, to reverse the encryption. Negative values are incremented by 26 to make them correspond to real letters. For examlpe, `I` - `P` = `8` - `15` = `-7`, -`-7` + `26` = `19` = `T`
```
  2 15  8 11 4 ( C P I L E )
- 0 15 15 11 4 ( A P P L E )
--------------
  2  0 19  0 0 ( C A T A A )
```

Sarah-OTP uses the [exclusive-or](https://en.wikipedia.org/wiki/Exclusive_or) operation rather than the substitution cipher to encrypt each letter as a byte between 0-255. This operation is similar to the that of the substitution cipher, but requires no special logic to reverse because bitwise (base 2) addition is performed on each byte.

### Usage

1. Create a key.
1. Share the key with the other user over safe channel.
1. Use the key to encrypt a message to create a cipher.
1. Pass the the cipher to the other user over a potentially compromised channel.
1. The other user decrypts the cipher with the copy of the key to reveal the message.
1. Discard the key.

### Safety Considerations

* The key is randomized, but it still created using the browser's random number generator, which is not truly random. This means that the key is not technically secure. 
* Do not use a key to encrypt multiple messages. If an adversary obtains multiple messages encrypted with the same key, he will be able to determine what the key is.
* Keep the key secret until it is used. Destroy it afterwards.
* No warranty is provided for Sarah-OTP, use at your own risk. See the [LICENSE](LICENSE) page.

## Build/Run

### HTTPS

The app requires HTTP TLS (HTTPS) to run. Insecure http requests are redirected to https.

#### localhost

Use [mkcert](https://github.com/FiloSottile/mkcert) to configure a development machine to accept local certificates.
```bash
go get github.com/FiloSottile/mkcert
mkcert -install
```
Generate certificates for localhost at 127.0.0.1
```bash
mkcert 127.0.0.1
```
Then, add the certificate files to the run environment configuration in `.env`.  The certificate files should be in the root of the application, but are aliased to be up a directory since the server runs in the build folder when running locally. 
```
TLS_CERT_FILE=../127.0.0.1.pem
TLS_KEY_FILE=../127.0.0.1-key.pem
```

### Server Ports

By default, the server will run on ports 80 and 443 for http and https traffic.  All http traffic is redirected to https.  To override the ports, use the HTTP_PORT and HTTPS_PORT flags.

If the server handles HTTPS by providing its own certificate, use the `PORT` variable to specify the https port.  When PORT is defined, no HTTP server will be started from `HTTP_PORT` and certificates are not read from the `TLS_CERT_FILE` and `TLS_KEY_FILE` flags.

##### Local Default TCP HTTP Ports

Run `make serve-tcp` to run on port 80 for HTTP and port 443 for HTTPS (default TCP ports).  Using these ports requires `sudo` (root) access.

### Make

The [Makefile](Makefile) runs the application locally.  This requires Go and a Postgres database to be installed.  [Node](https://github.com/nodejs) is needed to run WebAssembly tests.  Run `make serve` to build and run the application.

### Docker

Launching the application with [Docker](https://www.docker.com) requires minimal configuration.

1. Install [docker-compose](https://github.com/docker/compose)
1. Ensure the files for the `TLS_CERT_FILE` and `TLS_KEY_FILE` environment variables are **NOT** aliased relative to the build folder.  Instead, they should be aliased relative to them by the project folder.  Ideally, refer to them by their absolute paths.
1. Run `docker-compose up` to launch the application.
1. Access application by opening <http://localhost:8000>.