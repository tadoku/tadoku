import type {SidebarsConfig} from '@docusaurus/plugin-content-docs';
import authorizationApiSidebar from './docs/api/authorization/sidebar';
import contentApiSidebar from './docs/api/content/sidebar';
import immersionApiSidebar from './docs/api/immersion/sidebar';
import profileApiSidebar from './docs/api/profile/sidebar';

const sidebars: SidebarsConfig = {
  docs: [
    {
      type: 'category',
      label: 'Introduction',
      items: ['getting-started'],
    },
    {
      type: 'category',
      label: 'Development',
      items: ['architecture', 'adr', 'local-environment'],
    },
    {
      type: 'category',
      label: 'Services',
      items: [
        'services/immersion-api',
        'services/content-api',
        'services/authorization',
        'services/s2s-auth',
      ],
    },
    {
      type: 'category',
      label: 'API Reference',
      link: {
        type: 'doc',
        id: 'api/index',
      },
      items: [
        {
          type: 'category',
          label: 'Immersion API',
          link: {
            type: 'doc',
            id: 'api/immersion/immersion-api',
          },
          items: immersionApiSidebar.slice(1),
        },
        {
          type: 'category',
          label: 'Content API',
          link: {
            type: 'doc',
            id: 'api/content/content-api',
          },
          items: contentApiSidebar.slice(1),
        },
        {
          type: 'category',
          label: 'Profile API',
          link: {
            type: 'doc',
            id: 'api/profile/profile-api',
          },
          items: profileApiSidebar.slice(1),
        },
        {
          type: 'category',
          label: 'Authorization API',
          link: {
            type: 'doc',
            id: 'api/authorization/authz-api',
          },
          items: authorizationApiSidebar.slice(1),
        },
      ],
    },
    {
      type: 'category',
      label: 'Frontend',
      items: [
        'frontend/auth',
        'frontend/styleguide',
        'frontend/webv2',
      ],
    },
    {
      type: 'category',
      label: 'Jobs',
      items: ['jobs/postgres-backup'],
    },
    {
      type: 'category',
      label: 'Administration',
      items: ['account-deletion', 'database-migration-recovery'],
    },
  ],
};

export default sidebars;
