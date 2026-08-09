import type { SidebarsConfig } from "@docusaurus/plugin-content-docs";

const sidebar: SidebarsConfig = {
  apisidebar: [
    {
      type: "doc",
      id: "api/profile/profile-api",
    },
    {
      type: "category",
      label: "metadata",
      items: [
        {
          type: "doc",
          id: "api/profile/ping",
          label: "Checks if service is responsive",
          className: "api-method get",
        },
      ],
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
