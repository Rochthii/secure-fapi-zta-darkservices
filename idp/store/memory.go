package store

import (
	"sync"
	"time"
)

type AuthCodeInfo struct {
	CodeChallenge string
	ExpiresAt     time.Time
}

type RAMStore struct {
	codes sync.Map // Map chứa authorization_code -> AuthCodeInfo
	jtis  sync.Map // Map chứa jti -> thời gian hết hạn (dùng để chống replay attack)
}

var (
	instance *RAMStore
	once     sync.Once
)

// GetStore trả về singleton instance của RAMStore
func GetStore() *RAMStore {
	once.Do(func() {
		instance = &RAMStore{}
		// Khởi động Goroutine định kỳ dọn dẹp các JTI và Auth Code đã hết hạn
		go instance.startCleanupTicker()
	})
	return instance
}

// SaveAuthCode lưu mã authorization code kèm code_challenge và thời gian hết hạn
func (s *RAMStore) SaveAuthCode(code, challenge string, ttl time.Duration) {
	s.codes.Store(code, AuthCodeInfo{
		CodeChallenge: challenge,
		ExpiresAt:     time.Now().Add(ttl),
	})
}

// GetAndRemoveAuthCode lấy code_challenge từ code và xóa code đó ngay lập tức (Chỉ dùng 1 lần)
func (s *RAMStore) GetAndRemoveAuthCode(code string) (string, bool) {
	val, ok := s.codes.Load(code)
	if !ok {
		return "", false
	}
	
	// Xóa ngay lập tức khỏi bộ nhớ
	s.codes.Delete(code)
	
	info := val.(AuthCodeInfo)
	if time.Now().After(info.ExpiresAt) {
		return "", false // Đã hết hạn
	}
	
	return info.CodeChallenge, true
}

// IsJTIUsedAndSave kiểm tra xem JTI (DPoP Proof ID) đã được sử dụng chưa
// Nếu chưa, lưu lại JTI kèm TTL và trả về false. Nếu đã dùng, trả về true.
func (s *RAMStore) IsJTIUsedAndSave(jti string, ttl time.Duration) bool {
	now := time.Now()
	val, loaded := s.jtis.LoadOrStore(jti, now.Add(ttl))
	if loaded {
		// Kiểm tra xem JTI lưu trước đó đã thực sự hết hạn chưa
		expireTime := val.(time.Time)
		if now.Before(expireTime) {
			return true // Đang hoạt động -> Bị trùng lặp (Replay attack)
		}
		// Đã hết hạn, cập nhật lại thời gian mới
		s.jtis.Store(jti, now.Add(ttl))
		return false
	}
	return false
}

// startCleanupTicker định kỳ dọn dẹp bộ nhớ RAMStore
func (s *RAMStore) startCleanupTicker() {
	ticker := time.NewTicker(1 * time.Minute)
	for range ticker.C {
		now := time.Now()
		
		// 1. Dọn dẹp Auth Codes hết hạn
		s.codes.Range(func(key, val interface{}) bool {
			info := val.(AuthCodeInfo)
			if now.After(info.ExpiresAt) {
				s.codes.Delete(key)
			}
			return true
		})
		
		// 2. Dọn dẹp JTI cache hết hạn
		s.jtis.Range(func(key, val interface{}) bool {
			expireTime := val.(time.Time)
			if now.After(expireTime) {
				s.jtis.Delete(key)
			}
			return true
		})
	}
}
