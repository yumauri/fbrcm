import { mkdir, readFile, writeFile } from 'node:fs/promises'
import { dirname, posix, resolve } from 'node:path'

import { listMarkdownPages, renderMarkdownPage } from './markdown-pages.mjs'

const documentationSections = ['guide/', 'tui/', 'cli/', 'automation/', 'reference/']

export function isLlmsFullPage(relativePath) {
  return documentationSections.some((section) => relativePath.startsWith(section))
}

export function isRedirectPage(source) {
  const frontmatter = source.match(/^---\r?\n([\s\S]*?)\r?\n---(?:\r?\n|$)/)?.[1]
  return frontmatter
    ? /^\s*(?:-\s*)?http-equiv:\s*refresh\s*$/im.test(frontmatter)
    : false
}

function compareDocumentationPages(left, right) {
  const leftSection = documentationSections.findIndex((section) =>
    left.startsWith(section)
  )
  const rightSection = documentationSections.findIndex((section) =>
    right.startsWith(section)
  )

  if (leftSection !== rightSection) return leftSection - rightSection
  if (left.endsWith('/index.md') !== right.endsWith('/index.md')) {
    return left.endsWith('/index.md') ? -1 : 1
  }

  return left.localeCompare(right)
}

function absoluteLinkTarget(target, relativePath) {
  if (
    target.startsWith('#') ||
    target.startsWith('/') ||
    target.startsWith('//') ||
    /^[a-z][a-z\d+.-]*:/i.test(target)
  ) {
    return target
  }

  const fragmentIndex = target.indexOf('#')
  const fragment = fragmentIndex >= 0 ? target.slice(fragmentIndex) : ''
  const pathAndQuery = fragmentIndex >= 0 ? target.slice(0, fragmentIndex) : target
  const queryIndex = pathAndQuery.indexOf('?')
  const query = queryIndex >= 0 ? pathAndQuery.slice(queryIndex) : ''
  const path = queryIndex >= 0 ? pathAndQuery.slice(0, queryIndex) : pathAndQuery

  return `${posix.resolve('/', posix.dirname(relativePath), path)}${query}${fragment}`
}

export function absolutizeMarkdownLinks(markdown, relativePath) {
  return markdown.replace(/\]\(([^)]+)\)/g, (link, destination) => {
    const match = destination.match(/^(\S+)([\s\S]*)$/)
    if (!match) return link

    return `](${absoluteLinkTarget(match[1], relativePath)}${match[2]})`
  })
}

export async function renderLlmsFull(sourceDirectory) {
  const pages = (await listMarkdownPages(sourceDirectory))
    .filter(isLlmsFullPage)
    .sort(compareDocumentationPages)
  const documents = (
    await Promise.all(
      pages.map(async (page) => {
        const source = await readFile(resolve(sourceDirectory, ...page.split('/')), 'utf8')
        if (isRedirectPage(source)) return undefined

        return {
          page,
          markdown: absolutizeMarkdownLinks(
            renderMarkdownPage(source, page),
            page
          ).trim()
        }
      })
    )
  ).filter(Boolean)

  const sections = [
    '# fbrcm documentation',
    '> Complete website documentation for fbrcm, generated from the same Markdown sources as the human-readable site.',
    'For a curated entry point, see [llms.txt](/llms.txt).',
    ...documents.map(
      ({ page, markdown }) => `---\n\nSource: [/${page}](/${page})\n\n${markdown}`
    )
  ]

  return `${sections.join('\n\n')}\n`
}

export async function writeLlmsFull(sourceDirectory, outputPath) {
  await mkdir(dirname(outputPath), { recursive: true })
  await writeFile(outputPath, await renderLlmsFull(sourceDirectory))
}
