# Connect the Operator to an SSL-Enabled FE HOWTO

The operator talks to FE over the MySQL protocol to run maintenance statements such as
`SHOW COMPUTE NODES`, `ALTER SYSTEM DROP COMPUTE NODE`, and `DROP WAREHOUSE`. Before this feature the
operator could only open a **plaintext** connection. If FE was configured with
`ssl_force_secure_transport = true` in `fe.conf`, it rejected every one of these connections with
error 5205 (`Connections using insecure transport are prohibited`) and logged a warning each time.

This document describes the `--fe-ssl-mode` flag (Helm value `starrocksOperator.feSslMode`) that lets
the operator negotiate TLS for this connection, and how to choose the right value for your cluster.

## 1. How It Works

`--fe-ssl-mode` accepts three values, matched case-insensitively:

| Value | Behavior |
|---|---|
| `DISABLED` | Always plaintext. An FE with `ssl_force_secure_transport = true` rejects the connection with error 5205. |
| `PREFERRED` (default) | Encrypt when FE advertises SSL support; stay plaintext when it does not. A cluster without SSL configured on FE behaves exactly as before this feature. |
| `REQUIRED` | Always encrypt; fail the connection when FE does not support SSL. |

`VERIFY_CA` and `VERIFY_IDENTITY` are recognized names — they exist in the MySQL client's
`--ssl-mode` vocabulary that this flag mirrors — but the operator refuses to start if either is
configured; see [Certificate Verification](#2-certificate-verification) for why. An unrecognized or
empty value also stops the operator at startup: a typo in this flag should be loud, not silently
fall back to plaintext.

## 2. Certificate Verification

**The server certificate is never verified, in any mode**, including `REQUIRED`. FE's own
documentation for enabling SSL has you generate a self-signed keystore, for example:

```bash
keytool -genkeypair -alias starrocks -keyalg RSA -keysize 1024 -validity 365
```

A self-signed certificate has no CA to chain to, and its certificate name never matches the
in-cluster service DNS name the operator dials to reach FE. Verifying either one would fail
against the exact keystore FE's documentation tells you to create.

Because of this, `PREFERRED` and `REQUIRED` protect the connection from **passive eavesdropping
only**. They do not protect against an **active man-in-the-middle**, since the operator never
confirms it is actually talking to your FE. Do not treat `REQUIRED` as a strong security guarantee;
treat it as "always encrypted, certificate unchecked."

## 3. Choosing a Mode

- **Most users should change nothing.** The default, `PREFERRED`, already handles both cases: it
  encrypts against an FE with SSL enabled and stays plaintext against an FE without it.
- **Use `REQUIRED`** if you need a guarantee that the operator never talks to FE in plaintext — for
  example, a compliance posture that does not accept `PREFERRED`'s silent fallback to plaintext when
  FE does not advertise SSL.
- **Fall back to `DISABLED`** if FE advertises SSL but the TLS handshake cannot be negotiated (for
  example, a TLS version or cipher mismatch between the operator's MySQL driver and FE's keystore).
  `PREFERRED` does **not** fall back to plaintext in that situation — it only falls back when FE does
  not advertise SSL at all, so a handshake failure with `PREFERRED` still fails the connection.
  `DISABLED` is the escape hatch while you fix the handshake, or if you decide not to run FE with SSL.

## 4. FE-Side Settings

These `fe.conf` settings on the FE side determine whether FE advertises and accepts SSL, and are
what `--fe-ssl-mode` reacts to:

- `ssl_keystore_location` — path to the keystore file.
- `ssl_keystore_password` — password for the keystore.
- `ssl_key_password` — password for the private key inside the keystore.
- `ssl_force_secure_transport` — when `true`, FE rejects plaintext connections outright (error 5205).

Refer to FE's own documentation for how to generate the keystore and set these values. Two traps to
watch for while doing so:

- **`ssl_key_password` must equal `ssl_keystore_password`.** JDK 9+ defaults `keytool` to the PKCS12
  keystore format, which does not support a separate per-key password: `-storepass A -keypass B`
  silently keeps the private key's password as `A` and only prints a warning. FE still starts and
  logs the three `ssl_*` settings as if healthy, but cannot read its private key, and the operator in
  `REQUIRED` mode fails with no obvious cause.
- **Verify enforcement through the FE Service DNS name, not loopback.** FE exempts `127.0.0.1`
  connections from `ssl_force_secure_transport`, so `kubectl exec ... -- mysql -h127.0.0.1` still
  succeeds in plaintext even with enforcement on. Test from the FE's Service DNS name or another pod.

## 5. Configure via Helm

Set `starrocksOperator.feSslMode` when installing or upgrading the operator (or the `kube-starrocks`
umbrella chart, where the same key lives under `operator.starrocksOperator.feSslMode`):

```bash
helm upgrade starrocks-operator starrocks/operator \
  --reuse-values \
  --set starrocksOperator.feSslMode=REQUIRED
```

Or set it in `values.yaml`:

```yaml
starrocksOperator:
  feSslMode: REQUIRED
```

If you installed the `kube-starrocks` umbrella chart instead of the standalone `operator` chart, the
same value lives one level deeper, under the `operator` subchart:

```bash
helm upgrade starrocks starrocks/kube-starrocks \
  --reuse-values \
  --set operator.starrocksOperator.feSslMode=REQUIRED
```

## 6. Configure via Raw Manifest

If you deploy the operator from `deploy/operator.yaml` instead of Helm, edit the `--fe-ssl-mode`
argument that already ships in the manager container's `args`:

```yaml
        args:
        - --leader-elect
        - --zap-time-encoding=iso8601
        - --zap-encoder=console
        - --volume-name-with-hash=true
        - --fe-ssl-mode=REQUIRED
```

Then re-apply the manifest:

```bash
kubectl apply -f deploy/operator.yaml
```

## 7. Restart Behavior

`--fe-ssl-mode` is read once at process start, so changing it (through either path above) **restarts
the operator pod**. It does **not** restart FE, BE, or CN pods — those keep running unaffected, and
the new mode only takes effect for the operator's own FE connections after its pod comes back up.
