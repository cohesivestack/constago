import { defineRouteMiddleware } from '@astrojs/starlight/route-data';

export const onRequest = defineRouteMiddleware((context) => {
  if (context.url.pathname !== '/') return;

  context.locals.starlightRoute.head.push({
    tag: 'script',
    attrs: {
      type: 'application/ld+json',
    },
    content: JSON.stringify({
      '@context': 'https://schema.org',
      '@type': 'WebSite',
      name: 'Constago',
      alternateName: ['Constago Go code generator', 'Constago'],
      url: 'https://constago.build/',
    }),
  });
});
