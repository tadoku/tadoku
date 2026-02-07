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

// Contest namespace - for future contest-level permissions
// Note: Most contests are public. This namespace is for private/invite-only contests.
class Contest implements Namespace {
  related: {
    // Contest owners (full control)
    owners: User[]
    // Contest editors (can modify settings, admins inherit this)
    editors: (User | App["admins"])[]
    // Contest participants (can log entries)
    participants: User[]
  }

  permits = {
    // Full control over contest
    manage: (ctx: Context) =>
      this.related.owners.includes(ctx.subject) ||
      this.related.editors.includes(ctx.subject),
    // Can participate in contest
    participate: (ctx: Context) =>
      this.permits.manage(ctx) || this.related.participants.includes(ctx.subject),
    // View is typically public - handle at app level, not in Keto
  }
}

// Content namespace - for CMS content (pages, posts)
class Content implements Namespace {
  related: {
    // Content authors
    authors: (User | App["admins"])[]
    // Content editors
    editors: (User | App["admins"])[]
  }

  permits = {
    // Can create/edit content
    write: (ctx: Context) =>
      this.related.authors.includes(ctx.subject) ||
      this.related.editors.includes(ctx.subject),
    // Can view content (public)
    view: (ctx: Context) => true,
  }
}
