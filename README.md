# Security of Computer Systems – Project

## Emulating Environment with Trusted Third Party and Client-Server Data Exchange Scenario

This project implements a simulated secure communication environment based on a **Trusted Third Party** architecture. The system demonstrates how a client and a target server can establish mutual trust, authenticate each other, receive a shared session key and then exchange encrypted application data.

The project was created as part of the **Security of Computer Systems** course and focuses on the practical implementation of cryptographic mechanisms used in distributed systems.

## Project Overview

The system consists of three main components:

- **Trusted Third Party**
- **Client**
- **Target Server**

The Trusted Third Party acts as a central trust authority. It registers entities, issues X.509 certificates, validates identities and distributes session keys. After successful authentication, the client and the server communicate directly using symmetric encryption.

The implementation includes:

- RSA key generation
- X.509 certificate issuing
- entity registration in the TTP
- identity protection with RSA-OAEP and SHA-256
- digital signatures
- certificate validation
- public-key comparison against registration data
- session-key generation and encrypted delivery
- AES-GCM encrypted client-server communication
- documentation generated with Doxygen, GoDoc and Go-Markdoc

## Architecture

```text
+----------+                      +-----------------------+
|  Client  |                      |     Target Server     |
+----------+                      +-----------------------+
     |                                      |
     |          registration                |
     |------------------------------------->|
     |                                      |
     v                                      v
+--------------------------------------------------------+
|                  Trusted Third Party                   |
|                                                        |
|  - stores registered entities                          |
|  - issues X.509 certificates                           |
|  - validates public keys and certificates              |
|  - generates and distributes session keys              |
+--------------------------------------------------------+
     |                                      |
     |        encrypted session key         |
     |------------------------------------->|
     |                                      |
     v                                      v

After authentication:

+----------+        AES-GCM encrypted data        +-----------------------+
|  Client  | <----------------------------------> |     Target Server     |
+----------+                                     +-----------------------+