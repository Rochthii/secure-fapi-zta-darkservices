import { NextResponse } from "next/server";
import crypto from "crypto";

// Server-side cache
let cachedToken: string | null = null;
let cachedDpopKey: crypto.KeyObject | null = null;
let cachedDpopKeyPublic: crypto.KeyObject | null = null;
let cachedJwk: any = null;

// Initialize cryptographic key pair once
function getDPoPKeyPair() {
  if (!cachedDpopKey) {
    const { privateKey, publicKey } = crypto.generateKeyPairSync("ec", {
      namedCurve: "P-256",
    });
    cachedDpopKey = privateKey;
    cachedDpopKeyPublic = publicKey;
    
    const jwk = publicKey.export({ format: "jwk" });
    cachedJwk = {
      kty: "EC",
      crv: "P-256",
      x: jwk.x,
      y: jwk.y,
    };
  }
  return { privateKey: cachedDpopKey, jwk: cachedJwk };
}

// Convert DER signature format returned by Node.js crypto to JOSE/RAW signature format (64 bytes)
function derToJose(signature: Buffer): string {
  const rLen = signature[3];
  let r = signature.subarray(4, 4 + rLen);
  if (r.length === 33 && r[0] === 0) {
    r = r.subarray(1);
  }

  const sTagPos = 4 + rLen;
  const sLen = signature[sTagPos + 1];
  let s = signature.subarray(sTagPos + 2, sTagPos + 2 + sLen);
  if (s.length === 33 && s[0] === 0) {
    s = s.subarray(1);
  }

  const rPad = Buffer.alloc(32);
  r.copy(rPad, 32 - r.length);

  const sPad = Buffer.alloc(32);
  s.copy(sPad, 32 - s.length);

  return Buffer.concat([rPad, sPad]).toString("base64url");
}

// Generate DPoP proof
function generateDPoPProof(method: string, uri: string, accessToken?: string): string {
  const { privateKey, jwk } = getDPoPKeyPair();
  
  const header = {
    typ: "dpop+jwt",
    alg: "ES256",
    jwk: jwk,
  };

  const payload: any = {
    htm: method,
    htu: uri,
    iat: Math.floor(Date.now() / 1000),
    jti: crypto.randomBytes(16).toString("hex"),
  };

  if (accessToken) {
    const hash = crypto.createHash("sha256").update(accessToken).digest();
    payload.ath = hash.toString("base64url");
  }

  const headerB64 = Buffer.from(JSON.stringify(header)).toString("base64url");
  const payloadB64 = Buffer.from(JSON.stringify(payload)).toString("base64url");
  const signInput = `${headerB64}.${payloadB64}`;

  const signature = crypto.sign("sha256", Buffer.from(signInput), privateKey);
  const rawSignature = derToJose(signature);

  return `${signInput}.${rawSignature}`;
}

// Helper to obtain DPoP-bound Token
async function fetchToken(): Promise<string> {
  // PKCE Generation
  const verifier = crypto.randomBytes(32).toString("base64url");
  const challenge = crypto.createHash("sha256").update(verifier).digest().toString("base64url");

  const idpURL = "http://localhost:8081";
  
  // 1. Authorize
  const authURL = `${idpURL}/authorize?response_type=code&client_id=client-alice&client_secret=alice-secure-secret-2026&code_challenge=${challenge}&code_challenge_method=S256&redirect_uri=http://localhost:8080/callback`;
  const authRes = await fetch(authURL, { 
    headers: {
      "Accept": "application/json"
    },
    cache: "no-store" 
  });
  if (!authRes.ok) {
    const body = await authRes.text();
    throw new Error(`Authorize failed: Status ${authRes.status}, Body: ${body}`);
  }
  const authJson = await authRes.json();
  const code = authJson.code;

  // 2. Token Exchange
  const tokenURL = `${idpURL}/token`;
  const dpopProof = generateDPoPProof("POST", tokenURL);

  const params = new URLSearchParams();
  params.set("grant_type", "authorization_code");
  params.set("code", code);
  params.set("code_verifier", verifier);
  params.set("client_secret", "alice-secure-secret-2026");

  const tokenRes = await fetch(tokenURL, {
    method: "POST",
    headers: {
      "Content-Type": "application/x-www-form-urlencoded",
      "DPoP": dpopProof,
    },
    body: params.toString(),
    cache: "no-store",
  });

  if (!tokenRes.ok) {
    const body = await tokenRes.text();
    throw new Error(`Token exchange failed: Status ${tokenRes.status}, Body: ${body}`);
  }

  const tokenJson = await tokenRes.json();
  return tokenJson.access_token;
}

export async function POST() {
  try {
    // Check or fetch token
    if (!cachedToken) {
      cachedToken = await fetchToken();
    }

    const gatewayURL = "http://localhost:8080/api/balance";
    const requestCount = 50; // Safety cap to avoid DoS while giving solid metrics
    const latencies: number[] = [];
    let successCount = 0;
    
    const startTime = performance.now();

    // Perform requests concurrently in batches
    const runRequest = async () => {
      const dpopProof = generateDPoPProof("GET", gatewayURL, cachedToken!);
      const reqStart = performance.now();
      
      try {
        const res = await fetch(gatewayURL, {
          method: "GET",
          headers: {
            "Authorization": `DPoP ${cachedToken}`,
            "DPoP": dpopProof,
          },
          cache: "no-store",
        });
        
        const reqDuration = performance.now() - reqStart;
        latencies.push(reqDuration);

        if (res.status === 200) {
          successCount++;
        } else {
          if (res.status === 401 || res.status === 403) {
            cachedToken = null; // Invalidate expired/invalid token
          }
          const body = await res.text();
          console.log(`[DEBUG BENCHMARK] Request failed: Status ${res.status}, Body: ${body}`);
        }
      } catch (err) {
        latencies.push(performance.now() - reqStart);
      }
    };

    // Run batch requests
    const promises = Array.from({ length: requestCount }, () => runRequest());
    await Promise.all(promises);

    const totalTimeMs = performance.now() - startTime;

    // Calculations
    const sortedLatencies = [...latencies].sort((a, b) => a - b);
    const avgLatency = latencies.reduce((a, b) => a + b, 0) / latencies.length;
    const minLatency = sortedLatencies[0] || 0;
    const maxLatency = sortedLatencies[sortedLatencies.length - 1] || 0;
    const p95Latency = sortedLatencies[Math.floor(sortedLatencies.length * 0.95)] || 0;
    
    const rps = (successCount / totalTimeMs) * 1000;

    return NextResponse.json({
      status: "success",
      data: {
        rps: parseFloat(rps.toFixed(1)),
        avgLatencyMs: parseFloat(avgLatency.toFixed(2)),
        minLatencyMs: parseFloat(minLatency.toFixed(2)),
        maxLatencyMs: parseFloat(maxLatency.toFixed(2)),
        p95LatencyMs: parseFloat(p95Latency.toFixed(2)),
        successRate: parseFloat(((successCount / requestCount) * 10000 / 100).toFixed(1)),
        totalRequests: requestCount,
        successRequests: successCount,
        durationMs: parseFloat(totalTimeMs.toFixed(1)),
      }
    });

  } catch (err: any) {
    // If token error occurred, clear cached token to retry on next run
    cachedToken = null;
    return NextResponse.json(
      { status: "error", message: err.message || "Benchmark execution failed" },
      { status: 500 }
    );
  }
}
