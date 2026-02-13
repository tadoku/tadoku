import type {SidebarsConfig} from '@docusaurus/plugin-content-docs';

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
      items: ['account-deletion'],
    },
  ],
};

export default sidebars;
