import type { SidebarsConfig } from "@docusaurus/plugin-content-docs";

const sidebar: SidebarsConfig = {
  apisidebar: [
    {
      type: "doc",
      id: "api/authorization/authz-api",
    },
    {
      type: "category",
      label: "metadata",
      items: [
        {
          type: "doc",
          id: "api/authorization/ping",
          label: "Checks if service is responsive",
          className: "api-method get",
        },
      ],
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
    {
      type: "category",
      label: "permissions",
      items: [
        {
          type: "doc",
          id: "api/authorization/permission-check",
          label: "Checks if the current user has a specific permission",
          className: "api-method post",
        },
      ],
    },
  ],
};

export default sidebar.apisidebar;
