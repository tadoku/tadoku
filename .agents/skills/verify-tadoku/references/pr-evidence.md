# Publishing PR evidence

Use this when delivering verification on an authorized Tadoku PR. Put one
evidence block in the PR description. The existing DevCLI/browser workflow
supplies the evidence; GitHub hosts the media. Do not add screenshots, videos
or one-off capture scripts to source commits, and do not create another upload
service. Posting evidence does not authorize merging the PR.

## Capture the behavior

- Record the tested commit, any uncommitted changes, branch/owner, selected
  development URLs, actual frontend/API backend identity, fixtures, exact commands
  and expected/actual assertions. Link to the relevant feature-map journey.
- For visual changes, capture the relevant states. For multi-step workflows,
  record a short video showing the trigger, interaction and stable result in one
  sequence. Include before/after evidence for regressions when available; never
  label an unobserved baseline as reproduced. Backend-only changes may be better
  demonstrated by sanitized request/response assertions than a decorative video.
- Use the available browser recorder or Playwright `recordVideo` on the same
  context that exercised the selected environment. Close that context to finalize
  the video before uploading. Wait for observable state, not arbitrary sleeps.
  Inspect screenshots and play the recording; a blank/truncated video is not proof.
- Keep originals outside tracked source. Show synthetic development data only.
  Exclude passwords, cookies, authorization headers, recovery links, tokens,
  Secret contents and authenticated browser-storage dumps, even on private PRs.
  Review the full recording, not only its final frame. If redacting/trimming,
  disclose edits and preserve the tested behavior's meaning; don't hide failures.

## Format the evidence block

Keep the result scannable before any section is expanded:

1. Start with the outcome, tested commit and environment.
2. Use a compact table with one row per meaningful check: **Check**, **Result**
   and **Proof**. Link each proof to the relevant attachment, trace or CI run.
   State failures and skipped checks in the table; never turn a skip into a pass.
3. Put a visible GitHub alert immediately below the table. Use `> [!NOTE]` for
   coverage limits and `> [!WARNING]` when a gap blocks confidence in the
   change. Name untested states and viewports, mocks, TLS bypasses and skipped
   checks with their reasons. Do not hide material limitations in a collapsed
   section.
4. Add a closed `<details>` section for visual evidence. Use a table of states,
   small linked previews and what each image proves. Link the recording with a
   descriptive label instead of leaving a large player in the main reading
   flow. For nonvisual changes, link the relevant request/response artifact or
   trace instead of adding decorative media.
5. Add a second closed `<details>` section for reproduction and provenance:
   command, route, fixture, selected backend identity, steps and exact
   assertions. Keep long service names and setup detail here.

For example, replace the placeholders and omit media that was not captured:

```markdown
## Verification

**Passed** on `TESTED_SHA` in `ENVIRONMENT` using `FIXTURE`; no mocked responses.

| Check | Result | Proof |
| --- | --- | --- |
| User journey | Pass — expected result after the action | [Result screenshot](SCREENSHOT_URL) |
| Interaction | Pass — downstream state was visible | [Recording](VIDEO_URL) |
| Local build | Skipped — state the reason | [CI run](CI_URL) |

> [!NOTE]
> **Coverage limits:** Name untested states and viewports, skipped checks, mocks
> or TLS bypasses. State what was actually verified.

<details>
<summary>Visual evidence · 1 screenshot and 1 recording</summary>

| State | Preview | What it proves |
| --- | --- | --- |
| Result | <a href="SCREENSHOT_URL"><img src="SCREENSHOT_URL" alt="Describe the visible result" width="160"></a> | The expected state is visible. |

[Watch the interaction recording](VIDEO_URL).

</details>

<details>
<summary>Reproduce and provenance · setup, steps and assertions</summary>

| Item | Detail |
| --- | --- |
| Revision | `TESTED_SHA`; name any uncommitted diff |
| Setup | Command, branch/owner, selected URLs and actual backend identities |
| Journey | Entry point, fixture, user actions and expected result |
| Assertions | Observed result, persistence/downstream effect and failure state |

</details>
```

Keep the tables short enough to read on a narrow screen. Use descriptive link
labels and image alt text. Link previews to the full-size hosted attachment;
use text links when a thumbnail would not help. GitHub supports
[tables](https://docs.github.com/en/get-started/writing-on-github/working-with-advanced-formatting/organizing-information-with-tables),
[collapsed sections](https://docs.github.com/en/get-started/writing-on-github/working-with-advanced-formatting/organizing-information-with-collapsed-sections)
and [alerts](https://docs.github.com/en/get-started/writing-on-github/getting-started-with-writing-and-formatting-on-github/basic-writing-and-formatting-syntax#alerts).

## Upload through GitHub CLI

GitHub CLI introduced attachments in 2.99.0. Check the executable actually on PATH:

```sh
gh version
gh pr edit --help
gh pr view 123 --repo tadoku/tadoku --json url,headRefOid
```

Confirm `--attach` is available, the repository/PR is correct and the tested SHA
matches the intended revision. Uploads need push access. Never print tokens or
change credentials to bypass a denial. If the tool is too old, use an approved
updated client or the authenticated GitHub browser uploader; otherwise retain the
files and report the blocker rather than claiming delivery.

Fetch the current PR body and preserve its summary, issue links and other
sections. Add or update its single `## Verification` block, then write the full
body to a temporary file outside the repository. To upload new media in the same
edit, put local Markdown references in that block where the media belongs:

```markdown
[Result screenshot](./result.png)

[Interaction recording](./workflow.webm)
```

Then edit the PR description with only the files actually captured:

```sh
gh pr edit 123 --repo tadoku/tadoku \
  --body-file /tmp/pr-body.md \
  --attach ./result.png \
  --attach ./workflow.webm
```

GitHub rewrites matching Markdown paths to hosted attachment URLs. To replace
those links with small linked HTML previews, first read the uploaded URLs from
the PR body, then edit that same body with the hosted URLs in both `href` and
`src`. Do not use local paths in HTML: the CLI attachment rewrite covers
Markdown references. For already-hosted media, reuse its URLs and edit the body
once. Avoid duplicate media references. `gh pr create --body-file ... --attach`
also works when opening a new PR.

Use PNG/JPEG for screenshots, or MP4/WebM for video (H.264 MP4 is GitHub's
compatibility recommendation). Keep each file below 10 MB unless the repository's
higher video allowance is confirmed. CLI attachment support is for images/videos,
not arbitrary trace ZIPs or logs; use existing authorized artifact storage for
those, or retain locally and explicitly say they are not attached.

## Verify delivery

Re-fetch the PR body and check that local paths were replaced by hosted URLs.
Open the posted PR in an authenticated browser and confirm the table, alert and
collapsed sections render, image links open and the video plays through the
result. If browser access is unavailable, distinguish successful upload from
unverified rendering/playback in the handoff. After an ambiguous upload error,
inspect the PR before retrying to avoid duplicate attachments. Do not fall back
to a public gist/host for private evidence.

Return the PR link, not just a local filename. Record unverified
boundaries and capture failures honestly. If code changed after recording, rerun
affected verification or label the evidence with its older SHA; do not imply that
a recording proves a different revision.

Reference: [GitHub CLI attachments](https://docs.github.com/en/github-cli/github-cli/attaching-files-with-github-cli)
and [file types, limits and access](https://docs.github.com/en/get-started/writing-on-github/working-with-advanced-formatting/attaching-files).
