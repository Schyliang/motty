# VNC security negotiation

Motty supports the common RSA-AES and certificate-based VeNCrypt security
schemes below. Server product versions, compiled features, and configuration
determine which schemes a server actually offers.

| Security scheme | RFB type | Authentication | Session encryption |
| --- | --- | --- | --- |
| None | 1 | None | None |
| VNC Authentication | 2 | Traditional VNC password | None |
| RA2 | 5 | Username/password or VNC password | AES-128-EAX |
| RA256 (`RA2_256`) | 129 | Username/password or VNC password | AES-256-EAX |
| RA2ne | 6 | Username/password or VNC password | None; only the authentication exchange is encrypted |
| RA256ne (`RA2ne_256`) | 130 | Username/password or VNC password | None; only the authentication exchange is encrypted |
| VeNCrypt 0.2, X509None | 19 / 260 | None | TLS 1.2 or newer |
| VeNCrypt 0.2, X509Vnc | 19 / 261 | Traditional VNC password | TLS 1.2 or newer |
| VeNCrypt 0.2, X509Plain | 19 / 262 | Username/password | TLS 1.2 or newer |

Full session encryption is preferred: RA256, then RA2, then VeNCrypt. An
identity rejection, cancellation, or cryptographic error ends the attempt; no
weaker authentication is tried after a failure. Legacy-only servers retain
their None/VNC Authentication selection behavior.

## Connecting and trust

For account-based authentication, enter the optional VNC username or answer the
username/password prompt during connection. A deliberately empty username is
allowed for servers such as WayVNC configured without a username. Account-based
passwords are stored separately from traditional VNC passwords, including the
empty-username account.

When no system account was configured and exactly one nonempty system username
is saved for the host, Motty reuses that account and password after verifying
the server's identity. This also keeps an SSH tunnel's account separate from
the VNC system account. Multiple saved accounts are left for the user to choose.

Confirm the RSA key or TLS certificate before sending credentials. The identity
dialog allows rejection, trust for this connection, and trust with saving. A
saved identity change stops the connection, as a changed SSH host key does.
Pins are public fingerprints in the host's configuration record, so they follow
the host through backup, sync, copying and deletion. They remain scoped to the
entered server host and port even when an SSH tunnel dials loopback. VNC RSA and
TLS identities are separate; neither can reuse an SSH or RDP pin.

Duplicate hosts carry compatible RSA/TLS pins and saved system accounts to the
surviving catalog ID. Conflicting identities or different authentication-only
permissions keep separate records; a failed password migration keeps the
original hosts and secrets for retry.

Library callers without a TLS trust callback use normal certificate-chain and
hostname verification. The desktop application supplies an explicit certificate
trust/pinning callback for private or self-signed VNC certificates.

RA2ne and RA256ne require the VNC host's **Allow encrypted login with an
unencrypted desktop** option, off by default. These modes protect authentication
credentials only: desktop, keyboard, pointer, and clipboard messages remain
unencrypted. They never report full session encryption. Through SSH, the outer
tunnel still protects the traffic and the status reports SSH protection.

Traditional VNC Authentication is a DES challenge-response scheme. It does not
encrypt the session. The application warns about unencrypted direct sessions.

## Protocol and cryptographic boundaries

Known RealVNC `RFB 004.001` and `RFB 005.000` banners negotiate to `RFB 003.008`.
RFB 3.3, 3.7, 3.8, and newer 3.x minor versions retain their existing version
selection. Other major versions are rejected. Extended security types require
RFB 3.7 or newer; RFB 3.3 supports only None/VNC Authentication.

RSA-AES exchanges RSAES-PKCS1-v1.5-encrypted random values and binds public keys
with authenticated transcript proofs. RA2 uses 16-byte random values and SHA-1;
RA256 uses 32-byte values and SHA-256. Every attempt generates a 2048-bit client
RSA key and accepts valid 1024–8192-bit server keys. Each EAX record authenticates
its length and ciphertext before making plaintext available. Directions have
separate little-endian 128-bit counters, shared by authentication and session
messages. The `ne` modes return to the original stream after credentials and
reject unexpected encrypted plaintext at that boundary.

Usernames/passwords use UTF-8 byte lengths, each limited to 255 bytes in the
current implementation. Oversized or invalid UTF-8 values are rejected without
truncation. VeNCrypt X509Plain encodes these lengths as 32-bit values.

VeNCrypt version 0.1, bare Plain(256), anonymous TLSNone/TLSVnc/TLSPlain(257–259),
standalone ANONTLS(18), UltraVNC DSM/MS-Logon, Apple DH(30), RealVNC private type
13, and unspecified vendor extensions are not supported. Go's TLS library does
not implement anonymous DH/ECDH suites. A VeNCrypt server offering only anonymous
TLS is rejected; it is not automatically retried using a plaintext alternative.
Configure an X.509 mode for certificate-based TLS interoperability.

A server announcing a newer VeNCrypt version is offered 0.2. The connection
continues only if the server explicitly accepts that version; Motty does not
interpret a newer wire format as 0.2.

Text clipboard sharing retains its existing Latin-1 behavior. UTF-8 clipboard
support belongs to the separate clipboard branch.

## Verification boundaries

Automated coverage includes published AES-128/256-EAX vectors, all four RSA-AES
types and both credential subtypes, encrypted and plaintext session boundaries,
framebuffer traffic, TLS certificate authentication, security preference,
identity rejection/change, cancellation, fragmentation, tampering, replay,
truncation, and concurrent reads/writes. Workbench tests cover explicit
authentication-only permission, empty usernames, separate credential/identity
storage, host copying, and preservation of the source password.

On 2026-11-06, the user tested `motty-vnc-ra2-autologin.exe` against a real
Raspberry Pi. System-account login and desktop display succeeded, and the saved
account/password were reused automatically without a credential dialog.
The user also reported progressive, scan-like desktop updates during window
movement and scrolling. This observation has not been independently reproduced;
no rendering change is included in this authentication work.

On 2026-11-07, independent TigerVNC 1.15 virtual-machine testing passed RA256,
RA2ne and RA256ne with both system-account and VNC-password authentication, and
all three X.509 VeNCrypt subtypes. The Windows desktop backend connected to each
of the nine services three times. Linux tests verified desktop encodings,
pointer/key input, resize, bidirectional Latin-1 clipboard traffic, continuous
records, credential and identity rejection, and default refusal of ne modes.
All three X.509 subtypes also passed forced TLS 1.2 and TLS 1.3 handshakes.

Other RealVNC versions, WayVNC/NeatVNC and x11vnc require separate live checks.
These tests do not establish native application key persistence or restart,
physical keyboard/IME and OS clipboard notification behavior, hours-long
sessions, SSH transport or the revised executable's UI status indicators.

Old WayVNC 0.6.2 with NeatVNC 0.6.x uses VeNCrypt/X509Plain for authenticated
connections. Newer NeatVNC versions can additionally offer RA2/RA256. The actual
installed library and build features must be checked before claiming support
for a particular distribution or server version.

Protocol references: [TigerVNC RSA-AES](https://github.com/TigerVNC/tigervnc/blob/009df04af8fe3843176636da5ea297cf6b7c7a98/common/rfb/CSecurityRSAAES.cxx),
[TigerVNC VeNCrypt](https://github.com/TigerVNC/tigervnc/blob/009df04af8fe3843176636da5ea297cf6b7c7a98/common/rfb/CSecurityVeNCrypt.cxx),
[NeatVNC X509Plain](https://github.com/any1/neatvnc/blob/175676b6f4505c8d51bfac9d1c7b75271f9e43a3/src/auth/vencrypt.c),
and the [EAX specification](https://web.cs.ucdavis.edu/~rogaway/papers/eax.pdf).
