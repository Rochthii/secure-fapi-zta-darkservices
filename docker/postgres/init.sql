-- Khởi tạo Database FAPI-ZTA & WORM Audit Ledger
CREATE EXTENSION IF NOT EXISTS pgcrypto;

-- 1. Tạo role cho ứng dụng kết nối (Không dùng superuser)
-- Mật khẩu sẽ được quản lý qua biến môi trường, mặc định ở đây là 'app_secure_password_2026'
DO $$
BEGIN
    IF NOT EXISTS (SELECT FROM pg_catalog.pg_roles WHERE rolname = 'app_user') THEN
        CREATE ROLE app_user WITH LOGIN PASSWORD 'app_secure_password_2026';
    END IF;
END
$$;

-- 2. Tạo bảng Giao dịch (transactions)
CREATE TABLE IF NOT EXISTS transactions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL,
    amount DECIMAL(15, 2) NOT NULL,
    description TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

-- Bật Row-Level Security (RLS) cho bảng transactions
ALTER TABLE transactions ENABLE ROW LEVEL SECURITY;

-- Tạo RLS Policy cho transactions: Chỉ cho phép truy cập dữ liệu của chính Tenant của mình
CREATE POLICY tenant_isolation_policy ON transactions
    FOR ALL
    TO app_user
    USING (tenant_id = NULLIF(current_setting('app.tenant_id', true), '')::uuid)
    WITH CHECK (tenant_id = NULLIF(current_setting('app.tenant_id', true), '')::uuid);

-- Cấp quyền cho app_user trên bảng transactions
GRANT SELECT, INSERT ON transactions TO app_user;

-- 3. Tạo bảng Nhật ký kiểm toán bất biến (audit_logs)
CREATE TABLE IF NOT EXISTS audit_logs (
    id BIGSERIAL PRIMARY KEY,
    timestamp TIMESTAMPTZ NOT NULL DEFAULT now(),
    actor_id UUID NOT NULL,
    tenant_id UUID NOT NULL,
    action VARCHAR(50) NOT NULL,
    resource VARCHAR(100) NOT NULL,
    details JSONB,
    prev_hash CHAR(64) NOT NULL,
    block_hash CHAR(64) NOT NULL
);

-- Bật RLS cho bảng audit_logs để Tenant chỉ xem được log của chính họ
ALTER TABLE audit_logs ENABLE ROW LEVEL SECURITY;

CREATE POLICY audit_tenant_isolation ON audit_logs
    FOR ALL
    TO app_user
    USING (tenant_id = NULLIF(current_setting('app.tenant_id', true), '')::uuid)
    WITH CHECK (tenant_id = NULLIF(current_setting('app.tenant_id', true), '')::uuid);

-- Cấp quyền ghi và đọc cho app_user trên bảng audit_logs và sequence đi kèm
GRANT SELECT, INSERT ON audit_logs TO app_user;
GRANT USAGE, SELECT ON SEQUENCE audit_logs_id_seq TO app_user;

-- 4. RÀNG BUỘC WORM (Write Once, Read Many) - Chặn UPDATE & DELETE
CREATE OR REPLACE FUNCTION prevent_audit_tampering()
RETURNS TRIGGER AS $$
BEGIN
    RAISE EXCEPTION 'Audit logs are immutable. UPDATE and DELETE operations are strictly prohibited.';
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER audit_log_immutable
    BEFORE UPDATE OR DELETE ON audit_logs
    FOR EACH ROW
    EXECUTE FUNCTION prevent_audit_tampering();

-- 5. HÀM TỰ ĐỘNG TÍNH HASH-CHAIN (HMAC-SHA256)
CREATE OR REPLACE FUNCTION hash_audit_record()
RETURNS TRIGGER AS $$
DECLARE
    last_block_hash CHAR(64);
    concat_text TEXT;
    audit_secret TEXT;
BEGIN
    -- Lấy audit secret từ context session
    audit_secret := NULLIF(current_setting('app.audit_secret', true), '');
    IF audit_secret IS NULL THEN
        RAISE EXCEPTION 'Audit secret is missing or invalid. Integrity check failed.';
    END IF;

    -- Lấy hash của bản ghi liền trước (bản ghi có id lớn nhất)
    SELECT block_hash INTO last_block_hash
    FROM audit_logs
    ORDER BY id DESC
    LIMIT 1;

    -- Nếu là bản ghi đầu tiên, thiết lập hash mặc định gồm 64 ký tự '0'
    IF last_block_hash IS NULL THEN
        last_block_hash := '0000000000000000000000000000000000000000000000000000000000000000';
    END IF;

    -- Gán prev_hash cho bản ghi mới
    NEW.prev_hash := last_block_hash;

    -- Chuỗi ghép để tính hash (id + timestamp + actor + action + resource + details + prev_hash)
    concat_text := COALESCE(NEW.timestamp::text, '') || '|' ||
                   COALESCE(NEW.actor_id::text, '') || '|' ||
                   COALESCE(NEW.tenant_id::text, '') || '|' ||
                   COALESCE(NEW.action, '') || '|' ||
                   COALESCE(NEW.resource, '') || '|' ||
                   COALESCE(NEW.details::text, '{}') || '|' ||
                   NEW.prev_hash;

    -- Tính toán mã hash HMAC-SHA256 bằng pgcrypto sử dụng secret key
    NEW.block_hash := encode(hmac(concat_text, audit_secret, 'sha256'), 'hex');

    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER audit_log_hash_chain
    BEFORE INSERT ON audit_logs
    FOR EACH ROW
    EXECUTE FUNCTION hash_audit_record();
