export const commandIndexLink = '/reference/commands/'
export const commandBehaviorLink = '/reference/command-behavior'

const commandGroups = [
  ['Start and get help', ['root', 'help']],
  ['Parameters and structure', ['add', 'duplicate', 'get', 'explain', 'update', 'delete', 'groups']],
  ['Conditions', ['conditions']],
  ['Drafts', ['draft']],
  ['Plans', ['plan', 'apply']],
  ['Projects and templates', ['project', 'projects']],
  ['Firebase applications', ['apps']],
  ['History and managed features', ['versions', 'experiments', 'rollouts', 'personalizations']],
  ['Local environment', ['setup', 'doctor', 'cache', 'config', 'hooks', 'auth', 'profile', 'theme', 'completion']],
  ['Machine discovery', ['mcp', 'capabilities', 'schema']]
]

export function groupCommands(capabilities) {
  const groups = new Map(commandGroups.map(([title]) => [title, []]))
  for (const { id, path, summary } of capabilities) {
    const name = ['fbrcm', ...path].join(' ')
    const key = path[0] ?? 'root'
    const title = commandGroups.find(([, keys]) => keys.includes(key))?.[0]
    if (!title) throw new Error(`Missing command index group for ${name}`)

    groups.get(title).push({
      id,
      name,
      text: path.join(' ') || name,
      link: commandIndexLink + (path.join('/') || 'root'),
      summary
    })
  }
  return [...groups].filter(([, commands]) => commands.length)
}

export function commandSidebar(capabilities) {
  return {
    text: 'Command index',
    link: commandIndexLink,
    collapsed: true,
    items: [
      ...groupCommands(capabilities).map(([text, commands]) => ({
        text,
        collapsed: true,
        items: commands.map(({ text, link }) => ({ text, link }))
      })),
      { text: 'Shared command behavior', link: commandBehaviorLink }
    ]
  }
}
