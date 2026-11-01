# sendlog-tui — Feature Document

**Objective**: Rebuild the Send-Log-TCP Wails app as a native Go terminal UI (Bubble Tea + Bubbles + Lipgloss) under `cmd/sendlog`, reusing the extracted core in `internal/{syslog,framing,storage,stress,logimport}`. GUI desktop phase (native, no Wails) is deferred until the core + TUI are validated; then toolkit decision (primarily Windows + Linux).
**Branch**: `native-rewrite` (in-place repo; `main` stays frozen at v1.8.0 Wails app).
**Delivery**: work-unit commits on this branch, one per task, Conventional Commits. No push/PR until user decides.
**Test runner**: `go test ./...` (TUI models are unit-tested through `Update`/`View` with plain Go tests; the full module suite must stay green for the untouched Wails/CLI paths).

## Specs

S1. User's request (verbatim): "lo q quiero es hacer esta aplicacion de manera nativa, es decir quiero rescribirla completamente de manera correcta en go, empezando x la cli, hacer una cli moderna con bubletea q funcione de manera detalla para lo q esta hecha la aplicacion como tal q su objetivo principal es poder enviar logs a un server syslog y poder recibir esos logs en un siem y ver q todo funcione correctamente. Luego desarrollar una UI moderna pero nativa, sin wails x en medio, q crees?"

S2. Scope decisions (verbatim, 2026-07-06): repo = "In-place, rama nueva" → work on branch `native-rewrite`, `main` untouched. GUI toolkit = "Decidir después" → TUI is the current deliverable; GUI must be native and multiplatform, "principalmente windows y linux" (decided later, after core + TUI validation).

S3. The TUI must cover the full functionality of the current app (S1 "de manera detallada para lo que está hecha la aplicación como tal"): connection settings, message composition, multiple messages, batch import, sending with progress/results, stress test with live stats, profiles CRUD, templates CRUD, config import/export. All sending goes through the existing `internal/syslog`, `internal/framing`, `internal/stress`, `internal/storage`, `internal/logimport` packages — no reimplementation of protocol logic.

S4. Primary objective end-to-end (S1): "poder enviar logs a un server syslog y poder recibir esos logs en un siem y ver q todo funcione correctamente" → the TUI must let the user point at a host, send messages, and see per-message success/failure; a `Test Connection` action must give the user an explicit pass/fail before bulk sending.

S5. Connection screen fields (assumption: mirror the Wails form exactly): host, port, protocol (tcp/udp), TCP framing method (octet-counting / non-transparent, shown only for tcp), TLS toggle, TLS verify toggle, CA certificate path, client cert + key paths (mTLS).

S6. Messages screen (assumption: mirror Wails): message format (rfc5424 / rfc3164), facility (0-23 named), severity (0-7 named), hostname, appname, raw PRI mode (prepend `<PRI>` verbatim, no header), add/remove/edit multiple messages.

S7. Batch import (actual behavior of Wails `batch-import.tsx` + `internal/logimport`): file path input; JSON (string array, or array of objects with `message`/`msg`/`log` field, or NDJSON), CSV (auto-detect header row with message/msg/log column), TXT/LOG (one message per line); shows total lines / valid messages / errors preview; then user chooses **Append** or **Replace All** on the current message list. No per-row facility/severity overrides — imported rows become plain message strings (verified in internal/logimport/import.go).

S8. Send flow (assumption: mirror Wails `SendMessages`): per-message progress, per-message result (ok/error with reason), final summary (sent/failed totals).

S9. Stress screen (assumption: mirror Wails `StartStressTest`): message, messages/sec, duration (s), concurrent connections, start/stop, live stats (elapsed, sent, current avg rate).

S10. Profiles screen (actual `storage` API: GetProfiles, SaveProfile upsert-by-ID, DeleteProfile; GUI has no rename/duplicate verbs — those were my invention, dropped): list, save new, load into current form, overwrite existing (edit → SaveProfile with same ID), delete.

S11. Templates screen (actual `storage` API: GetTemplates, SaveTemplate upsert-by-ID, DeleteTemplate): list, save new, apply to current form (sets message, facility, severity, appname, format), overwrite existing, delete. `LogTemplate` carries Message/Facility/Severity/Appname/MessageFormat/UseRFC5424.

S12. Config screen (actual API: `storage.Export()` returns JSON of all profiles+templates; `storage.Import(json, merge bool)`). The Wails GUI exposes the bindings but has NO UI wiring (verified: no .tsx calls ExportConfig/ImportConfig) — the TUI is where this lands: export to a user-chosen path, import from a path with merge/replace choice.

S13. UI structure (assumption, my design): Bubble Tea root model with a left navigation menu (arrow keys move, enter selects) and 8 screens: Connection, Messages, Import File (batch logs, S7), Send (S8), Stress, Profiles, Templates, Config (S12 profile/template import/export). Global keys: `ctrl+c` quit, `esc`/`q` back to nav from a screen, `?` help. Lipgloss styling, single consistent dark-friendly palette, 100%-wide layout.

S14. New command name: `sendlog` (distinct from the existing `sendlog-cli` which stays untouched on this branch).

S15. Code quality (user, verbatim 2026-07-06): "tenemos q conrruir esta app usando buenas practicas de go y clean arquitecture, ademas de hacer el codigo sencillo y entendible sin mucho sobre ingieneria, usando la skill minimal comment para escribir code" → layering: `cmd/sendlog` = wiring only, `internal/tui` = presentation (Bubble Tea models call domain services directly, no DI interfaces), `internal/{syslog,framing,stress,storage,logimport}` = domain (untouched); simple readable code, no over-engineering; zero comments by default per minimal-comments skill (names carry intent; WHY-only comments when genuinely non-obvious).

## Tasks

- T1 (S2,S13,S14,S15): scaffold — inline. Deps + `cmd/sendlog` main + root model + nav menu + all 7 screens stubbed + styles. Route: inline (parent). Status: pending.
- T2 (S4,S5,S15): shared `AppState` (connection + messages + stress fields, JSON tags matching S12) + reusable `OptionSelector` component + Connection screen completed (fields, validation, Test Connection via `syslog.Dial` + close) as the exemplar pattern for T3-T6. Route: inline. Status: pending.
- T3 (S6,S7,S8,S15): Messages + Import + Send screens (format, facility/severity, raw PRI, add/remove/edit, batch import via `logimport.Import` with Append/Replace, per-message progress/results via `SendSyslogMessages` Response, summary). Route: delegated parallel writer. Status: pending.
- T4 (S9,S15): Stress screen (message, rate, duration, start/stop, live stats via `stress.Engine` onStats callback bridged to tea.Msg ticker). Route: delegated parallel writer. Status: pending.
- T5 (S10,S11,S15): Profiles + Templates screens (list, save upsert, load/apply to form, delete, `storage` reuse). Route: delegated parallel writer. Status: pending.
- T6 (S12,S15): Config screen (Export to path, Import from path with merge/replace, `storage` reuse). Route: delegated parallel writer. Status: pending.
- T7 (S1-S15): README section for the TUI (build, run, keybindings, examples incl. SIEM end-to-end check) + CI job building `cmd/sendlog` (pure Go, no CGO). Route: inline. Status: pending.

## Log

L1 (user, 2026-07-06, verbatim): "lo q quiero es hacer esta aplicacion de manera nativa, es decir quiero rescribirla completamente de manera correcta en go, empezando x la cli, hacer una cli moderna con bubletea q funcione de manera detalla para lo q esta hecha la aplicacion como tal q su objetivo principal es poder enviar logs a un server syslog y poder recibir esos logs en un siem y ver q todo funcione correctamente. Luego desarrollar una UI moderna pero nativa, sin wails x en medio, q crees?"

L2 (user decisions, 2026-07-06): repo strategy = in-place, rama nueva (`native-rewrite`); GUI toolkit = "esto tenemos q discutirlo xq tiene q ser nativo y multiplataforma, principalmente windows y linux" → deferred ("Decidir después"); Fyne v2 presented as recommendation, eframe as runner-up, Gio rejected for form-heavy UI.

L3 (2026-07-06): Branch `native-rewrite` created from `main` @ d6c7a67. No `odd/` dir existed in repo; creating it now for this feature. Writer trigger (a) parallelism for T3-T6: after T1+T2 establish the root registry, AppState, and OptionSelector, the four screen units have disjoint edit surfaces (one file + tests each, no shared-file edits) and each needs several minutes of work → launched together in background after T2 lands.

L4 (user, 2026-07-06, verbatim): "tenemos q conrruir esta app usando buenas practicas de go y clean arquitecture, ademas de hacer el codigo sencillo y entendible sin mucho sobre ingieneria, usando la skill minimal comment para escribir code" → recorded as S15; all writers receive the minimal-comments skill path. Layering per S15: cmd/sendlog (wiring) → internal/tui (presentation) → internal/* (domain, untouched). No DI abstractions, no interfaces beyond what the models need.

L5 (2026-07-06, spec corrections after reading the actual core APIs — previous summaries of internal/* in session memory were wrong and are superseded): (a) S7 rewritten: logimport returns `Result{Messages []string, TotalLines int, Errors []string}` with NO per-row facility/severity; GUI offers Append vs Replace All. (b) S10/S11 rewritten: storage has no Rename/Duplicate; only Get*/Save*(upsert by ID)/Delete*. (c) S12 rewritten: no form-snapshot export exists in the GUI (bindings ExportConfig/ImportConfig are never called from .tsx — half-finished feature); TUI covers it with storage.Export/Import (profiles+templates JSON, merge flag). (d) S9: `stress.ContinuousSendConfig` has no ConcurrentConnections field — it uses one connection with a rate ticker; the Wails GUI stress form is what the TUI mirrors (verified fields: message, messagesPerSec, duration, maxMessages, randomizeData available in config but GUI subset). The exact API reference is recorded in the Engram mirror under odd/sendlog-tui/tasks.
