import { existsSync, readFileSync, readdirSync } from 'node:fs';
import { extname, join, relative, resolve } from 'node:path';

const dist = resolve(import.meta.dirname, '..', 'dist');

if (!existsSync(dist)) {
  throw new Error('dist does not exist; run npm run build first');
}

function filesUnder(directory) {
  return readdirSync(directory, { withFileTypes: true }).flatMap((entry) => {
    const path = join(directory, entry.name);
    return entry.isDirectory() ? filesUnder(path) : [path];
  });
}

function targetFile(pathname) {
  const clean = decodeURIComponent(pathname).replace(/^\/+/, '');
  const direct = join(dist, clean);

  if (extname(clean)) return direct;
  if (existsSync(direct) && !clean.endsWith('/')) return direct;
  return join(direct, 'index.html');
}

const htmlFiles = filesUnder(dist).filter((file) => file.endsWith('.html'));
const failures = [];

for (const source of htmlFiles) {
  const html = readFileSync(source, 'utf8');
  const hrefs = [...html.matchAll(/\shref=(?:"([^"]+)"|'([^']+)')/g)].map(
    (match) => match[1] ?? match[2],
  );

  for (const href of hrefs) {
    if (
      href.startsWith('#') ||
      href.startsWith('http://') ||
      href.startsWith('https://') ||
      href.startsWith('mailto:') ||
      href.startsWith('tel:')
    ) {
      continue;
    }

    const url = new URL(href, 'https://constago.build/');
    const target = targetFile(url.pathname);
    if (!existsSync(target)) {
      failures.push(`${relative(dist, source)} -> ${href}`);
      continue;
    }

    if (url.hash && target.endsWith('.html')) {
      const id = decodeURIComponent(url.hash.slice(1));
      const targetHtml = readFileSync(target, 'utf8');
      const escaped = id.replace(/[.*+?^${}()|[\]\\]/g, '\\$&');
      if (!new RegExp(`\\sid=(?:"${escaped}"|'${escaped}')`).test(targetHtml)) {
        failures.push(`${relative(dist, source)} -> ${href} (missing anchor)`);
      }
    }
  }
}

if (failures.length > 0) {
  throw new Error(`Broken internal links:\n${failures.join('\n')}`);
}

console.log(`Checked internal links in ${htmlFiles.length} HTML files.`);
