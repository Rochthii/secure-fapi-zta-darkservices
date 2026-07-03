import { NextResponse } from "next/server";
import { X509Certificate } from "crypto";
import fs from "fs";
import path from "path";

export async function GET() {
  try {
    const certsDir = "e:/Projects/Project_TN/secure-fapi-zta-darkservices/certs";
    const certFiles = [
      { name: "Root CA", path: path.join(certsDir, "ca/root-ca.crt") },
      { name: "Intermediate CA", path: path.join(certsDir, "ca/intermediate-ca.crt") },
      { name: "Client Alice", path: path.join(certsDir, "clients/client-alice.crt") }
    ];

    const results = [];
    let expiringCount = 0;

    for (const file of certFiles) {
      if (fs.existsSync(file.path)) {
        const raw = fs.readFileSync(file.path);
        const cert = new X509Certificate(raw);
        
        const validTo = new Date(cert.validTo);
        const validFrom = new Date(cert.validFrom);
        const daysRemaining = Math.ceil((validTo.getTime() - Date.now()) / (1000 * 60 * 60 * 24));
        
        if (daysRemaining < 30) {
          expiringCount++;
        }

        results.push({
          name: file.name,
          subject: cert.subject,
          issuer: cert.issuer,
          validFrom: validFrom.toISOString(),
          validTo: validTo.toISOString(),
          daysRemaining: daysRemaining,
          fingerprint: cert.fingerprint
        });
      }
    }

    return NextResponse.json({
      status: "success",
      data: {
        certificates: results,
        expiringCount: expiringCount
      }
    });
  } catch (err: any) {
    return NextResponse.json({ status: "error", message: err.message }, { status: 500 });
  }
}
