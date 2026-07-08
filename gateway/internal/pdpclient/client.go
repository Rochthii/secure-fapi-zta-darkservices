// Package pdpclient cung cấp gRPC client để giao tiếp với standalone-policy-engine (PDP).
//
// Kiến trúc PEP → PDP:
//   - Gateway (PEP) nhận HTTP request → xác thực DPoP/mTLS → gọi PDP.CheckAccess()
//   - PDP (standalone-policy-engine) đánh giá in-memory Trie + AST → trả ALLOW/DENY
//   - Gateway thực thi quyết định: nếu DENY → 403 Forbidden, nếu PDP lỗi → 503 fail-closed
//
// Transport hỗ trợ:
//   - Insecure TCP (local dev, khi PDP_TLS_CERT rỗng)
//   - mTLS (production nội bộ)
//   - Phase 2: OpenZiti dark channel (khi PDP_VIA_ZITI=true)
package pdpclient

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"time"

	policyv1 "gateway/internal/pdpclient/pb"

	zititransport "github.com/openziti/sdk-golang/ziti"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/keepalive"
)

const (
	// defaultRequestTimeout là timeout tối đa cho mỗi gRPC CheckAccess call.
	// Phải nhỏ hơn timeout của PDP server (100ms) để Gateway kiểm soát được.
	defaultRequestTimeout = 80 * time.Millisecond

	// defaultDialTimeout là timeout khi thiết lập kết nối ban đầu.
	defaultDialTimeout = 5 * time.Second
)

// PDPClient quản lý kết nối gRPC tới Policy Decision Point.
// Thread-safe: có thể dùng chung từ nhiều goroutine đồng thời.
type PDPClient struct {
	conn   *grpc.ClientConn
	client policyv1.PolicyDecisionPointClient

	// requestTimeout giới hạn thời gian chờ mỗi request CheckAccess.
	requestTimeout time.Duration

	// failOpen nếu true cho phép tất cả request đi qua khi PDP không phản hồi.
	// CẢNH BÁO: chỉ dùng trong môi trường dev. Production phải để false (fail-closed).
	failOpen bool

	// zCtx lưu trữ OpenZiti context để đóng giải phóng tài nguyên khi client đóng kết nối.
	zCtx zititransport.Context
}

// Config chứa cấu hình khởi tạo PDPClient.
type Config struct {
	// Addr là địa chỉ của PDP gRPC server, ví dụ: "localhost:50051" hoặc "policy-engine:50051"
	Addr string

	// TLSCertFile, TLSKeyFile, TLSCAFile là đường dẫn cert cho mTLS client.
	// Để rỗng cả 3 để dùng insecure (chỉ dev local).
	TLSCertFile string
	TLSKeyFile  string
	TLSCAFile   string

	// RequestTimeout ghi đè defaultRequestTimeout nếu khác 0.
	RequestTimeout time.Duration

	// FailOpen nếu true cho phép request khi PDP không khả dụng.
	// Mặc định: false (fail-closed, chuẩn Zero Trust).
	FailOpen bool

	// Cấu hình OpenZiti cho Phase 2 (kết nối ẩn hoàn toàn)
	UseZiti          bool
	ZitiIdentityFile string
	ZitiServiceName  string
}

// New khởi tạo PDPClient với kết nối gRPC persistent, Keep-Alive, và mTLS tùy chọn.
// Trả về lỗi nếu không thể thiết lập kết nối trong thời gian dialTimeout.
func New(cfg Config) (*PDPClient, error) {
	// Khi chạy chế độ thường (non-Ziti), địa chỉ Addr bắt buộc phải có
	if !cfg.UseZiti && cfg.Addr == "" {
		return nil, fmt.Errorf("pdpclient: PDP address must not be empty")
	}

	timeout := defaultRequestTimeout
	if cfg.RequestTimeout > 0 {
		timeout = cfg.RequestTimeout
	}

	// Cấu hình Keep-Alive để duy trì kết nối persistent với PDP.
	kaParams := keepalive.ClientParameters{
		Time:                10 * time.Second,
		Timeout:             3 * time.Second,
		PermitWithoutStream: true,
	}

	var dialOpts []grpc.DialOption
	dialOpts = append(dialOpts, grpc.WithKeepaliveParams(kaParams))

	var zCtx zititransport.Context
	var err error

	if cfg.UseZiti {
		log.Printf("[PDPClient] Đang kết nối gRPC tới PDP qua OpenZiti overlay. Service: '%s', Identity: %s", cfg.ZitiServiceName, cfg.ZitiIdentityFile)
		if _, err := os.Stat(cfg.ZitiIdentityFile); os.IsNotExist(err) {
			return nil, fmt.Errorf("pdpclient: ziti identity file not found at: %s", cfg.ZitiIdentityFile)
		}

		zCfg, err := zititransport.NewConfigFromFile(cfg.ZitiIdentityFile)
		if err != nil {
			return nil, fmt.Errorf("pdpclient: failed to load client ziti config: %w", err)
		}

		zCtx, err = zititransport.NewContext(zCfg)
		if err != nil {
			return nil, fmt.Errorf("pdpclient: failed to create client ziti context: %w", err)
		}

		if err := zCtx.Authenticate(); err != nil {
			zCtx.Close()
			return nil, fmt.Errorf("pdpclient: failed to authenticate client with ziti controller: %w", err)
		}

		// Định tuyến kết nối gRPC quay qua OpenZiti dialer
		dialOpts = append(dialOpts, grpc.WithContextDialer(func(ctx context.Context, addr string) (net.Conn, error) {
			return zCtx.Dial(cfg.ZitiServiceName)
		}))

		// OpenZiti tự động mã hóa mTLS 2 chiều ở lớp overlay mạng ảo (ZTA), nên sử dụng insecure ở mức ứng dụng
		dialOpts = append(dialOpts, grpc.WithTransportCredentials(insecure.NewCredentials()))
	} else {
		// Chọn transport credentials dựa trên cấu hình truyền thống.
		if cfg.TLSCertFile != "" && cfg.TLSKeyFile != "" && cfg.TLSCAFile != "" {
			creds, err := loadMTLSCredentials(cfg.TLSCertFile, cfg.TLSKeyFile, cfg.TLSCAFile)
			if err != nil {
				return nil, fmt.Errorf("pdpclient: failed to load mTLS credentials: %w", err)
			}
			dialOpts = append(dialOpts, grpc.WithTransportCredentials(creds))
			log.Printf("[PDPClient] mTLS enabled — connecting to PDP at %s", cfg.Addr)
		} else {
			// Insecure — chỉ dùng trong local dev.
			dialOpts = append(dialOpts, grpc.WithTransportCredentials(insecure.NewCredentials()))
			log.Printf("[PDPClient] WARNING: insecure transport — PDP at %s. Use mTLS in production!", cfg.Addr)
		}
	}

	// Thiết lập kết nối với timeout.
	dialCtx, cancel := context.WithTimeout(context.Background(), defaultDialTimeout)
	defer cancel()

	// ForceCodec "json" — phải khớp với codec của PDP server
	// Nếu không set, gRPC sẽ dùng default protobuf codec và struct không implement proto.Message sẽ fail
	dialOpts = append(dialOpts, grpc.WithDefaultCallOptions(
		grpc.ForceCodec(policyv1.JSONCodec()),
	))

	// Khi kết nối qua Ziti, target address của grpc.DialContext không quan trọng vì dialer sẽ override và route
	// tất cả packet thẳng tới Ziti service. Tuy nhiên, ta vẫn dùng tên service làm address tượng trưng.
	targetAddr := cfg.Addr
	if cfg.UseZiti {
		targetAddr = "passthrough:///" + cfg.ZitiServiceName
	}

	//nolint:staticcheck // grpc.DialContext là API chuẩn với version grpc hiện tại trong go.mod
	conn, err := grpc.DialContext(dialCtx, targetAddr, dialOpts...)
	if err != nil {
		if zCtx != nil {
			zCtx.Close()
		}
		return nil, fmt.Errorf("pdpclient: failed to dial PDP: %w", err)
	}

	return &PDPClient{
		conn:           conn,
		client:         policyv1.NewPolicyDecisionPointClient(conn),
		requestTimeout: timeout,
		failOpen:       cfg.FailOpen,
		zCtx:           zCtx,
	}, nil
}

// CheckAccess gọi gRPC PDP.CheckAccess và trả về quyết định phân quyền.
//
// Trả về:
//   - allow=true nếu PDP quyết định ALLOW
//   - matchedPolicyID là ID của luật đã kích hoạt quyết định (trống nếu DENY mặc định)
//   - httpStatus là HTTP status code phù hợp để trả về cho client:
//     403 nếu DENY, 503 nếu PDP không phản hồi (fail-closed), 0 nếu ALLOW
//   - err khác nil nếu có lỗi hạ tầng (timeout, connection refused)
//
// Hành vi fail-closed/fail-open được kiểm soát bởi PDPClient.failOpen.
func (c *PDPClient) CheckAccess(
	ctx context.Context,
	tenantID, subject, action, resource string,
	ctxMap map[string]string,
) (allow bool, matchedPolicyID string, httpStatus int, err error) {

	// Áp dụng timeout riêng cho request này.
	reqCtx, cancel := context.WithTimeout(ctx, c.requestTimeout)
	defer cancel()

	req := &policyv1.CheckAccessRequest{
		TenantId: tenantID,
		Subject:  subject,
		Action:   action,
		Resource: resource,
		Context:  ctxMap,
	}

	resp, err := c.client.CheckAccess(reqCtx, req)
	if err != nil {
		// Lỗi gRPC (timeout, connection refused, v.v.)
		if c.failOpen {
			// fail-open: ghi log cảnh báo và cho phép request đi qua.
			log.Printf("[PDPClient] WARNING: PDP unavailable (fail-open mode), allowing request. subject=%s action=%s resource=%s err=%v",
				subject, action, resource, err)
			return true, "", 0, nil
		}
		// fail-closed (mặc định): từ chối request và báo lỗi 503.
		log.Printf("[PDPClient] ERROR: PDP unavailable (fail-closed), denying request. subject=%s action=%s resource=%s err=%v",
			subject, action, resource, err)
		return false, "", http.StatusServiceUnavailable, fmt.Errorf("policy decision point unavailable: %w", err)
	}

	if resp.Decision == policyv1.CheckAccessResponse_ALLOW {
		return true, resp.MatchedPolicyId, 0, nil
	}

	// DENY — trả về 403 Forbidden.
	return false, resp.MatchedPolicyId, http.StatusForbidden, nil
}

// Close đóng kết nối gRPC và giải phóng tài nguyên mạng ảo OpenZiti context.
func (c *PDPClient) Close() error {
	var err error
	if c.conn != nil {
		err = c.conn.Close()
	}
	if c.zCtx != nil {
		c.zCtx.Close()
	}
	return err
}

// loadMTLSCredentials tải mTLS credentials từ file cert/key/CA.
// CA cert được dùng để xác thực server certificate của PDP.
func loadMTLSCredentials(certFile, keyFile, caFile string) (credentials.TransportCredentials, error) {
	// Tải client cert (dùng để PDP xác thực Gateway là PEP hợp lệ).
	clientCert, err := tls.LoadX509KeyPair(certFile, keyFile)
	if err != nil {
		return nil, fmt.Errorf("failed to load client cert/key pair (%s, %s): %w", certFile, keyFile, err)
	}

	// Tải CA cert để xác thực server cert của PDP.
	caCert, err := os.ReadFile(caFile)
	if err != nil {
		return nil, fmt.Errorf("failed to read CA cert file %s: %w", caFile, err)
	}

	certPool := x509.NewCertPool()
	if !certPool.AppendCertsFromPEM(caCert) {
		return nil, fmt.Errorf("failed to parse CA cert from %s: invalid PEM format", caFile)
	}

	tlsCfg := &tls.Config{
		Certificates: []tls.Certificate{clientCert},
		RootCAs:      certPool,
		// MinVersion tối thiểu TLS 1.2, khuyến nghị 1.3 cho production.
		MinVersion: tls.VersionTLS12,
	}

	return credentials.NewTLS(tlsCfg), nil
}
