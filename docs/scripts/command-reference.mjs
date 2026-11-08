import { mkdir, readFile, rm, writeFile } from 'node:fs/promises'
import { dirname, resolve } from 'node:path'
import { createMarkdownRenderer } from 'vitepress'

import { commandBehaviorLink, commandIndexLink, groupCommands } from './command-navigation.mjs'
import { listMarkdownPages } from './markdown-pages.mjs'

const indexDirectory = commandIndexLink.slice(1)

const sharedSections = [
  ['Positional project resolution', /^(?:project\.(?:show|open|export|import|defaults)|projects\.(?:diff|promote))$/],
  ['Remote Config cache freshness', /^(?:get|explain|conditions\.(?:list|show)|groups\.list|(?:experiments|rollouts|personalizations)\.(?:list|show))$/],
  ['Condition mutations', /^conditions\.(?:add|edit|rename|move|delete)$/],
  ['Group mutations', /^groups\.(?:add|edit|rename|delete)$/],
  ['Project quota-project overrides', /^project\.quota-project\./],
  ['Auth identity quota-project defaults', /^auth\.quota-project\./],
  ['Firebase applications', /^apps\./],
  ['Remote Config managed features', /^(?:experiments|rollouts|personalizations)\./],
  ['Remote Config version history', /^versions\./],
  ['Persistent project visibility', /^projects\.(?:blacklist|whitelist)\./],
  ['Shell completion', /^completion\./]
]

const siteLinks = new Map([
  ['TUI.md#context', '/tui/context'],
  ['MCP.md', '/automation/mcp'],
  ['EXPR.md', '/reference/filtering'],
  ['#filter-queries', '/reference/filtering#query-modes'],
  ['theming.md', '/reference/themes'],
  ['cli-contract.md#parameter-explanations', '/automation/json-contract#parameter-explanations'],
  ...sharedSections.map(([title]) => [`#${slug(title)}`, `${commandBehaviorLink}#${slug(title)}`])
])

function slug(title) {
  return title.toLowerCase().replaceAll(' ', '-')
}

function rewriteLinks(markdown) {
  return markdown.replace(/\]\(([^)\s]+)\)/g, (link, target) =>
    siteLinks.has(target) ? `](${siteLinks.get(target)})` : link
  )
}

function sectionsByHeading(markdown, renderer) {
  const lines = markdown.split(/\r?\n/)
  const tokens = renderer.parse(markdown, {})
  const headings = tokens.flatMap((token, index) =>
    token.type === 'heading_open' && token.level === 0
      ? [{ title: tokens[index + 1].content, depth: Number(token.tag[1]), line: token.map[0] }]
      : []
  )
  const sections = new Map()
  for (const [index, heading] of headings.entries()) {
    if (heading.depth !== 3) continue
    const following = headings.slice(index + 1)
    const end = following.find(({ depth }) => depth <= heading.depth)?.line ?? lines.length
    const body = lines.slice(heading.line + 1, end)
    for (const { line } of following) {
      if (line >= end) break
      body[line - heading.line - 1] = lines[line].replace(/^##/, '')
    }
    if (sections.has(heading.title)) throw new Error(`Duplicate CLI reference heading: ${heading.title}`)
    sections.set(heading.title, rewriteLinks(body.join('\n').trim()))
  }
  return sections
}

function commandSections(sections) {
  const commands = new Map()
  for (const [heading, body] of sections) {
    if (!heading.startsWith('`fbrcm')) continue
    const usage = heading.slice(1, -1)
    const name = usage.match(/^fbrcm(?: [a-z][a-z0-9-]*)*/)[0]
    if (commands.has(name)) throw new Error(`Multiple CLI reference sections for ${name}`)
    commands.set(name, { usage, body })
  }
  return commands
}

function commandPage({ id, name }, section) {
  const availability = section.body.match(/^\*\*(Since \d+\.\d+\.\d+)\*\*\n\n/)
  const related = sharedSections
    .filter(([, match]) => match.test(id))
    .map(([title]) => `[${title}](${commandBehaviorLink}#${slug(title)})`)
  if (id === 'explain') related.push('[Explain a parameter value](/cli/explain)')

  return [
    '---\neditLink: false\n---',
    ...(availability ? [`<Badge type="info" text="${availability[1]}" />`] : []),
    `# \`${name}\``,
    `\`\`\`text\n${section.usage}\n\`\`\``,
    section.body.slice(availability?.[0].length ?? 0),
    `## See also\n\n${[...related, `[Command index](${commandIndexLink})`].map((link) => `- ${link}`).join('\n')}`
  ].join('\n\n') + '\n'
}

export async function renderCommandReference(cliMarkdown, capabilities, sourceDirectory) {
  const renderer = await createMarkdownRenderer(sourceDirectory)
  const sections = sectionsByHeading(cliMarkdown, renderer)
  const commands = commandSections(sections)
  const pages = new Map()
  const groups = groupCommands(capabilities)

  for (const entry of groups.flatMap(([, entries]) => entries)) {
    const path = `${entry.link.slice(1)}.md`
    if (pages.has(path)) throw new Error(`Duplicate command reference route for ${entry.name}`)
    const section = commands.get(entry.name)
    if (!section?.body) throw new Error(`Missing CLI reference section for ${entry.name}`)
    commands.delete(entry.name)
    pages.set(path, commandPage(entry, section))
  }
  if (commands.size) throw new Error(`CLI reference sections without commands: ${[...commands.keys()].join(', ')}`)

  const index = [
    '---\neditLink: false\n---',
    '# Command index',
    'Each command links to its syntax, options, and behavior. Use `fbrcm help <command>` to see the flags supported by your installed version.',
    `For common rules on selectors, caches, and command families, see [Shared command behavior](${commandBehaviorLink}).`,
    'For first-time setup, follow [Getting started](/guide/). For machine discovery, use [`fbrcm capabilities`](/reference/commands/capabilities) with `--json`.',
    ...groups.map(([title, entries]) =>
      [
        `## ${title}`, '', '| Command | Purpose |', '| --- | --- |',
        ...entries.map(({ text, link, summary }) => `| [\`${text}\`](${link}) | ${summary.replaceAll('|', '\\|')} |`)
      ].join('\n')
    )
  ]
  pages.set(`${indexDirectory}index.md`, index.join('\n\n') + '\n')

  const shared = sharedSections.map(([title]) => {
    const section = sections.get(title)
    if (!section) throw new Error(`Missing shared CLI reference section: ${title}`)
    return `## ${title}\n\n${section}`
  })
  pages.set(`${commandBehaviorLink.slice(1)}.md`, [
    '---\neditLink: false\n---',
    '# Shared command behavior',
    `Use the [command index](${commandIndexLink}) for command-specific syntax, options, and examples.`,
    ...shared
  ].join('\n\n') + '\n')
  return pages
}

export async function writeCommandReference(siteDirectory, pages) {
  const commandDirectory = resolve(siteDirectory, indexDirectory)
  await mkdir(commandDirectory, { recursive: true })
  const existing = await listMarkdownPages(commandDirectory)

  await Promise.all([...pages].map(async ([path, markdown]) => {
    const target = resolve(siteDirectory, path)
    const current = await readFile(target, 'utf8').catch((error) => {
      if (error.code !== 'ENOENT') throw error
    })
    if (current === markdown) return

    await mkdir(dirname(target), { recursive: true })
    await writeFile(target, markdown)
  }))

  await Promise.all(existing
    .filter((path) => !pages.has(indexDirectory + path))
    .map((path) => rm(resolve(commandDirectory, path), { force: true })))
}
