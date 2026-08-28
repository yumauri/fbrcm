export function pageRoute(relativePath) {
  const normalizedPath = relativePath.replaceAll('\\', '/')

  if (normalizedPath === 'index.md') return '/'
  if (normalizedPath.endsWith('/index.md')) {
    return `/${normalizedPath.slice(0, -'index.md'.length)}`
  }

  return `/${normalizedPath.slice(0, -'.md'.length)}`
}

export function markdownPathForRoute(route) {
  const pathname = route.split(/[?#]/, 1)[0] || '/'
  return pathname.endsWith('/') ? `${pathname}index.md` : `${pathname}.md`
}

export function pageHeadDevPlugin(siteUrl) {
  return {
    name: 'fbrcm-page-head',
    apply: 'serve',
    transformIndexHtml(_html, context) {
      const route = context.originalUrl ?? context.path ?? '/'
      const pathname = route.split(/[?#]/, 1)[0] || '/'
      const tags = [
        {
          tag: 'link',
          attrs: {
            rel: 'alternate',
            type: 'text/markdown',
            href: markdownPathForRoute(route)
          },
          injectTo: 'head'
        },
        {
          tag: 'link',
          attrs: { rel: 'describedby', type: 'text/plain', href: '/llms.txt' },
          injectTo: 'head'
        }
      ]

      if (siteUrl) {
        tags.unshift({
          tag: 'link',
          attrs: { rel: 'canonical', href: `${siteUrl}${pathname}` },
          injectTo: 'head'
        })
      }

      return tags
    }
  }
}
