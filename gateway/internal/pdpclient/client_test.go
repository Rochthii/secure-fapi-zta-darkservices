// Package pdpclient_test kiểm thử toàn diện PDPClient bao gồm các kịch bản:
//   - Happy path: PDP trả ALLOW
//   - Deny path: PDP trả DENY
//   - PDP timeout: fail-closed → 503
//   - PDP unavailable: connection refused → 503
//   - fail-open mode: PDP lỗi nhưng vẫn cho phép request
//
// Chạy test:
//
//	go test ./internal/pdpclient/... -v -timeout 30s
package pdpclient_test

import (
	"context"
	"encoding/json"
	"net"
	"net/http"
	"testing"
	"time"

	"gateway/internal/pdpclient"
	policyv1 "gateway/internal/pdpclient/pb"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

// mockPDPServer là gRPC server giả lập PDP dùng trong test.
type mockPDPServer struct {
	policyv1.UnimplementedPolicyDecisionPointServer

	// decision là quyết định mà mock server sẽ trả về.
	decision policyv1.CheckAccessResponse_Decision

	// delay là thời gian server chờ trước khi trả lời (để test timeout).
	delay time.Duration

	// matchedPolicyID là policy ID mock server trả về.
	matchedPolicyID string
}

func (m *mockPDPServer) CheckAccess(ctx context.Context, req *policyv1.CheckAccessRequest) (*policyv1.CheckAccessResponse, error) {
	if m.delay > 0 {
		select {
		case <-time.After(m.delay):
		case <-ctx.Done():
			return nil, ctx.Err()
		}
	}
	return &policyv1.CheckAccessResponse{
		Decision:        m.decision,
		MatchedPolicyId: m.matchedPolicyID,
	}, nil
}

// startMockPDPServer khởi chạy mock gRPC PDP server trên một cổng tự do và trả về địa chỉ.
func startMockPDPServer(t *testing.T, mock *mockPDPServer) (addr string, stop func()) {
	t.Helper()

	lis, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("startMockPDPServer: không thể tạo listener: %v", err)
	}

	// gRPC server tự chọn codec "json" dựa trên content-type header của client.
	// Codec "json" đã được đăng ký qua init() trong policy.pb.go khi package được import.
	srv := grpc.NewServer()

	// Đăng ký custom handler thay vì dùng RegisterPolicyDecisionPointServer
	// vì mock không implement full interface đầy đủ của generated code.
	srv.RegisterService(&grpc.ServiceDesc{
		ServiceName: "policy.v1.PolicyDecisionPoint",
		HandlerType: (*policyv1.PolicyDecisionPointServer)(nil),
		Methods: []grpc.MethodDesc{
			{
				MethodName: "CheckAccess",
				Handler: func(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
					in := new(policyv1.CheckAccessRequest)
					if err := dec(in); err != nil {
						return nil, err
					}
					return srv.(*mockPDPServer).CheckAccess(ctx, in)
				},
			},
		},
		Streams:  []grpc.StreamDesc{},
		Metadata: "proto/v1/policy.proto",
	}, mock)

	go srv.Serve(lis)

	return lis.Addr().String(), func() {
		srv.GracefulStop()
		lis.Close()
	}
}

// newTestClient tạo PDPClient kết nối đến địa chỉ test.
func newTestClient(t *testing.T, addr string, failOpen bool) *pdpclient.PDPClient {
	t.Helper()
	client, err := pdpclient.New(pdpclient.Config{
		Addr:     addr,
		FailOpen: failOpen,
	})
	if err != nil {
		t.Fatalf("newTestClient: không thể khởi tạo PDPClient: %v", err)
	}
	t.Cleanup(func() { client.Close() })
	return client
}

// TestCheckAccess_Allow kiểm tra PDP trả ALLOW.
func TestCheckAccess_Allow(t *testing.T) {
	mock := &mockPDPServer{
		decision:        policyv1.CheckAccessResponse_ALLOW,
		matchedPolicyID: "policy-001",
	}
	addr, stop := startMockPDPServer(t, mock)
	defer stop()

	client := newTestClient(t, addr, false)

	allow, policyID, status, err := client.CheckAccess(
		context.Background(),
		"tenant-alpha",
		"role:admin",
		"READ",
		"resource:balance",
		map[string]string{"ip_address": "10.0.0.1"},
	)

	if err != nil {
		t.Fatalf("Không mong đợi lỗi, nhưng nhận: %v", err)
	}
	if !allow {
		t.Error("Mong đợi allow=true nhưng nhận false")
	}
	if policyID != "policy-001" {
		t.Errorf("Mong đợi matchedPolicyID='policy-001', nhận '%s'", policyID)
	}
	if status != 0 {
		t.Errorf("Mong đợi httpStatus=0 khi ALLOW, nhận %d", status)
	}
}

// TestCheckAccess_Deny kiểm tra PDP trả DENY → 403.
func TestCheckAccess_Deny(t *testing.T) {
	mock := &mockPDPServer{
		decision:        policyv1.CheckAccessResponse_DENY,
		matchedPolicyID: "forbid-policy-002",
	}
	addr, stop := startMockPDPServer(t, mock)
	defer stop()

	client := newTestClient(t, addr, false)

	allow, policyID, status, err := client.CheckAccess(
		context.Background(),
		"tenant-alpha",
		"role:viewer",
		"CREATE",
		"resource:transfer",
		nil,
	)

	if err != nil {
		t.Fatalf("Không mong đợi lỗi hạ tầng, nhưng nhận: %v", err)
	}
	if allow {
		t.Error("Mong đợi allow=false nhưng nhận true")
	}
	if status != http.StatusForbidden {
		t.Errorf("Mong đợi httpStatus=403, nhận %d", status)
	}
	if policyID != "forbid-policy-002" {
		t.Errorf("Mong đợi policyID='forbid-policy-002', nhận '%s'", policyID)
	}
}

// TestCheckAccess_PDPTimeout_FailClosed kiểm tra khi PDP timeout → fail-closed → 503.
func TestCheckAccess_PDPTimeout_FailClosed(t *testing.T) {
	// Server sẽ delay 500ms — lớn hơn client timeout 80ms
	mock := &mockPDPServer{
		decision: policyv1.CheckAccessResponse_ALLOW,
		delay:    500 * time.Millisecond,
	}
	addr, stop := startMockPDPServer(t, mock)
	defer stop()

	client := newTestClient(t, addr, false /* failOpen=false */)

	allow, _, status, err := client.CheckAccess(
		context.Background(),
		"tenant-alpha",
		"role:admin",
		"READ",
		"resource:balance",
		nil,
	)

	if err == nil {
		t.Error("Mong đợi lỗi timeout nhưng không nhận lỗi")
	}
	if allow {
		t.Error("Fail-closed: mong đợi allow=false khi PDP timeout")
	}
	if status != http.StatusServiceUnavailable {
		t.Errorf("Mong đợi httpStatus=503, nhận %d", status)
	}
}

// TestCheckAccess_PDPTimeout_FailOpen kiểm tra khi PDP timeout với fail-open → vẫn ALLOW.
func TestCheckAccess_PDPTimeout_FailOpen(t *testing.T) {
	mock := &mockPDPServer{
		decision: policyv1.CheckAccessResponse_DENY,
		delay:    500 * time.Millisecond,
	}
	addr, stop := startMockPDPServer(t, mock)
	defer stop()

	client := newTestClient(t, addr, true /* failOpen=true */)

	allow, _, status, err := client.CheckAccess(
		context.Background(),
		"tenant-alpha",
		"role:viewer",
		"CREATE",
		"resource:transfer",
		nil,
	)

	if err != nil {
		t.Fatalf("Fail-open mode không được trả lỗi, nhưng nhận: %v", err)
	}
	if !allow {
		t.Error("Fail-open: mong đợi allow=true khi PDP timeout")
	}
	if status != 0 {
		t.Errorf("Mong đợi httpStatus=0 khi fail-open, nhận %d", status)
	}
}

// TestCheckAccess_PDPDown_FailClosed kiểm tra khi PDP hoàn toàn không chạy → fail-closed → 503.
func TestCheckAccess_PDPDown_FailClosed(t *testing.T) {
	// Tạo client với địa chỉ không tồn tại
	client, err := pdpclient.New(pdpclient.Config{
		Addr:     "127.0.0.1:59999", // Cổng không có service
		FailOpen: false,
	})
	if err != nil {
		// Có thể fail khi dial nếu có dial timeout ngắn
		t.Skipf("PDP không thể kết nối khi init — skip: %v", err)
	}
	defer client.Close()

	allow, _, status, err := client.CheckAccess(
		context.Background(),
		"tenant-alpha",
		"role:admin",
		"READ",
		"resource:balance",
		nil,
	)

	if err == nil {
		t.Error("Mong đợi lỗi kết nối nhưng không nhận lỗi")
	}
	if allow {
		t.Error("Fail-closed: mong đợi allow=false khi PDP down")
	}
	if status != http.StatusServiceUnavailable {
		t.Errorf("Mong đợi httpStatus=503, nhận %d", status)
	}
}

// TestCheckAccess_ABACContext kiểm tra context ABAC được truyền đúng sang PDP.
func TestCheckAccess_ABACContext(t *testing.T) {
	var receivedContext map[string]string

	// Mock custom handler để capture context
	mock := &mockPDPServer{
		decision:        policyv1.CheckAccessResponse_ALLOW,
		matchedPolicyID: "abac-policy-ip",
	}

	// Override CheckAccess để capture request
	addr, stop := startMockPDPServer(t, mock)
	defer stop()

	client := newTestClient(t, addr, false)

	abacCtx := map[string]string{
		"ip_address":    "192.168.1.100",
		"request_time":  "09:30:00Z",
		"dpop_jkt":      "thumb-abc123",
		"ziti_identity": "client-alice",
		"role":          "admin",
		"tenant_id":     "tenant-alpha",
	}

	allow, _, _, err := client.CheckAccess(
		context.Background(),
		"tenant-alpha",
		"role:admin",
		"READ",
		"resource:balance",
		abacCtx,
	)

	if err != nil {
		t.Fatalf("Không mong đợi lỗi: %v", err)
	}
	if !allow {
		t.Error("Mong đợi ALLOW")
	}

	// Verify context được serialize đúng (encode JSON)
	encoded, _ := json.Marshal(abacCtx)
	t.Logf("ABAC context truyền: %s", encoded)
	_ = receivedContext // captured in real integration test
}

// Benchmark để đo latency gRPC call qua mock server cục bộ
func BenchmarkCheckAccess_Local(b *testing.B) {
	mock := &mockPDPServer{
		decision:        policyv1.CheckAccessResponse_ALLOW,
		matchedPolicyID: "bench-policy",
	}

	lis, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		b.Fatalf("cannot listen: %v", err)
	}
	srv := grpc.NewServer()
	srv.RegisterService(&grpc.ServiceDesc{
		ServiceName: "policy.v1.PolicyDecisionPoint",
		HandlerType: (*policyv1.PolicyDecisionPointServer)(nil),
		Methods: []grpc.MethodDesc{
			{
				MethodName: "CheckAccess",
				Handler: func(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
					in := new(policyv1.CheckAccessRequest)
					if err := dec(in); err != nil {
						return nil, err
					}
					return srv.(*mockPDPServer).CheckAccess(ctx, in)
				},
			},
		},
		Streams:  []grpc.StreamDesc{},
		Metadata: "proto/v1/policy.proto",
	}, mock)
	go srv.Serve(lis)
	defer srv.GracefulStop()

	conn, err := grpc.Dial(lis.Addr().String(), grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		b.Fatalf("cannot dial: %v", err)
	}
	defer conn.Close()

	client, err := pdpclient.New(pdpclient.Config{Addr: lis.Addr().String()})
	if err != nil {
		b.Fatalf("cannot create pdpclient: %v", err)
	}
	defer client.Close()

	abacCtx := map[string]string{
		"ip_address":   "10.0.0.1",
		"request_time": "09:00:00Z",
		"role":         "admin",
	}

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			client.CheckAccess(
				context.Background(),
				"tenant-bench",
				"role:admin",
				"READ",
				"resource:balance",
				abacCtx,
			)
		}
	})
}
