import { readFile, readdir } from 'node:fs/promises'
import { dirname, join, resolve, sep } from 'node:path'
import { fileURLToPath } from 'node:url'

const scriptDirectory = dirname(fileURLToPath(import.meta.url))
const docsDirectory = resolve(scriptDirectory, '..')
const interLicensePath = resolve(docsDirectory, 'licenses', 'Inter-OFL-1.1.txt')
const noticeFileName = 'THIRD_PARTY_NOTICES.txt'

const sharedLicensePaths = [
  {
    matches: (name) => name.startsWith('@docsearch/'),
    name: 'algolia-docsearch-MIT.txt',
    path: resolve(docsDirectory, 'licenses', 'algolia-docsearch-MIT.txt')
  },
  {
    matches: (name) => name.startsWith('@algolia/autocomplete-'),
    name: 'algolia-autocomplete-MIT.txt',
    path: resolve(docsDirectory, 'licenses', 'algolia-autocomplete-MIT.txt')
  }
]

const licensePrefixes = ['LICENSE', 'COPYING', 'NOTICE', 'PATENTS']
const compatiblePackageLicenses = new Set(['MIT'])

export function thirdPartyNoticesPlugin() {
  return {
    name: 'fbrcm-third-party-notices',
    apply: 'build',
    async generateBundle(_options, bundle) {
      const packageRoots = new Set()
      for (const artifact of Object.values(bundle)) {
        if (artifact.type !== 'chunk') continue

        for (const moduleId of Object.keys(artifact.modules)) {
          const packageRoot = packageRootFromModuleId(moduleId)
          if (packageRoot) packageRoots.add(packageRoot)
        }
      }

      const components = await Promise.all(
        [...packageRoots].sort().map((packageRoot) => packageComponent(packageRoot))
      )
      const interLicense = await readFile(interLicensePath, 'utf8')
      components.push({
        name: 'Inter font',
        version: 'bundled by VitePress',
        source: 'https://github.com/rsms/inter',
        license: 'OFL-1.1',
        files: [{ name: 'Inter-OFL-1.1.txt', content: interLicense }]
      })
      components.sort((left, right) => left.name.localeCompare(right.name))

      this.emitFile({
        type: 'asset',
        fileName: noticeFileName,
        source: renderThirdPartyNotices(components)
      })
    }
  }
}

export function packageRootFromModuleId(moduleId) {
  const normalized = moduleId.split('?', 1)[0].split('\\').join('/')
  const marker = '/node_modules/'
  const markerIndex = normalized.lastIndexOf(marker)
  if (markerIndex < 0) return undefined

  const modulesRoot = normalized.slice(0, markerIndex + marker.length)
  const pathWithinModules = normalized.slice(markerIndex + marker.length)
  const parts = pathWithinModules.split('/')
  const packageParts = parts[0]?.startsWith('@') ? parts.slice(0, 2) : parts.slice(0, 1)
  if (packageParts.length === 0 || packageParts.some((part) => !part)) return undefined

  return (modulesRoot + packageParts.join('/')).split('/').join(sep)
}

async function packageComponent(packageRoot) {
  const packageJSON = JSON.parse(await readFile(join(packageRoot, 'package.json'), 'utf8'))
  if (!compatiblePackageLicenses.has(packageJSON.license)) {
    throw new Error(
      `${packageJSON.name ?? packageRoot} uses unreviewed license ${JSON.stringify(packageJSON.license)}`
    )
  }
  const entries = await readdir(packageRoot, { withFileTypes: true })
  const licenseNames = entries
    .filter(
      (entry) =>
        entry.isFile() &&
        licensePrefixes.some((prefix) => entry.name.toUpperCase().startsWith(prefix))
    )
    .map((entry) => entry.name)
    .sort((left, right) => left.localeCompare(right))

  let files
  if (licenseNames.length === 0) {
    const sharedLicense = sharedLicensePaths.find(({ matches }) => matches(packageJSON.name))
    if (sharedLicense) {
      files = [
        {
          name: sharedLicense.name,
          content: await readFile(sharedLicense.path, 'utf8')
        }
      ]
    }
  } else {
    files = await Promise.all(
      licenseNames.map(async (name) => ({
        name,
        content: await readFile(join(packageRoot, name), 'utf8')
      }))
    )
  }

  if (!files) {
    throw new Error(
      `${packageJSON.name ?? packageRoot} has no LICENSE, COPYING, NOTICE, or PATENTS file`
    )
  }
  return {
    name: packageJSON.name,
    version: packageJSON.version,
    source: packageSource(packageJSON),
    license: packageJSON.license ?? 'See included license text',
    files
  }
}

function packageSource(packageJSON) {
  const repository = packageJSON.repository
  if (typeof repository === 'string') return repository
  if (repository && typeof repository.url === 'string') return repository.url
  return packageJSON.homepage ?? `https://www.npmjs.com/package/${packageJSON.name}`
}

export function renderThirdPartyNotices(components) {
  let output = `fbrcm documentation third-party notices
===========================================

The generated fbrcm documentation site includes the software and font components
listed below. These components remain subject to their respective license terms.
`
  for (const component of components) {
    output += `

======================================================================
Component: ${component.name}
Version: ${component.version}
License: ${component.license}
Source: ${component.source}
`
    for (const file of component.files) {
      output += `
--- ${file.name} ---

${file.content.trim()}
`
    }
  }
  return output
}
