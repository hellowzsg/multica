import { describe, expect, it } from "vitest";
import { providerSupportsMcpConfig } from "./mcp-support";

describe("providerSupportsMcpConfig", () => {
  // Original 7 backends that consume ExecOptions.McpConfig — kept here as
  // a regression fence so a future refactor of the allow-list can't drop
  // one of them without the test catching it.
  it.each([
    ["claude"],
    ["codex"],
    ["hermes"],
    ["kimi"],
    ["kiro"],
    ["opencode"],
    ["openclaw"],
  ])("supports built-in provider %s", (provider) => {
    expect(providerSupportsMcpConfig(provider)).toBe(true);
  });

  // PR #3094 variant providers — they reuse claude / codex backends via
  // ProtocolFamily, so MCP wiring works for free. codebuddy's own CLI
  // also exposes `--mcp-config <fileOrString>` so the daemon's flag
  // append survives the dispatch.
  it.each([
    ["claude-internal"],
    ["codebuddy"],
    ["codex-internal"],
  ])("supports claude/codex family variant %s", (provider) => {
    expect(providerSupportsMcpConfig(provider)).toBe(true);
  });

  // gemini-internal is intentionally NOT supported: the gemini backend
  // ignores ExecOptions.McpConfig, so its variant inherits the same
  // limitation. Adding it would re-create the "UI saves it, runtime
  // drops it" footgun the allow-list exists to prevent.
  it("does not support gemini-internal because gemini backend ignores mcp_config", () => {
    expect(providerSupportsMcpConfig("gemini-internal")).toBe(false);
  });

  it.each([
    ["gemini"],
    ["antigravity"],
    ["pi"],
    ["copilot"],
    ["cursor"],
    ["unknown-provider"],
  ])("does not support %s", (provider) => {
    expect(providerSupportsMcpConfig(provider)).toBe(false);
  });

  it("returns false for nullish input", () => {
    expect(providerSupportsMcpConfig(null)).toBe(false);
    expect(providerSupportsMcpConfig(undefined)).toBe(false);
    expect(providerSupportsMcpConfig("")).toBe(false);
  });
});
