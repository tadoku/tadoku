import type { SidebarsConfig } from "@docusaurus/plugin-content-docs";

const sidebar: SidebarsConfig = {
  apisidebar: [
    {
      type: "doc",
      id: "api/profile/profile-api",
    },
    {
      type: "category",
      label: "admin",
      items: [
        {
          type: "doc",
          id: "api/profile/users-list",
          label: "Lists all users (admin only)",
          className: "api-method get",
        },
        {
          type: "doc",
          id: "api/profile/account-deletion-create",
          label: "Request the deletion of a user's account (admin only)",
          className: "api-method post",
        },
      ],
    },
  ],
};

export default sidebar.apisidebar;
