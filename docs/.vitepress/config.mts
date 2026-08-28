import { defineConfig } from 'vitepress'
import { writeFile } from 'node:fs/promises'
import { resolve } from 'node:path'
import { fileURLToPath } from 'node:url'

import { writeMarkdownPages } from '../scripts/markdown-pages.mjs'
import { pageHeadDevPlugin, pageRoute } from '../scripts/page-head.mjs'
import { renderRobotsTxt, robotsDevPlugin } from '../scripts/robots.mjs'

const repository = 'https://github.com/yumauri/fbrcm'
const siteUrl = process.env.DOCS_SITE_URL?.replace(/\/+$/, '')
const socialImage = siteUrl ? `${siteUrl}/og.png` : undefined
const siteDirectory = fileURLToPath(new URL('../site', import.meta.url))

export default defineConfig({
  title: 'fbrcm',
  titleTemplate: 'fbrcm - :title',
  description:
    'A TUI and CLI for managing Firebase Remote Config across Google Cloud projects.',
  lang: 'en-US',
  srcDir: 'site',
  base: '/',
  vite: {
    plugins: [robotsDevPlugin(siteUrl), pageHeadDevPlugin(siteUrl)]
  },
  cleanUrls: true,
  lastUpdated: true,
  sitemap: siteUrl ? { hostname: `${siteUrl}/` } : undefined,
  buildEnd: async ({ outDir }) => {
    await Promise.all([
      writeMarkdownPages(siteDirectory, outDir),
      writeFile(resolve(outDir, 'robots.txt'), renderRobotsTxt(siteUrl))
    ])
  },
  transformHead: ({ pageData }) => {
    if (pageData.isNotFound) return

    const markdownPath = `/${pageData.relativePath.replaceAll('\\', '/')}`
    const pageHead = [
      ['link', { rel: 'alternate', type: 'text/markdown', href: markdownPath }],
      ['link', { rel: 'describedby', type: 'text/plain', href: '/llms.txt' }]
    ]

    if (siteUrl) {
      pageHead.unshift([
        'link',
        { rel: 'canonical', href: `${siteUrl}${pageRoute(pageData.relativePath)}` }
      ])
    }

    return pageHead
  },
  head: [
    ['link', { rel: 'icon', href: '/favicon.ico', sizes: '32x32' }],
    ['link', { rel: 'icon', href: '/icon.svg', type: 'image/svg+xml' }],
    ['link', { rel: 'apple-touch-icon', href: '/apple-touch-icon.png' }],
    ['link', { rel: 'manifest', href: '/manifest.webmanifest' }],
    ['meta', { name: 'theme-color', content: '#f57c32' }],
    ['meta', { property: 'og:type', content: 'website' }],
    ['meta', { property: 'og:title', content: 'fbrcm · Firebase Remote Config manager' }],
    [
      'meta',
      {
        property: 'og:description',
        content:
          'Manage Firebase Remote Config across Google Cloud projects with a TUI or CLI.'
      }
    ],
    ['meta', { name: 'twitter:card', content: 'summary_large_image' }],
    ['meta', { name: 'twitter:title', content: 'fbrcm · Firebase Remote Config manager' }],
    [
      'meta',
      {
        name: 'twitter:description',
        content:
          'Manage Firebase Remote Config across Google Cloud projects with a TUI or CLI.'
      }
    ],
    ...(socialImage
      ? [
          ['meta', { property: 'og:image', content: socialImage }],
          ['meta', { name: 'twitter:image', content: socialImage }]
        ]
      : [])
  ],
  markdown: {
    languageAlias: {
      expr: 'javascript'
    },
    lineNumbers: true
  },
  themeConfig: {
    logo: { src: '/fbrcm-logo.svg', alt: 'fbrcm' },
    siteTitle: false,
    nav: [
      { text: 'Guide', link: '/guide/' },
      { text: 'TUI', link: '/tui/' },
      { text: 'CLI', link: '/cli/' },
      {
        text: 'Automation',
        items: [
          { text: 'Agent workflows', link: '/automation/' },
          { text: 'JSON contract', link: '/automation/json-contract' }
        ]
      }
    ],
    sidebar: [
      {
        text: 'Start here',
        items: [
          { text: 'Getting started', link: '/guide/' },
          { text: 'Authentication and discovery', link: '/guide/authentication' },
          { text: 'How fbrcm works', link: '/guide/mental-model' },
          { text: 'Safe change workflow', link: '/guide/safe-changes' }
        ]
      },
      {
        text: 'Terminal UI',
        items: [
          { text: 'TUI overview', link: '/tui/' },
          { text: 'Editing and drafts', link: '/tui/editing' }
        ]
      },
      {
        text: 'CLI workflows',
        items: [
          { text: 'CLI overview', link: '/cli/' },
          { text: 'Parameters and conditions', link: '/cli/parameters' },
          { text: 'Projects and templates', link: '/cli/projects' },
          { text: 'Drafts', link: '/cli/drafts' },
          { text: 'History and managed features', link: '/cli/history' }
        ]
      },
      {
        text: 'Automation',
        items: [
          { text: 'Agent workflows', link: '/automation/' },
          { text: 'JSON contract', link: '/automation/json-contract' }
        ]
      },
      {
        text: 'Reference',
        items: [
          { text: 'Filtering', link: '/reference/filtering' },
          { text: 'Configuration', link: '/reference/configuration' },
          { text: 'Themes', link: '/reference/themes' },
          { text: 'Hooks', link: '/reference/hooks' },
          { text: 'Command index', link: '/reference/commands' },
          { text: 'Troubleshooting', link: '/reference/troubleshooting' }
        ]
      }
    ],
    socialLinks: [{ icon: 'github', link: repository }],
    editLink: {
      pattern: `${repository}/edit/main/docs/site/:path`,
      text: 'Edit this page on GitHub'
    },
    search: {
      provider: 'local',
      options: {
        detailedView: true
      }
    },
    outline: {
      level: [2, 3],
      label: 'On this page'
    },
    docFooter: {
      prev: 'Previous page',
      next: 'Next page'
    },
    footer: {
      message:
        '<a href="/privacy-policy">Privacy policy</a> · <a href="/llms.txt">llms.txt</a> · <a href="/llms-full.txt">llms-full.txt</a> · <a href="/LICENSE.txt">MIT License</a> · 2026'
    }
  }
})
