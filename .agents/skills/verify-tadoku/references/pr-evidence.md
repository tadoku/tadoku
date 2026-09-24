# Publishing PR evidence

Use this when delivering verification on an authorized Tadoku PR. The existing
DevCLI/browser workflow supplies the evidence; GitHub hosts the media. Do not add
screenshots, videos or one-off capture scripts to source commits, and do not create
another upload service. Posting evidence does not authorize merging the PR.

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

## Upload through GitHub CLI

GitHub CLI introduced attachments in 2.99.0. Check the executable actually on PATH:

```sh
gh version
gh pr comment --help
gh pr view 123 --repo tadoku/tadoku --json url,headRefOid
```

Confirm `--attach` is available, the repository/PR is correct and the tested SHA
matches the intended revision. Uploads need push access. Never print tokens or
change credentials to bypass a denial. If the tool is too old, use an approved
updated client or the authenticated GitHub browser uploader; otherwise retain the
files and report the blocker rather than claiming delivery.

From the evidence directory, write `verification.md` with a concise test summary:
tested SHA/environment, steps and assertions, pass/fail, limitations (including
TLS bypasses or mocks), and links to any existing authorized detailed artifacts.
Place these references where the media belongs, replacing filenames as needed:

```markdown
![The expected result after completing the workflow](./result.png)

![](./workflow.webm)
```

The standalone video reference becomes a player. End the comment with
`**Posted by MODEL_NAME**` on its own line after a blank line, replacing
`MODEL_NAME` with the exact authoring model. Keep it after the media references.
Then post only the files actually captured:

```sh
gh pr comment 123 --repo tadoku/tadoku \
  --body-file verification.md \
  --attach ./result.png \
  --attach ./workflow.webm
```

GitHub rewrites matching local paths to hosted attachment URLs. `gh pr create`
and `gh pr edit` also accept `--attach` for a PR description. Preserve existing
body content when editing. Prefer updating a known evidence comment over repeated
posts; inspect its ownership first, since `--edit-last` targets the authenticated
account's last comment, not necessarily this agent's evidence.

Use PNG/JPEG for screenshots, or MP4/WebM for video (H.264 MP4 is GitHub's
compatibility recommendation). Keep each file below 10 MB unless the repository's
higher video allowance is confirmed. CLI attachment support is for images/videos,
not arbitrary trace ZIPs or logs; use existing authorized artifact storage for
those, or retain locally and explicitly say they are not attached.

## Verify delivery

Re-fetch the PR/comment and check that local paths were replaced by hosted URLs.
Open the posted PR in an authenticated browser and confirm images render and the
video plays through the result. If browser access is unavailable, distinguish
successful upload from unverified inline rendering/playback in the handoff.
After an ambiguous upload error, inspect the PR before retrying to avoid duplicate
comments or attachments. Do not fall back to a public gist/host for private evidence.

Return the PR or evidence-comment link, not just a local filename. Record unverified
boundaries and capture failures honestly. If code changed after recording, rerun
affected verification or label the evidence with its older SHA; do not imply that
a recording proves a different revision.

Reference: [GitHub CLI attachments](https://docs.github.com/en/github-cli/github-cli/attaching-files-with-github-cli)
and [file types, limits and access](https://docs.github.com/en/get-started/writing-on-github/working-with-advanced-formatting/attaching-files).
