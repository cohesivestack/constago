import { defineConfig } from 'astro/config';
import sitemap from '@astrojs/sitemap';
import starlight from '@astrojs/starlight';
import tailwindcss from '@tailwindcss/vite';
import { docsSidebar } from './src/data/sidebar.mjs';
import { starlightSeo } from './src/plugins/starlight-seo.mjs';

export default defineConfig({
  site: 'https://constago.build',
  redirects: {
    '/getting-started/': '/',
  },
  integrations: [
    sitemap(),
    starlight({
      title: 'Constago',
      description: 'Constago generates type-safe Go constants, field accessors, and getter methods from struct fields and tags.',
      favicon: '/favicon.png',
      social: [
        {
          icon: 'github',
          label: 'GitHub',
          href: 'https://github.com/cohesivestack/constago',
        },
      ],
      customCss: ['./src/styles/global.css'],
      plugins: [starlightSeo()],
      sidebar: docsSidebar,
    }),
  ],
  vite: {
    plugins: [tailwindcss()],
  },
});
