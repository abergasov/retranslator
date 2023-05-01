package executor

import (
	"context"
	"crypto/tls"
	"fmt"
	"net"
	"net/http"

	cloudflarebp "github.com/DaRealFreak/cloudflare-bp-go"
	"github.com/jellydator/ttlcache/v3"
	utls "github.com/refraction-networking/utls"
	"golang.org/x/net/http2"
)

func (s *Service) getRequest(ctx context.Context, targetURL, cookie string) (*http.Request, error) {
	key := targetURL + cookie
	item := s.cacheRequests.Get(key)
	if item != nil {
		return item.Value().WithContext(ctx), nil
	}
	req, err := http.NewRequest(http.MethodGet, targetURL, http.NoBody)
	if err != nil {
		return &http.Request{}, fmt.Errorf("failed to get url: %w", err)
	}
	s.cacheRequests.Set(key, req, ttlcache.DefaultTTL)
	return req.WithContext(ctx), err
}

func (s *Service) getClient(targetURL, cookie string) *http.Client {
	key := fmt.Sprintf("%s-%s", targetURL, cookie)
	s.clientsMU.RLock()
	client, ok := s.clients[key]
	s.clientsMU.RUnlock()
	if ok {
		return client
	}
	s.clientsMU.Lock()
	defer s.clientsMU.Unlock()

	http2Transport := &http2.Transport{
		DialTLSContext:     utlsDial,
		DisableCompression: false,
		AllowHTTP:          false,
	}
	httpTransport := &http.Transport{
		DialTLSContext: func(ctx context.Context, network, addr string) (net.Conn, error) {
			// Note that hardcoding the address is not necessary here. Only
			// do that if you want to ignore the DNS lookup that already
			// happened behind the scenes.

			tcpConn, err := (&net.Dialer{}).DialContext(ctx, network, addr)
			if err != nil {
				return nil, err
			}
			config := utls.Config{
				ServerName: addr[:len(addr)-4],
			}
			//tlsConn := utls.UClient(tcpConn, &config, utls.HelloCustom)
			tlsConn := utls.UClient(tcpConn, &config, utls.HelloChrome_102)
			//clientHelloSpec := utls.ClientHelloSpec{
			//	CipherSuites: []uint16{
			//		utls.GREASE_PLACEHOLDER,
			//		utls.TLS_AES_128_GCM_SHA256,
			//		utls.TLS_AES_256_GCM_SHA384,
			//		utls.TLS_CHACHA20_POLY1305_SHA256,
			//		utls.TLS_ECDHE_ECDSA_WITH_AES_128_GCM_SHA256,
			//		utls.TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256,
			//		utls.TLS_ECDHE_ECDSA_WITH_AES_256_GCM_SHA384,
			//		utls.TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384,
			//		utls.TLS_ECDHE_ECDSA_WITH_CHACHA20_POLY1305,
			//		utls.TLS_ECDHE_RSA_WITH_CHACHA20_POLY1305,
			//		utls.TLS_ECDHE_RSA_WITH_AES_128_CBC_SHA,
			//		utls.TLS_ECDHE_RSA_WITH_AES_256_CBC_SHA,
			//		utls.TLS_RSA_WITH_AES_128_GCM_SHA256,
			//		utls.TLS_RSA_WITH_AES_256_GCM_SHA384,
			//		utls.TLS_RSA_WITH_AES_128_CBC_SHA,
			//		utls.TLS_RSA_WITH_AES_256_CBC_SHA,
			//	},
			//	CompressionMethods: []byte{
			//		0x00, // compressionNone
			//	},
			//	Extensions: []utls.TLSExtension{
			//		&utls.UtlsGREASEExtension{},
			//		&utls.SNIExtension{},
			//		&utls.UtlsExtendedMasterSecretExtension{},
			//		&utls.RenegotiationInfoExtension{Renegotiation: utls.RenegotiateOnceAsClient},
			//		&utls.SupportedCurvesExtension{Curves: []utls.CurveID{
			//			utls.CurveID(utls.GREASE_PLACEHOLDER),
			//			utls.X25519,
			//			utls.CurveP256,
			//			utls.CurveP384,
			//		}},
			//		&utls.SupportedPointsExtension{SupportedPoints: []byte{
			//			0x00, // pointFormatUncompressed
			//		}},
			//		&utls.SessionTicketExtension{},
			//		&utls.ALPNExtension{AlpnProtocols: []string{"h2", "http/1.1"}},
			//		&utls.StatusRequestExtension{},
			//		&utls.SignatureAlgorithmsExtension{SupportedSignatureAlgorithms: []utls.SignatureScheme{
			//			utls.ECDSAWithP256AndSHA256,
			//			utls.PSSWithSHA256,
			//			utls.PKCS1WithSHA256,
			//			utls.ECDSAWithP384AndSHA384,
			//			utls.PSSWithSHA384,
			//			utls.PKCS1WithSHA384,
			//			utls.PSSWithSHA512,
			//			utls.PKCS1WithSHA512,
			//		}},
			//		&utls.SCTExtension{},
			//		&utls.KeyShareExtension{KeyShares: []utls.KeyShare{
			//			{Group: utls.CurveID(utls.GREASE_PLACEHOLDER), Data: []byte{0}},
			//			{Group: utls.X25519},
			//		}},
			//		&utls.PSKKeyExchangeModesExtension{Modes: []uint8{
			//			utls.PskModeDHE,
			//		}},
			//		&utls.SupportedVersionsExtension{Versions: []uint16{
			//			utls.GREASE_PLACEHOLDER,
			//			utls.VersionTLS13,
			//			utls.VersionTLS12,
			//		}},
			//		&utls.UtlsCompressCertExtension{},
			//		//&utls.GenericExtension{Id: 0x4469}, // WARNING: UNKNOWN EXTENSION, USE AT YOUR OWN RISK
			//		&utls.UtlsGREASEExtension{},
			//		&utls.UtlsPaddingExtension{GetPaddingLen: utls.BoringPaddingStyle},
			//	},
			//}
			//if err = tlsConn.ApplyPreset(&clientHelloSpec); err != nil {
			//	return nil, fmt.Errorf("uTlsConn.ApplyPreset() error: %w", err)
			//}

			if err = tlsConn.Handshake(); err != nil {
				return nil, fmt.Errorf("uTlsConn.Handshake() error: %w", err)
			}

			return tlsConn, nil
		},
	}
	//httpTransport := &http.Transport{
	//	TLSClientConfig: &tls.Config{
	//		CurvePreferences: []tls.CurveID{tls.CurveP256, tls.CurveP384, tls.CurveP521, tls.X25519},
	//		CipherSuites: []uint16{
	//			tls.TLS_AES_128_GCM_SHA256,
	//			tls.TLS_AES_256_GCM_SHA384,
	//			tls.TLS_CHACHA20_POLY1305_SHA256,
	//			tls.TLS_ECDHE_ECDSA_WITH_AES_128_GCM_SHA256,
	//			tls.TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256,
	//			tls.TLS_ECDHE_ECDSA_WITH_AES_256_GCM_SHA384,
	//			tls.TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384,
	//			tls.TLS_ECDHE_ECDSA_WITH_CHACHA20_POLY1305_SHA256,
	//			tls.TLS_ECDHE_RSA_WITH_CHACHA20_POLY1305_SHA256,
	//			tls.TLS_ECDHE_RSA_WITH_AES_128_CBC_SHA,
	//			tls.TLS_ECDHE_RSA_WITH_AES_256_CBC_SHA,
	//			tls.TLS_RSA_WITH_AES_128_GCM_SHA256,
	//			tls.TLS_RSA_WITH_AES_256_GCM_SHA384,
	//			tls.TLS_RSA_WITH_AES_128_CBC_SHA,
	//			tls.TLS_RSA_WITH_AES_256_CBC_SHA,
	//		},
	//	},
	//	ForceAttemptHTTP2:   true,
	//	MaxIdleConns:        30,
	//	MaxIdleConnsPerHost: 30,
	//	IdleConnTimeout:     90 * time.Second,
	//}
	//println(httpTransport.TLSClientConfig.CipherSuites)
	if http2Transport.AllowHTTP {

	}
	if httpTransport.ForceAttemptHTTP2 {

	}
	client = &http.Client{
		//Transport: http2Transport,
	}
	client.Transport = cloudflarebp.AddCloudFlareByPass(client.Transport)
	s.clients[key] = client
	return client
}

func utlsDial(_ context.Context, _, addr string, _ *tls.Config) (net.Conn, error) {
	tcpConn, err := net.Dial("tcp4", addr)
	if err != nil {
		return nil, err
	}

	tlsConn := utls.UClient(tcpConn, &utls.Config{
		ServerName: addr[:len(addr)-4],
		//}, utls.HelloChrome_102)
	}, utls.HelloIOS_13)

	err = tlsConn.Handshake()
	if err != nil {
		return nil, err
	}

	return tlsConn, nil
}
