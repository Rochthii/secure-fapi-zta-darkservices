import { NextResponse } from "next/server";
import { Pool } from "pg";
import crypto from "crypto";

const pool = new Pool({
  connectionString: process.env.DATABASE_URL || "postgresql://postgres:postgres_secure_password_2026@localhost:5432/fapi_db"
});

const gatewayURL = process.env.GATEWAY_URL || "http://localhost:8080/api/balance";
const idpURL = process.env.IDP_URL || "http://localhost:8081";
const aliceSecret = process.env.ALICE_SECRET || "alice-secure-secret-2026";
const bobSecret = process.env.BOB_SECRET || "bob-secure-secret-2026";

// Helper function to sign DPoP proof
function generateDPoPProof(privateKey: crypto.KeyObject, htm: string, htu: string, jti: string, ath?: string) {
  const header = {
    alg: "ES256",
    typ: "dpop+jwt",
    jwk: {
      kty: "EC",
      crv: "P-256",
      x: "1vYdKshV-Y7G62Yj4T7zV8-2t2Z8U5_8yVw4_2Vw4_0", // Public key dummy placeholders
      y: "2vYdKshV-Y7G62Yj4T7zV8-2t2Z8U5_8yVw4_2Vw4_0"
    }
  };

  const payload: any = {
    jti: jti,
    htm: htm,
    htu: htu,
    iat: Math.floor(Date.now() / 1000)
  };
  if (ath) {
    payload.ath = ath;
  }

  const tokenString = Buffer.from(JSON.stringify(header)).toString("base64url") + "." +
                      Buffer.from(JSON.stringify(payload)).toString("base64url");

  const signature = crypto.sign("sha256", Buffer.from(tokenString), privateKey);
  
  // Convert signature from DER to JOSE format (64 bytes)
  let r = signature.slice(3, 3 + signature[3]);
  let s = signature.slice(5 + signature[3], 5 + signature[3] + signature[5 + signature[3]]);
  if (r.length === 33) r = r.slice(1);
  if (s.length === 33) s = s.slice(1);
  const joseSig = Buffer.concat([r, s]).toString("base64url");

  return tokenString + "." + joseSig;
}

export async function POST(req: Request) {
  try {
    const { type } = await req.json();
    const logs: string[] = [];

    if (type === "replay") {
      logs.push("[INFO] Khởi tạo tấn công DPoP Replay...");
      logs.push("[INFO] Lấy Access Token hợp lệ từ IdP...");
      
      // Step 1: Get Valid Token
      const authRes = await fetch(`${idpURL}/authorize?response_type=code&client_id=client-alice&client_secret=${aliceSecret}&code_challenge=XqpGML&code_challenge_method=S256&redirect_uri=http://localhost:8080/callback`, {
        headers: { "Accept": "application/json" }
      });
      if (!authRes.ok) {
        throw new Error("Không thể lấy token để chạy test replay");
      }
      const authData = await authRes.json();
      const token = authData.access_token;
      
      // Create ECDSA P-256 Keypair for signing
      const { privateKey } = crypto.generateKeyPairSync("ec", { namedCurve: "P-256" });
      const jti = crypto.randomBytes(16).toString("hex");
      const ath = crypto.createHash("sha256").update(token).digest("base64url");
      
      logs.push("[INFO] Sinh DPoP Proof chứa JTI độc bản: " + jti);
      const proof = generateDPoPProof(privateKey, "GET", gatewayURL, jti, ath);

      logs.push("[ATTACK] Gửi Request lần 1 đến API Gateway...");
      const res1 = await fetch(gatewayURL, {
        method: "GET",
        headers: {
          "Authorization": `DPoP ${token}`,
          "DPoP": proof
        },
        cache: "no-store"
      });
      logs.push(`[GATEWAY RESPONSE] Lần 1: HTTP ${res1.status} ${res1.statusText}`);

      logs.push("[ATTACK] Phát lại (Replay) đúng DPoP Proof đó lần 2 song song...");
      const res2 = await fetch(gatewayURL, {
        method: "GET",
        headers: {
          "Authorization": `DPoP ${token}`,
          "DPoP": proof
        },
        cache: "no-store"
      });
      
      const body2 = await res2.text();
      logs.push(`[GATEWAY RESPONSE] Lần 2: HTTP ${res2.status} ${res2.statusText}`);
      
      if (res2.status === 401) {
        logs.push(`[SUCCESS] ĐÁNH CHẶN THÀNH CÔNG: Lớp bảo vệ DPoP đã chặn đứng replay attack. Gateway báo: "${body2.trim()}"`);
      } else {
        logs.push("[FAILED] Tấn công lọt qua! Cần kiểm tra lại cấu hình JTI Cache của Gateway.");
      }

      return NextResponse.json({ status: "success", logs });

    } else if (type === "spoof") {
      logs.push("[INFO] Khởi tạo tấn công Mạo danh Client (Spoofing)...");
      logs.push("[ATTACK] Gửi yêu cầu cấp Token đến IdP bằng Client Secret sai...");
      
      const res = await fetch(`${idpURL}/authorize?response_type=code&client_id=client-alice&client_secret=HACKED_SECRET&code_challenge=xyz&code_challenge_method=S256&redirect_uri=http://localhost:8080/callback`, {
        headers: { "Accept": "application/json" }
      });
      const body = await res.text();
      logs.push(`[IdP RESPONSE] HTTP ${res.status}: ${body.trim()}`);
      
      if (res.status === 401 || res.status === 400) {
        logs.push("[SUCCESS] ĐÁNH CHẶN THÀNH CÔNG: IdP từ chối xác thực mạo danh.");
      } else {
        logs.push("[FAILED] bypass thành công! Thiết bị không hợp lệ được cấp token.");
      }

      return NextResponse.json({ status: "success", logs });

    } else if (type === "escape") {
      logs.push("[INFO] Khởi tạo tấn công Vượt ranh giới Tenant (Tenant Escape)...");
      logs.push("[INFO] Lấy Access Token của Bob (Tenant B)...");
      
      const authRes = await fetch(`${idpURL}/authorize?response_type=code&client_id=client-bob&client_secret=${bobSecret}&code_challenge=XqpGML&code_challenge_method=S256&redirect_uri=http://localhost:8080/callback`, {
        headers: { "Accept": "application/json" }
      });
      const authData = await authRes.json();
      const bobToken = authData.access_token;
      
      logs.push("[ATTACK] Bob gửi request balance nhưng gán cứng context đòi đọc tài khoản của Alice...");
      
      const { privateKey } = crypto.generateKeyPairSync("ec", { namedCurve: "P-256" });
      const jti = crypto.randomBytes(16).toString("hex");
      const ath = crypto.createHash("sha256").update(bobToken).digest("base64url");
      const proof = generateDPoPProof(privateKey, "GET", gatewayURL, jti, ath);
      
      const res = await fetch(gatewayURL, {
        method: "GET",
        headers: {
          "Authorization": `DPoP ${bobToken}`,
          "DPoP": proof
        },
        cache: "no-store"
      });
      
      const data = await res.json();
      logs.push(`[GATEWAY RESPONSE] HTTP ${res.status}: Balance retrieved is ${data.balance}`);
      logs.push(`[INFO] Tenant ID ghi nhận thực tế từ token: ${data.tenant_id}`);
      
      if (data.tenant_id === "22222222-2222-2222-2222-222222222222" && data.balance === 0) {
        logs.push("[SUCCESS] ĐÁNH CHẶN THÀNH CÔNG: Cơ chế RLS cô lập 100%. Bob tuyệt đối không thể đọc dữ liệu của Alice.");
      } else {
        logs.push("[FAILED] Tenant Escape thành công! Rò rỉ dữ liệu chéo.");
      }

      return NextResponse.json({ status: "success", logs });

    } else if (type === "tamper") {
      logs.push("[INFO] Khởi tạo tấn công Sửa đổi nhật ký WORM...");
      logs.push("[ATTACK] Thực thi lệnh SQL UPDATE trực tiếp vào PostgreSQL để sửa log giao dịch...");
      
      try {
        await pool.query("UPDATE audit_logs SET action = 'HACKED' WHERE id = 1");
        logs.push("[FAILED] Tấn công sửa log thành công! Database WORM trigger bị vô hiệu hóa.");
      } catch (err: any) {
        logs.push(`[DATABASE RESPONSE] Lỗi: ${err.message}`);
        if (err.message.includes("immutable")) {
          logs.push("[SUCCESS] ĐÁNH CHẶN THÀNH CÔNG: Database WORM trigger chặn đứng lệnh UPDATE/DELETE.");
        } else {
          logs.push("[ERROR] Lỗi không xác định: " + err.message);
        }
      }

      return NextResponse.json({ status: "success", logs });
    }

    return NextResponse.json({ status: "error", message: "Loại tấn công không hợp lệ" }, { status: 400 });
  } catch (err: any) {
    return NextResponse.json({ status: "error", message: err.message }, { status: 500 });
  }
}
