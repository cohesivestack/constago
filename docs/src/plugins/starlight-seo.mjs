/**
 * Adds route-level SEO metadata that depends on the final Starlight route.
 */
export function starlightSeo() {
  return {
    name: 'constago-starlight-seo',
    hooks: {
      'config:setup'({ addRouteMiddleware }) {
        addRouteMiddleware({
          entrypoint: new URL('../route-data/seo.mjs', import.meta.url).pathname,
          order: 'post',
        });
      },
    },
  };
}
