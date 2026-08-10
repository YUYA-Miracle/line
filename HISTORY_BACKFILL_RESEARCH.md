# LINE startup history backfill: research and current limits

Context: YUYA-OS issue [#307](https://github.com/YUYA-Miracle/YUYA-OS/issues/307)
found that 380/413 (92%) of LINE portals synced by this bridge have zero
messages, because startup backfill fetches at most `startupBackfillMessageLimit`
(default 50) messages per chat via `TalkService.getRecentMessagesV2`, with no
way to request anything older.

## What this fork changes

1. `startupBackfillMessageLimit` is resolvable from the
   `LINE_STARTUP_BACKFILL_LIMIT` environment variable (clamped to `[1, 1000]`,
   default unchanged at 50). This does **not** change default behavior; it
   only makes the limit adjustable without a rebuild, for use once a real
   server-side ceiling has been measured (see "What still needs live
   verification" below).
2. `backfillRecentMessages` now logs `requested_limit` and a
   `possibly_truncated` boolean (`fetched == limit`) at `info` level instead
   of only `debug`. This is a **signal, not a fact**: `getRecentMessagesV2`
   has no documented "has more" indicator, so a full batch only means the
   requested window was filled — it does not prove older history exists, and
   it must never be read as `history.complete` for a chat.

Neither change touches `FetchMessages`'s backward-pagination stub
(`params.Forward == false` still returns an immediately-final empty batch),
because no pagination RPC to implement was found — see below.

## What we looked for: an older-history / cursor RPC

`GetRecentMessagesV2(chatMid string, limit int)` (`pkg/line/methods.go`) wraps
the single LINE RPC `TalkService.getRecentMessagesV2(chatMid, limit)`. It
takes no cursor, sequence number, or "before message ID" argument, and the
response carries no continuation token — confirmed by reading the response
struct (`[]*Message`, no pagination wrapper) and by upstream's own comment in
`pkg/connector/sync.go` ("there is no older-history pagination, so backward
fetches return an empty, final batch").

To check whether some *other* RPC on the same LINE private protocol exposes
older-message pagination that this bridge simply never called, we reviewed:

- This repo's own `pkg/line/*.go`: every message-fetch path
  (`GetRecentMessagesV2`, `GetMessageBoxes`) was enumerated. `GetMessageBoxes`
  takes `LastMessagesPerMessageBoxCount`, a per-box *recent* message count —
  not a history cursor — and `MinChatID` paginates the *list of chats*, not
  messages within a chat.
- Publicly available reverse-engineered LINE `TalkService`/`MessageService`
  Thrift IDL (e.g. `winbotscript/line-protocol`, `fadhiilrachman/line-protocol`,
  `ruyaoyao/LINE-instant-messenger-protocol`). Findings:
  - `MessageService.fetchMessageOperations(localRevision, lastOpTimestamp,
    count)` is a forward-only revision/operation-log sync used for catching a
    device up on operations since a point in time (this is effectively what
    `ListenSSE`/`GetLastOpRevision` already use in this bridge for live sync).
    It is not a backward browse of arbitrary chat history, and LINE's
    operation log has its own (short, undocumented-here) retention window.
  - `SquareService.FetchSquareChatEventsRequest(subscriptionId, squareChatMid,
    syncToken, limit, direction)` **does** support a `FetchDirection`
    (forward/backward) with a sync token. However, Square is LINE's separate
    "OpenChat"/community feature with its own service and MID namespace — it
    is not applicable to regular 1:1 or group `TalkService` chats, which is
    what this bridge and the 380 empty portals are.
  - No `getPreviousMessages*`, `getOlderMessages*`, or equivalent
    cursor-based `TalkService`/`MessageService` RPC for regular chats appears
    in any of the IDL sources reviewed.

**Conclusion: no known, documented, or reverse-engineered RPC exists in
LINE's private protocol for paginating backward through a regular chat's
message history from the server.** This is consistent with LINE's own
architecture: unlike WhatsApp/Signal multi-device, LINE historically treats
message history as primarily client-local (each official client keeps its
own on-device store), and the server-side retention exposed to a newly
authorized "device" (which is what this bridge is, from LINE's perspective)
appears to be shallow. We did not invent or implement a pagination API that
we could not verify exists.

## What still needs live verification (not done here — no real LINE account
available in this environment)

- **Server-side ceiling for `getRecentMessagesV2`.** The current 50-message
  limit is a bridge-side choice (`startupBackfillMessageLimit`), not a proven
  server minimum. It is unknown whether requesting 100, 300, or 1000 returns
  more messages, the same 50, an error, or a rate limit. This must be probed
  read-only, on a low-traffic test chat, before raising
  `LINE_STARTUP_BACKFILL_LIMIT` in production — confirm the response does not
  trigger read receipts, delivery receipts, or device notifications first.
- **How far back the server's retention actually goes**, independent of the
  limit parameter — i.e. whether raising the limit surfaces messages older
  than the current oldest (2026-07-25 per the issue's production
  measurement), or whether the account's history genuinely starts there.
- Whether `getRecentMessagesV2`'s behavior differs for DMs vs. groups, or for
  chats the bridge has never previously fetched vs. ones it has.

None of this can be safely established without calling the live LINE API
against a real, authenticated account — which is explicitly out of scope for
this change per the issue's read-only/non-destructive constraints.

## Practical mitigation available today (partial, not a fix)

Because no pagination path was found, the only lever that exists without a
new upstream capability is the single-shot `limit` argument. Raising
`LINE_STARTUP_BACKFILL_LIMIT` (once its real ceiling is known) would let
chats whose full history is <= that ceiling be captured in one shot instead
of being truncated at 50; it would not help chats whose history genuinely
exceeds whatever the server allows in one request. This is an incremental
improvement, not a resolution of issue #307's root cause.
