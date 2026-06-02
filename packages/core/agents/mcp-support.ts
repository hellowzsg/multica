// The set of runtime providers whose backend reads `agent.mcp_config` and
// forwards MCP servers to the underlying CLI. The MCP config tab is hidden
// for every other provider so a user can't save a value the runtime will
// silently ignore. Keep this list in sync with the backends in
// `server/pkg/agent/` that read `ExecOptions.McpConfig`, plus the OpenClaw
// per-task wrapper preparer in `server/internal/daemon/execenv/` which
// materialises `mcp.servers` into the synthesised config rather than going
// through ExecOptions.
//
// Variant providers (codebuddy, claude-internal, codex-internal) inherit
// from claude / codex backends via ProtocolFamily — they get MCP wiring
// for free, and codebuddy's own CLI exposes the same `--mcp-config` flag.
// gemini-internal is intentionally absent: the gemini backend does not
// consume mcp_config, so neither does its variant.
const MCP_SUPPORTED_PROVIDERS = new Set([
  "claude",
  "claude-internal",
  "codebuddy",
  "codex",
  "codex-internal",
  "hermes",
  "kimi",
  "kiro",
  "opencode",
  "openclaw",
]);

export function providerSupportsMcpConfig(provider: string | undefined | null): boolean {
  if (!provider) return false;
  return MCP_SUPPORTED_PROVIDERS.has(provider);
}
