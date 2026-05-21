# Lessons Learned

## 2026-04-12 - Mermaid Render Reliability
- Correction received: `Reconciliation Components` diagram failed to render.
- Pattern: flowchart diagrams with parser-sensitive constructs (`\\n` label escapes, slash-heavy labels, and ambiguous subgraph IDs) can fail across Mermaid renderers.
- Rule:
  - Use conservative Mermaid syntax in docs: quoted labels, simple alphanumeric subgraph IDs, and plain-text node labels.
  - Avoid escaped newlines and special-character-heavy labels in node text when portability matters.

## 2026-04-14 - Interview Script Abstraction Level
- Correction received: interview script was too heavy on code-level evidence (file and line references).
- Pattern: for interview delivery docs, excessive proof citations reduce storytelling clarity and timing control.
- Rule:
  - Default interview scripts to system-first narrative: core architecture, flow, consistency model, tradeoffs, and roadmap.
  - Avoid file path and line-by-line references unless explicitly requested.
  - Add one compact core diagram after each major pattern to support whiteboard-style explanation.

## 2026-04-14 - Senior Interview Narrative Depth
- Correction received: script needed to be more detailed and smarter at senior SWE level.
- Pattern: system-first scripts can still feel shallow if they describe flow without design rationale and invariants.
- Rule:
  - For senior interview docs, include explicit sections for: why this architecture, invariants, trade-offs, and evolution path.
  - Tie narrative directly to CV claims so each bullet is defended with architectural reasoning.
  - Prefer high-signal Q&A that tests judgment (boundary decisions, failure semantics, scaling triggers).

## 2026-04-14 - Commit Phase Clarity
- Correction received: commit phase explanation was still too high-level for confident interview delivery.
- Pattern: interview flow sections become hard to present when commit semantics are not shown as concrete runtime steps.
- Rule:
  - In reconciliation talks, always describe commit as ordered runtime steps: tx begin, per-resource execution loop, state/index update, persist, release.
  - Explicitly include success and failure semantics (stop point, partial progress, next reconcile behavior).
  - Provide a 45-60s spoken version for high-pressure interview answers.

## 2026-04-14 - External README Adaptation
- Correction received: user provided an external README and asked to personalize its useful parts into current project docs.
- Pattern: when a reference doc is provided, a generic rewrite is insufficient unless structure and best sections are explicitly mapped into local project context.
- Rule:
  - Extract structure from external docs (features, workflow, lifecycle, cases), then remap names, commands, and boundaries to the current repository.
  - Avoid cross-project branding leakage (e.g., old binary names) in final docs.
  - Keep the output interview-oriented when the active document is an interview support file.

## 2026-04-14 - Preserve User Draft Style
- Correction received: user wanted incremental additions on top of the old draft, not a full rewrite from scratch.
- Pattern: even when content quality improves, changing the draft style too much can violate user intent.
- Rule:
  - When user says "bổ sung thiếu" or equivalent, preserve the existing structure and only add missing sections.
  - Prioritize minimal-diff edits over wholesale rewrites unless explicitly asked.
  - If source reference is provided (another README), extract missing pieces only (e.g., commands), then merge into current draft style.
