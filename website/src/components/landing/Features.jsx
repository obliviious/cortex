import { FiZap, FiGitBranch, FiUsers, FiFolder, FiSend, FiClock } from 'react-icons/fi'

export default function Features() {
  const features = [
    {
      icon: FiZap,
      title: 'Parallel Execution',
      description: 'Run independent tasks concurrently for maximum efficiency.',
    },
    {
      icon: FiGitBranch,
      title: 'Task Dependencies',
      description: 'Chain tasks with `needs` and pass outputs via templates.',
    },
    {
      icon: FiUsers,
      title: 'Multi-Agent Support',
      description: 'Use Claude Code, OpenCode, or other AI CLIs together.',
    },
    {
      icon: FiFolder,
      title: 'Multi-Project',
      description: 'Orchestrate multiple Cortexfiles with MasterCortex.yml.',
    },
    {
      icon: FiSend,
      title: 'Webhooks',
      description: 'Get notified on task completion or failure.',
    },
    {
      icon: FiClock,
      title: 'Session Tracking',
      description: 'View and manage past run sessions easily.',
    },
  ]

  return (
    <section id="features" className="py-20">
      <div className="container-custom">
        {/* Section Header */}
        <div className="text-center mb-16">
          <h2 className="font-display text-4xl md:text-5xl text-dark-400 mb-4">
            Everything you need to orchestrate AI
          </h2>
          <p className="text-dark-100 text-lg max-w-2xl mx-auto">
            Cortex provides all the tools you need to run complex AI workflows without the complexity.
          </p>
        </div>

        {/* Features Grid */}
        <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-8">
          {features.map((feature, index) => (
            <div
              key={index}
              className="bg-white rounded-2xl p-8 hover:shadow-lg transition-shadow duration-300"
            >
              <div className="w-12 h-12 bg-coral-500/10 rounded-xl flex items-center justify-center mb-6">
                <feature.icon className="w-6 h-6 text-coral-500" />
              </div>
              <h3 className="font-semibold text-xl text-dark-400 mb-3">
                {feature.title}
              </h3>
              <p className="text-dark-100">
                {feature.description}
              </p>
            </div>
          ))}
        </div>
      </div>
    </section>
  )
}
