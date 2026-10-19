import type { SidebarsConfig } from "@docusaurus/plugin-content-docs";

const sidebar: SidebarsConfig = {
  apisidebar: [
    {
      type: "doc",
      id: "api/identity/identity-api",
    },
    {
      type: "category",
      label: "identity",
      items: [
        {
          type: "doc",
          id: "api/identity/identity-registration-create",
          label: "Register an account and create a session",
          className: "api-method post",
        },
        {
          type: "doc",
          id: "api/identity/identity-login-create",
          label: "Log in or re-authenticate",
          className: "api-method post",
        },
        {
          type: "doc",
          id: "api/identity/identity-logout",
          label: "Revoke the current session and clear its cookie",
          className: "api-method post",
        },
        {
          type: "doc",
          id: "api/identity/identity-session-get",
          label: "Resolve the tenant account app and current account",
          className: "api-method get",
        },
        {
          type: "doc",
          id: "api/identity/identity-recovery-create",
          label: "Request recovery without revealing account existence",
          className: "api-method post",
        },
        {
          type: "doc",
          id: "api/identity/identity-recovery-consume",
          label: "Consume a recovery token and redirect to Settings",
          className: "api-method get",
        },
        {
          type: "doc",
          id: "api/identity/identity-profile-update",
          label: "Update profile; email changes require a recent login",
          className: "api-method patch",
        },
        {
          type: "doc",
          id: "api/identity/identity-password-update",
          label: "Update password after a recent login and revoke other sessions",
          className: "api-method put",
        },
      ],
    },
  ],
};

export default sidebar.apisidebar;
