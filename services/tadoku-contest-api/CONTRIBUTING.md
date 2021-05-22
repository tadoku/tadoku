# How to contribute

I'm really glad you're reading this, because volunteer developers are always welcome to improve Tadoku.

If you haven't already, come find us on our [Discord server](https://discord.gg/Dd8t9WB). We want you working on things you're excited about.

## Architecture

TODO: please ask on Discord if you need more info and this hasn't been written yet.

## Testing

Nearly all code should be tested. Please include a test so your contribution can be quickly reviewed. We're not aiming for 100% coverage here, just enough so we can refactor swiftly and have faith in the test suite.

## Submitting changes

Please send a [GitHub Pull Request](https://github.com/tadoku/api/pull/new/master) with a clear list of what you've done (read more about [pull requests](http://help.github.com/pull-requests/)).

Always write a clear log message for your commits without the use of capital letters.

    $ git commit -m "a brief summary of the commit"

## Coding conventions

Start reading our code and you'll get the hang of it. We optimize for readability:

  * Code should be formatted with `gofmt`
  * Code should pass the linting tests: `$ make lint`
  * The project is structured in layers, it's roughly an implementation of clean architecture
    * It's okay to depend on domain in infra, but the other way around is not okay
  * Use interfaces instead of depending on external types
    * Even internally this should be done so that we can mock them
