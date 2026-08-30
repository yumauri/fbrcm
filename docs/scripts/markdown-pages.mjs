import { mkdir, readdir, readFile, writeFile } from 'node:fs/promises'
import { dirname, relative, resolve, sep } from 'node:path'

const frontmatterPattern = /^---\r?\n([\s\S]*?)\r?\n---(?:\r?\n|$)/

function cleanScalar(value) {
  const trimmed = value.trim()
  const quote = trimmed[0]

  if ((quote === '"' || quote === "'") && trimmed.at(-1) === quote) {
    return trimmed.slice(1, -1)
  }

  return trimmed
}

function frontmatterScalar(frontmatter, key) {
  const escapedKey = key.replace(/[.*+?^${}()|[\]\\]/g, '\\$&')
  const match = frontmatter.match(
    new RegExp(`^\\s*(?:-\\s*)?${escapedKey}:\\s*(.+)$`, 'm')
  )
  return match ? cleanScalar(match[1]) : undefined
}

function splitFrontmatter(source) {
  const match = source.match(frontmatterPattern)

  if (!match) return { body: source, frontmatter: '' }

  return {
    body: source.slice(match[0].length),
    frontmatter: match[1]
  }
}

function headingDepthBefore(source, offset) {
  const headings = [...source.slice(0, offset).matchAll(/^(#{1,6})\s+\S.*$/gm)]
  return headings.at(-1)?.[1].length ?? 1
}

function flattenContentTabs(markdown) {
  return markdown.replace(
    /<ContentTabs\b([\s\S]*?)>\s*([\s\S]*?)<\/ContentTabs>/g,
    (component, attributes, content, offset, source) => {
      const labels = new Map()

      for (const match of attributes.matchAll(
        /\{\s*id:\s*['"]([^'"]+)['"]\s*,\s*label:\s*['"]([^'"]+)['"]/g
      )) {
        labels.set(match[1], match[2])
      }

      const headingDepth = Math.min(headingDepthBefore(source, offset) + 1, 6)
      const headingPrefix = '#'.repeat(headingDepth)

      return content
        .replace(
          /<template\s+#([\w-]+)>\s*([\s\S]*?)<\/template>/g,
          (template, id, body) =>
            `${headingPrefix} ${labels.get(id) ?? id}\n\n${body.trim()}\n`
        )
        .trim()
    }
  )
}

function flattenBadges(markdown) {
  return markdown.replace(
    /<Badge\b[^>]*\btext=(['"])(.*?)\1[^>]*\/>/g,
    (_badge, _quote, label) => `**${label}**`
  )
}

function markdownTarget(target) {
  if (
    target.startsWith('#') ||
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
  let path = queryIndex >= 0 ? pathAndQuery.slice(0, queryIndex) : pathAndQuery

  if (!path || path.endsWith('.md')) return target

  if (path.endsWith('/')) {
    path += 'index.md'
  } else {
    const finalSegment = path.slice(path.lastIndexOf('/') + 1)
    if (!finalSegment.includes('.')) path += '.md'
  }

  return path + query + fragment
}

function rewriteInternalLinks(markdown) {
  return markdown.replace(/\]\(([^)]+)\)/g, (link, destination) => {
    const match = destination.match(/^(\S+)([\s\S]*)$/)
    if (!match) return link

    return `](${markdownTarget(match[1])}${match[2]})`
  })
}

function fallbackHeading(relativePath, frontmatter) {
  const title = frontmatterScalar(frontmatter, 'title')
  if (!title) return undefined

  return relativePath === 'index.md' ? `fbrcm - ${title}` : title
}

export function renderMarkdownPage(source, relativePath) {
  const { body: sourceBody, frontmatter } = splitFrontmatter(source)
  let body = rewriteInternalLinks(
    flattenBadges(flattenContentTabs(sourceBody))
  ).trim()

  if (!/^#\s+\S/m.test(body)) {
    const heading = fallbackHeading(relativePath, frontmatter)
    const description =
      frontmatterScalar(frontmatter, 'tagline') ??
      frontmatterScalar(frontmatter, 'description')
    const actionText = frontmatterScalar(frontmatter, 'text')
    const actionLink = frontmatterScalar(frontmatter, 'link')
    const sections = []

    if (heading) sections.push(`# ${heading}`)
    if (description) sections.push(description)
    if (actionText && actionLink) {
      sections.push(`[${actionText}](${markdownTarget(actionLink)})`)
    }
    if (body) sections.push(body)

    body = sections.join('\n\n')
  }

  return `${body}\n`
}

export async function listMarkdownPages(sourceDirectory) {
  const pages = []

  async function visit(directory) {
    const entries = await readdir(directory, { withFileTypes: true })
    entries.sort((left, right) => left.name.localeCompare(right.name))

    for (const entry of entries) {
      if (directory === sourceDirectory && entry.name === 'public') continue

      const path = resolve(directory, entry.name)
      if (entry.isDirectory()) {
        await visit(path)
      } else if (entry.isFile() && entry.name.endsWith('.md')) {
        pages.push(relative(sourceDirectory, path).split(sep).join('/'))
      }
    }
  }

  await visit(sourceDirectory)
  return pages
}

function sourcePathForRoute(sourceDirectory, routePath) {
  if (!routePath.endsWith('.md')) return undefined

  const segments = routePath.split('/')
  if (segments.some((segment) => !segment || segment === '.' || segment === '..')) {
    return undefined
  }

  const sourceRoot = resolve(sourceDirectory)
  const sourcePath = resolve(sourceRoot, ...segments)
  if (!sourcePath.startsWith(sourceRoot + sep)) return undefined

  return sourcePath
}

export async function readMarkdownPage(sourceDirectory, routePath) {
  const sourcePath = sourcePathForRoute(sourceDirectory, routePath)
  if (!sourcePath) return undefined

  try {
    const source = await readFile(sourcePath, 'utf8')
    return renderMarkdownPage(source, routePath)
  } catch (error) {
    if (error?.code === 'ENOENT') return undefined
    throw error
  }
}

export async function writeMarkdownPages(sourceDirectory, outputDirectory) {
  const pages = await listMarkdownPages(sourceDirectory)

  await Promise.all(
    pages.map(async (page) => {
      const markdown = await readMarkdownPage(sourceDirectory, page)
      const outputPath = resolve(outputDirectory, ...page.split('/'))

      await mkdir(dirname(outputPath), { recursive: true })
      await writeFile(outputPath, markdown)
    })
  )

  return pages.length
}
