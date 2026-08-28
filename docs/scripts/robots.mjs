export function renderRobotsTxt(siteUrl) {
  const lines = ['User-agent: *', 'Allow: /']

  if (siteUrl) {
    lines.push('', `Sitemap: ${siteUrl}/sitemap.xml`)
  }

  return `${lines.join('\n')}\n`
}

export function robotsDevPlugin(siteUrl) {
  return {
    name: 'fbrcm-robots',
    configureServer(server) {
      server.middlewares.use((request, response, next) => {
        const pathname = request.url?.split('?', 1)[0]
        if (pathname !== '/robots.txt') return next()

        response.statusCode = 200
        response.setHeader('Content-Type', 'text/plain; charset=utf-8')
        response.end(renderRobotsTxt(siteUrl))
      })
    }
  }
}
