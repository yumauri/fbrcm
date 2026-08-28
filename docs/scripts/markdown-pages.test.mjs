import assert from 'node:assert/strict'
import test from 'node:test'

import { renderMarkdownPage } from './markdown-pages.mjs'

test('removes frontmatter and creates useful landing-page Markdown', () => {
  const source = `---
layout: home
title: Firebase Remote Config Manager
hero:
  tagline: Manage Remote Config from your terminal
  actions:
    - text: Get started
      link: /guide/
---
`

  assert.equal(
    renderMarkdownPage(source, 'index.md'),
    `# fbrcm - Firebase Remote Config Manager

Manage Remote Config from your terminal

[Get started](/guide/index.md)
`
  )
})

test('flattens content tabs into named Markdown sections', () => {
  const source = `# Install

<ContentTabs
  :tabs="[
    { id: 'macos', label: 'macOS' },
    { id: 'linux', label: 'Linux' }
  ]"
>
<template #macos>

Run macOS command.

</template>
<template #linux>

Run Linux command.

</template>
</ContentTabs>
`

  assert.equal(
    renderMarkdownPage(source, 'guide/index.md'),
    `# Install

## macOS

Run macOS command.

## Linux

Run Linux command.
`
  )
})

test('rewrites documentation links while preserving assets and external links', () => {
  const source = `# Links

- [Guide](/guide/)
- [Authentication](./authentication#quota)
- [License](/LICENSE.txt)
- [Website](https://example.com/docs)
- ![Demo](/demo.gif)
`

  assert.equal(
    renderMarkdownPage(source, 'guide/index.md'),
    `# Links

- [Guide](/guide/index.md)
- [Authentication](./authentication.md#quota)
- [License](/LICENSE.txt)
- [Website](https://example.com/docs)
- ![Demo](/demo.gif)
`
  )
})
