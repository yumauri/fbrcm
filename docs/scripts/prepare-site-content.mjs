import { mkdir, readFile, writeFile } from 'node:fs/promises'
import { dirname, resolve } from 'node:path'
import { fileURLToPath } from 'node:url'

const scriptDirectory = dirname(fileURLToPath(import.meta.url))
const docsDirectory = resolve(scriptDirectory, '..')
const repositoryRoot = resolve(docsDirectory, '..')
const siteDirectory = resolve(docsDirectory, 'site')

const privacySource = resolve(repositoryRoot, 'PRIVACY.md')
const llmsSource = resolve(repositoryRoot, 'llms.txt')
const licenseSource = resolve(repositoryRoot, 'LICENSE')
const privacyPage = resolve(siteDirectory, 'privacy-policy.md')
const llmsAsset = resolve(siteDirectory, 'public', 'llms.txt')
const licenseAsset = resolve(siteDirectory, 'public', 'LICENSE.txt')

const privacyFrontmatter = `---
title: Privacy Policy
description: How fbrcm accesses, uses, stores, and protects Google user data.
sidebar: false
editLink: false
---

`

const [privacyMarkdown, llmsText, licenseText] = await Promise.all([
  readFile(privacySource, 'utf8'),
  readFile(llmsSource),
  readFile(licenseSource)
])

await mkdir(dirname(llmsAsset), { recursive: true })
await Promise.all([
  writeFile(privacyPage, privacyFrontmatter + privacyMarkdown),
  writeFile(llmsAsset, llmsText),
  writeFile(licenseAsset, licenseText)
])
