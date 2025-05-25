# Auth frontend

This is the identity portal of the Tadoku website. It contains all functionality to:

- Sign up to the website
- Log in to the website
- Update your profile (email, display name, password)
- Log out from the website
- Activate accounts after sign up

It should not contain any non-identity related functionality.

## Architecture

- A [Next.js](https://nextjs.org/) app written using TypeScript
- The app is a frontend for [Ory Kratos](https://github.com/ory/kratos)
- The app uses Tailwind CSS for styling
- Uses the Tadoku component library "ui" from within the workspace
  - Refer to `frontend/apps/styleguide` for a reference on how to use this component library
- All forms are written with [react-hook-form](https://react-hook-form.com/)

## Important links

- [Source code](https://github.com/tadoku/tadoku/tree/main/frontend/apps/auth)
- [Production](https://account.tadoku.app/)
