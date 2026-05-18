package execenv

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// TestBuildCommentReplyInstructionsCodexLinux pins that the strong
// "MUST use --content-stdin + HEREDOC" mandate stays alive for Codex on
// non-Windows hosts. Codex's habit of emitting literal `\n` inside
// `--content "..."` is the original reason this mandate exists
// (#1795 / #1851); on Linux/macOS stdin is the right answer.
//
// Not parallel: mutates the package-level runtimeGOOS.
func TestBuildCommentReplyInstructionsCodexLinux(t *testing.T) {
	saved := runtimeGOOS
	t.Cleanup(func() { runtimeGOOS = saved })
	runtimeGOOS = "linux"

	issueID := "11111111-1111-1111-1111-111111111111"
	triggerID := "22222222-2222-2222-2222-222222222222"

	got := BuildCommentReplyInstructions("codex", issueID, triggerID)

	for _, want := range []string{
		"multica issue comment add " + issueID + " --parent " + triggerID + " --content-stdin",
		"Always use `--content-stdin`",
		"even when the reply is a single line",
		"<<'COMMENT'",
		"Do NOT write literal `\\n` escapes to simulate line breaks",
		"do NOT reuse --parent values from previous turns",
	} {
		if !strings.Contains(got, want) {
			t.Fatalf("codex/linux reply instructions missing %q\n---\n%s", want, got)
		}
	}

	if strings.Contains(got, "--content \"...\"") {
		t.Fatalf("codex reply instructions should not offer inline --content form\n---\n%s", got)
	}
}

// TestBuildCommentReplyInstructionsNonCodexLinux pins that every non-Codex
// provider on Linux/macOS gets the lightweight pre-#1795 inline template.
// The "MUST stdin" mandate was originally a Codex-specific fix that
// #1795 / #1851 accidentally spread to every provider, breaking Windows
// non-ASCII for all of them (#2198 / #2236 / #2376). Non-Codex providers
// handle inline escaping correctly and the CLI server-decodes `\n` etc.,
// so the inline template works on every non-Windows platform.
//
// Not parallel: mutates the package-level runtimeGOOS.
func TestBuildCommentReplyInstructionsNonCodexLinux(t *testing.T) {
	saved := runtimeGOOS
	t.Cleanup(func() { runtimeGOOS = saved })

	issueID := "11111111-1111-1111-1111-111111111111"
	triggerID := "22222222-2222-2222-2222-222222222222"

	for _, host := range []string{"linux", "darwin"} {
		for _, provider := range []string{"claude", "opencode", "openclaw", "hermes", "kimi", "kiro", "cursor", "gemini"} {
			name := provider + "/" + host
			t.Run(name, func(t *testing.T) {
				runtimeGOOS = host
				got := BuildCommentReplyInstructions(provider, issueID, triggerID)

				for _, want := range []string{
					"multica issue comment add " + issueID + " --parent " + triggerID + " --content \"...\"",
					"do NOT reuse --parent values from previous turns",
					"If you decide to reply",
				} {
					if !strings.Contains(got, want) {
						t.Errorf("%s reply instructions missing %q\n---\n%s", name, want, got)
					}
				}

				// Non-Codex / non-Windows providers must NOT receive the
				// Codex-specific "MUST stdin" mandate or its HEREDOC
				// template — that was the over-spread of #1795 / #1851.
				for _, banned := range []string{
					"Always use `--content-stdin`",
					"<<'COMMENT'",
					"--parent " + triggerID + " --content-stdin",
					"--parent " + triggerID + " --content-file",
				} {
					if strings.Contains(got, banned) {
						t.Errorf("%s reply instructions still steers at codex template: %q\n---\n%s", name, banned, got)
					}
				}
			})
		}
	}
}

// TestBuildCommentReplyInstructionsWindowsUsesContentFile pins that on
// Windows every provider — Codex AND non-Codex — gets the
// `--content-file` template. The bug is shell-layer, not provider-layer:
// any agent on Windows piping HEREDOC through PowerShell loses non-ASCII
// bytes (PS 5.1's `$OutputEncoding` defaults to ASCIIEncoding). Issues
// #2198 (Chinese, Codex), #2236 (Chinese, Codex), #2376 (Cyrillic,
// non-Codex agent name) all match this signature.
//
// Not parallel: mutates the package-level runtimeGOOS.
func TestBuildCommentReplyInstructionsWindowsUsesContentFile(t *testing.T) {
	saved := runtimeGOOS
	t.Cleanup(func() { runtimeGOOS = saved })
	runtimeGOOS = "windows"

	issueID := "11111111-1111-1111-1111-111111111111"
	triggerID := "22222222-2222-2222-2222-222222222222"

	for _, provider := range []string{"codex", "claude", "opencode", "openclaw", "hermes", "kimi", "kiro", "cursor", "gemini"} {
		t.Run(provider+"/windows", func(t *testing.T) {
			got := BuildCommentReplyInstructions(provider, issueID, triggerID)
			for _, want := range []string{
				"multica issue comment add " + issueID + " --parent " + triggerID + " --content-file",
				"On Windows, write the reply body to a UTF-8 file",
				"Do NOT pipe via `--content-stdin`",
				"silently drops non-ASCII",
				"$OutputEncoding",
			} {
				if !strings.Contains(got, want) {
					t.Errorf("%s reply instructions missing %q\n---\n%s", provider, want, got)
				}
			}
			for _, banned := range []string{
				"<<'COMMENT'",
				"--parent " + triggerID + " --content-stdin",
				"cat <<",
			} {
				if strings.Contains(got, banned) {
					t.Errorf("%s/windows reply instructions should not contain %q\n---\n%s", provider, banned, got)
				}
			}
		})
	}
}

func TestBuildCommentReplyInstructionsEmptyWhenNoTrigger(t *testing.T) {
	t.Parallel()

	for _, provider := range []string{"codex", "claude", "opencode"} {
		if got := BuildCommentReplyInstructions(provider, "issue-id", ""); got != "" {
			t.Fatalf("expected empty string when triggerCommentID is empty for %s, got %q", provider, got)
		}
	}
}

func TestBuildIssueResultInstructionsCodexLinux(t *testing.T) {
	saved := runtimeGOOS
	t.Cleanup(func() { runtimeGOOS = saved })
	runtimeGOOS = "linux"

	issueID := "11111111-1111-1111-1111-111111111111"

	got := BuildIssueResultInstructions("codex", issueID)

	for _, want := range []string{
		"multica issue comment add " + issueID + " --content-stdin",
		"always use `--content-stdin` with a HEREDOC",
		"<<'COMMENT'",
		"Do NOT use inline `--content`",
	} {
		if !strings.Contains(got, want) {
			t.Fatalf("codex/linux issue result instructions missing %q\n---\n%s", want, got)
		}
	}
}

func TestBuildIssueResultInstructionsNonCodexLinux(t *testing.T) {
	saved := runtimeGOOS
	t.Cleanup(func() { runtimeGOOS = saved })
	runtimeGOOS = "linux"

	issueID := "11111111-1111-1111-1111-111111111111"
	got := BuildIssueResultInstructions("claude", issueID)

	for _, want := range []string{
		"multica issue comment add " + issueID + " --content \"...\"",
		"prefer `--content-stdin`",
		"`--content-file <path>`",
	} {
		if !strings.Contains(got, want) {
			t.Fatalf("non-codex/linux issue result instructions missing %q\n---\n%s", want, got)
		}
	}
	for _, banned := range []string{
		"<<'COMMENT'",
		"--content-file ./result.md",
	} {
		if strings.Contains(got, banned) {
			t.Fatalf("non-codex/linux issue result instructions should not contain %q\n---\n%s", banned, got)
		}
	}
}

func TestBuildIssueResultInstructionsWindowsUsesContentFile(t *testing.T) {
	saved := runtimeGOOS
	t.Cleanup(func() { runtimeGOOS = saved })
	runtimeGOOS = "windows"

	issueID := "11111111-1111-1111-1111-111111111111"
	for _, provider := range []string{"codex", "claude", "opencode"} {
		t.Run(provider, func(t *testing.T) {
			got := BuildIssueResultInstructions(provider, issueID)
			for _, want := range []string{
				"multica issue comment add " + issueID + " --content-file ./result.md",
				"write the comment body to a UTF-8 file",
				"Do NOT pipe via `--content-stdin` on Windows",
			} {
				if !strings.Contains(got, want) {
					t.Errorf("%s/windows issue result instructions missing %q\n---\n%s", provider, want, got)
				}
			}
			for _, banned := range []string{
				"<<'COMMENT'",
				"cat <<",
			} {
				if strings.Contains(got, banned) {
					t.Errorf("%s/windows issue result instructions should not contain %q\n---\n%s", provider, banned, got)
				}
			}
		})
	}
}

// Pins runtimeGOOS to "linux" so the AGENTS.md surface is deterministic.
// Comment-trigger reply commands now live only in the per-turn prompt.
// Not parallel: mutates runtimeGOOS.
func TestInjectRuntimeConfigCommentTriggerOmitsPerTurnReplyCommand(t *testing.T) {
	saved := runtimeGOOS
	t.Cleanup(func() { runtimeGOOS = saved })
	runtimeGOOS = "linux"

	issueID := "11111111-1111-1111-1111-111111111111"
	triggerID := "22222222-2222-2222-2222-222222222222"

	for _, provider := range []string{"claude", "codex", "opencode"} {
		t.Run(provider, func(t *testing.T) {
			dir := t.TempDir()
			ctx := TaskContextForEnv{
				IssueID:          issueID,
				TriggerCommentID: triggerID,
			}
			if _, err := InjectRuntimeConfig(dir, provider, ctx); err != nil {
				t.Fatalf("InjectRuntimeConfig failed: %v", err)
			}

			fileName := "CLAUDE.md"
			if provider != "claude" {
				fileName = "AGENTS.md"
			}
			content, err := os.ReadFile(filepath.Join(dir, fileName))
			if err != nil {
				t.Fatalf("read %s: %v", fileName, err)
			}
			s := string(content)

			for _, banned := range []string{
				triggerID,
				"--parent " + triggerID,
				"<<'COMMENT'",
				"Find the triggering comment",
				"always use `--content-stdin` with a HEREDOC",
			} {
				if strings.Contains(s, banned) {
					t.Errorf("%s should not contain per-turn reply command fragment %q\n---\n%s", fileName, banned, s)
				}
			}
		})
	}
}

// TestInjectRuntimeConfigWindowsCommentTriggerHasNoPerTurnReplyCommand asserts
// the end-to-end CLAUDE.md / AGENTS.md surface for a comment-triggered task on
// a Windows daemon has no parent-specific reply command. The Windows-safe
// command lives in BuildPrompt for the current turn.
//
// Not parallel: mutates the package-level runtimeGOOS.
func TestInjectRuntimeConfigWindowsCommentTriggerHasNoPerTurnReplyCommand(t *testing.T) {
	saved := runtimeGOOS
	t.Cleanup(func() { runtimeGOOS = saved })
	runtimeGOOS = "windows"

	issueID := "11111111-1111-1111-1111-111111111111"
	triggerID := "22222222-2222-2222-2222-222222222222"
	ctx := TaskContextForEnv{
		IssueID:          issueID,
		TriggerCommentID: triggerID,
	}

	for _, provider := range []string{"claude", "codex", "opencode"} {
		t.Run(provider, func(t *testing.T) {
			dir := t.TempDir()
			if _, err := InjectRuntimeConfig(dir, provider, ctx); err != nil {
				t.Fatalf("InjectRuntimeConfig failed: %v", err)
			}
			fileName := "CLAUDE.md"
			if provider != "claude" {
				fileName = "AGENTS.md"
			}
			data, err := os.ReadFile(filepath.Join(dir, fileName))
			if err != nil {
				t.Fatalf("read %s: %v", fileName, err)
			}
			s := string(data)

			for _, want := range []string{
				"--content-file",
				"--description-file",
				"The per-turn prompt is the source of truth",
			} {
				if !strings.Contains(s, want) {
					t.Errorf("%s missing stable guidance %q\n---\n%s", fileName, want, s)
				}
			}

			for _, banned := range []string{
				triggerID,
				"--parent " + triggerID + " --content-file",
				"--parent " + triggerID + " --content-stdin",
				"always use `--content-stdin` with a HEREDOC, even for short single-line replies",
				"MUST pipe via stdin",
				"use `--description-stdin` and pipe a HEREDOC",
				"<<'COMMENT'",
				"Agent-authored comments should always pipe content via stdin",
			} {
				if strings.Contains(s, banned) {
					t.Errorf("%s still steers agent at stdin: %q\n---\n%s", fileName, banned, s)
				}
			}
		})
	}
}
