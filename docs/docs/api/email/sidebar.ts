import type { SidebarsConfig } from "@docusaurus/plugin-content-docs";

const sidebar: SidebarsConfig = {
  apisidebar: [
    {
      type: "doc",
      id: "api/email/email-api",
    },
    {
      type: "category",
      label: "emails",
      items: [
        {
          type: "doc",
          id: "api/email/email-list",
          label: "List retained tenant emails, newest first",
          className: "api-method get",
        },
        {
          type: "doc",
          id: "api/email/email-get",
          label: "Read a retained tenant email including parameters",
          className: "api-method get",
        },
      ],
    },
  ],
};

export default sidebar.apisidebar;
