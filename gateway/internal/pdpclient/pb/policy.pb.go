// Code được sao chép và điều chỉnh từ standalone-policy-engine/proto/v1/policy.pb.go
// để tránh tạo dependency vòng giữa 2 Go module.
// Sử dụng JSON-over-gRPC codec tương thích với PDP server.
// Phiên bản proto gốc: proto/v1/policy.proto (policy.v1.PolicyDecisionPoint)
package policyv1

import (
	"encoding/json"

	"google.golang.org/grpc/encoding"
)

// CheckAccessRequest là yêu cầu kiểm tra quyền truy cập gửi lên PDP.
type CheckAccessRequest struct {
	TenantId string            `json:"tenant_id,omitempty"`
	Subject  string            `json:"subject,omitempty"`
	Action   string            `json:"action,omitempty"`
	Resource string            `json:"resource,omitempty"`
	Context  map[string]string `json:"context,omitempty"`
}

// CheckAccessResponse_Decision là enum quyết định của PDP.
type CheckAccessResponse_Decision int32

const (
	CheckAccessResponse_DENY  CheckAccessResponse_Decision = 0
	CheckAccessResponse_ALLOW CheckAccessResponse_Decision = 1
)

// CheckAccessResponse là phản hồi từ PDP.
// Lưu ý: decision KHÔNG dùng omitempty vì giá trị 0 (DENY) là hợp lệ và cần được serialize.
type CheckAccessResponse struct {
	Decision        CheckAccessResponse_Decision `json:"decision"`
	MatchedPolicyId string                       `json:"matched_policy_id,omitempty"`
}

// ExplainRequest là yêu cầu giải thích chi tiết quyết định.
type ExplainRequest struct {
	TenantId string            `json:"tenant_id,omitempty"`
	Subject  string            `json:"subject,omitempty"`
	Action   string            `json:"action,omitempty"`
	Resource string            `json:"resource,omitempty"`
	Context  map[string]string `json:"context,omitempty"`
}

// ExplainResponse_Decision là enum quyết định trong ExplainResponse.
type ExplainResponse_Decision int32

const (
	ExplainResponse_DENY  ExplainResponse_Decision = 0
	ExplainResponse_ALLOW ExplainResponse_Decision = 1
)

// ExplainResponse là phản hồi giải thích chi tiết từ PDP.
// Lưu ý: decision KHÔNG dùng omitempty vì giá trị 0 (DENY) là hợp lệ.
type ExplainResponse struct {
	Decision    ExplainResponse_Decision `json:"decision"`
	FinalReason string                   `json:"final_reason,omitempty"`
	Matched     []*PolicyMetadata        `json:"matched,omitempty"`
}

// PolicyMetadata chứa thông tin chi tiết về một policy đã khớp.
type PolicyMetadata struct {
	PolicyId   string `json:"policy_id,omitempty"`
	Effect     string `json:"effect,omitempty"`
	PolicyText string `json:"policy_text,omitempty"`
}

// Các phương thức tương thích protobuf interface cơ bản để gRPC có thể compile.
func (x *CheckAccessRequest) Reset()         { *x = CheckAccessRequest{} }
func (x *CheckAccessRequest) String() string  { return "" }
func (*CheckAccessRequest) ProtoMessage()     {}

func (x *CheckAccessResponse) Reset()         { *x = CheckAccessResponse{} }
func (x *CheckAccessResponse) String() string  { return "" }
func (*CheckAccessResponse) ProtoMessage()     {}

func (x *ExplainRequest) Reset()         { *x = ExplainRequest{} }
func (x *ExplainRequest) String() string  { return "" }
func (*ExplainRequest) ProtoMessage()     {}

func (x *ExplainResponse) Reset()         { *x = ExplainResponse{} }
func (x *ExplainResponse) String() string  { return "" }
func (*ExplainResponse) ProtoMessage()     {}

func (x *PolicyMetadata) Reset()         { *x = PolicyMetadata{} }
func (x *PolicyMetadata) String() string  { return "" }
func (*PolicyMetadata) ProtoMessage()     {}

// JSONCodec là JSON codec tùy chỉnh để thay thế protobuf binary encoding.
// PHẢI khớp với PDP server — cả hai dùng cùng codec name "json".
type jsonCodec struct{}

func (jsonCodec) Marshal(v interface{}) ([]byte, error) {
	return json.Marshal(v)
}

func (jsonCodec) Unmarshal(data []byte, v interface{}) error {
	return json.Unmarshal(data, v)
}

func (jsonCodec) Name() string {
	return "json"
}

// JSONCodec trả về instance của json codec để dùng với grpc.ForceCodec() ở phía client.
// Cả server (PDP) và client (Gateway) đều phải sử dụng cùng codec này.
func JSONCodec() jsonCodec {
	return jsonCodec{}
}

// init đăng ký jsonCodec vào registry của gRPC.
// Đây là điểm khởi tạo duy nhất — được gọi tự động khi package được import.
func init() {
	encoding.RegisterCodec(jsonCodec{})
}
