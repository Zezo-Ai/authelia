---
title: "Server"
description: "Configuring the Server Settings."
summary: "Authelia runs an internal web server. This section describes how to configure and tune this."
date: 2024-03-14T06:00:14+11:00
draft: false
images: []
weight: 199200
toc: true
aliases:
  - /c/server
  - /docs/configuration/server.html
seo:
  title: "" # custom title (optional)
  description: "" # custom description (recommended)
  canonical: "" # custom canonical URL (optional)
  noindex: false # false (default) or true
---

## Configuration

{{< config-alert-example >}}

```yaml {title="configuration.yml"}
server:
  address: 'tcp://:{{< sitevar name="port" nojs="9091" >}}/'
  disable_healthcheck: false
  tls:
    key: ''
    certificate: ''
    client_certificates: []
  headers:
    csp_template: ''
  buffers:
    read: 4096
    write: 4096
  timeouts:
    read: '6s'
    write: '6s'
    idle: '30s'
  endpoints:
    enable_pprof: false
    enable_expvars: false
    authz: {} ## See the dedicated "Server Authz Endpoints" configuration guide.
    rate_limits: {} ## See the dedicated "Server Endpoint Rate Limits" configuration guide.
```

## Options

### address

{{< confkey type="string" syntax="address" default="tcp://:9091/" required="no" >}}

{{< callout context="danger" title="Important Notes" icon="outline/alert-octagon" >}}
The [Proxy Integration](../../integration/proxies/introduction.md#important-notes) documentation has important notes on
this option for when integrating it with a proxy.
{{< /callout >}}

Configures the listener address for the Main HTTP Server. The address itself is a listener and the scheme must be the
`unix` scheme, the `fd` scheme, or one of the `tcp` schemes. It can configure the host, port, and path the listener
responds to.

To configure the path for a `unix` or `fd` scheme see the address syntax documentation linked above.

If the path is configured to anything other than `/` requests will be handled for both `/` and the configured path.
For example if configured to `tcp://:{{< sitevar name="port" nojs="9091" >}}/authelia` then requests will be handled for both the `/` and `/authelia/`
path.

#### Examples

```yaml {title="configuration.yml"}
server:
  address: 'tcp://127.0.0.1:{{< sitevar name="port" nojs="9091" >}}/'
```

```yaml {title="configuration.yml"}
server:
  address: 'tcp://127.0.0.1:{{< sitevar name="port" nojs="9091" >}}/subpath'
```

```yaml {title="configuration.yml"}
server:
  address: 'unix:///var/run/authelia.sock'
```

```yaml
# When running "systemd-socket-activate -l 9091 go run ./cmd/authelia", the connections to port 9091 will be forwaded to file descriptor 3.
server:
  address: fd://:3
```

### asset_path

{{< confkey type="string" required="no" >}}

Authelia by default serves all static assets from an embedded file system in the Go binary.

Modifying this setting will allow you to override and serve specific assets for Authelia from a specified path. All
assets that can be overridden must be placed in the `asset_path`. The structure of this directory and the assets which
can be overridden is documented in the
[Sever Asset Overrides Reference Guide](../../reference/guides/server-asset-overrides.md).

### disable_healthcheck

{{< confkey type="boolean" default="false" required="no" >}}

On startup Authelia checks for the existence of /app/healthcheck.sh and /app/.healthcheck.env and if both of these exist
it writes the configuration vars for the healthcheck to the /app/.healthcheck.env file. In instances where this is not
desirable, it's possible to disable these interactions entirely.

An example situation where this is the case is in Kubernetes when set security policies that prevent writing to the
ephemeral storage of a container or just don't want to enable the internal health check.

### tls

Authelia typically listens for plain unencrypted connections. This is by design as most environments allow to
security on lower areas of the OSI model. However it required, if you specify both the [tls key](#key) and
[tls certificate](#certificate) options, Authelia will listen for TLS connections.

The key must be generated by the administrator and can be done by following the
[Generating an RSA Self Signed Certificate](../../reference/guides/generating-secure-values.md#generating-an-rsa-self-signed-certificate)
guide provided a self-signed certificate is fit for purpose. If a self-signed certificate is fit for purpose is beyond
the scope of the documentation and if it is not fit for purpose we instead recommend generating a certificate signing
request or obtaining a certificate signed by one of the many ACME certificate providers. Methods to achieve this are
beyond the scope of this guide.

#### key

{{< confkey type="string" required="situational" >}}

The path to the private key for TLS connections. Must be in DER base64/PEM format and must be encoded per the [PKCS#8],
[PKCS#1], or [SECG1] specifications.

[PKCS#8]: https://datatracker.ietf.org/doc/html/rfc5208
[PKCS#1]: https://datatracker.ietf.org/doc/html/rfc8017
[SECG1]: https://datatracker.ietf.org/doc/html/rfc5915

#### certificate

{{< confkey type="string" required="situational" >}}

The path to the public certificate for TLS connections. Must be in DER base64/PEM format.

#### client_certificates

{{< confkey type="list(string)" required="situational" >}}

The list of file paths to certificates used for authenticating clients. Those certificates can be root
or intermediate certificates. If no item is provided mutual TLS is disabled.

### headers

#### csp_template

{{< confkey type="string" required="no" >}}

This customizes the value of the Content-Security-Policy header. It will replace all instances of the below placeholder
with the nonce value of the Authelia react bundle. This is an advanced option to customize, and you should do
sufficient research about how browsers utilize and understand this header before attempting to customize it.

{{< csp >}}

### buffers

{{< confkey type="structure" structure="server-buffers" required="no" >}}

Configures the server buffers.

### timeouts

{{< confkey type="structure" structure="server-timeouts" required="no" >}}

Configures the server timeouts.

### endpoints

#### enable_pprof

{{< confkey type="boolean" default="false" required="no" >}}

{{< callout context="danger" title="Security Note" icon="outline/alert-octagon" >}}
This is a developer endpoint. __DO NOT__ enable it unless you know why you're enabling it.
__DO NOT__ enable this in production.
{{< /callout >}}

Enables the go [pprof](https://pkg.go.dev/net/http/pprof) endpoints.

#### enable_expvars

{{< confkey type="boolean" default="false" required="no" >}}

{{< callout context="danger" title="Security Note" icon="outline/alert-octagon" >}}
This is a developer endpoint. __DO NOT__ enable it unless you know why you're enabling it.
__DO NOT__ enable this in production.
{{< /callout >}}

Enables the go [expvar](https://pkg.go.dev/expvar) endpoints.

#### authz

This is an *__advanced__* option allowing configuration of the authorization endpoints and has its own section.
Generally this does not need to be configured for most use cases. See the
[Server Authz Endpoints](./server-endpoints-authz.md) configuration guide for more information.

#### rate_limits

This is an *__advanced__* option allowing configuration of the endpoint rate limits and has its own section.
Generally this does not need to be configured for most use cases. See the
[Server Endpoint Rate Limits](./server-endpoint-rate-limits.md) configuration guide for more information.

## Additional Notes

### Buffer Sizes

The read and write buffer sizes generally should be the same. This is because when Authelia verifies
if the user is authorized to visit a URL, it also sends back nearly the same size response as the request. However,
you're able to tune these individually depending on your needs.

### Asset Overrides

If replacing the Logo for your Authelia portal, it is recommended to upload a transparent PNG of your desired logo.
Authelia will automatically resize the logo to an appropriate size to present in the frontend.
