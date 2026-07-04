import { NextResponse } from "next/server";
import { Pool } from "pg";

const pool = new Pool({
  connectionString: process.env.DATABASE_URL || "postgresql://postgres:postgres_secure_password_2026@localhost:5432/fapi_db"
});

export async function GET() {
  try {
    const result = await pool.query(
      "SELECT id, timestamp, actor_id::text, tenant_id::text, action, resource, details, prev_hash, block_hash FROM audit_logs ORDER BY id ASC"
    );
    const rows = result.rows;

    const steps: any[] = [];
    let isValid = true;

    for (let i = 0; i < rows.length; i++) {
      const current = rows[i];
      let stepValid = true;
      let error = "";

      if (i === 0) {
        const expectedStartHash = "0000000000000000000000000000000000000000000000000000000000000000";
        if (current.prev_hash !== expectedStartHash) {
          stepValid = false;
          error = `Genesis block prev_hash is modified: ${current.prev_hash}`;
        }
      } else {
        const previous = rows[i - 1];
        if (current.prev_hash.trim() !== previous.block_hash.trim()) {
          stepValid = false;
          error = `Chain broken at Block #${current.id}: prev_hash does not match previous block_hash.`;
        }
      }

      if (!stepValid) {
        isValid = false;
      }

      steps.push({
        blockId: current.id,
        action: current.action,
        resource: current.resource,
        prevHash: current.prev_hash,
        blockHash: current.block_hash,
        valid: stepValid,
        error: error
      });
    }

    return NextResponse.json({
      status: "success",
      data: {
        isValid: isValid,
        totalBlocks: rows.length,
        steps: steps
      }
    });
  } catch (err: any) {
    return NextResponse.json({ status: "error", message: err.message }, { status: 500 });
  }
}
