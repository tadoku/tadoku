import { Namespace, Context } from "@ory/keto-namespace-types"

class User implements Namespace {}

class app implements Namespace {
  related: {
    admins: User[]
    banned: User[]
    parents: app[]
    testers: User[]
  }

  permits = {
    admin: (ctx: Context) =>
      this.related.admins.includes(ctx.subject) ||
      this.related.parents.traverse((p) => p.permits.admin(ctx)),
    is_banned: (ctx: Context) =>
      this.related.banned.includes(ctx.subject) ||
      this.related.parents.traverse((p) => p.permits.is_banned(ctx)),
    access: (ctx: Context) =>
      this.permits.admin(ctx) || this.related.testers.includes(ctx.subject),
  }
}
