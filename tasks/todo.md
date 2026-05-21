# Task Plan: Code-Accurate Architecture Diagrams

## Checklist
- [x] Inspect code paths for `plan`, `apply`, `destroy`, graph/reference resolution, provider catalog, and state transaction flow
- [x] Draft Mermaid diagrams directly from current implementation details
- [x] Create `docs/architecture.md` with context/container/component and runtime flow diagrams
- [x] Verify repository still passes tests (`go test ./...`)
- [x] Write review/results section

## Notes
- Scope: documentation-only changes based on current repository behavior.
- Diagram set: C4-style context/container/component + apply/destroy sequence + state transaction flow.

## Review
- Status: Complete
- Verification:
  - [x] `go test ./...` (pass)
- Result summary:
  - Added `docs/architecture.md` with five Mermaid diagrams based on implemented flows and boundaries.
  - Included C4 context/container/component views and concrete runtime sequences for `apply` and `destroy`.
  - Documented state lock/checksum/migration transaction behavior in a dedicated flowchart.

---

# Task Plan: Enrich Core Diagrams

## Checklist
- [x] Identify additional core internals worth documenting beyond existing C4 and command sequences
- [x] Add new renderer-safe Mermaid core diagrams to `docs/architecture.md`
- [x] Validate documentation consistency against current code paths
- [x] Write review/results section

## Notes
- Goal: provide deeper internals without overloading one diagram.
- Constraint: keep Mermaid syntax conservative for broad renderer compatibility.

## Review
- Status: Complete
- Verification:
  - [x] `go test ./...` (pass)
- Result summary:
  - Expanded architecture docs with six additional core diagrams for command dispatch, diff decisions, reference resolution, dependency ordering semantics, provider action lifecycle, and state model.
  - Kept Mermaid syntax conservative to improve render compatibility across engines.

---

# Task Plan: 12-minute Interview Script

## Checklist
- [x] Re-scan entrypoints, core reconciler flow, state manager, provider catalog, and tests
- [x] Extract evidence-backed points for transaction boundary, ordering, locking, idempotency scope, and consistency model
- [x] Write a usable 10-12 minute speaking script in Vietnamese
- [x] Include concise Q and A section for technical drill-down

## Notes
- User requested only the 12-minute version.
- Claims are constrained to evidence from current repo code.

## Review
- Status: Complete
- Result summary:
- Added `docs/interview-script-12min.md` with a complete 12-minute talk track.
- Included As-is, Gap/Risk, Next-step, edge cases, and evidence references per claim.

---

# Task Plan: Refine Interview Script for System-Level Delivery

## Checklist
- [x] Reframe the 12-minute script from code-evidence style to core-system narrative
- [x] Remove file and line level proof references, keep only architecture and flow explanation
- [x] Add compact core Mermaid diagrams after each major pattern section
- [x] Keep timing and speaking flow interview-friendly (about 12 minutes total)
- [x] Add review/results section after final rewrite

## Notes
- User intent: prioritize interview storytelling (core, system, flow) over code citation depth.
- Scope: `docs/interview-script-12min.md` and documentation planning notes only.

## Review
- Status: Complete
- Verification:
  - [x] Manual read-through for pacing and abstraction level (system-first, no code-line citations)
  - [x] Mermaid blocks added after each major pattern section
- Result summary:
  - Rewrote the interview script into a 12-minute system narrative focused on architecture, flow, consistency, and tradeoffs.
  - Removed file and line evidence sections and replaced them with interview-ready talking points.
  - Added core diagrams for each major pattern: runtime architecture, apply pipeline, dependency ordering, transaction boundary, consistency model, destroy flow, and roadmap.

---

# Task Plan: Upgrade Interview Script to Senior SWE Depth

## Checklist
- [x] Reframe speaking flow to align tightly with CV claims (layered architecture + two-phase reconciliation)
- [x] Add senior-level design rationale for each core layer and boundary
- [x] Expand consistency and failure semantics with explicit invariants and trade-offs
- [x] Add deeper execution semantics (ordering, interpolation timing, operation lifecycle)
- [x] Upgrade diagrams to reflect system reasoning rather than only process narration
- [x] Add high-signal Q&A with architect-level responses
- [x] Add review/results section

## Notes
- User intent: "chi tiet hon, thong minh hon" for interview delivery as senior SWE.
- Scope: `docs/interview-script-12min.md` only plus planning notes.

## Review
- Status: Complete
- Verification:
  - [x] Manual read-through against CV claims and interview storytelling flow
  - [x] Confirmed section coverage: layered architecture, two-phase reconciliation, ordering, transaction boundary, failure model, extensibility
  - [x] Confirmed core diagram exists after each major pattern section
- Result summary:
  - Rewrote `docs/interview-script-12min.md` into a senior-level script centered on architecture decisions and system invariants.
  - Added deeper reasoning on why each boundary exists, not just what the flow does.
  - Upgraded Q&A to architect-level prompts with concise, defensible responses.

---

# Task Plan: Clarify Commit Phase for Interview Delivery

## Checklist
- [x] Expand `Commit phase` section into explicit step-by-step execution flow
- [x] Add commit-focused diagram with success and error semantics
- [x] Add concise speaking script (45-60s) for interview delivery
- [x] Keep wording system-level and easy to present verbally
- [x] Add review/results section

## Notes
- User intent: understand commit internals clearly and present them confidently in interview.
- Scope: `docs/interview-script-12min.md` and planning notes only.

## Review
- Status: Complete
- Verification:
  - [x] Confirmed commit section now includes execution order, interpolation timing, tx boundary, success path, and failure path
  - [x] Confirmed commit-specific sequence diagram added
  - [x] Confirmed interview-ready 45-60s speaking track added
- Result summary:
  - Upgraded `Commit phase` from high-level bullets to a clear step-by-step runtime flow.
  - Added explicit invariants and practical failure semantics for senior-level discussion.
  - Added a concise spoken script you can reuse directly in interview.

---

# Task Plan: Create Reconciliation Usage Principles File

## Checklist
- [x] Create a new doc describing what inputs are used and what decisions are based on
- [x] Add case matrix for create/update/delete/noop and error scenarios
- [x] Add detailed commit execution guide with ordered runtime steps
- [x] Add interview-ready talking points and short answer patterns
- [x] Add review/results section

## Notes
- User intent: one standalone file explaining "dùng cái gì, dựa trên cái gì, các case, commit như thế nào".
- Scope: new markdown doc under `docs/` plus planning notes.

## Review
- Status: Complete
- Verification:
  - [x] Confirmed doc includes inputs, decision principles, case matrix, commit algorithm, invariants, and interview script
  - [x] Confirmed commit sequence diagram includes both success and failure paths
- Result summary:
  - Added `docs/reconciliation-usage-principles.md` as a standalone interview support file.
  - Structured content around practical interview prompts: what data is used, what decisions are based on, key cases, and commit execution flow.

---

# Task Plan: Personalize Principles Doc Using External README

## Checklist
- [x] Read external source `README (2).md` and extract reusable structure and concepts
- [x] Adapt terminology and commands to MiniIaC context (not copy project-specific branding)
- [x] Enrich `docs/reconciliation-usage-principles.md` with usage flow, decision rules, and commit algorithm details
- [x] Keep interview-oriented framing with case matrix and short speaking script
- [x] Add review/results section

## Notes
- User intent: merge useful patterns from another project's README and personalize for current MiniIaC doc.
- Scope: `docs/reconciliation-usage-principles.md` and planning notes only.

## Review
- Status: Complete
- Verification:
  - [x] Confirmed command references and wording use MiniIaC context
  - [x] Confirmed document now contains sections derived from README style: workflow table, lifecycle, failure/recovery guidance
- Result summary:
  - Upgraded principles doc from a short note into a full interview playbook with command usage, decision principles, case matrix, commit sequence, invariants, and Q&A.
  - Reused the strong structure from external README while keeping content specific to MiniIaC.

---

# Task Plan: Restore Previous Style and Add Missing Commands Only

## Checklist
- [x] Refactor `docs/reconciliation-usage-principles.md` back to the previous concise structure
- [x] Keep existing core sections, avoid full-from-scratch rewrite style
- [x] Add missing command usage section (`init/plan/apply/destroy/state show`, flags) explicitly
- [x] Preserve commit flow detail and interview-oriented quick script
- [x] Add review/results section

## Notes
- User correction: prefer "giữ bản cũ + bổ sung thiếu" over full rewritten playbook style.
- Scope: `docs/reconciliation-usage-principles.md` and planning notes only.

## Review
- Status: Complete
- Verification:
  - [x] Confirmed doc structure returned to concise principle-first format
  - [x] Confirmed missing commands and flags section added without expanding into full README style
  - [x] Confirmed commit detail, case matrix, and interview script remain intact
- Result summary:
  - Restored the prior concise format of `docs/reconciliation-usage-principles.md`.
  - Added the missing operational command guidance (`init/plan/apply/destroy/state show`, log flags, auto-approve) as requested.
