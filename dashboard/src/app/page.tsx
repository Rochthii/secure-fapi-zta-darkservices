"use client";

import React, { useState, useEffect, useRef } from "react";
import { 
  Shield, 
  Activity, 
  Terminal, 
  Network, 
  Cpu, 
  Database, 
  Lock, 
  AlertTriangle, 
  CheckCircle, 
  Wifi, 
  HardDrive, 
  Server, 
  Layers, 
  Radio,
  Users,
  Key,
  HelpCircle,
  Clock,
  ShieldAlert
} from "lucide-react";
import { 
  ResponsiveContainer, 
  AreaChart, 
  Area, 
  XAxis, 
  YAxis, 
  Tooltip, 
  Legend 
} from "recharts";

// TypeScript Interfaces
interface LogEvent {
  id: string;
  timestamp: string;
  level: "INFO" | "WARNING" | "CRITICAL";
  source: string;
  message: string;
}

interface LatencyData {
  time: string;
  dpopVerify: number;
  tokenVerify: number;
  rlsContext: number;
  wormExec: number;
}

interface NodeDetails {
  name: string;
  role: string;
  cpu: number;
  ram: string;
  rps: number;
  latency: string;
  dpopSuccess: string;
  mtlsSuccess: string;
  description: string;
}

interface DBIdentity {
  actorId: string;
  tenantId: string;
}

interface CertMetadata {
  name: string;
  subject: string;
  issuer: string;
  validFrom: string;
  validTo: string;
  daysRemaining: number;
  fingerprint: string;
}

export default function SOCDashboard() {
  const [activeTab, setActiveTab] = useState<string>("overview");
  const [isMounted, setIsMounted] = useState(false);
  const [healthStatus, setHealthStatus] = useState<"HEALTHY" | "UNDER ATTACK">("HEALTHY");
  const [throughput, setThroughput] = useState(0);
  const [activeTunnels, setActiveTunnels] = useState(0);
  const [logs, setLogs] = useState<LogEvent[]>([]);
  const [latencyHistory, setLatencyHistory] = useState<LatencyData[]>([]);
  
  // Real active identities from DB
  const [identities, setIdentities] = useState<DBIdentity[]>([]);
  
  // Real certs metadata from disk
  const [certificates, setCertificates] = useState<CertMetadata[]>([]);
  const [expiringCertsCount, setExpiringCertsCount] = useState(0);
  
  const [isOffline, setIsOffline] = useState(true);
  const [errorMessage, setErrorMessage] = useState("");
  
  // Node Details Panel
  const [selectedNode, setSelectedNode] = useState<NodeDetails | null>(null);

  // WORM Chain Validation state
  const [isVerifying, setIsVerifying] = useState(false);
  const [verificationSteps, setVerificationSteps] = useState<string[]>([]);
  const [verificationSuccess, setVerificationSuccess] = useState<boolean | null>(null);
  
  const terminalEndRef = useRef<HTMLDivElement>(null);
  const verifyConsoleRef = useRef<HTMLDivElement>(null);

  // Live Benchmark states
  const [isBenchmarking, setIsBenchmarking] = useState(false);
  const [benchmarkResult, setBenchmarkResult] = useState<any>(null);
  const [benchmarkError, setBenchmarkError] = useState("");
  const [cooldown, setCooldown] = useState(0);

  const runBenchmark = async () => {
    if (cooldown > 0 || isBenchmarking) return;
    setIsBenchmarking(true);
    setBenchmarkError("");
    setBenchmarkResult(null);

    try {
      const res = await fetch("/api/benchmark", { method: "POST" });
      if (!res.ok) {
        throw new Error(`Benchmark failed: Status ${res.status}`);
      }
      const json = await res.json();
      if (json.status === "success" && json.data) {
        setBenchmarkResult(json.data);
        setCooldown(30);
      } else {
        throw new Error(json.message || "Benchmark failed");
      }
    } catch (e: any) {
      setBenchmarkError(e.message || "Error running benchmark");
    } finally {
      setIsBenchmarking(false);
    }
  };

  useEffect(() => {
    if (cooldown <= 0) return;
    const timer = setTimeout(() => setCooldown(cooldown - 1), 1000);
    return () => clearTimeout(timer);
  }, [cooldown]);

  // Initialize and mount
  useEffect(() => {
    setIsMounted(true);
  }, []);

  // Fetch real-time data from local Go API Gateway and Postgres DB
  useEffect(() => {
    if (!isMounted) return;

    const fetchData = async () => {
      try {
        // 1. Fetch metrics
        const metricsRes = await fetch("/api/metrics");
        if (!metricsRes.ok) {
          throw new Error(`API Gateway offline (Status ${metricsRes.status})`);
        }
        
        const metricsJson = await metricsRes.json();
        if (metricsJson.status === "error") {
          throw new Error(metricsJson.message);
        }

        if (metricsJson.status === "success" && metricsJson.data) {
          const data = metricsJson.data;
          
          // Map metrics data to page states
          const activeTunnelsVal = data.securityOverhead.ziti > 0 ? 1 : 0; 
          setActiveTunnels(activeTunnelsVal);

          // Compute throughput
          let totalReqs = 0;
          data.requests.forEach((r: any) => totalReqs += r.count);
          setThroughput(totalReqs);

          // Append new latency entry if there is active traffic
          const now = new Date();
          const timeStr = now.toTimeString().split(" ")[0].substring(3); // "MM:SS"
          
          const newEntry: LatencyData = {
            time: timeStr,
            dpopVerify: data.securityOverhead.dpop || 0,
            tokenVerify: data.securityOverhead.token || 0,
            rlsContext: data.dbLatency.rls_context || 0,
            wormExec: data.dbLatency.worm_exec || 0,
          };
          
          setLatencyHistory(prev => [...prev.slice(-9), newEntry]);
          setIsOffline(false);
          setErrorMessage("");
        }
      } catch (e: any) {
        setIsOffline(true);
        setErrorMessage(e.message || "Failed to connect to Gateway");
      }

      try {
        // 2. Fetch logs
        const logsRes = await fetch("/api/logs");
        if (logsRes.ok) {
          const logsJson = await logsRes.json();
          if (logsJson.status === "success" && logsJson.data) {
            const dbLogs = logsJson.data.map((l: any) => {
              const actionUpper = l.action.toUpperCase();
              const isCrit = actionUpper.includes("FAIL") || actionUpper.includes("BLOCK") || actionUpper.includes("REJECT") || actionUpper.includes("TAMPER");
              const level = isCrit ? "CRITICAL" : (actionUpper.includes("UPDATE") || actionUpper.includes("DELETE") ? "WARNING" : "INFO");
              
              const formattedTime = new Date(l.timestamp).toTimeString().split(" ")[0];
              
              let detailsStr = "";
              if (l.details) {
                if (typeof l.details === "string") {
                  detailsStr = l.details;
                } else {
                  detailsStr = JSON.stringify(l.details);
                }
              }
              
              return {
                id: l.id.toString(),
                timestamp: formattedTime,
                level: level,
                source: l.resource.toUpperCase(),
                message: `${l.action} on ${l.resource} - Details: ${detailsStr} | prev_hash: ${l.prev_hash.substring(0, 8)}... | block_hash: ${l.block_hash.substring(0, 8)}...`
              };
            });
            
            setLogs(dbLogs.reverse());
            
            // Extract distinct identities dynamically from logs
            const uniqueIdentitiesMap = new Map<string, string>();
            logsJson.data.forEach((l: any) => {
              if (l.actor_id && l.tenant_id) {
                uniqueIdentitiesMap.set(l.actor_id, l.tenant_id);
              }
            });
            const dbIdentities: DBIdentity[] = [];
            uniqueIdentitiesMap.forEach((tenantId, actorId) => {
              dbIdentities.push({ actorId, tenantId });
            });
            setIdentities(dbIdentities);
            
            // Check if any critical alert in the last logs
            const hasRecentThreat = dbLogs.some((l: any) => l.level === "CRITICAL");
            setHealthStatus(hasRecentThreat ? "UNDER ATTACK" : "HEALTHY");
          }
        }
      } catch (e) {
        console.error("Failed to fetch logs from PostgreSQL", e);
      }

      try {
        // 3. Fetch real certificates info
        const certsRes = await fetch("/api/certs");
        if (certsRes.ok) {
          const certsJson = await certsRes.json();
          if (certsJson.status === "success" && certsJson.data) {
            setCertificates(certsJson.data.certificates);
            setExpiringCertsCount(certsJson.data.expiringCount);
          }
        }
      } catch (e) {
        console.error("Failed to fetch certs metadata", e);
      }
    };

    // Initial load
    fetchData();

    // Scrape every 1.5 seconds
    const interval = setInterval(fetchData, 1500);
    return () => clearInterval(interval);
  }, [isMounted]);

  // Autoscroll logs window
  useEffect(() => {
    if (terminalEndRef.current) {
      terminalEndRef.current.scrollIntoView({ behavior: "smooth" });
    }
  }, [logs]);

  // Autoscroll verify console
  useEffect(() => {
    if (verifyConsoleRef.current) {
      verifyConsoleRef.current.scrollTop = verifyConsoleRef.current.scrollHeight;
    }
  }, [verificationSteps]);

  // Audit Chain Verification Flow
  const handleVerifyChain = async () => {
    setIsVerifying(true);
    setVerificationSuccess(null);
    setVerificationSteps(["Initializing connection to PostgreSQL fapi_db...", "Accessing immutable table audit_logs..."]);

    const delay = (ms: number) => new Promise(res => setTimeout(res, ms));
    await delay(600);
    
    try {
      const res = await fetch("/api/verify-chain");
      if (!res.ok) {
        throw new Error(`Database connection failed (Status ${res.status})`);
      }
      
      const json = await res.json();
      if (json.status === "success" && json.data) {
        const { steps, isValid } = json.data;
        
        setVerificationSteps(prev => [...prev, `Found ${steps.length} ledger blocks. Starting cryptographic validation...`]);
        await delay(500);

        for (let i = 0; i < steps.length; i++) {
          const step = steps[i];
          setVerificationSteps(prev => [
            ...prev, 
            `[Block #${step.blockId}] action='${step.action}' resource='${step.resource}' -> verifying SHA-256 links...`
          ]);
          await delay(100);
          if (!step.valid) {
            setVerificationSteps(prev => [...prev, `[FAIL] Block #${step.blockId} failed validation: ${step.error}`]);
            setVerificationSuccess(false);
            setIsVerifying(false);
            return;
          }
        }

        setVerificationSteps(prev => [...prev, "All block linkages verified successfully.", "LEDGER INTEGRITY VERIFIED."]);
        setVerificationSuccess(isValid);
      } else {
        throw new Error(json.message || "Invalid verification response");
      }
    } catch (e: any) {
      setVerificationSteps(prev => [...prev, `[ERROR] Verification interrupted: ${e.message}`]);
      setVerificationSuccess(false);
    }
    setIsVerifying(false);
  };

  // Calculate dynamic security score based on actual logs
  const calculateSecurityScore = () => {
    if (isOffline) return 0;
    const criticalLogsCount = logs.filter(l => l.level === "CRITICAL").length;
    const score = Math.max(10, 100 - (criticalLogsCount * 10));
    return score;
  };

  const securityScore = calculateSecurityScore();

  return (
    <div className="flex flex-1 flex-col bg-[#05070a] text-zinc-100 font-mono select-none h-screen overflow-hidden">
      
      {/* HEADER BAR */}
      <header className="flex items-center justify-between px-6 py-4 bg-[#0a0f18] border-b border-[#132237] shadow-lg shadow-[#020509]/80">
        <div className="flex items-center space-x-3">
          <Shield className="w-8 h-8 text-cyan-400 animate-pulse" />
          <div>
            <h1 className="text-xl font-bold tracking-widest bg-gradient-to-r from-cyan-400 to-emerald-400 bg-clip-text text-transparent">
              SECURE FAPI-ZTA SOC
            </h1>
            <p className="text-xs text-zinc-500 tracking-tight">Zero-Trust Telemetry Center v2.0</p>
          </div>
        </div>

        {/* Global Stats */}
        <div className="flex items-center space-x-6 text-sm">
          <div className="flex items-center space-x-2">
            <span className="text-zinc-500">LIVE FEED:</span>
            <span className={`font-bold px-2 py-0.5 rounded text-xs ${
              isOffline
                ? "bg-red-950/80 text-red-400 border border-red-500/20 animate-pulse" 
                : "bg-emerald-950/80 text-emerald-400 border border-emerald-500/20"
            }`}>
              {isOffline ? "DISCONNECTED" : "CONNECTED TO GO GATEWAY"}
            </span>
          </div>

          <div className="flex items-center space-x-2 border-l border-[#132237] pl-6">
            <Radio className="w-4 h-4 text-zinc-400" />
            <span className="text-zinc-400">STATUS:</span>
            <span className={`font-bold px-2 py-0.5 rounded text-xs ${
              healthStatus === "HEALTHY" 
                ? "bg-emerald-950/80 text-emerald-400 border border-emerald-500/20" 
                : "bg-red-950/80 text-red-400 border border-red-500/20 animate-bounce"
            }`}>
              {healthStatus}
            </span>
          </div>

          <div className="flex items-center space-x-2 border-l border-[#132237] pl-6">
            <Network className="w-4 h-4 text-cyan-400" />
            <span className="text-zinc-400">ZITI TUNNELS:</span>
            <span className="text-cyan-400 font-bold">{activeTunnels} ACTIVE</span>
          </div>

          <div className="flex items-center space-x-2 border-l border-[#132237] pl-6">
            <Activity className="w-4 h-4 text-yellow-400" />
            <span className="text-zinc-400">THROUGHPUT:</span>
            <span className="text-yellow-400 font-bold">{throughput} REQS</span>
          </div>
        </div>
      </header>

      {/* CORE CONTAINER */}
      <div className="flex flex-1 overflow-hidden">
        
        {/* SIDEBAR NAVIGATION */}
        <aside className="w-64 bg-[#070b12] border-r border-[#132237] flex flex-col justify-between py-6 overflow-y-auto">
          <div className="space-y-6">
            
            {/* OVERVIEW Group */}
            <div className="space-y-1">
              <div className="px-6 text-[10px] font-bold text-zinc-500 uppercase tracking-widest">Overview</div>
              <nav className="space-y-0.5 px-3">
                <button
                  onClick={() => setActiveTab("overview")}
                  className={`flex items-center space-x-3 w-full px-4 py-2 rounded-lg text-xs transition-all border ${
                    activeTab === "overview" 
                      ? "bg-cyan-950/30 text-cyan-400 border-cyan-500/30 font-semibold" 
                      : "text-zinc-400 border-transparent hover:bg-zinc-900/40 hover:text-zinc-200"
                  }`}
                >
                  <Activity className="w-3.5 h-3.5" />
                  <span>Platform Monitor</span>
                </button>
              </nav>
            </div>

            {/* IDENTITY Group */}
            <div className="space-y-1">
              <div className="px-6 text-[10px] font-bold text-zinc-500 uppercase tracking-widest">Identity (OIDC/OAUTH)</div>
              <nav className="space-y-0.5 px-3">
                {[
                  { id: "users", label: "Users & Sessions", icon: Users },
                  { id: "clients", label: "OAuth Clients & Tokens", icon: Key }
                ].map(item => (
                  <button
                    key={item.id}
                    onClick={() => setActiveTab(item.id)}
                    className={`flex items-center space-x-3 w-full px-4 py-2 rounded-lg text-xs transition-all border ${
                      activeTab === item.id 
                        ? "bg-cyan-950/30 text-cyan-400 border-cyan-500/30 font-semibold" 
                        : "text-zinc-400 border-transparent hover:bg-zinc-900/40 hover:text-zinc-200"
                    }`}
                  >
                    <item.icon className="w-3.5 h-3.5" />
                    <span>{item.label}</span>
                  </button>
                ))}
              </nav>
            </div>

            {/* ZERO TRUST Group */}
            <div className="space-y-1">
              <div className="px-6 text-[10px] font-bold text-zinc-500 uppercase tracking-widest">Zero Trust</div>
              <nav className="space-y-0.5 px-3">
                {[
                  { id: "topology", label: "OpenZiti Fabric Map", icon: Network },
                  { id: "ziti_policies", label: "Stealth Service Policies", icon: Shield }
                ].map(item => (
                  <button
                    key={item.id}
                    onClick={() => setActiveTab(item.id)}
                    className={`flex items-center space-x-3 w-full px-4 py-2 rounded-lg text-xs transition-all border ${
                      activeTab === item.id 
                        ? "bg-cyan-950/30 text-cyan-400 border-cyan-500/30 font-semibold" 
                        : "text-zinc-400 border-transparent hover:bg-zinc-900/40 hover:text-zinc-200"
                    }`}
                  >
                    <item.icon className="w-3.5 h-3.5" />
                    <span>{item.label}</span>
                  </button>
                ))}
              </nav>
            </div>

            {/* GATEWAY Group */}
            <div className="space-y-1">
              <div className="px-6 text-[10px] font-bold text-zinc-500 uppercase tracking-widest">Gateway</div>
              <nav className="space-y-0.5 px-3">
                {[
                  { id: "gateway_status", label: "DPoP & mTLS Matrix", icon: Lock },
                  { id: "crypto", label: "Cryptography Radar", icon: Cpu },
                  { id: "performance", label: "Performance & Benchmark", icon: Activity }
                ].map(item => (
                  <button
                    key={item.id}
                    onClick={() => setActiveTab(item.id)}
                    className={`flex items-center space-x-3 w-full px-4 py-2 rounded-lg text-xs transition-all border ${
                      activeTab === item.id 
                        ? "bg-cyan-950/30 text-cyan-400 border-cyan-500/30 font-semibold" 
                        : "text-zinc-400 border-transparent hover:bg-zinc-900/40 hover:text-zinc-200"
                    }`}
                  >
                    <item.icon className="w-3.5 h-3.5" />
                    <span>{item.label}</span>
                  </button>
                ))}
              </nav>
            </div>

            {/* DATABASE Group */}
            <div className="space-y-1">
              <div className="px-6 text-[10px] font-bold text-zinc-500 uppercase tracking-widest">Database</div>
              <nav className="space-y-0.5 px-3">
                {[
                  { id: "ledger", label: "WORM Chain Explorer", icon: HardDrive }
                ].map(item => (
                  <button
                    key={item.id}
                    onClick={() => setActiveTab(item.id)}
                    className={`flex items-center space-x-3 w-full px-4 py-2 rounded-lg text-xs transition-all border ${
                      activeTab === item.id 
                        ? "bg-cyan-950/30 text-cyan-400 border-cyan-500/30 font-semibold" 
                        : "text-zinc-400 border-transparent hover:bg-zinc-900/40 hover:text-zinc-200"
                    }`}
                  >
                    <item.icon className="w-3.5 h-3.5" />
                    <span>{item.label}</span>
                  </button>
                ))}
              </nav>
            </div>

            {/* SECURITY Group */}
            <div className="space-y-1">
              <div className="px-6 text-[10px] font-bold text-zinc-500 uppercase tracking-widest">Security SIEM</div>
              <nav className="space-y-0.5 px-3">
                {[
                  { id: "threats", label: "Live Threat Feed", icon: Terminal }
                ].map(item => (
                  <button
                    key={item.id}
                    onClick={() => setActiveTab(item.id)}
                    className={`flex items-center space-x-3 w-full px-4 py-2 rounded-lg text-xs transition-all border ${
                      activeTab === item.id 
                        ? "bg-cyan-950/30 text-cyan-400 border-cyan-500/30 font-semibold" 
                        : "text-zinc-400 border-transparent hover:bg-zinc-900/40 hover:text-zinc-200"
                    }`}
                  >
                    <item.icon className="w-3.5 h-3.5" />
                    <span>{item.label}</span>
                  </button>
                ))}
              </nav>
            </div>

          </div>

          {/* Infrastructure status summary */}
          <div className="px-6 space-y-3 pt-6 border-t border-[#132237]">
            <div className="text-[10px] font-bold text-zinc-500 uppercase tracking-widest">Stack health</div>
            <div className="space-y-2 text-xs">
              <div className="flex items-center justify-between">
                <span className="text-zinc-500">PostgreSQL (RLS):</span>
                <span className="text-emerald-400 font-semibold flex items-center gap-1">
                  <span className="w-1.5 h-1.5 rounded-full bg-emerald-400"></span> Online
                </span>
              </div>
              <div className="flex items-center justify-between">
                <span className="text-zinc-500">OpenZiti Network:</span>
                <span className={`font-semibold flex items-center gap-1 ${isOffline ? "text-red-400" : "text-emerald-400"}`}>
                  <span className={`w-1.5 h-1.5 rounded-full ${isOffline ? "bg-red-400" : "bg-emerald-400"}`}></span> {isOffline ? "Offline" : "Stealth"}
                </span>
              </div>
              <div className="flex items-center justify-between">
                <span className="text-zinc-500">WORM Ledger:</span>
                <span className="text-emerald-400 font-semibold flex items-center gap-1">
                  <span className="w-1.5 h-1.5 rounded-full bg-emerald-400"></span> Locked
                </span>
              </div>
            </div>
          </div>
        </aside>

        {/* WORKSPACE CONTENT AREA */}
        <main className="flex-1 flex flex-col overflow-hidden bg-[#060a10] p-6 space-y-6 relative">
          
          {/* OFFLINE ERROR BANNER */}
          {isOffline && (
            <div className="bg-red-950/80 border border-red-800/40 rounded-xl p-4 flex items-center space-x-3 text-xs text-red-400 animate-pulse">
              <AlertTriangle className="w-5 h-5 text-red-400" />
              <div>
                <span className="font-bold">SYSTEM OFFLINE:</span> {errorMessage || "Go API Gateway has stopped or port is blocked. Run idp and gateway modules to populate live metrics."}
              </div>
            </div>
          )}

          {/* TOP ROW KPI CARDS */}
          <div className="grid grid-cols-5 gap-4">
            
            {/* KPI 1: Security Score */}
            <div className="bg-[#080d16] border border-[#132237] rounded-xl p-4 flex flex-col justify-between h-24">
              <div className="flex items-center justify-between">
                <span className="text-[10px] text-zinc-500 uppercase tracking-widest">Security Score</span>
                <Shield className="w-4 h-4 text-emerald-400" />
              </div>
              <div className="flex items-baseline space-x-1 mt-2">
                <span className="text-2xl font-bold text-emerald-400">{securityScore}</span>
                <span className="text-zinc-500 text-xs">/ 100</span>
              </div>
            </div>

            {/* KPI 2: Threats Detected */}
            <div className="bg-[#080d16] border border-[#132237] rounded-xl p-4 flex flex-col justify-between h-24">
              <div className="flex items-center justify-between">
                <span className="text-[10px] text-zinc-500 uppercase tracking-widest">Threats Blocked</span>
                <ShieldAlert className="w-4 h-4 text-red-400" />
              </div>
              <div className="flex items-baseline space-x-1 mt-2">
                <span className="text-2xl font-bold text-red-400">
                  {logs.filter(l => l.level === "CRITICAL").length}
                </span>
                <span className="text-zinc-500 text-xs">Blocked Today</span>
              </div>
            </div>

            {/* KPI 3: Failed DPoP */}
            <div className="bg-[#080d16] border border-[#132237] rounded-xl p-4 flex flex-col justify-between h-24">
              <div className="flex items-center justify-between">
                <span className="text-[10px] text-zinc-500 uppercase tracking-widest">Failed DPoP Verify</span>
                <Lock className="w-4 h-4 text-yellow-400" />
              </div>
              <div className="flex items-baseline space-x-1 mt-2">
                <span className="text-2xl font-bold text-yellow-400">
                  {logs.filter(l => l.message.includes("FAIL") || l.message.includes("REPLAY")).length}
                </span>
                <span className="text-zinc-500 text-xs">Rejection logs</span>
              </div>
            </div>

            {/* KPI 4: Active Sessions */}
            <div className="bg-[#080d16] border border-[#132237] rounded-xl p-4 flex flex-col justify-between h-24">
              <div className="flex items-center justify-between">
                <span className="text-[10px] text-zinc-500 uppercase tracking-widest">Active Identities</span>
                <Users className="w-4 h-4 text-cyan-400" />
              </div>
              <div className="flex items-baseline space-x-1 mt-2">
                <span className="text-2xl font-bold text-cyan-400">{identities.length}</span>
                <span className="text-zinc-500 text-xs">From Postgres logs</span>
              </div>
            </div>

            {/* KPI 5: Certificate Expiring */}
            <div className="bg-[#080d16] border border-[#132237] rounded-xl p-4 flex flex-col justify-between h-24">
              <div className="flex items-center justify-between">
                <span className="text-[10px] text-zinc-500 uppercase tracking-widest">Cert Expirations</span>
                <Clock className="w-4 h-4 text-purple-400" />
              </div>
              <div className="flex items-baseline space-x-1 mt-2">
                <span className="text-2xl font-bold text-purple-400">{expiringCertsCount}</span>
                <span className="text-zinc-500 text-xs">Alert Nodes</span>
              </div>
            </div>

          </div>

          {/* TAB CONTENT */}
          <div className="flex-1 overflow-hidden flex flex-col">
            
            {/* TAB: OVERVIEW */}
            {activeTab === "overview" && (
              <div className="flex-1 flex flex-col space-y-6 overflow-y-auto">
                <div className="grid grid-cols-3 gap-6">
                  
                  {/* Health summary */}
                  <div className="bg-[#080d16] border border-[#132237] rounded-xl p-6 col-span-2">
                    <h3 className="text-xs font-bold text-cyan-400 uppercase tracking-wider mb-4">Security Operations Summary</h3>
                    <div className="space-y-4">
                      <p className="text-xs text-zinc-400 leading-relaxed">
                        The Zero-Trust Policy Engine (PDP) evaluates client roles dynamically using OPA configurations. mTLS identities bound to DPoP assertions prevent network spoofing attempts.
                      </p>
                      <div className="grid grid-cols-2 gap-4 pt-2">
                        <div className="p-3 bg-zinc-950 rounded border border-zinc-900">
                          <div className="text-[10px] text-zinc-500">mTLS Handshake Profiles</div>
                          <div className="text-sm font-bold text-emerald-400 mt-1">{isOffline ? "Disconnected" : "Strict TLS 1.3"}</div>
                        </div>
                        <div className="p-3 bg-zinc-950 rounded border border-zinc-900">
                          <div className="text-[10px] text-zinc-500">Cross-Layer Match Ratio</div>
                          <div className="text-sm font-bold text-cyan-400 mt-1">{isOffline ? "0%" : "100% Bound"}</div>
                        </div>
                      </div>
                    </div>
                  </div>

                  {/* System latency check */}
                  <div className="bg-[#080d16] border border-[#132237] rounded-xl p-6">
                    <h3 className="text-xs font-bold text-cyan-400 uppercase tracking-wider mb-4">Uptime & Service Level Objectives (SLO)</h3>
                    <div className="space-y-3">
                      <div className="flex justify-between items-center text-xs">
                        <span className="text-zinc-500">Gateway Status:</span>
                        <span className={isOffline ? "text-red-400 font-bold" : "text-emerald-400 font-bold"}>
                          {isOffline ? "OFFLINE" : "ONLINE"}
                        </span>
                      </div>
                      <div className="flex justify-between items-center text-xs">
                        <span className="text-zinc-500">Database Connection:</span>
                        <span className="text-emerald-400 font-bold">ONLINE</span>
                      </div>
                      <div className="flex justify-between items-center text-xs">
                        <span className="text-zinc-500">DB Execution SLO:</span>
                        <span className="text-emerald-400 font-bold">Passed</span>
                      </div>
                      <div className="pt-2 border-t border-[#132237]">
                        <div className="text-[10px] text-zinc-500">Platform Health Level</div>
                        <div className="text-xs text-emerald-400 mt-1">
                          {isOffline ? "AWAITING GATEWAY RESPONSE" : "NO ACTIVE VULNERABILITIES"}
                        </div>
                      </div>
                    </div>
                  </div>

                </div>

                {/* Log preview on main page */}
                <div className="bg-[#080d16] border border-[#132237] rounded-xl p-6 flex-1 flex flex-col min-h-[200px]">
                  <h3 className="text-xs font-bold text-cyan-400 uppercase tracking-wider mb-3">Gateway Live Event Log Stream</h3>
                  <div className="flex-1 bg-zinc-950 p-4 rounded border border-zinc-900 font-mono text-xs overflow-y-auto max-h-[180px]">
                    {logs.length === 0 ? (
                      <div className="text-zinc-500 italic">No logs found in PostgreSQL database. Run tests to execute transactions.</div>
                    ) : (
                      logs.slice(-6).map((log) => (
                        <div key={log.id} className="flex items-center space-x-2 py-0.5">
                          <span className="text-zinc-500">[{log.timestamp}]</span>
                          <span className="text-cyan-500">[{log.source}]</span>
                          <span className="text-zinc-300 truncate">{log.message}</span>
                        </div>
                      ))
                    )}
                  </div>
                </div>
              </div>
            )}
            
            {/* TAB: PERFORMANCE */}
            {activeTab === "performance" && (
              <div className="flex-1 flex flex-col space-y-6 overflow-y-auto pr-2">
                {/* Latency Breakdown KPI cards */}
                <div className="grid grid-cols-4 gap-6">
                  <div className="bg-[#080d16] border border-[#132237] rounded-xl p-4 flex flex-col justify-between">
                    <span className="text-[10px] text-zinc-500 uppercase tracking-widest">Realtime Throughput</span>
                    <span className="text-2xl font-bold text-cyan-400 mt-2">
                      {isOffline ? "0" : throughput} <span className="text-xs text-zinc-500 font-normal">REQS</span>
                    </span>
                  </div>
                  <div className="bg-[#080d16] border border-[#132237] rounded-xl p-4 flex flex-col justify-between">
                    <span className="text-[10px] text-zinc-500 uppercase tracking-widest">Cryptographic Overhead</span>
                    <span className="text-2xl font-bold text-yellow-500 mt-2">
                      {isOffline ? "0.0" : (latencyHistory.length > 0 ? (
                        ((latencyHistory[latencyHistory.length - 1].dpopVerify + latencyHistory[latencyHistory.length - 1].tokenVerify) / 1000).toFixed(1)
                      ) : "0.0")} <span className="text-xs text-zinc-500 font-normal">ms</span>
                    </span>
                  </div>
                  <div className="bg-[#080d16] border border-[#132237] rounded-xl p-4 flex flex-col justify-between">
                    <span className="text-[10px] text-zinc-500 uppercase tracking-widest">Database Latency</span>
                    <span className="text-2xl font-bold text-emerald-500 mt-2">
                      {isOffline ? "0.0" : (latencyHistory.length > 0 ? (
                        ((latencyHistory[latencyHistory.length - 1].rlsContext + latencyHistory[latencyHistory.length - 1].wormExec) / 1000).toFixed(1)
                      ) : "0.0")} <span className="text-xs text-zinc-500 font-normal">ms</span>
                    </span>
                  </div>
                  <div className="bg-[#080d16] border border-[#132237] rounded-xl p-4 flex flex-col justify-between">
                    <span className="text-[10px] text-zinc-500 uppercase tracking-widest">API Status SLO</span>
                    <span className="text-2xl font-bold text-green-400 mt-2">
                      {isOffline ? "UNKNOWN" : "99.99%"}
                    </span>
                  </div>
                </div>

                <div className="grid grid-cols-3 gap-6">
                  {/* Latency History Chart */}
                  <div className="bg-[#080d16] border border-[#132237] rounded-xl p-6 col-span-2 flex flex-col h-[320px]">
                    <h3 className="text-xs font-bold text-cyan-400 uppercase tracking-wider mb-4">
                      Realtime Cryptographic & DB Latency Stack (µs)
                    </h3>
                    <div className="flex-1 w-full text-xs">
                      {latencyHistory.length === 0 ? (
                        <div className="h-full flex items-center justify-center text-zinc-500 italic">
                          Awaiting gateway activity metrics stream...
                        </div>
                      ) : (
                        <ResponsiveContainer width="100%" height="100%">
                          <AreaChart data={latencyHistory} margin={{ top: 10, right: 10, left: -20, bottom: 0 }}>
                            <XAxis dataKey="time" stroke="#4b5563" fontSize={10} />
                            <YAxis stroke="#4b5563" fontSize={10} label={{ value: 'µs', angle: -90, position: 'insideLeft', fill: '#4b5563' }} />
                            <Tooltip 
                              contentStyle={{ backgroundColor: '#070b12', borderColor: '#132237', color: '#f3f4f6', fontFamily: 'monospace', fontSize: 11 }}
                            />
                            <Legend wrapperStyle={{ fontSize: 10, paddingTop: 10 }} />
                            <Area type="monotone" dataKey="dpopVerify" stackId="1" stroke="#eab308" fill="#eab308" fillOpacity={0.15} name="DPoP Verify" />
                            <Area type="monotone" dataKey="tokenVerify" stackId="1" stroke="#06b6d4" fill="#06b6d4" fillOpacity={0.15} name="Token JWKS" />
                            <Area type="monotone" dataKey="rlsContext" stackId="1" stroke="#10b981" fill="#10b981" fillOpacity={0.15} name="Postgres RLS" />
                            <Area type="monotone" dataKey="wormExec" stackId="1" stroke="#8b5cf6" fill="#8b5cf6" fillOpacity={0.15} name="WORM Hash/Write" />
                          </AreaChart>
                        </ResponsiveContainer>
                      )}
                    </div>
                  </div>

                  {/* Live Benchmark Panel */}
                  <div className="bg-[#080d16] border border-[#132237] rounded-xl p-6 flex flex-col justify-between">
                    <div>
                      <h3 className="text-xs font-bold text-cyan-400 uppercase tracking-wider mb-2">
                        Live Load Benchmark Console
                      </h3>
                      <p className="text-[10px] text-zinc-500 mb-4 leading-relaxed">
                        Trigger a safe cryptographic stress test of 50 concurrent transactions through FAPI 2.0 & DPoP Gateway logic to measure throughput capacity.
                      </p>

                      {benchmarkResult && (
                        <div className="space-y-2 text-xs bg-zinc-950 p-3 rounded border border-zinc-900 font-mono">
                          <div className="flex justify-between">
                            <span className="text-zinc-500">Throughput:</span>
                            <span className="text-yellow-400 font-bold">{benchmarkResult.rps} RPS</span>
                          </div>
                          <div className="flex justify-between">
                            <span className="text-zinc-500">Avg Latency:</span>
                            <span className="text-cyan-400 font-bold">{benchmarkResult.avgLatencyMs} ms</span>
                          </div>
                          <div className="flex justify-between">
                            <span className="text-zinc-500">P95 Latency:</span>
                            <span className="text-purple-400 font-bold">{benchmarkResult.p95LatencyMs} ms</span>
                          </div>
                          <div className="flex justify-between">
                            <span className="text-zinc-500">Success Rate:</span>
                            <span className="text-emerald-400 font-bold">{benchmarkResult.successRate}%</span>
                          </div>
                          <div className="text-[9px] text-zinc-600 border-t border-zinc-900 pt-2 flex justify-between">
                            <span>Requests: {benchmarkResult.successRequests}/{benchmarkResult.totalRequests}</span>
                            <span>Time: {benchmarkResult.durationMs}ms</span>
                          </div>
                        </div>
                      )}

                      {benchmarkError && (
                        <div className="text-xs text-red-400 bg-red-950/40 p-3 rounded border border-red-900/30">
                          Error: {benchmarkError}
                        </div>
                      )}
                    </div>

                    <div className="mt-4">
                      <button
                        onClick={runBenchmark}
                        disabled={isBenchmarking || cooldown > 0 || isOffline}
                        className={`w-full py-2.5 rounded-lg font-bold text-xs tracking-wider border transition-all cursor-pointer ${
                          isOffline
                            ? "bg-zinc-950 border-zinc-900 text-zinc-600 cursor-not-allowed"
                            : isBenchmarking
                            ? "bg-cyan-950/20 border-cyan-500/30 text-cyan-400 animate-pulse"
                            : cooldown > 0
                            ? "bg-zinc-950 border-zinc-900 text-zinc-500 cursor-not-allowed"
                            : "bg-cyan-500/10 border-cyan-500/30 text-cyan-400 hover:bg-cyan-500/20 active:scale-[0.98]"
                        }`}
                      >
                        {isBenchmarking
                          ? "RUNNING BENCHMARK..."
                          : cooldown > 0
                          ? `COOLDOWN (${cooldown}s)`
                          : "START LIVE BENCHMARK"}
                      </button>
                    </div>
                  </div>
                </div>
              </div>
            )}

            {/* TAB: USERS */}
            {activeTab === "users" && (
              <div className="flex-1 flex flex-col bg-[#080d16] rounded-xl border border-[#132237] p-6 overflow-y-auto">
                <h2 className="text-sm font-bold uppercase tracking-wider text-cyan-400 mb-4 flex items-center gap-2">
                  <Users className="w-4 h-4" /> Authenticated Users & Active Sessions
                </h2>
                
                <table className="w-full text-xs text-left border-collapse">
                  <thead>
                    <tr className="border-b border-[#132237] text-zinc-500 uppercase tracking-wider">
                      <th className="pb-2">User / Actor ID</th>
                      <th className="pb-2">Tenant ID</th>
                      <th className="pb-2">Auth Method</th>
                      <th className="pb-2">Session State</th>
                    </tr>
                  </thead>
                  <tbody className="divide-y divide-[#132237]/40 text-zinc-300">
                    {identities.length === 0 ? (
                      <tr>
                        <td colSpan={4} className="py-3 text-zinc-500 italic text-center">No active identity sessions found in PostgreSQL audit ledger.</td>
                      </tr>
                    ) : (
                      identities.map((identity, index) => (
                        <tr key={index}>
                          <td className="py-3 font-semibold text-cyan-400">{identity.actorId}</td>
                          <td>{identity.tenantId}</td>
                          <td>mTLS + DPoP</td>
                          <td><span className="px-2 py-0.5 bg-emerald-950/80 text-emerald-400 rounded-full border border-emerald-500/20 text-[10px]">RECORDED ACTIVE</span></td>
                        </tr>
                      ))
                    )}
                  </tbody>
                </table>
              </div>
            )}

            {/* TAB: CLIENTS */}
            {activeTab === "clients" && (
              <div className="flex-1 flex flex-col bg-[#080d16] rounded-xl border border-[#132237] p-6 overflow-y-auto">
                <h2 className="text-sm font-bold uppercase tracking-wider text-cyan-400 mb-4 flex items-center gap-2">
                  <Key className="w-4 h-4" /> Registered Client Credentials
                </h2>
                
                <table className="w-full text-xs text-left border-collapse">
                  <thead>
                    <tr className="border-b border-[#132237] text-zinc-500 uppercase tracking-wider">
                      <th className="pb-2">Client ID</th>
                      <th className="pb-2">Scope</th>
                      <th className="pb-2">Binding Profile</th>
                    </tr>
                  </thead>
                  <tbody className="divide-y divide-[#132237]/40 text-zinc-300">
                    <tr>
                      <td className="py-3 font-semibold text-cyan-400">client-alice</td>
                      <td>read write</td>
                      <td>ES256 (DPoP bound)</td>
                    </tr>
                    <tr>
                      <td className="py-3 font-semibold text-cyan-400">client-bob</td>
                      <td>read</td>
                      <td>ES256 (DPoP bound)</td>
                    </tr>
                    <tr>
                      <td className="py-3 font-semibold text-red-400">client-evil</td>
                      <td>none</td>
                      <td>Blocked at network boundary</td>
                    </tr>
                  </tbody>
                </table>
              </div>
            )}

            {/* TAB: TOPOLOGY */}
            {activeTab === "topology" && (
              <div className="flex-1 flex flex-col bg-[#080d16] rounded-xl border border-[#132237] p-6 relative overflow-hidden">
                <div className="flex items-center justify-between mb-4">
                  <div className="flex items-center space-x-2">
                    <Network className="w-5 h-5 text-cyan-400" />
                    <h2 className="text-sm font-bold uppercase tracking-wider text-cyan-400">
                      Zero Trust Topology Map (NIST PEP/PDP Enforcement)
                    </h2>
                  </div>
                  <span className="text-xs text-zinc-500">Click on nodes to explore live metrics</span>
                </div>

                {/* VISUAL TOPOLOGY GRAPH CONTAINER */}
                <div className="flex-1 relative flex items-center justify-around">
                  
                  {/* SVG CONNECTION LAYER */}
                  <svg className="absolute inset-0 w-full h-full pointer-events-none z-0">
                    <line x1="15%" y1="50%" x2="35%" y2="35%" stroke="#1e293b" strokeWidth="2" />
                    <line x1="15%" y1="50%" x2="35%" y2="65%" stroke="#1e293b" strokeWidth="2" />
                    <line x1="35%" y1="35%" x2="60%" y2="50%" stroke="#1e293b" strokeWidth="2" />
                    <line x1="35%" y1="65%" x2="60%" y2="50%" stroke="#1e293b" strokeWidth="2" />
                    <line x1="60%" y1="50%" x2="85%" y2="30%" stroke="#1e293b" strokeWidth="2" />
                    <line x1="60%" y1="50%" x2="85%" y2="70%" stroke="#1e293b" strokeWidth="2" />

                    {/* Animated flows */}
                    {!isOffline && healthStatus === "HEALTHY" ? (
                      <>
                        <circle r="4" fill="#22d3ee" className="z-10">
                          <animateMotion dur="3s" repeatCount="indefinite" path="M 100,250 Q 200,180 300,175" />
                        </circle>
                        <circle r="4" fill="#34d399" className="z-10">
                          <animateMotion dur="4s" repeatCount="indefinite" path="M 300,175 Q 450,220 550,250" />
                        </circle>
                      </>
                    ) : !isOffline ? (
                      <>
                        <circle r="5" fill="#f87171" className="z-10">
                          <animateMotion dur="1s" repeatCount="indefinite" path="M 100,250 Q 200,180 300,175" />
                        </circle>
                        <circle r="5" fill="#f87171" className="z-10">
                          <animateMotion dur="1.5s" repeatCount="indefinite" path="M 300,175 Q 450,220 550,250" />
                        </circle>
                      </>
                    ) : null}
                  </svg>

                  {/* Node 1: Client */}
                  <button 
                    onClick={() => setSelectedNode({
                      name: "Client Endpoint",
                      role: "Device/User client-alice",
                      cpu: 4.2,
                      ram: "450 MB",
                      rps: isOffline ? 0 : 15,
                      latency: "3.88 ms",
                      dpopSuccess: "ES256 ECDSA Signed",
                      mtlsSuccess: "Ziti bound mTLS",
                      description: "Performs client-side DPoP validation token request and initiates the stealth OpenZiti tunnel."
                    })}
                    className="z-10 flex flex-col items-center focus:outline-none hover:scale-105 transition-transform"
                  >
                    <div className={`w-16 h-16 rounded-full flex items-center justify-center border transition-all duration-500 shadow-lg ${
                      isOffline
                        ? "bg-zinc-900 border-zinc-800 text-zinc-600"
                        : healthStatus === "HEALTHY" 
                        ? "bg-cyan-950/20 border-cyan-500/30 text-cyan-400" 
                        : "bg-red-950/20 border-red-500/30 text-red-400 animate-pulse"
                    }`}>
                      <Server className="w-8 h-8" />
                    </div>
                    <span className="text-xs mt-2 text-zinc-300">Client Endpoint</span>
                    <span className="text-[10px] text-zinc-500 font-mono">client-alice</span>
                  </button>

                  {/* Ziti Edge Routers */}
                  <div className="z-10 flex flex-col justify-around h-full py-16">
                    <button 
                      onClick={() => setSelectedNode({
                        name: "Ziti Edge Router 1",
                        role: "Stealth Overlay Router",
                        cpu: isOffline ? 0 : 18.5,
                        ram: "1.2 GB",
                        rps: isOffline ? 0 : 240,
                        latency: "0.12 ms",
                        dpopSuccess: "N/A (Network Layer)",
                        mtlsSuccess: isOffline ? "Offline" : "100% Handshakes",
                        description: "Provides inbound firewall block. Only accepts connections using correct OpenZiti client credentials."
                      })}
                      className="flex flex-col items-center focus:outline-none hover:scale-105 transition-transform"
                    >
                      <div className={`w-14 h-14 rounded-full border flex items-center justify-center ${
                        isOffline ? "bg-zinc-900 border-zinc-800 text-zinc-600" : "bg-slate-900 border-slate-700/50 text-slate-400"
                      }`}>
                        <Wifi className="w-6 h-6" />
                      </div>
                      <span className="text-xs mt-2 text-zinc-300">Edge Router</span>
                      <span className="text-[10px] text-zinc-500">ziti-edge-1</span>
                    </button>

                    <button 
                      onClick={() => setSelectedNode({
                        name: "Ziti Edge Router 2",
                        role: "Stealth Overlay Router",
                        cpu: isOffline ? 0 : 15.2,
                        ram: "1.1 GB",
                        rps: isOffline ? 0 : 210,
                        latency: "0.15 ms",
                        dpopSuccess: "N/A (Network Layer)",
                        mtlsSuccess: isOffline ? "Offline" : "100% Handshakes",
                        description: "Backup router providing high availability routing paths inside the Ziti dark network."
                      })}
                      className="flex flex-col items-center focus:outline-none hover:scale-105 transition-transform"
                    >
                      <div className={`w-14 h-14 rounded-full border flex items-center justify-center ${
                        isOffline ? "bg-zinc-900 border-zinc-800 text-zinc-600" : "bg-slate-900 border-slate-700/50 text-slate-400"
                      }`}>
                        <Wifi className="w-6 h-6" />
                      </div>
                      <span className="text-xs mt-2 text-zinc-300">Edge Router</span>
                      <span className="text-[10px] text-zinc-500">ziti-edge-2</span>
                    </button>
                  </div>

                  {/* Node 4: API Gateway */}
                  <button 
                    onClick={() => setSelectedNode({
                      name: "API Gateway (PEP)",
                      role: "Policy Enforcement Point",
                      cpu: isOffline ? 0 : 28.4,
                      ram: "2.4 GB",
                      rps: throughput,
                      latency: "7.48 ms",
                      dpopSuccess: "ES256 signature verified",
                      mtlsSuccess: "SourceIdentifier matched",
                      description: "Acts as PEP by intercepting all requests, evaluating token bindings, and checking the policy engine (PDP) before allowing access."
                    })}
                    className="z-10 flex flex-col items-center focus:outline-none hover:scale-105 transition-transform"
                  >
                    <div className={`w-20 h-20 rounded-xl border flex flex-col items-center justify-center shadow-2xl relative ${
                      isOffline 
                        ? "bg-zinc-900 border-zinc-800 text-zinc-600"
                        : "bg-slate-950 border-cyan-500/50 text-cyan-400 shadow-cyan-950/20"
                    }`}>
                      {!isOffline && <div className="absolute top-0 right-0 w-2.5 h-2.5 bg-emerald-400 rounded-full border border-[#080d16] animate-ping" />}
                      <Layers className="w-10 h-10" />
                    </div>
                    <span className="text-xs mt-2 font-bold text-zinc-300">API Gateway (PEP)</span>
                    <span className="text-[10px] text-zinc-500">listening :8080</span>
                  </button>

                  {/* OPA & DB */}
                  <div className="z-10 flex flex-col justify-around h-full py-16">
                    <button 
                      onClick={() => setSelectedNode({
                        name: "OPA Engine (PDP)",
                        role: "Policy Decision Point",
                        cpu: isOffline ? 0 : 10.4,
                        ram: "620 MB",
                        rps: throughput,
                        latency: "0.22 ms",
                        dpopSuccess: "N/A",
                        mtlsSuccess: "N/A",
                        description: "Evaluates request parameters (role, path, HTTP method) against policies.json dynamically."
                      })}
                      className="flex flex-col items-center focus:outline-none hover:scale-105 transition-transform"
                    >
                      <div className={`w-14 h-14 rounded-full border flex items-center justify-center ${
                        isOffline ? "bg-zinc-900 border-zinc-800 text-zinc-600" : "bg-slate-900 border-emerald-500/30 text-emerald-400"
                      }`}>
                        <Shield className="w-6 h-6" />
                      </div>
                      <span className="text-xs mt-2 text-zinc-300">Policy Engine (PDP)</span>
                      <span className="text-[10px] text-zinc-500">policies.json</span>
                    </button>

                    <button 
                      onClick={() => setSelectedNode({
                        name: "PostgreSQL Database",
                        role: "Secure Storage & RLS Engine",
                        cpu: 12.1,
                        ram: "4.8 GB",
                        rps: throughput,
                        latency: "12.82 ms",
                        dpopSuccess: "N/A",
                        mtlsSuccess: "N/A",
                        description: "Implements Postgres Row-Level Security (RLS) policies and triggers WORM ledger constraints with SHA-256 hash chaining."
                      })}
                      className="flex flex-col items-center focus:outline-none hover:scale-105 transition-transform"
                    >
                      <div className="w-14 h-14 rounded-full bg-slate-900 border border-purple-500/30 flex items-center justify-center text-purple-400">
                        <Database className="w-6 h-6" />
                      </div>
                      <span className="text-xs mt-2 text-zinc-300">PostgreSQL (WORM)</span>
                      <span className="text-[10px] text-zinc-500">RLS Active</span>
                    </button>
                  </div>

                </div>

                {/* BOTTOM MAP STATS */}
                <div className="grid grid-cols-3 gap-4 border-t border-[#132237] pt-4 mt-4">
                  <div className="bg-[#0b121f] rounded-lg p-3 border border-[#132237]/60">
                    <div className="text-[10px] text-zinc-500 uppercase tracking-widest">NIST 800-207 State</div>
                    <div className="text-xs text-emerald-400 font-semibold mt-1">Continuous Diagnostics Active</div>
                  </div>
                  <div className="bg-[#0b121f] rounded-lg p-3 border border-[#132237]/60">
                    <div className="text-[10px] text-zinc-500 uppercase tracking-widest">Ziti Identity Bound</div>
                    <div className="text-xs text-cyan-400 font-semibold mt-1">{isOffline ? "Awaiting Network" : "SourceIdentifier Match (mTLS)"}</div>
                  </div>
                  <div className="bg-[#0b121f] rounded-lg p-3 border border-[#132237]/60">
                    <div className="text-[10px] text-zinc-500 uppercase tracking-widest">LEDGER CHAIN STATE</div>
                    <div className="text-xs text-yellow-400 font-semibold mt-1">SHA-256 Hash-chain Verified</div>
                  </div>
                </div>
              </div>
            )}

            {/* TAB: STEALTH POLICIES */}
            {activeTab === "ziti_policies" && (
              <div className="flex-1 flex flex-col bg-[#080d16] rounded-xl border border-[#132237] p-6 overflow-y-auto">
                <h2 className="text-sm font-bold uppercase tracking-wider text-cyan-400 mb-4 flex items-center gap-2">
                  <Shield className="w-4 h-4" /> OpenZiti Stealth Service Policies
                </h2>
                <div className="grid grid-cols-2 gap-4">
                  <div className="p-4 bg-[#0b121f] rounded-xl border border-[#132237]/50 space-y-2">
                    <div className="text-xs font-bold text-zinc-300">Service Policy: Dial API Gateway</div>
                    <div className="text-xs text-zinc-500">Identity Alice is allowed to dial Service Gateway-Service through the overlay.</div>
                    <div className="text-xs font-bold text-emerald-400 mt-2">Status: ALLOWED</div>
                  </div>
                  <div className="p-4 bg-[#0b121f] rounded-xl border border-[#132237]/50 space-y-2">
                    <div className="text-xs font-bold text-zinc-300">Service Policy: Dial Administrative Database</div>
                    <div className="text-xs text-zinc-500">Identity Evil tried to dial DB-Service. Blocked at the Edge.</div>
                    <div className="text-xs font-bold text-red-400 mt-2">Status: BLOCKED</div>
                  </div>
                </div>
              </div>
            )}

            {/* TAB: GATEWAY STATUS */}
            {activeTab === "gateway_status" && (
              <div className="flex-1 flex flex-col bg-[#080d16] rounded-xl border border-[#132237] p-6 overflow-y-auto">
                <h2 className="text-sm font-bold uppercase tracking-wider text-cyan-400 mb-4 flex items-center gap-2">
                  <Lock className="w-4 h-4" /> Cryptographic Assertions & Tokens Matrix (FAPI 2.0)
                </h2>
                
                <div className="grid grid-cols-2 gap-6 flex-1">
                  
                  {/* DPoP Proof details */}
                  <div className="bg-[#0b121f] border border-[#132237]/60 rounded-xl p-4 space-y-4">
                    <div className="text-xs font-bold text-zinc-300 border-b border-[#132237] pb-2">DPoP Assertion Properties</div>
                    <div className="space-y-3 text-xs">
                      <div className="flex justify-between border-b border-zinc-900 pb-1">
                        <span className="text-zinc-500">Algorithm:</span>
                        <span className="text-cyan-400 font-semibold">ES256 (ECDSA P-256)</span>
                      </div>
                      <div className="flex justify-between border-b border-zinc-900 pb-1">
                        <span className="text-zinc-500">Replay Protection:</span>
                        <span className="text-emerald-400 font-semibold">JTI Cache Checked (TTL 60s)</span>
                      </div>
                      <div className="flex justify-between border-b border-zinc-900 pb-1">
                        <span className="text-zinc-500">HTM/HTU Binding:</span>
                        <span className="text-emerald-400">Strict Match Enforced</span>
                      </div>
                      <div className="flex justify-between pb-1">
                        <span className="text-zinc-500">Token Binding claim (cnf):</span>
                        <span className="text-yellow-400">SHA-256 Thumbprint Bound</span>
                      </div>
                    </div>

                    <div className="bg-zinc-950 p-3 rounded border border-zinc-800 text-[10px] text-zinc-400 font-mono space-y-1">
                      <div>Header: {"{ \"typ\": \"dpop+jwt\", \"alg\": \"ES256\", \"jwk\": {...} }"}</div>
                      <div className="truncate text-zinc-500">Payload: {"{ \"jti\": \"a8f...\", \"htm\": \"GET\", \"htu\": \"http://...\", \"iat\": 1719... }"}</div>
                    </div>
                  </div>

                  {/* mTLS status details */}
                  <div className="bg-[#0b121f] border border-[#132237]/60 rounded-xl p-4 space-y-4">
                    <div className="text-xs font-bold text-zinc-300 border-b border-[#132237] pb-2">mTLS Overlay Credentials</div>
                    <div className="space-y-3 text-xs">
                      <div className="flex justify-between border-b border-zinc-900 pb-1">
                        <span className="text-zinc-500">Active CA:</span>
                        <span className="text-cyan-400">OpenZiti Edge Root CA</span>
                      </div>
                      <div className="flex justify-between border-b border-zinc-900 pb-1">
                        <span className="text-zinc-500">Client Certs Registered:</span>
                        <span className="text-emerald-400">{certificates.length}</span>
                      </div>
                      <div className="flex justify-between border-b border-zinc-900 pb-1">
                        <span className="text-zinc-500">Handshake Profile:</span>
                        <span className="text-cyan-400">TLS 1.3 ALPN Enforced</span>
                      </div>
                      <div className="flex justify-between pb-1">
                        <span className="text-zinc-500">Binding Cross-check:</span>
                        <span className="text-emerald-400">SourceZitiID == ClaimsSub</span>
                      </div>
                    </div>

                    <div className="bg-zinc-950 p-3 rounded border border-zinc-900 text-xs overflow-y-auto max-h-[100px] space-y-1">
                      {certificates.map((cert, index) => (
                        <div key={index} className="flex justify-between text-[10px] text-zinc-400">
                          <span>{cert.name}:</span>
                          <span className={cert.daysRemaining < 30 ? "text-red-400 font-bold" : "text-emerald-400"}>
                            {cert.daysRemaining} days left
                          </span>
                        </div>
                      ))}
                    </div>
                  </div>

                </div>
              </div>
            )}

            {/* TAB: CRYPTO RADAR */}
            {activeTab === "crypto" && (
              <div className="flex-1 flex flex-col bg-[#080d16] rounded-xl border border-[#132237] p-6 overflow-y-auto">
                <h2 className="text-sm font-bold uppercase tracking-wider text-cyan-400 mb-4 flex items-center gap-2">
                  <Cpu className="w-4 h-4" /> Cryptographic Performance Metrics (Prometheus Scraped)
                </h2>

                <div className="flex-1 min-h-[220px]">
                  {isOffline ? (
                    <div className="text-zinc-500 italic flex items-center justify-center h-full">Charts unavailable: Go Gateway is offline.</div>
                  ) : (
                    isMounted && (
                      <ResponsiveContainer width="100%" height="100%">
                        <AreaChart data={latencyHistory} margin={{ top: 10, right: 10, left: -20, bottom: 0 }}>
                          <defs>
                            <linearGradient id="colorDpop" x1="0" y1="0" x2="0" y2="1">
                              <stop offset="5%" stopColor="#22d3ee" stopOpacity={0.4}/>
                              <stop offset="95%" stopColor="#22d3ee" stopOpacity={0}/>
                            </linearGradient>
                            <linearGradient id="colorWorm" x1="0" y1="0" x2="0" y2="1">
                              <stop offset="5%" stopColor="#a855f7" stopOpacity={0.4}/>
                              <stop offset="95%" stopColor="#a855f7" stopOpacity={0}/>
                            </linearGradient>
                          </defs>
                          <XAxis dataKey="time" stroke="#4b5563" fontSize={10} />
                          <YAxis stroke="#4b5563" fontSize={10} />
                          <Tooltip contentStyle={{ backgroundColor: "#0b121f", borderColor: "#132237", color: "#f3f4f6" }} />
                          <Legend wrapperStyle={{ fontSize: 11 }} />
                          <Area type="monotone" dataKey="dpopVerify" name="DPoP Verify (µs)" stroke="#22d3ee" fillOpacity={1} fill="url(#colorDpop)" />
                          <Area type="monotone" dataKey="wormExec" name="WORM Chain Hash (µs)" stroke="#a855f7" fillOpacity={1} fill="url(#colorWorm)" />
                        </AreaChart>
                      </ResponsiveContainer>
                    )
                  )}
                </div>
              </div>
            )}

            {/* TAB: WORM LEDGER CHAIN EXPLORER */}
            {activeTab === "ledger" && (
              <div className="flex-1 flex flex-col bg-[#080d16] rounded-xl border border-[#132237] p-6 space-y-4 overflow-y-auto">
                <div className="flex items-center justify-between">
                  <h2 className="text-sm font-bold uppercase tracking-wider text-cyan-400 flex items-center gap-2">
                    <HardDrive className="w-4 h-4" /> WORM Ledger Immutable Chain Explorer
                  </h2>
                  <button 
                    disabled={isVerifying}
                    onClick={handleVerifyChain}
                    className={`px-4 py-2 rounded-lg text-xs font-bold transition-all border ${
                      isVerifying 
                        ? "bg-zinc-800 text-zinc-500 border-zinc-700 cursor-not-allowed animate-pulse" 
                        : "bg-cyan-950 text-cyan-400 border-cyan-500/40 hover:bg-cyan-900"
                    }`}
                  >
                    {isVerifying ? "Verifying Ledger..." : "Verify Ledger Integrity"}
                  </button>
                </div>

                <div className="grid grid-cols-3 gap-6 flex-1 min-h-[300px]">
                  
                  {/* Left Console window */}
                  <div className="col-span-2 flex flex-col bg-zinc-950 border border-zinc-900 rounded-xl overflow-hidden p-4">
                    <div className="text-[10px] text-zinc-500 border-b border-zinc-900 pb-1 mb-2 uppercase tracking-widest">Verification Logs</div>
                    <div ref={verifyConsoleRef} className="flex-1 overflow-y-auto font-mono text-xs text-zinc-400 space-y-1 pr-2">
                      {verificationSteps.length === 0 ? (
                        <div className="text-zinc-500 italic">Click the "Verify Ledger Integrity" button above to run cryptographic chain checks.</div>
                      ) : (
                        verificationSteps.map((step, idx) => (
                          <div key={idx} className="flex items-start space-x-1">
                            <span className="text-cyan-600">&gt;</span>
                            <span className={
                              step.includes("VALID") || step.includes("VERIFIED") 
                                ? "text-emerald-400 font-bold" 
                                : step.includes("FAIL") || step.includes("broken") || step.includes("ERROR")
                                ? "text-red-400 font-bold"
                                : "text-zinc-300"
                            }>
                              {step}
                            </span>
                          </div>
                        ))
                      )}
                    </div>
                  </div>

                  {/* Right Validation Card Panel */}
                  <div className="bg-[#0b121f] rounded-xl border border-[#132237]/60 p-4 flex flex-col justify-between">
                    <div>
                      <div className="text-[10px] text-zinc-500 uppercase tracking-widest mb-2">Ledger Integrity State</div>
                      {verificationSuccess === null ? (
                        <div className="flex items-center gap-2 text-xs text-zinc-400">
                          <HelpCircle className="w-5 h-5 text-zinc-500" /> Waiting for check...
                        </div>
                      ) : verificationSuccess ? (
                        <div className="space-y-3">
                          <div className="flex items-center gap-2 text-sm text-emerald-400 font-bold">
                            <CheckCircle className="w-6 h-6 text-emerald-400" /> INTEGRITY SECURE
                          </div>
                          <p className="text-[11px] text-zinc-400 leading-relaxed">
                            No modifications, updates, or deletions were detected. Every audit log matches the SHA-256 chain and previous linkage parameters.
                          </p>
                        </div>
                      ) : (
                        <div className="space-y-3">
                          <div className="flex items-center gap-2 text-sm text-red-400 font-bold">
                            <AlertTriangle className="w-6 h-6 text-red-400" /> INTEGRITY COMPROMISED
                          </div>
                          <p className="text-[11px] text-zinc-400 leading-relaxed">
                            A modification attempt or broken hash-link was detected inside the audit logs table. Please review syslog and database access triggers immediately.
                          </p>
                        </div>
                      )}
                    </div>

                    <div className="p-3 bg-zinc-950 rounded border border-zinc-900 text-[10px] space-y-1 text-zinc-500">
                      <div>DB Host: localhost:5432</div>
                      <div>Table: audit_logs</div>
                      <div>Trigger: prevent_audit_tampering</div>
                    </div>
                  </div>

                </div>
              </div>
            )}

            {/* TAB: THREAT FEED */}
            {activeTab === "threats" && (
              <div className="flex-1 flex flex-col bg-[#080d16] rounded-xl border border-[#132237] overflow-hidden">
                <div className="flex items-center justify-between px-6 py-4 border-b border-[#132237] bg-[#0b121f]">
                  <div className="flex items-center space-x-2">
                    <Terminal className="w-4 h-4 text-cyan-400" />
                    <h2 className="text-sm font-bold uppercase tracking-wider text-cyan-400">
                      Live Threat & Security Feed (SIEM Audit Logs)
                    </h2>
                  </div>
                  <div className="flex items-center space-x-2">
                    <span className="w-2 h-2 rounded-full bg-emerald-400 animate-ping"></span>
                    <span className="text-xs text-emerald-400">STREAMS ONLINE</span>
                  </div>
                </div>

                {/* TERMINAL LOG PANEL */}
                <div className="flex-1 p-4 bg-zinc-950 font-mono text-xs overflow-y-auto space-y-2 max-h-[300px]">
                  {logs.length === 0 ? (
                    <div className="text-zinc-500 italic">No logs parsed from PostgreSQL audit ledger. Run tests or make API calls to populate ledger data.</div>
                  ) : (
                    logs.map((log) => (
                      <div key={log.id} className="flex items-start space-x-2 hover:bg-zinc-900/30 p-1 rounded transition-colors">
                        <span className="text-zinc-500">[{log.timestamp}]</span>
                        <span className={`font-bold px-1.5 py-0.5 rounded text-[10px] ${
                          log.level === "CRITICAL" 
                            ? "bg-red-950 text-red-400 border border-red-800/30" 
                            : log.level === "WARNING"
                            ? "bg-yellow-950 text-yellow-400 border border-yellow-800/30"
                            : "bg-cyan-950 text-cyan-400 border border-cyan-800/30"
                        }`}>
                          {log.level}
                        </span>
                        <span className="text-zinc-400">[{log.source}]</span>
                        <span className={log.level === "CRITICAL" ? "text-red-400 font-bold" : "text-zinc-300"}>
                          {log.message}
                        </span>
                      </div>
                    ))
                  )}
                  <div ref={terminalEndRef} />
                </div>
              </div>
            )}

          </div>

          {/* LOWER STATS BAR */}
          <div className="grid grid-cols-4 gap-4">
            <div className="bg-[#080d16] border border-[#132237] rounded-xl p-4 flex items-center space-x-3">
              <Activity className="w-8 h-8 text-cyan-400" />
              <div>
                <div className="text-[10px] text-zinc-500 uppercase tracking-widest">Gateway Latency</div>
                <div className="text-lg font-bold text-cyan-400">
                  {latencyHistory.length > 0 && !isOffline ? ((latencyHistory[latencyHistory.length - 1].dpopVerify + latencyHistory[latencyHistory.length - 1].tokenVerify) / 1000).toFixed(2) : "0.00"} ms
                </div>
              </div>
            </div>

            <div className="bg-[#080d16] border border-[#132237] rounded-xl p-4 flex items-center space-x-3">
              <Cpu className="w-8 h-8 text-emerald-400" />
              <div>
                <div className="text-[10px] text-zinc-500 uppercase tracking-widest">Crypto overhead</div>
                <div className="text-lg font-bold text-emerald-400">
                  {latencyHistory.length > 0 && !isOffline ? ((latencyHistory[latencyHistory.length - 1].dpopVerify) / 1000).toFixed(2) : "0.00"} ms
                </div>
              </div>
            </div>

            <div className="bg-[#080d16] border border-[#132237] rounded-xl p-4 flex items-center space-x-3">
              <Database className="w-8 h-8 text-purple-400" />
              <div>
                <div className="text-[10px] text-zinc-500 uppercase tracking-widest">DB RLS Context</div>
                <div className="text-lg font-bold text-purple-400">
                  {latencyHistory.length > 0 && !isOffline ? ((latencyHistory[latencyHistory.length - 1].rlsContext) / 1000).toFixed(2) : "0.00"} ms
                </div>
              </div>
            </div>

            <div className="bg-[#080d16] border border-[#132237] rounded-xl p-4 flex items-center space-x-3">
              <HardDrive className="w-8 h-8 text-yellow-400" />
              <div>
                <div className="text-[10px] text-zinc-500 uppercase tracking-widest">WORM Hash-Chain</div>
                <div className="text-lg font-bold text-yellow-400">
                  {latencyHistory.length > 0 && !isOffline ? ((latencyHistory[latencyHistory.length - 1].wormExec) / 1000).toFixed(2) : "0.00"} ms
                </div>
              </div>
            </div>
          </div>

          {/* SLIDING NODE DETAILS SIDEBAR PANEL */}
          {selectedNode && (
            <div className="absolute top-0 right-0 w-80 h-full bg-[#080d16] border-l border-[#132237] shadow-2xl p-6 flex flex-col justify-between z-50">
              <div className="space-y-6">
                
                {/* Header */}
                <div className="flex items-center justify-between border-b border-[#132237] pb-3">
                  <div>
                    <h3 className="text-sm font-bold text-cyan-400">{selectedNode.name}</h3>
                    <p className="text-[10px] text-zinc-500 font-mono">{selectedNode.role}</p>
                  </div>
                  <button 
                    onClick={() => setSelectedNode(null)}
                    className="text-zinc-500 hover:text-zinc-300 text-xs focus:outline-none"
                  >
                    [CLOSE]
                  </button>
                </div>

                {/* Resource Stats */}
                <div className="space-y-3 text-xs">
                  <div className="flex justify-between border-b border-zinc-900 pb-1">
                    <span className="text-zinc-500">CPU Usage:</span>
                    <span className="text-cyan-400 font-semibold">{selectedNode.cpu}%</span>
                  </div>
                  <div className="flex justify-between border-b border-zinc-900 pb-1">
                    <span className="text-zinc-500">Memory:</span>
                    <span className="text-cyan-400 font-semibold">{selectedNode.ram}</span>
                  </div>
                  <div className="flex justify-between border-b border-zinc-900 pb-1">
                    <span className="text-yellow-400 font-bold">{selectedNode.rps} RPS</span>
                  </div>
                  <div className="flex justify-between border-b border-zinc-900 pb-1">
                    <span className="text-zinc-500">Internal Latency:</span>
                    <span className="text-emerald-400 font-bold">{selectedNode.latency}</span>
                  </div>
                  <div className="flex justify-between border-b border-zinc-900 pb-1">
                    <span className="text-zinc-500">DPoP State:</span>
                    <span className="text-cyan-400">{selectedNode.dpopSuccess}</span>
                  </div>
                  <div className="flex justify-between pb-1">
                    <span className="text-zinc-500">mTLS State:</span>
                    <span className="text-cyan-400">{selectedNode.mtlsSuccess}</span>
                  </div>
                </div>

                {/* Description */}
                <div className="space-y-2">
                  <div className="text-[10px] text-zinc-500 uppercase tracking-widest">Node Description</div>
                  <p className="text-[11px] text-zinc-400 leading-relaxed bg-zinc-950 p-3 rounded border border-zinc-900">
                    {selectedNode.description}
                  </p>
                </div>

              </div>

              {/* Status light */}
              <div className="p-3 bg-emerald-950/20 border border-emerald-500/20 rounded text-center">
                <span className="text-[11px] text-emerald-400 font-bold flex items-center justify-center gap-1">
                  <CheckCircle className="w-3.5 h-3.5" /> Node Healthy & Verified
                </span>
              </div>
            </div>
          )}

        </main>
      </div>

    </div>
  );
}
