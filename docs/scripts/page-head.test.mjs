import assert from 'node:assert/strict'
import test from 'node:test'

import {
  markdownPathForRoute,
  pageHeadDevPlugin,
  pageRoute
} from './page-head.mjs'

test('maps Markdown source paths to clean website routes', () => {
  assert.equal(pageRoute('index.md'), '/')
  assert.equal(pageRoute('guide/index.md'), '/guide/')
  assert.equal(pageRoute('reference/configuration.md'), '/reference/configuration')
  assert.equal(pageRoute('privacy-policy.md'), '/privacy-policy')
})

test('normalizes source paths from Windows builds', () => {
  assert.equal(pageRoute('reference\\configuration.md'), '/reference/configuration')
})

test('maps clean development routes back to Markdown files', () => {
  assert.equal(markdownPathForRoute('/'), '/index.md')
  assert.equal(markdownPathForRoute('/guide/'), '/guide/index.md')
  assert.equal(
    markdownPathForRoute('/automation/json-contract?preview=1'),
    '/automation/json-contract.md'
  )
})

test('adds page discovery links to development HTML', () => {
  const tags = pageHeadDevPlugin('https://docs.example.com').transformIndexHtml('', {
    originalUrl: '/automation/json-contract?preview=1'
  })

  assert.deepEqual(tags, [
    {
      tag: 'link',
      attrs: {
        rel: 'canonical',
        href: 'https://docs.example.com/automation/json-contract'
      },
      injectTo: 'head'
    },
    {
      tag: 'link',
      attrs: {
        rel: 'alternate',
        type: 'text/markdown',
        href: '/automation/json-contract.md'
      },
      injectTo: 'head'
    },
    {
      tag: 'link',
      attrs: { rel: 'describedby', type: 'text/plain', href: '/llms.txt' },
      injectTo: 'head'
    }
  ])
})
