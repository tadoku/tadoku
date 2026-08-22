## Ory

### Oathkeeper

The local Flipt operator UI uses the dedicated host configured as
`hosts.flipt` in the ignored `tilt_config.json`. The host resolves to
Oathkeeper, which requires a Kratos session and sends only the Kratos subject
to authz-api's fixed Tadoku-admin check. There is no anonymous authenticator on
this rule. Authz-api returns 200 for an allowed administrator, 403 for a denial,
and 503 when Keto is unavailable; Oathkeeper's `remote_json` authorizer requires
the 200 response to allow the request.

Generate the callback credential before starting Tilt and set it as
`oathkeeper_authz_token` in the active `local` or `shared` configuration:

```sh
openssl rand -hex 32
```

Rendering fails when this value is missing, a placeholder, shorter than 32
characters, or not URL-safe. The rendered files under `.tilt/` are ignored and
the committed templates contain no credential. Tilt creates the
`dev-oathkeeper-authz` Secret in both the Oathkeeper and authz-api namespaces.
Oathkeeper reads it into `OATHKEEPER_AUTHZ_TOKEN` and templates it into the
authorization callback's bearer header at runtime; authz-api reads the same
Secret key through `API_OATHKEEPER_AUTHZ_TOKEN`.

Application SDKs use the internal base URL
`http://oathkeeper-proxy.default:4455/flipt`. They first exchange their
projected Kubernetes service-account token at
`/token-exchange/flipt-evaluation/immersion-api` or
`/token-exchange/flipt-evaluation/frontend-webv2`, then send the returned
Oathkeeper-signed JWT as a bearer token. The evaluation rule accepts only
`GET` on the `default` namespace snapshot endpoint and strips `/flipt` before
forwarding to Flipt. Management paths and the streaming endpoint are not
included.

### Production feature-state contract

Production Flipt uses the private `tadoku/feature-flags` repository as its
only state store and commits ordinary changes directly to `main`. The branch
allows direct pushes without a pull-request requirement. Branch protection is
not part of the accepted production contract. The repository is initialized
with a normal non-Flipt commit; do not add sample features, environment
branches, pull-request templates, lifecycle manifests, or bootstrap flag
state. Creating the first production feature remains an explicit operator
action in Flipt.

Current state verified on 21 August 2026: the repository is private, its
default branch is `main`, and root commit
`d373e92e8b3ad2e34a4029a8cc3166eb15f0f1a1` adds only `README.md`. The branch
is intentionally unprotected, and this does not block the Phase 1 gate.

#### References

* https://k8s.ory.sh/helm/oathkeeper.html
* https://www.ory.sh/docs/oathkeeper
* https://github.com/ory/kratos/blob/master/contrib/quickstart/oathkeeper/oathkeeper.yml
* https://www.ory.sh/docs/oathkeeper/pipeline/authz
* https://www.ory.sh/zero-trust-api-security-ory-tutorial/
* https://docs.mojaloop.io/business-operations-framework-docs/guide/SecurityBC.html
* https://gruchalski.com/posts/2021-05-20-figuring-out-oathkeeper/
