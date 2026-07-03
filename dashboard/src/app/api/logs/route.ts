import { NextResponse } from "next/server";
import { Pool } from "pg";

const pool = new Pool({
  connectionString: "postgresql://postgres:postgres_secure_password_2026@localhost:5432/fapi_db"
});

export async function GET() {
  try {
    const result = await pool.query(
      "SELECT id, timestamp, actor_id::text, tenant_id::text, action, resource, details, prev_hash, block_hash FROM audit_logs ORDER BY id DESC LIMIT 50"
    );
    return NextResponse.json({ status: "success", data: result.rows });
  } catch (err: any) {
    return NextResponse.json({ status: "error", message: err.message }, { status: 500 });
  }
}
