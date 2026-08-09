import type { SidebarsConfig } from "@docusaurus/plugin-content-docs";

const sidebar: SidebarsConfig = {
  apisidebar: [
    {
      type: "doc",
      id: "api/content/content-api",
    },
    {
      type: "category",
      label: "pages",
      items: [
        {
          type: "doc",
          id: "api/content/page-find-by-slug",
          label: "Returns page content for a given slug",
          className: "api-method get",
        },
        {
          type: "doc",
          id: "api/content/page-update",
          label: "Updates an existing page",
          className: "api-method put",
        },
        {
          type: "doc",
          id: "api/content/page-delete",
          label: "Deletes an existing page",
          className: "api-method delete",
        },
        {
          type: "doc",
          id: "api/content/page-create",
          label: "Creates a new page",
          className: "api-method post",
        },
        {
          type: "doc",
          id: "api/content/page-list",
          label: "lists all pages",
          className: "api-method get",
        },
        {
          type: "doc",
          id: "api/content/page-version-list",
          label: "Lists all versions of a page",
          className: "api-method get",
        },
        {
          type: "doc",
          id: "api/content/page-version-get",
          label: "Gets a specific version of a page",
          className: "api-method get",
        },
      ],
    },
    {
      type: "category",
      label: "posts",
      items: [
        {
          type: "doc",
          id: "api/content/post-find-by-slug",
          label: "Returns page content for a given slug",
          className: "api-method get",
        },
        {
          type: "doc",
          id: "api/content/post-update",
          label: "Updates an existing post",
          className: "api-method put",
        },
        {
          type: "doc",
          id: "api/content/post-delete",
          label: "Deletes an existing post",
          className: "api-method delete",
        },
        {
          type: "doc",
          id: "api/content/post-create",
          label: "Creates a new post",
          className: "api-method post",
        },
        {
          type: "doc",
          id: "api/content/post-list",
          label: "lists all posts",
          className: "api-method get",
        },
        {
          type: "doc",
          id: "api/content/post-version-list",
          label: "Lists all versions of a post",
          className: "api-method get",
        },
        {
          type: "doc",
          id: "api/content/post-version-get",
          label: "Gets a specific version of a post",
          className: "api-method get",
        },
      ],
    },
    {
      type: "category",
      label: "announcements",
      items: [
        {
          type: "doc",
          id: "api/content/announcement-list-active",
          label: "Lists currently active announcements",
          className: "api-method get",
        },
        {
          type: "doc",
          id: "api/content/announcement-create",
          label: "Creates a new announcement",
          className: "api-method post",
        },
        {
          type: "doc",
          id: "api/content/announcement-list",
          label: "Lists all announcements",
          className: "api-method get",
        },
        {
          type: "doc",
          id: "api/content/announcement-find-by-id",
          label: "Gets an announcement by ID",
          className: "api-method get",
        },
        {
          type: "doc",
          id: "api/content/announcement-update",
          label: "Updates an existing announcement",
          className: "api-method put",
        },
        {
          type: "doc",
          id: "api/content/announcement-delete",
          label: "Deletes an existing announcement",
          className: "api-method delete",
        },
      ],
    },
    {
      type: "category",
      label: "metadata",
      items: [
        {
          type: "doc",
          id: "api/content/ping",
          label: "Checks if service is responsive",
          className: "api-method get",
        },
      ],
    },
  ],
};

export default sidebar.apisidebar;
