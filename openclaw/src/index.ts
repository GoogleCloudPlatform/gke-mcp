import { definePluginEntry } from "openclaw/plugin-sdk/plugin-entry";
import path from "path";
import { fileURLToPath } from "url";
import os from "os";
import fs from "fs";
import { mutateConfigFile } from "openclaw/plugin-sdk/config-mutation";
import { runExec } from "openclaw/plugin-sdk/process-runtime";

const __dirname = path.dirname(fileURLToPath(import.meta.url));

const homeDir = os.homedir();
const localBinDir = path.join(homeDir, ".local", "bin");
const gkeMcpPath = path.join(localBinDir, "gke-mcp");

// Define paths relative to the plugin root
const pluginDir = path.resolve(__dirname, "..");
const pluginAgentsDir = path.join(pluginDir, "agents", "gke-sre");

async function ensureBinaryInstalled() {
  if (fs.existsSync(gkeMcpPath)) {
    return;
  }

  console.log(`[gke-agent] Binary not found at ${gkeMcpPath}. Running installation script...`);
  const scriptPath = path.join(pluginDir, "scripts", "install-gke-mcp.sh");
  
  try {
    const { stdout, stderr } = await runExec("bash", [scriptPath], { timeoutMs: 60000 });
    if (stdout.trim()) {
      console.log(`[gke-agent] MCP Binary Installation Output:\n${stdout.trim()}`);
    }
    if (stderr && stderr.trim()) {
      console.warn(`[gke-agent] MCP Binary Installation Warnings:\n${stderr.trim()}`);
    }
  } catch (e: any) {
    const errorMsg = e.stderr || e.stdout || e.message || String(e);
    console.error(`[gke-agent] Failed to install gke-mcp binary:`, errorMsg);
    throw new Error(`gke-mcp binary installation failed: ${errorMsg}`);
  }
}

async function ensureAgentRegisteredProgrammatically() {
  const agentId = "gke-sre";
  const workspaceDir = path.join(homeDir, ".openclaw", "workspace", "agents", agentId);

  // 1. Populate workspace first (standard filesystem ops, no deadlock)
  if (fs.existsSync(pluginAgentsDir)) {
    if (!fs.existsSync(workspaceDir)) {
      fs.mkdirSync(workspaceDir, { recursive: true });
    }
    
    // Simple recursive copy function
    const copyRecursive = (srcDir: string, destDir: string) => {
      const entries = fs.readdirSync(srcDir);
      for (const entry of entries) {
        const srcPath = path.join(srcDir, entry);
        const destPath = path.join(destDir, entry);
        
        if (fs.statSync(srcPath).isDirectory()) {
          if (!fs.existsSync(destPath)) {
            fs.mkdirSync(destPath, { recursive: true });
          }
          copyRecursive(srcPath, destPath);
        } else {
          fs.copyFileSync(srcPath, destPath);
        }
      }
    };
    
    copyRecursive(pluginAgentsDir, workspaceDir);
  }

  // 2. Mutate config safely via SDK
  await mutateConfigFile({
    mutate: (config) => {
      let changed = false;

      // Ensure agents list exists
      if (!config.agents) config.agents = {};
      if (!config.agents.list) config.agents.list = [];

      // Add agent if missing
      const agentExists = config.agents.list.some(a => a.id === agentId);
      if (!agentExists) {
        config.agents.list.push({
          id: agentId,
          workspace: workspaceDir,
          identity: {
            name: "GKE Expert SRE",
            emoji: "🚢",
            theme: "blue",
          }
        });
        changed = true;
        console.log(`[gke-agent] Programmatically added agent '${agentId}' to config.`);
      }

      // Configure semantic routing via Subagents
      if (!config.agents.defaults) config.agents.defaults = {};
      if (!config.agents.defaults.subagents) config.agents.defaults.subagents = {};
      if (!config.agents.defaults.subagents.allowAgents) config.agents.defaults.subagents.allowAgents = [];
      
      const allowed = config.agents.defaults.subagents.allowAgents;
      if (!allowed.includes("*") && !allowed.includes(agentId)) {
        allowed.push(agentId);
        changed = true;
        console.log(`[gke-agent] Added '${agentId}' to allowed subagents for semantic routing.`);
      }

      // Add MCP server if missing
      if (!config.mcp) config.mcp = {};
      if (!config.mcp.servers) config.mcp.servers = {};
      
      if (!config.mcp.servers["gke-mcp"]) {
        config.mcp.servers["gke-mcp"] = {
          command: gkeMcpPath,
          args: [],
          env: { ...process.env } as any
        };
        changed = true;
        console.log(`[gke-agent] Programmatically added MCP server 'gke-mcp' to config.`);
      }

      if (!changed) {
        // No-op if already configured
        return;
      }
    }
  });
}

export default definePluginEntry({
  id: "gke-agent",
  name: "GKE Agent Plugin",
  description: "Plugin to manage GKE clusters.",
  register(api) {
    api.on("gateway_start", async () => {
      try {
        await ensureBinaryInstalled();
        await ensureAgentRegisteredProgrammatically();
      } catch (e) {
        console.error(`[gke-agent] Setup failed:`, e);
      }
    });
  },
});