import {themes as prismThemes} from 'prism-react-renderer';
import type {Config} from '@docusaurus/types';
import type * as Preset from '@docusaurus/preset-classic';
import type * as OpenApiPlugin from 'docusaurus-plugin-openapi-docs';

const config: Config = {
  title: 'Tadoku',
  tagline: 'Tadoku developer documentation',
  favicon: 'img/favicon.ico',

  url: 'https://tadoku.github.io',
  baseUrl: '/tadoku/',

  organizationName: 'tadoku',
  projectName: 'tadoku',

  onBrokenLinks: 'warn',

  markdown: {
    hooks: {
      onBrokenMarkdownLinks: 'warn',
    },
  },

  i18n: {
    defaultLocale: 'en',
    locales: ['en'],
  },

  presets: [
    [
      'classic',
      {
        docs: {
          sidebarPath: './sidebars.ts',
          editUrl: 'https://github.com/tadoku/tadoku/tree/main/docs/',
          routeBasePath: '/',
          docItemComponent: '@theme/ApiItem',
        },
        blog: false,
        theme: {
          customCss: './src/css/custom.css',
        },
      } satisfies Preset.Options,
    ],
  ],

  plugins: [
    'docusaurus-plugin-sass',
    [
      'docusaurus-plugin-openapi-docs',
      {
        id: 'openapi',
        docsPluginId: 'classic',
        config: {
          immersion: {
            specPath: '../services/immersion-api/http/rest/openapi/api.yaml',
            outputDir: 'docs/api/immersion',
            label: 'Immersion API',
            downloadUrl:
              'https://raw.githubusercontent.com/tadoku/tadoku/main/services/immersion-api/http/rest/openapi/api.yaml',
            hideSendButton: true,
            sidebarOptions: {
              groupPathsBy: 'tag',
              categoryLinkSource: 'tag',
            },
          } satisfies OpenApiPlugin.Options,
          content: {
            specPath: '../services/content-api/http/rest/openapi/api.yaml',
            outputDir: 'docs/api/content',
            label: 'Content API',
            downloadUrl:
              'https://raw.githubusercontent.com/tadoku/tadoku/main/services/content-api/http/rest/openapi/api.yaml',
            hideSendButton: true,
            sidebarOptions: {
              groupPathsBy: 'tag',
              categoryLinkSource: 'tag',
            },
          } satisfies OpenApiPlugin.Options,
          profile: {
            specPath: '../services/profile-api/http/rest/openapi/api.yaml',
            outputDir: 'docs/api/profile',
            label: 'Profile API',
            downloadUrl:
              'https://raw.githubusercontent.com/tadoku/tadoku/main/services/profile-api/http/rest/openapi/api.yaml',
            hideSendButton: true,
            sidebarOptions: {
              groupPathsBy: 'tag',
              categoryLinkSource: 'tag',
            },
          } satisfies OpenApiPlugin.Options,
          authorization: {
            specPath: '../services/authz-api/http/rest/openapi/api.yaml',
            outputDir: 'docs/api/authorization',
            label: 'Authorization API',
            downloadUrl:
              'https://raw.githubusercontent.com/tadoku/tadoku/main/services/authz-api/http/rest/openapi/api.yaml',
            hideSendButton: true,
            sidebarOptions: {
              groupPathsBy: 'tag',
              categoryLinkSource: 'tag',
            },
          } satisfies OpenApiPlugin.Options,
        },
      },
    ],
  ],

  themes: ['docusaurus-theme-openapi-docs'],

  themeConfig: {
    colorMode: {
      respectPrefersColorScheme: true,
    },
    navbar: {
      title: 'Tadoku',
      items: [
        {
          type: 'docSidebar',
          sidebarId: 'docs',
          position: 'left',
          label: 'Docs',
        },
        {
          to: '/api/',
          position: 'left',
          label: 'API Reference',
        },
        {
          href: 'https://github.com/tadoku/tadoku',
          label: 'GitHub',
          position: 'right',
        },
      ],
    },
    footer: {
      style: 'dark',
      links: [
        {
          title: 'Documentation',
          items: [
            {
              label: 'Getting Started',
              to: '/',
            },
            {
              label: 'Architecture',
              to: '/architecture',
            },
          ],
        },
        {
          title: 'Links',
          items: [
            {
              label: 'Tadoku',
              href: 'https://tadoku.app',
            },
            {
              label: 'GitHub',
              href: 'https://github.com/tadoku/tadoku',
            },
          ],
        },
      ],
      copyright: `Copyright \u00a9 ${new Date().getFullYear()} Tadoku. Built with Docusaurus.`,
    },
    prism: {
      theme: prismThemes.github,
      darkTheme: prismThemes.dracula,
      additionalLanguages: ['bash', 'sql', 'go', 'yaml', 'json'],
    },
    api: {
      authPersistance: false,
      requestCredentials: 'omit',
    },
  } satisfies Preset.ThemeConfig,
};

export default config;
