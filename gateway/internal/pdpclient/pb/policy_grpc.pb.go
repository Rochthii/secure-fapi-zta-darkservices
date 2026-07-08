// Code được sao chép và điều chỉnh từ standalone-policy-engine/proto/v1/policy_grpc.pb.go
// Chứa gRPC client/server stub cho PolicyDecisionPoint service.
// Service method path phải khớp tuyệt đối với PDP server: "/policy.v1.PolicyDecisionPoint/..."
package policyv1

import (
	"context"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// PolicyDecisionPointClient là interface client cho dịch vụ PDP.
type PolicyDecisionPointClient interface {
	CheckAccess(ctx context.Context, in *CheckAccessRequest, opts ...grpc.CallOption) (*CheckAccessResponse, error)
	ExplainDecision(ctx context.Context, in *ExplainRequest, opts ...grpc.CallOption) (*ExplainResponse, error)
}

type policyDecisionPointClient struct {
	cc grpc.ClientConnInterface
}

// NewPolicyDecisionPointClient tạo mới gRPC client instance.
func NewPolicyDecisionPointClient(cc grpc.ClientConnInterface) PolicyDecisionPointClient {
	return &policyDecisionPointClient{cc}
}

func (c *policyDecisionPointClient) CheckAccess(ctx context.Context, in *CheckAccessRequest, opts ...grpc.CallOption) (*CheckAccessResponse, error) {
	out := new(CheckAccessResponse)
	// Path PHẢI khớp với ServiceDesc trong PDP server.
	err := c.cc.Invoke(ctx, "/policy.v1.PolicyDecisionPoint/CheckAccess", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *policyDecisionPointClient) ExplainDecision(ctx context.Context, in *ExplainRequest, opts ...grpc.CallOption) (*ExplainResponse, error) {
	out := new(ExplainResponse)
	err := c.cc.Invoke(ctx, "/policy.v1.PolicyDecisionPoint/ExplainDecision", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// PolicyDecisionPointServer là interface server cho dịch vụ PDP (dùng cho test mock).
type PolicyDecisionPointServer interface {
	CheckAccess(context.Context, *CheckAccessRequest) (*CheckAccessResponse, error)
	ExplainDecision(context.Context, *ExplainRequest) (*ExplainResponse, error)
}

// UnimplementedPolicyDecisionPointServer có thể được nhúng để có forward compatibility.
type UnimplementedPolicyDecisionPointServer struct{}

func (UnimplementedPolicyDecisionPointServer) CheckAccess(context.Context, *CheckAccessRequest) (*CheckAccessResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method CheckAccess not implemented")
}

func (UnimplementedPolicyDecisionPointServer) ExplainDecision(context.Context, *ExplainRequest) (*ExplainResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method ExplainDecision not implemented")
}
