import assert from 'node:assert/strict'
import test from 'node:test'

import { renderRobotsTxt, robotsDevPlugin } from './robots.mjs'

test('allows all crawlers and advertises the absolute sitemap', () => {
  assert.equal(
    renderRobotsTxt('https://docs.example.com'),
    `User-agent: *
Allow: /

Sitemap: https://docs.example.com/sitemap.xml
`
  )
})

test('allows all crawlers when the site URL is not configured', () => {
  assert.equal(
    renderRobotsTxt(undefined),
    `User-agent: *
Allow: /
`
  )
})

test('serves robots.txt from the development server', () => {
  let middleware
  const headers = new Map()
  let body

  robotsDevPlugin('https://docs.example.com').configureServer({
    middlewares: {
      use(handler) {
        middleware = handler
      }
    }
  })

  middleware(
    { url: '/robots.txt?cache-bust=1' },
    {
      setHeader(name, value) {
        headers.set(name, value)
      },
      end(value) {
        body = value
      }
    },
    () => assert.fail('robots.txt should be handled by the plugin')
  )

  assert.equal(headers.get('Content-Type'), 'text/plain; charset=utf-8')
  assert.equal(body, renderRobotsTxt('https://docs.example.com'))
})

test('leaves other development routes to VitePress', () => {
  let middleware
  let continued = false

  robotsDevPlugin(undefined).configureServer({
    middlewares: {
      use(handler) {
        middleware = handler
      }
    }
  })

  middleware({}, {}, () => {
    continued = true
  })

  assert.equal(continued, true)
})
