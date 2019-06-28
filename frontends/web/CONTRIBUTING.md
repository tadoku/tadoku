# How to contribute

I'm really glad you're reading this, because volunteer developers are always welcome to improve Tadoku.

If you haven't already, come find us on our [Discord server](https://discord.gg/Dd8t9WB). We want you working on things you're excited about.

## Testing

No strategy has been decided yet. There are a handful of `jest` tests, but they're still quite limited. Feel free to discuss this on the Discord server to make this better.

## Submitting changes

Please send a [GitHub Pull Request](https://github.com/tadoku/web/pull/new/master) with a clear list of what you've done (read more about [pull requests](http://help.github.com/pull-requests/)).

Always write a clear log message for your commits without the use of capital letters.

    $ git commit -m "a brief summary of the commit"

## Coding conventions

Start reading our code and you'll get the hang of it. We optimize for readability:

  * Code should be formatted according to the prettier configuration
  * We indent using two spaces (soft tabs)
  * We use no semicolons
  * We try to avoid mutation as much as possible
    * Local mutation is fine, as long as it stays within a limited scope
  * Functionality is split across different domains inside the app folder
  * Functional components only
  * Redux: use hooks instead of `connect`
  * New code should be properly typed