export function isDev() {
  return import.meta.env.CF_PAGES_BRANCH !== 'main' || import.meta.env.DEV;
}