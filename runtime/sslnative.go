package runtime

// The native part of the ssl module, on Go's crypto/tls. A context is a
// tls.Config; a wrapped socket is a tls.Conn over the socket's descriptor,
// read and written with the socket's timeout (retryIO releases the GIL
// while it waits). Errors are returned as tuples
// (class, errno, reason, message, verify_code, verify_message) that ssl.py
// raises as SSLError and its subclasses, spelled as OpenSSL spells them.

import (
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/asn1"
	"encoding/hex"
	"encoding/pem"
	"errors"
	"fmt"
	"io"
	"net"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"syscall"
	"time"
)

type sslContext struct {
	serverSide    bool
	verifyMode    int // 0 CERT_NONE, 1 CERT_OPTIONAL, 2 CERT_REQUIRED
	checkHostname bool
	certs         []tls.Certificate
	roots         *x509.CertPool
	caCerts       []*x509.Certificate
	minVersion    uint16
	maxVersion    uint16
	alpn          []string
}

// fdConn is a socket descriptor as a net.Conn, with the socket's timeout.
type fdConn struct {
	fd      int
	timeout float64 // -1 blocks
}

func (c *fdConn) Read(b []byte) (int, error) {
	var n int
	err := retryIO(c.fd, false, c.timeout, func() (err error) {
		n, err = syscall.Read(c.fd, b)
		return
	})
	if err != nil {
		return 0, err
	}
	if n == 0 && len(b) > 0 {
		return 0, io.EOF
	}
	return n, nil
}

func (c *fdConn) Write(b []byte) (int, error) {
	total := 0
	for total < len(b) {
		var n int
		err := retryIO(c.fd, true, c.timeout, func() (err error) {
			n, err = sendNoPipe(c.fd, b[total:], msgNoSignal)
			return
		})
		if err != nil {
			return total, err
		}
		total += n
	}
	return total, nil
}

func (c *fdConn) Close() error                     { return nil } // the socket closes its descriptor
func (c *fdConn) LocalAddr() net.Addr              { return fdAddr{} }
func (c *fdConn) RemoteAddr() net.Addr             { return fdAddr{} }
func (c *fdConn) SetDeadline(time.Time) error      { return nil }
func (c *fdConn) SetReadDeadline(time.Time) error  { return nil }
func (c *fdConn) SetWriteDeadline(time.Time) error { return nil }

type fdAddr struct{}

func (fdAddr) Network() string { return "tcp" }
func (fdAddr) String() string  { return "" }

type sslConn struct {
	ctx        *sslContext
	raw        *fdConn
	conn       *tls.Conn
	serverSide bool
	hostname   string
	verified   bool // the peer's certificate was verified
}

var (
	sslMu       sync.Mutex
	sslContexts = map[int64]*sslContext{}
	sslConns    = map[int64]*sslConn{}
	sslNextID   int64
)

func sslNewID() int64 {
	sslNextID++
	return sslNextID
}

func sslCtx(v any) *sslContext {
	sslMu.Lock()
	defer sslMu.Unlock()
	return sslContexts[DynInt(v)]
}

func sslConnOf(v any) *sslConn {
	sslMu.Lock()
	defer sslMu.Unlock()
	c := sslConns[DynInt(v)]
	if c == nil {
		panic(valueError("the SSL connection is closed"))
	}
	return c
}

// sslErr is the tuple ssl.py raises.
func sslErr(class string, errno int, reason, msg string) any {
	return NewVarTuple[any](class, int64(errno), reason, msg, nil, nil)
}

func sslVerifyErr(code int, msg string) any {
	return NewVarTuple[any]("SSLCertVerificationError", int64(1), "CERTIFICATE_VERIFY_FAILED", "certificate verify failed: "+msg, int64(code), msg)
}

// alertReasons are OpenSSL's names of the alerts a peer sends.
var alertReasons = map[uint8][2]string{
	10:  {"SSLV3_ALERT_UNEXPECTED_MESSAGE", "ssl/tls alert unexpected message"},
	20:  {"SSLV3_ALERT_BAD_RECORD_MAC", "ssl/tls alert bad record mac"},
	40:  {"SSLV3_ALERT_HANDSHAKE_FAILURE", "ssl/tls alert handshake failure"},
	42:  {"SSLV3_ALERT_BAD_CERTIFICATE", "ssl/tls alert bad certificate"},
	43:  {"SSLV3_ALERT_UNSUPPORTED_CERTIFICATE", "ssl/tls alert unsupported certificate"},
	44:  {"SSLV3_ALERT_CERTIFICATE_REVOKED", "ssl/tls alert certificate revoked"},
	45:  {"SSLV3_ALERT_CERTIFICATE_EXPIRED", "ssl/tls alert certificate expired"},
	46:  {"SSLV3_ALERT_CERTIFICATE_UNKNOWN", "ssl/tls alert certificate unknown"},
	47:  {"SSLV3_ALERT_ILLEGAL_PARAMETER", "ssl/tls alert illegal parameter"},
	48:  {"TLSV1_ALERT_UNKNOWN_CA", "tlsv1 alert unknown ca"},
	49:  {"TLSV1_ALERT_ACCESS_DENIED", "tlsv1 alert access denied"},
	50:  {"TLSV1_ALERT_DECODE_ERROR", "tlsv1 alert decode error"},
	51:  {"TLSV1_ALERT_DECRYPT_ERROR", "tlsv1 alert decrypt error"},
	70:  {"TLSV1_ALERT_PROTOCOL_VERSION", "tlsv1 alert protocol version"},
	71:  {"TLSV1_ALERT_INSUFFICIENT_SECURITY", "tlsv1 alert insufficient security"},
	80:  {"TLSV1_ALERT_INTERNAL_ERROR", "tlsv1 alert internal error"},
	90:  {"SSLV3_ALERT_USER_CANCELLED", "ssl/tls alert user cancelled"},
	112: {"TLSV1_UNRECOGNIZED_NAME", "tlsv1 unrecognized name"},
	116: {"TLSV13_ALERT_CERTIFICATE_REQUIRED", "tlsv13 alert certificate required"},
	120: {"TLSV1_ALERT_NO_APPLICATION_PROTOCOL", "tlsv1 alert no application protocol"},
}

// sslErrorOf turns an error of crypto/tls into the tuple of the SSLError
// CPython would raise. A Python exception (a timeout) is raised as it is.
func (c *sslConn) sslErrorOf(err error) any {
	var pe *PyError
	if errors.As(err, &pe) {
		panic(pe)
	}
	var verr *tls.CertificateVerificationError
	if errors.As(err, &verr) {
		return verifyErrorOf(verr.Err, verr.UnverifiedCertificates, c.hostname)
	}
	var mine *sslVerifyFailure
	if errors.As(err, &mine) {
		return sslVerifyErr(mine.code, mine.msg)
	}
	var alert tls.AlertError
	if errors.As(err, &alert) {
		if r, ok := alertReasons[uint8(alert)]; ok {
			return sslErr("SSLError", 1, r[0], r[1])
		}
		return sslErr("SSLError", 1, "UNKNOWN_ALERT", "unknown alert")
	}
	var rh tls.RecordHeaderError
	if errors.As(err, &rh) {
		return sslErr("SSLError", 1, "WRONG_VERSION_NUMBER", "wrong version number")
	}
	if errors.Is(err, io.EOF) || errors.Is(err, io.ErrUnexpectedEOF) {
		return sslErr("SSLEOFError", 8, "UNEXPECTED_EOF_WHILE_READING", "EOF occurred in violation of protocol")
	}
	var errno syscall.Errno
	if errors.As(err, &errno) {
		panic(sysError(errno))
	}
	msg := err.Error()
	switch {
	case strings.Contains(msg, "protocol version not supported"), strings.Contains(msg, "unsupported versions"):
		return sslErr("SSLError", 1, "UNSUPPORTED_PROTOCOL", "unsupported protocol")
	case strings.Contains(msg, "no application protocol"):
		return sslErr("SSLError", 1, "NO_APPLICATION_PROTOCOL", "no application protocol")
	case strings.Contains(msg, "client didn't provide a certificate"):
		return sslErr("SSLError", 1, "PEER_DID_NOT_RETURN_A_CERTIFICATE", "peer did not return a certificate")
	case strings.Contains(msg, "no cipher suite supported"), strings.Contains(msg, "handshake failure"):
		return sslErr("SSLError", 1, "NO_SHARED_CIPHER", "no shared cipher")
	}
	return sslErr("SSLError", 1, "UNKNOWN", strings.TrimPrefix(msg, "tls: "))
}

// sslVerifyFailure is a verification error found by the checks done here
// (when hostname checking is off, crypto/tls does not verify).
type sslVerifyFailure struct {
	code int
	msg  string
}

func (e *sslVerifyFailure) Error() string { return e.msg }

// verifyErrorOf gives OpenSSL's verify code and message for a failed
// verification.
func verifyErrorOf(err error, chain []*x509.Certificate, hostname string) any {
	var ua x509.UnknownAuthorityError
	var he x509.HostnameError
	var ci x509.CertificateInvalidError
	switch {
	case errors.As(err, &ua):
		if len(chain) > 0 {
			last := chain[len(chain)-1]
			if selfSigned(last) {
				if len(chain) == 1 {
					return sslVerifyErr(18, "self-signed certificate")
				}
				return sslVerifyErr(19, "self-signed certificate in certificate chain")
			}
		}
		return sslVerifyErr(20, "unable to get local issuer certificate")
	case errors.As(err, &he):
		if net.ParseIP(he.Host) != nil {
			return sslVerifyErr(64, "IP address mismatch, certificate is not valid for '"+he.Host+"'.")
		}
		return sslVerifyErr(62, "Hostname mismatch, certificate is not valid for '"+he.Host+"'.")
	case errors.As(err, &ci):
		if ci.Reason == x509.Expired {
			if ci.Cert != nil && time.Now().Before(ci.Cert.NotBefore) {
				return sslVerifyErr(9, "certificate is not yet valid")
			}
			return sslVerifyErr(10, "certificate has expired")
		}
		return sslVerifyErr(26, "unsuitable certificate purpose")
	}
	return sslVerifyErr(1, "unspecified certificate verification error")
}

func selfSigned(c *x509.Certificate) bool {
	return string(c.RawSubject) == string(c.RawIssuer) && c.CheckSignatureFrom(c) == nil
}

// config is the tls.Config of a connection.
func (ctx *sslContext) config(serverSide bool, hostname string) *tls.Config {
	cfg := &tls.Config{
		Certificates: ctx.certs,
		MinVersion:   ctx.minVersion,
		MaxVersion:   ctx.maxVersion,
		NextProtos:   ctx.alpn,
	}
	roots := ctx.roots
	if roots == nil {
		roots = x509.NewCertPool() // nothing loaded: nothing is trusted
	}
	if serverSide {
		cfg.ClientCAs = roots
		switch ctx.verifyMode {
		case 1:
			cfg.ClientAuth = tls.VerifyClientCertIfGiven
		case 2:
			cfg.ClientAuth = tls.RequireAndVerifyClientCert
		}
		return cfg
	}
	if net.ParseIP(hostname) == nil {
		cfg.ServerName = hostname // SNI
	}
	if ctx.verifyMode != 0 && ctx.checkHostname && hostname != "" {
		cfg.RootCAs = roots
		cfg.ServerName = hostname
		return cfg
	}
	cfg.InsecureSkipVerify = true
	if ctx.verifyMode != 0 {
		// The chain is verified, the hostname is not.
		cfg.VerifyConnection = func(cs tls.ConnectionState) error {
			if len(cs.PeerCertificates) == 0 {
				return &sslVerifyFailure{50, "application verification failure"}
			}
			opts := x509.VerifyOptions{Roots: roots, Intermediates: x509.NewCertPool()}
			for _, c := range cs.PeerCertificates[1:] {
				opts.Intermediates.AddCert(c)
			}
			if _, err := cs.PeerCertificates[0].Verify(opts); err != nil {
				t := verifyErrorOf(err, cs.PeerCertificates, hostname).(VarTuple[any])
				return &sslVerifyFailure{int(DynInt(t.At(4))), Str(t.At(5))}
			}
			return nil
		}
	}
	return cfg
}

func (ctx *sslContext) addPEM(data []byte) (int, error) {
	n := 0
	for {
		var block *pem.Block
		block, data = pem.Decode(data)
		if block == nil {
			break
		}
		if block.Type != "CERTIFICATE" {
			continue
		}
		cert, err := x509.ParseCertificate(block.Bytes)
		if err != nil {
			return n, err
		}
		ctx.addCA(cert)
		n++
	}
	return n, nil
}

func (ctx *sslContext) addCA(cert *x509.Certificate) {
	if ctx.roots == nil {
		ctx.roots = x509.NewCertPool()
	}
	ctx.roots.AddCert(cert)
	ctx.caCerts = append(ctx.caCerts, cert)
}

var tlsVersionNames = map[uint16]string{
	tls.VersionTLS10: "TLSv1", tls.VersionTLS11: "TLSv1.1", tls.VersionTLS12: "TLSv1.2", tls.VersionTLS13: "TLSv1.3",
}

// opensslCipherNames are OpenSSL's names of the TLS 1.2 suites Go has; the
// TLS 1.3 names are the same in both.
var opensslCipherNames = map[uint16]string{
	tls.TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256:         "ECDHE-RSA-AES128-GCM-SHA256",
	tls.TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384:         "ECDHE-RSA-AES256-GCM-SHA384",
	tls.TLS_ECDHE_ECDSA_WITH_AES_128_GCM_SHA256:       "ECDHE-ECDSA-AES128-GCM-SHA256",
	tls.TLS_ECDHE_ECDSA_WITH_AES_256_GCM_SHA384:       "ECDHE-ECDSA-AES256-GCM-SHA384",
	tls.TLS_ECDHE_RSA_WITH_CHACHA20_POLY1305_SHA256:   "ECDHE-RSA-CHACHA20-POLY1305",
	tls.TLS_ECDHE_ECDSA_WITH_CHACHA20_POLY1305_SHA256: "ECDHE-ECDSA-CHACHA20-POLY1305",
	tls.TLS_ECDHE_RSA_WITH_AES_128_CBC_SHA:            "ECDHE-RSA-AES128-SHA",
	tls.TLS_ECDHE_RSA_WITH_AES_256_CBC_SHA:            "ECDHE-RSA-AES256-SHA",
	tls.TLS_ECDHE_ECDSA_WITH_AES_128_CBC_SHA:          "ECDHE-ECDSA-AES128-SHA",
	tls.TLS_ECDHE_ECDSA_WITH_AES_256_CBC_SHA:          "ECDHE-ECDSA-AES256-SHA",
	tls.TLS_RSA_WITH_AES_128_GCM_SHA256:               "AES128-GCM-SHA256",
	tls.TLS_RSA_WITH_AES_256_GCM_SHA384:               "AES256-GCM-SHA384",
	tls.TLS_RSA_WITH_AES_128_CBC_SHA:                  "AES128-SHA",
	tls.TLS_RSA_WITH_AES_256_CBC_SHA:                  "AES256-SHA",
}

func cipherBits(name string) int64 {
	switch {
	case strings.Contains(name, "256"), strings.Contains(name, "CHACHA20"):
		return 256
	case strings.Contains(name, "128"):
		return 128
	}
	return 0
}

// The names of the attributes of a certificate's subject and issuer.
var oidNames = map[string]string{
	"2.5.4.3": "commonName", "2.5.4.4": "surname", "2.5.4.5": "serialNumber", "2.5.4.6": "countryName",
	"2.5.4.7": "localityName", "2.5.4.8": "stateOrProvinceName", "2.5.4.9": "streetAddress",
	"2.5.4.10": "organizationName", "2.5.4.11": "organizationalUnitName", "2.5.4.12": "title",
	"2.5.4.15": "businessCategory", "2.5.4.17": "postalCode", "2.5.4.42": "givenName", "2.5.4.43": "initials",
	"2.5.4.46": "dnQualifier", "2.5.4.65": "pseudonym", "1.2.840.113549.1.9.1": "emailAddress",
	"0.9.2342.19200300.100.1.25": "domainComponent", "0.9.2342.19200300.100.1.1": "userId",
	"1.3.6.1.4.1.311.60.2.1.3": "jurisdictionCountryName", "1.3.6.1.4.1.311.60.2.1.2": "jurisdictionStateOrProvinceName",
	"1.3.6.1.4.1.311.60.2.1.1": "jurisdictionLocalityName",
}

func rdnTuple(raw []byte) any {
	var seq pkix.RDNSequence
	if _, err := asn1.Unmarshal(raw, &seq); err != nil {
		return NewVarTuple[any]()
	}
	var rdns []any
	for _, set := range seq {
		var attrs []any
		for _, a := range set {
			name, ok := oidNames[a.Type.String()]
			if !ok {
				name = a.Type.String()
			}
			attrs = append(attrs, NewVarTuple[any](name, fmt.Sprint(a.Value)))
		}
		rdns = append(rdns, NewVarTuple[any](attrs...))
	}
	return NewVarTuple[any](rdns...)
}

// asn1Time is a time as OpenSSL's ASN1_TIME_print writes it.
func asn1Time(t time.Time) string {
	return t.UTC().Format("Jan _2 15:04:05 2006 GMT")
}

// decodeCert is the dictionary getpeercert() gives for a certificate.
func decodeCert(c *x509.Certificate) *Dict[any, any] {
	d := NewDict[any, any]()
	d.Set("subject", rdnTuple(c.RawSubject))
	d.Set("issuer", rdnTuple(c.RawIssuer))
	d.Set("version", int64(c.Version))
	serial := strings.ToUpper(hex.EncodeToString(c.SerialNumber.Bytes()))
	if serial == "" {
		serial = "00"
	}
	d.Set("serialNumber", serial)
	d.Set("notBefore", asn1Time(c.NotBefore))
	d.Set("notAfter", asn1Time(c.NotAfter))
	var san []any
	for _, n := range c.DNSNames {
		san = append(san, NewVarTuple[any]("DNS", n))
	}
	for _, e := range c.EmailAddresses {
		san = append(san, NewVarTuple[any]("email", e))
	}
	for _, ip := range c.IPAddresses {
		s := ip.String()
		if ip.To4() == nil {
			// OpenSSL writes IPv6 addresses in full, upper case.
			b := ip.To16()
			var parts []string
			for i := 0; i < 16; i += 2 {
				parts = append(parts, fmt.Sprintf("%X", int(b[i])<<8|int(b[i+1])))
			}
			s = strings.Join(parts, ":")
		}
		san = append(san, NewVarTuple[any]("IP Address", s))
	}
	for _, u := range c.URIs {
		san = append(san, NewVarTuple[any]("URI", u.String()))
	}
	if len(san) > 0 {
		d.Set("subjectAltName", NewVarTuple[any](san...))
	}
	if len(c.OCSPServer) > 0 {
		d.Set("OCSP", stringTuple(c.OCSPServer))
	}
	if len(c.IssuingCertificateURL) > 0 {
		d.Set("caIssuers", stringTuple(c.IssuingCertificateURL))
	}
	if len(c.CRLDistributionPoints) > 0 {
		d.Set("crlDistributionPoints", stringTuple(c.CRLDistributionPoints))
	}
	return d
}

func stringTuple(xs []string) any {
	items := make([]any, len(xs))
	for i, x := range xs {
		items[i] = x
	}
	return NewVarTuple[any](items...)
}

func init() {
	RegisterNative("_ssl.context_new", func(args []any, kw []KW) any {
		ctx := &sslContext{serverSide: Truth(args[0])}
		if !ctx.serverSide {
			ctx.verifyMode, ctx.checkHostname = 2, true
		}
		sslMu.Lock()
		defer sslMu.Unlock()
		id := sslNewID()
		sslContexts[id] = ctx
		return id
	})
	// _ssl.context_set(ctx, name, value)
	RegisterNative("_ssl.context_set", func(args []any, kw []KW) any {
		ctx := sslCtx(args[0])
		switch Str(args[1]) {
		case "verify_mode":
			ctx.verifyMode = int(DynInt(args[2]))
		case "check_hostname":
			ctx.checkHostname = Truth(args[2])
		case "minimum_version":
			ctx.minVersion = tlsVersionOf(DynInt(args[2]))
		case "maximum_version":
			ctx.maxVersion = tlsVersionOf(DynInt(args[2]))
		case "alpn":
			ctx.alpn = nil
			for c := DynIter(args[2]).Iter(); c.Next(); {
				ctx.alpn = append(ctx.alpn, Str(c.Value()))
			}
		}
		return nil
	})
	// _ssl.load_cert_chain(ctx, certfile, keyfile, password) -> error or None
	RegisterNative("_ssl.load_cert_chain", func(args []any, kw []KW) any {
		ctx := sslCtx(args[0])
		certfile := Str(args[1])
		keyfile := certfile
		if !IsNone(args[2]) {
			keyfile = Str(args[2])
		}
		certPEM, err := os.ReadFile(certfile)
		if err != nil {
			panic(sysError(err, certfile))
		}
		keyPEM, err := os.ReadFile(keyfile)
		if err != nil {
			panic(sysError(err, keyfile))
		}
		cert, err := tls.X509KeyPair(certPEM, keyPEM)
		if err != nil {
			if strings.Contains(err.Error(), "private key does not match") {
				return sslErr("SSLError", 116, "KEY_VALUES_MISMATCH", "key values mismatch")
			}
			if strings.Contains(err.Error(), "PRIVATE KEY") || strings.Contains(err.Error(), "private key") {
				return sslErr("SSLError", 9, "PEM_LIB", "PEM lib")
			}
			return sslErr("SSLError", 9, "PEM_LIB", "PEM lib")
		}
		ctx.certs = append(ctx.certs, cert)
		return nil
	})
	// _ssl.load_verify_locations(ctx, cafile, capath, cadata) -> error or None
	RegisterNative("_ssl.load_verify_locations", func(args []any, kw []KW) any {
		ctx := sslCtx(args[0])
		if !IsNone(args[1]) {
			p := Str(args[1])
			data, err := os.ReadFile(p)
			if err != nil {
				panic(sysError(err, p))
			}
			if n, err := ctx.addPEM(data); err != nil || n == 0 {
				return sslErr("SSLError", 136, "NO_CERTIFICATE_OR_CRL_FOUND", "no certificate or crl found")
			}
		}
		if !IsNone(args[2]) {
			dir := Str(args[2])
			entries, _ := os.ReadDir(dir)
			for _, e := range entries {
				if data, err := os.ReadFile(filepath.Join(dir, e.Name())); err == nil {
					ctx.addPEM(data)
				}
			}
		}
		if !IsNone(args[3]) {
			switch v := BaseValue(Unbox(args[3])).(type) {
			case string:
				if n, err := ctx.addPEM([]byte(v)); err != nil || n == 0 {
					return sslErr("SSLError", 136, "NO_CERTIFICATE_OR_CRL_FOUND", "no certificate or crl found")
				}
			default:
				der := bytesArg(args[3])
				for len(der) > 0 {
					var raw asn1.RawValue
					rest, err := asn1.Unmarshal(der, &raw)
					if err != nil {
						return sslErr("SSLError", 13, "NOT_ENOUGH_DATA", "not enough data")
					}
					cert, err := x509.ParseCertificate(der[:len(der)-len(rest)])
					if err != nil {
						return sslErr("SSLError", 13, "NOT_ENOUGH_DATA", "not enough data")
					}
					ctx.addCA(cert)
					der = rest
				}
			}
		}
		return nil
	})
	RegisterNative("_ssl.load_default_certs", func(args []any, kw []KW) any {
		ctx := sslCtx(args[0])
		pool, err := x509.SystemCertPool()
		if err == nil && pool != nil {
			if ctx.roots == nil {
				ctx.roots = pool
			} else {
				for _, c := range ctx.caCerts {
					pool.AddCert(c)
				}
				ctx.roots = pool
			}
		}
		return nil
	})
	RegisterNative("_ssl.ca_certs", func(args []any, kw []KW) any {
		ctx := sslCtx(args[0])
		out := NewList[any]()
		for _, c := range ctx.caCerts {
			if Truth(args[1]) {
				out.Append(Bytes(string(c.Raw)))
			} else {
				out.Append(decodeCert(c))
			}
		}
		return out
	})
	// _ssl.wrap(ctx, fd, server_side, server_hostname) -> conn
	RegisterNative("_ssl.wrap", func(args []any, kw []KW) any {
		ctx := sslCtx(args[0])
		serverSide := Truth(args[2])
		hostname := ""
		if !IsNone(args[3]) {
			hostname = Str(args[3])
		}
		raw := &fdConn{fd: int(DynInt(args[1])), timeout: -1}
		c := &sslConn{ctx: ctx, raw: raw, serverSide: serverSide, hostname: hostname}
		cfg := ctx.config(serverSide, hostname)
		if serverSide {
			c.conn = tls.Server(raw, cfg)
		} else {
			c.conn = tls.Client(raw, cfg)
		}
		c.verified = ctx.verifyMode != 0
		sslMu.Lock()
		defer sslMu.Unlock()
		id := sslNewID()
		sslConns[id] = c
		return id
	})
	// _ssl.handshake(conn, timeout) -> error or None
	RegisterNative("_ssl.handshake", func(args []any, kw []KW) any {
		c := sslConnOf(args[0])
		c.raw.timeout = timeoutArg(args, 1)
		var err error
		err = c.conn.Handshake()
		if err != nil {
			return c.sslErrorOf(err)
		}
		return nil
	})
	// _ssl.read(conn, n, timeout, suppress_ragged_eofs) -> (bytes, error)
	RegisterNative("_ssl.read", func(args []any, kw []KW) any {
		c := sslConnOf(args[0])
		n := int(DynInt(args[1]))
		c.raw.timeout = timeoutArg(args, 2)
		buf := make([]byte, n)
		var got int
		var err error
		got, err = c.conn.Read(buf)
		if err != nil && got == 0 {
			if errors.Is(err, io.EOF) {
				return NewVarTuple[any](Bytes(""), nil) // close_notify
			}
			if errors.Is(err, io.ErrUnexpectedEOF) && Truth(args[3]) {
				return NewVarTuple[any](Bytes(""), nil)
			}
			return NewVarTuple[any](nil, c.sslErrorOf(err))
		}
		return NewVarTuple[any](Bytes(string(buf[:got])), nil)
	})
	// _ssl.write(conn, data, timeout) -> (n, error)
	RegisterNative("_ssl.write", func(args []any, kw []KW) any {
		c := sslConnOf(args[0])
		data := bytesArg(args[1])
		c.raw.timeout = timeoutArg(args, 2)
		var n int
		var err error
		n, err = c.conn.Write(data)
		if err != nil {
			return NewVarTuple[any](nil, c.sslErrorOf(err))
		}
		return NewVarTuple[any](int64(n), nil)
	})
	// _ssl.peercert(conn, binary) -> dict, bytes or None
	RegisterNative("_ssl.peercert", func(args []any, kw []KW) any {
		c := sslConnOf(args[0])
		cs := c.conn.ConnectionState()
		if !cs.HandshakeComplete {
			panic(valueError("handshake not done yet"))
		}
		if len(cs.PeerCertificates) == 0 {
			return nil
		}
		cert := cs.PeerCertificates[0]
		if Truth(args[1]) {
			return Bytes(string(cert.Raw))
		}
		if !c.verified {
			return NewDict[any, any]()
		}
		return decodeCert(cert)
	})
	// _ssl.info(conn) -> (version, cipher, bits, alpn) or None before the handshake
	RegisterNative("_ssl.info", func(args []any, kw []KW) any {
		c := sslConnOf(args[0])
		cs := c.conn.ConnectionState()
		if !cs.HandshakeComplete {
			return nil
		}
		name := tls.CipherSuiteName(cs.CipherSuite)
		if n, ok := opensslCipherNames[cs.CipherSuite]; ok {
			name = n
		}
		var alpn any
		if cs.NegotiatedProtocol != "" {
			alpn = cs.NegotiatedProtocol
		}
		return NewVarTuple[any](tlsVersionNames[cs.Version], name, cipherBits(name), alpn)
	})
	// _ssl.unwrap(conn, timeout): sends close_notify.
	RegisterNative("_ssl.unwrap", func(args []any, kw []KW) any {
		c := sslConnOf(args[0])
		c.raw.timeout = timeoutArg(args, 1)
		var err error
		err = c.conn.CloseWrite()
		if err != nil && !errors.Is(err, net.ErrClosed) {
			return c.sslErrorOf(err)
		}
		return nil
	})
	RegisterNative("_ssl.free", func(args []any, kw []KW) any {
		sslMu.Lock()
		defer sslMu.Unlock()
		delete(sslConns, DynInt(args[0]))
		return nil
	})
	// _ssl.decode_cert(path) -> dict: ssl._ssl._test_decode_cert.
	RegisterNative("_ssl.decode_cert", func(args []any, kw []KW) any {
		p := Str(args[0])
		data, err := os.ReadFile(p)
		if err != nil {
			panic(sysError(err, p))
		}
		block, _ := pem.Decode(data)
		if block == nil {
			panic(newError("SSLError", "Error decoding PEM-encoded file"))
		}
		cert, err := x509.ParseCertificate(block.Bytes)
		if err != nil {
			panic(newError("SSLError", "Error decoding PEM-encoded file"))
		}
		return decodeCert(cert)
	})
}

// tlsVersionOf maps ssl.TLSVersion values to crypto/tls versions; -2 and -1
// (MINIMUM_SUPPORTED, MAXIMUM_SUPPORTED) leave the default.
func tlsVersionOf(v int64) uint16 {
	switch v {
	case 0x0301:
		return tls.VersionTLS10
	case 0x0302:
		return tls.VersionTLS11
	case 0x0303:
		return tls.VersionTLS12
	case 0x0304:
		return tls.VersionTLS13
	}
	return 0
}
