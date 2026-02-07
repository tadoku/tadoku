// Ory Keto namespace configuration (OPL - Ory Permission Language)
// This defines the permission model for Tadoku

import { Namespace, Context } from "@ory/keto-namespace-types"

// User namespace - represents authenticated users
class User implements Namespace {}

// App namespace - application-level permissions (replaces the current role system)
// Note: "user" is the default state - anyone authenticated who isn't admin or banned.
// We don't store "users" explicitly; absence of admin/banned implies regular user.
class App implements Namespace {
  related: {
    // Users with admin privileges
    admins: User[]
    // Banned users (denied access)
    banned: User[]
  }

  permits = {
    // Check if user is an admin
    admin: (ctx: Context) => this.related.admins.includes(ctx.subject),
    // Check if user is banned
    is_banned: (ctx: Context) => this.related.banned.includes(ctx.subject),
  }
}
