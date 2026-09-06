import assert from 'node:assert/strict'
import test from 'node:test'
import { join } from 'node:path'

import { packageRootFromModuleId, renderThirdPartyNotices } from './third-party-notices.mjs'

test('packageRootFromModuleId resolves scoped and unscoped packages', () => {
  assert.equal(
    packageRootFromModuleId('/workspace/node_modules/vue/dist/vue.runtime.js'),
    join('/workspace', 'node_modules', 'vue')
  )
  assert.equal(
    packageRootFromModuleId('/workspace/node_modules/@vue/shared/dist/shared.js?commonjs-proxy'),
    join('/workspace', 'node_modules', '@vue', 'shared')
  )
  assert.equal(packageRootFromModuleId('/workspace/docs/local.ts'), undefined)
})

test('renderThirdPartyNotices emits component metadata and exact license text', () => {
  const rendered = renderThirdPartyNotices([
    {
      name: 'example-package',
      version: '1.2.3',
      license: 'MIT',
      source: 'https://example.com/package',
      files: [{ name: 'LICENSE', content: 'Example license\n' }]
    }
  ])

  assert.match(rendered, /Component: example-package/)
  assert.match(rendered, /Version: 1\.2\.3/)
  assert.match(rendered, /License: MIT/)
  assert.match(rendered, /--- LICENSE ---\n\nExample license/)
})
