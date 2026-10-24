import type { SidebarsConfig } from "@docusaurus/plugin-content-docs";

const sidebar: SidebarsConfig = {
  apisidebar: [
    {
      type: "doc",
      id: "api/authorization/authz-api",
    },
    {
      type: "category",
      label: "role",
      items: [
        {
          type: "doc",
          id: "api/authorization/role-get",
          label: "Fetches the role of the current user",
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
          id: "api/authorization/role-update",
          label: "Update user role (admin only)",
          className: "api-method put",
        },
      ],
    },
  ],
};

export default sidebar.apisidebar;
