import {writeFile} from 'node:fs/promises';
import {join} from 'node:path';
import {themes as prismThemes} from 'prism-react-renderer';
import type {Config, Plugin} from '@docusaurus/types';
import type {LoadedContent} from '@docusaurus/plugin-content-docs';
import type * as Preset from '@docusaurus/preset-classic';
import type * as OpenApiPlugin from 'docusaurus-plugin-openapi-docs';

const llmsSections: [string, string][] = [
  ['Start here', 'index'],
  ['Develop', 'develop'],
  ['Architecture', 'architecture'],
  ['Decisions', 'adr'],
  ['Tadoku API', 'tadoku-api'],
  ['Frontend', 'frontend'],
  ['Operations', 'operations'],
  ['API reference', 'api'],
];

const isOverview = (id: string, key: string) => id === key || id === `${key}/index`;

// Publishes llms.txt: an index of the hand-written pages for agents without a checkout.
function llmsTxt(): Plugin {
  let docs: LoadedContent['loadedVersions'][number]['docs'] = [];
  return {
    name: 'llms-txt',
    allContentLoaded({allContent}) {
      const content = allContent['docusaurus-plugin-content-docs'].default as LoadedContent;
      docs = content.loadedVersions[0].docs.filter(
        doc => !/^api\/(immersion|content|profile|authorization)\//.test(doc.id),
      );
    },
    async postBuild({outDir, siteConfig}) {
      const sections = llmsSections.map(([title, key]) => {
        const entries = docs
          .filter(doc => doc.id.split('/')[0] === key)
          .sort((a, b) => Number(!isOverview(a.id, key)) - Number(!isOverview(b.id, key)) || a.title.localeCompare(b.title))
          .map(doc => `- [${doc.title}](${siteConfig.url}${doc.permalink}): ${doc.description} Source: \`${doc.source.replace('@site/', 'docs/')}\``);
        return `## ${title}\n\n${entries.join('\n')}`;
      });
      const intro = `# ${siteConfig.title}\n\n> ${siteConfig.tagline}. Each page is the Markdown file named in its Source, in https://github.com/tadoku/tadoku. In a checkout, start with AGENTS.md.`;
      await writeFile(join(outDir, 'llms.txt'), `${intro}\n\n${sections.join('\n\n')}\n`);
    },
  };
}

const config: Config = {
  title: 'Tadoku',
  tagline: 'Tadoku developer documentation',
  favicon: 'img/favicon.png',

  url: 'https://tadoku.github.io',
  baseUrl: '/tadoku/',

  organizationName: 'tadoku',
  projectName: 'tadoku',

  onBrokenLinks: 'throw',
  onBrokenAnchors: 'throw',

  markdown: {
    hooks: {
      onBrokenMarkdownLinks: 'throw',
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
    llmsTxt,
    [
      'docusaurus-plugin-openapi-docs',
      {
        id: 'openapi',
        docsPluginId: 'classic',
        config: {
          immersion: {
            specPath: '.generated/api/immersion.yaml',
            outputDir: 'docs/api/immersion',
            label: 'Immersion API',
            hideSendButton: true,
            sidebarOptions: {
              groupPathsBy: 'tag',
              categoryLinkSource: 'tag',
            },
          } satisfies OpenApiPlugin.Options,
          content: {
            specPath: '.generated/api/content.yaml',
            outputDir: 'docs/api/content',
            label: 'Content API',
            hideSendButton: true,
            sidebarOptions: {
              groupPathsBy: 'tag',
              categoryLinkSource: 'tag',
            },
          } satisfies OpenApiPlugin.Options,
          profile: {
            specPath: '.generated/api/profile.yaml',
            outputDir: 'docs/api/profile',
            label: 'Profile API',
            hideSendButton: true,
            sidebarOptions: {
              groupPathsBy: 'tag',
              categoryLinkSource: 'tag',
            },
          } satisfies OpenApiPlugin.Options,
          authorization: {
            specPath: '.generated/api/authz.yaml',
            outputDir: 'docs/api/authorization',
            label: 'Authorization API',
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
              label: 'Start here',
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
