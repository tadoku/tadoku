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
      ],
    },
  ],
};

export default sidebar.apisidebar;
