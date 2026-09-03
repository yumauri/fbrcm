import assert from 'node:assert/strict'
import test from 'node:test'
import { fileURLToPath } from 'node:url'

import {
  absolutizeMarkdownLinks,
  isLlmsFullPage,
  isRedirectPage,
  renderLlmsFull
} from './llms-full.mjs'

test('includes documentation sections but not the landing or privacy pages', () => {
  assert.equal(isLlmsFullPage('guide/index.md'), true)
  assert.equal(isLlmsFullPage('automation/json-contract.md'), true)
  assert.equal(isLlmsFullPage('automation/mcp.md'), true)
  assert.equal(isLlmsFullPage('index.md'), false)
  assert.equal(isLlmsFullPage('privacy-policy.md'), false)
})

test('includes the MCP guide in the website documentation bundle', async () => {
  const sourceDirectory = fileURLToPath(new URL('../site', import.meta.url))
  const markdown = await renderLlmsFull(sourceDirectory)

  assert.match(markdown, /Source: \[\/automation\/mcp\.md\]\(\/automation\/mcp\.md\)/)
  assert.match(markdown, /# MCP server/)
  assert.match(markdown, /parameters\.get/)
  assert.match(markdown, /\[MCP server\]\(\/automation\/mcp\.md\)/)
})

test('makes page-relative links work from the root documentation bundle', () => {
  const markdown = `- [Authentication](./authentication.md#quota)
- [Hooks](../reference/hooks.md)
- [Guide](/guide/index.md)
- [External](https://example.com/docs)
`

  assert.equal(
    absolutizeMarkdownLinks(markdown, 'guide/index.md'),
    `- [Authentication](/guide/authentication.md#quota)
- [Hooks](/reference/hooks.md)
- [Guide](/guide/index.md)
- [External](https://example.com/docs)
`
  )
})

test('recognizes redirect-only compatibility pages', () => {
  assert.equal(
    isRedirectPage(`---
title: Moved
head:
  - - meta
    - http-equiv: refresh
      content: 0; url=/guide/
---

# Moved
`),
    true
  )
  assert.equal(isRedirectPage('# Current documentation\n'), false)
})
