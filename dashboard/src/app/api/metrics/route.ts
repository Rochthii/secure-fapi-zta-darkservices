import { NextResponse } from "next/server";

export async function GET() {
  try {
    const res = await fetch("http://localhost:8080/metrics", { cache: "no-store" });
    if (!res.ok) {
      throw new Error(`Gateway returned status ${res.status}`);
    }
    const text = await res.text();
    
    // Parse Prometheus text format
    const metrics: any = {
      requests: [],
      securityOverhead: {
        dpop: 0,
        token: 0,
        ziti: 0
      },
      dbLatency: {
        rls_context: 0,
        worm_exec: 0
      }
    };

    const lines = text.split("\n");
    for (const line of lines) {
      if (!line || line.startsWith("#")) continue;

      if (line.startsWith("gateway_requests_total")) {
        // gateway_requests_total{handler="/api/balance",status="200"} 2
        const match = line.match(/gateway_requests_total\{handler="([^"]+)",status="([^"]+)"\} (\d+)/);
        if (match) {
          metrics.requests.push({
            handler: match[1],
            status: match[2],
            count: parseInt(match[3], 10)
          });
        }
      } else if (line.startsWith("gateway_security_overhead_microseconds")) {
        // gateway_security_overhead_microseconds{stage="dpop"} 538
        const match = line.match(/gateway_security_overhead_microseconds\{stage="([^"]+)"\} (\d+)/);
        if (match) {
          metrics.securityOverhead[match[1]] = parseInt(match[2], 10);
        }
      } else if (line.startsWith("gateway_db_latency_microseconds")) {
        // gateway_db_latency_microseconds{operation="rls_context"} 1628
        const match = line.match(/gateway_db_latency_microseconds\{operation="([^"]+)"\} (\d+)/);
        if (match) {
          metrics.dbLatency[match[1]] = parseInt(match[2], 10);
        }
      }
    }

    return NextResponse.json({ status: "success", data: metrics });
  } catch (err: any) {
    return NextResponse.json({ status: "error", message: err.message }, { status: 500 });
  }
}
