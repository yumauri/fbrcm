import assert from 'node:assert/strict'
import { mkdtemp, readFile, rm, stat, utimes } from 'node:fs/promises'
import { tmpdir } from 'node:os'
import { join } from 'node:path'
import test from 'node:test'
import { fileURLToPath } from 'node:url'

import { commandSidebar } from './command-navigation.mjs'
import { renderCommandReference, writeCommandReference } from './command-reference.mjs'

const docsDirectory = fileURLToPath(new URL('..', import.meta.url))
const cliMarkdown = await readFile(new URL('../CLI.md', import.meta.url), 'utf8')
const capabilities = JSON.parse(await readFile(new URL('../../schemas/capabilities.json', import.meta.url), 'utf8'))
const indexPath = 'reference/commands/index.md'

test('the command index links to one dedicated page for every executable command', async () => {
  const pages = await renderCommandReference(cliMarkdown, capabilities, docsDirectory)
  const index = pages.get(indexPath)
  const expectedRoutes = capabilities.map(({ path }) => `/reference/commands/${path.join('/') || 'root'}`)
  const indexRows = [...index.matchAll(/^\| \[`[^`]+`\]\(([^)]+)\) \|/gm)].map(([, route]) => route)
  const sidebar = commandSidebar(capabilities)
  const sidebarGroups = sidebar.items.filter(({ items }) => items)

  assert.deepEqual([...indexRows].sort(), expectedRoutes.sort())
  assert.equal(sidebar.link, '/reference/commands/')
  assert.equal(sidebar.collapsed, true)
  assert.ok(sidebarGroups.every(({ collapsed }) => collapsed))
  assert.deepEqual(sidebarGroups.flatMap(({ items }) => items.map(({ link }) => link)), indexRows)
  for (const { path } of capabilities) {
    const name = ['fbrcm', ...path].join(' ')
    const route = `/reference/commands/${path.join('/') || 'root'}`
    const page = pages.get(`${route.slice(1)}.md`)
    assert.ok(page, `missing page for ${name}`)
    assert.ok(page.includes(`# \`${name}\``))
    assert.ok(page.includes(`\`\`\`text\n${name}`))
    assert.ok(page.includes('[Command index](/reference/commands/)'))
  }
})

test('keeps the command-specific reference and links to shared behavior', async () => {
  const pages = await renderCommandReference(cliMarkdown, capabilities, docsDirectory)
  const explain = pages.get('reference/commands/explain.md')
  const move = pages.get('reference/commands/conditions/move.md')
  const fish = pages.get('reference/commands/completion/fish.md')
  const shared = pages.get('reference/command-behavior.md')

  assert.match(explain, /<Badge type="info" text="Since 1\.1\.0" \/>\n\n# `fbrcm explain`/)
  assert.match(explain, /^## Supply request inputs$/m)
  assert.match(explain, /\]\(\/automation\/json-contract#parameter-explanations\)/)
  assert.match(move, /fbrcm conditions move <project> <condition> <priority>/)
  assert.match(move, /\/reference\/command-behavior#condition-mutations/)
  assert.doesNotMatch(move, /fbrcm conditions delete/)
  assert.match(shared, /^## Condition mutations$/m)
  assert.match(shared, /--plan-out <path>/)
  assert.match(fish, /fbrcm completion fish \| source/)
})

test('fails when a command is added without a CLI reference section', async () => {
  const missing = { id: 'groups.missing', path: ['groups', 'missing'], summary: 'Missing command documentation' }
  await assert.rejects(
    renderCommandReference(cliMarkdown, [...capabilities, missing], docsDirectory),
    /Missing CLI reference section for fbrcm groups missing/
  )
})

test('regeneration keeps existing pages in place and removes only obsolete pages', async (t) => {
  const site = await mkdtemp(join(tmpdir(), 'fbrcm-command-reference-'))
  t.after(() => rm(site, { recursive: true, force: true }))

  const explain = 'reference/commands/explain.md'
  const stale = 'reference/commands/projects/obsolete.md'
  await writeCommandReference(site, new Map([
    [explain, '# Explain\n'],
    [indexPath, '# Old index\n'],
    [stale, '# Obsolete\n']
  ]))
  const timestamp = new Date('2026-01-01T00:00:00Z')
  await utimes(join(site, explain), timestamp, timestamp)
  const previousIndex = await stat(join(site, indexPath))

  await writeCommandReference(site, new Map([
    [explain, '# Explain\n'],
    [indexPath, '# New index\n']
  ]))

  assert.equal((await stat(join(site, explain))).mtimeMs, timestamp.getTime())
  assert.equal((await stat(join(site, indexPath))).ino, previousIndex.ino)
  assert.equal(await readFile(join(site, indexPath), 'utf8'), '# New index\n')
  await assert.rejects(stat(join(site, stale)), { code: 'ENOENT' })
})
