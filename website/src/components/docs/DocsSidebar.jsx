import { Link, useParams } from 'react-router-dom'
import { FiBook, FiTerminal, FiSettings, FiCode, FiLayers, FiBell, FiFolder } from 'react-icons/fi'

export default function DocsSidebar() {
  const { section } = useParams()
  const currentSection = section || 'getting-started'

  const sections = [
    {
      title: 'Getting Started',
      items: [
        { id: 'getting-started', label: 'Quick Start', icon: FiBook },
        { id: 'installation', label: 'Installation', icon: FiTerminal },
      ],
    },
    {
      title: 'Configuration',
      items: [
        { id: 'cortexfile', label: 'Cortexfile.yml', icon: FiCode },
        { id: 'master-cortex', label: 'MasterCortex.yml', icon: FiLayers },
        { id: 'global-config', label: 'Global Config', icon: FiSettings },
      ],
    },
    {
      title: 'Features',
      items: [
        { id: 'templates', label: 'Template Variables', icon: FiCode },
        { id: 'webhooks', label: 'Webhooks', icon: FiBell },
        { id: 'sessions', label: 'Session Storage', icon: FiFolder },
      ],
    },
    {
      title: 'CLI Reference',
      items: [
        { id: 'commands', label: 'Commands', icon: FiTerminal },
      ],
    },
  ]

  return (
    <aside className="w-64 flex-shrink-0">
      <nav className="sticky top-8 space-y-8">
        {sections.map((group, groupIndex) => (
          <div key={groupIndex}>
            <h4 className="text-xs font-semibold text-dark-100 uppercase tracking-wider mb-3">
              {group.title}
            </h4>
            <ul className="space-y-1">
              {group.items.map((item) => (
                <li key={item.id}>
                  <Link
                    to={`/docs/${item.id}`}
                    className={`flex items-center gap-3 px-3 py-2 rounded-lg text-sm transition-colors ${
                      currentSection === item.id
                        ? 'bg-coral-500 text-white'
                        : 'text-dark-100 hover:bg-cream-300 hover:text-dark-400'
                    }`}
                  >
                    <item.icon className="w-4 h-4" />
                    {item.label}
                  </Link>
                </li>
              ))}
            </ul>
          </div>
        ))}
      </nav>
    </aside>
  )
}
