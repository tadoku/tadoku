# How to contribute

I'm really glad you're reading this, because volunteer developers are always welcome to improve Tadoku.

If you haven't already, come find us on our [Discord server](https://discord.gg/AsC9vZs2Ex). We want you working on things you're excited about.

## Architecture

TODO: please ask on Discord if you need more info and this hasn't been written yet.

### Design-system coexistence

The legacy `ui` package remains the only design system used by `admin`, `auth`, `webv2`, and the legacy `styleguide` until each application receives its complete Tadoku Paper cutover. The new `paper-styleguide` is the first and initially only `paper-ui` application.

Do not mix `ui` and `paper-ui` imports or styles in one application, import `paper-ui/src/*`, add Next.js or Headless UI dependencies to Paper, or load `paper-ui/styles.css` more than once. Run `cd frontend && pnpm check:paper-boundaries` when changing shared UI or an application boundary. The migration state is recorded in `frontend/paper-boundaries.json` and changes only in the coordinated cutover for an entire application.

## Testing

Nearly all code should be tested. Please include a test so your contribution can be quickly reviewed. We're not aiming for 100% coverage here, just enough so we can refactor swiftly and have faith in the test suite.

## Submitting changes

Please send a [GitHub Pull Request](https://github.com/tadoku/tadoku/pull/new/main) with a clear list of what you've done (read more about [pull requests](http://help.github.com/pull-requests/)).

Always write a clear log message for your commits without the use of capital letters.

    $ git commit -m "a brief summary of the commit"

## Coding conventions

Start reading our code and you'll get the hang of it. We optimize for readability:

  * Code should be formatted with `gofmt`
  * The project is structured in layers, it's roughly an implementation of clean architecture
    * It's okay to depend on domain in infra, but the other way around is not okay
  * Use interfaces instead of depending on external types
    * Even internally this should be done so that we can mock them
