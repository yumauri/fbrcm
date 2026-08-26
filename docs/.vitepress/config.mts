import { defineConfig } from 'vitepress'

const repository = 'https://github.com/yumauri/fbrcm'
const siteUrl = process.env.DOCS_SITE_URL?.replace(/\/+$/, '')
const socialImage = siteUrl ? `${siteUrl}/og.png` : undefined

export default defineConfig({
  title: 'fbrcm',
  titleTemplate: ':title · fbrcm',
  description:
    'A terminal manager for inspecting, comparing, and safely changing Firebase Remote Config across projects.',
  lang: 'en-US',
  srcDir: 'site',
  base: '/',
  cleanUrls: true,
  lastUpdated: true,
  sitemap: siteUrl ? { hostname: `${siteUrl}/` } : undefined,
  head: [
    ['meta', { name: 'theme-color', content: '#f57c32' }],
    ['meta', { property: 'og:type', content: 'website' }],
    ['meta', { property: 'og:title', content: 'fbrcm · Remote Config, under control.' }],
    [
      'meta',
      {
        property: 'og:description',
        content:
          'Inspect, compare, stage, and publish Firebase Remote Config changes from one terminal workflow.'
      }
    ],
    ['meta', { name: 'twitter:card', content: 'summary_large_image' }],
    ['meta', { name: 'twitter:title', content: 'fbrcm · Remote Config, under control.' }],
    [
      'meta',
      {
        name: 'twitter:description',
        content:
          'Inspect, compare, stage, and publish Firebase Remote Config changes from one terminal workflow.'
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
          { text: 'Mental model', link: '/guide/mental-model' },
          { text: 'Safe changes', link: '/guide/safe-changes' }
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
        text: 'Command line',
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
        collapsed: true,
        items: [
          { text: 'Filtering', link: '/reference/filtering' },
          { text: 'Configuration', link: '/reference/configuration' },
          { text: 'Themes', link: '/reference/themes' },
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
        '<a href="/privacy-policy">Privacy policy</a> · <a href="/llms.txt">llms.txt</a> · <a href="/LICENSE.txt">MIT License</a> · 2026'
    }
  }
})
