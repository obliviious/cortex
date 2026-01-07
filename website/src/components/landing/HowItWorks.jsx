import { FiCheck } from 'react-icons/fi'

export default function HowItWorks() {
  const steps = [
    {
      number: '01',
      title: 'Create a Cortexfile',
      description: 'Define your agents and tasks in a simple YAML file.',
      code: `agents:
  architect:
    tool: claude-code
    model: sonnet

  reviewer:
    tool: claude-code
    model: sonnet

tasks:
  analyze:
    agent: architect
    prompt: |
      Analyze the codebase structure.`,
    },
    {
      number: '02',
      title: 'Chain Tasks with Dependencies',
      description: 'Use template variables to pass outputs between tasks.',
      code: `tasks:
  implement:
    agent: architect
    needs: [analyze, review]
    write: true
    prompt: |
      Based on the analysis:
      {{outputs.analyze}}

      And the review:
      {{outputs.review}}

      Implement the changes.`,
    },
    {
      number: '03',
      title: 'Run the Workflow',
      description: 'Execute with a single command. Watch tasks run in parallel.',
      code: `$ cortex run

[1/3] analyze    running...
[2/3] review     running...
[3/3] implement  waiting...

All tasks completed successfully!`,
    },
  ]

  return (
    <section id="how-it-works" className="py-20 bg-cream-100">
      <div className="container-custom">
        {/* Section Header */}
        <div className="text-center mb-16">
          <h2 className="font-display text-4xl md:text-5xl text-dark-400 mb-4">
            How it works
          </h2>
          <p className="text-dark-100 text-lg max-w-2xl mx-auto">
            Get started in three simple steps. No complex setup required.
          </p>
        </div>

        {/* Steps */}
        <div className="space-y-16">
          {steps.map((step, index) => (
            <div
              key={index}
              className={`grid grid-cols-1 lg:grid-cols-2 gap-12 items-center ${
                index % 2 === 1 ? 'lg:flex-row-reverse' : ''
              }`}
            >
              {/* Content */}
              <div className={index % 2 === 1 ? 'lg:order-2' : ''}>
                <div className="inline-flex items-center gap-2 bg-coral-500/10 text-coral-500 px-3 py-1 rounded-full text-sm font-medium mb-4">
                  Step {step.number}
                </div>
                <h3 className="font-display text-3xl text-dark-400 mb-4">
                  {step.title}
                </h3>
                <p className="text-dark-100 text-lg mb-6">
                  {step.description}
                </p>
                <ul className="space-y-3">
                  <li className="flex items-center gap-3 text-dark-100">
                    <FiCheck className="w-5 h-5 text-coral-500" />
                    Easy to read and write
                  </li>
                  <li className="flex items-center gap-3 text-dark-100">
                    <FiCheck className="w-5 h-5 text-coral-500" />
                    No code required
                  </li>
                </ul>
              </div>

              {/* Code Block */}
              <div className={index % 2 === 1 ? 'lg:order-1' : ''}>
                <div className="bg-dark-400 rounded-2xl overflow-hidden shadow-xl">
                  {/* Code Header */}
                  <div className="flex items-center gap-2 px-4 py-3 bg-dark-300">
                    <div className="w-3 h-3 rounded-full bg-red-400"></div>
                    <div className="w-3 h-3 rounded-full bg-yellow-400"></div>
                    <div className="w-3 h-3 rounded-full bg-green-400"></div>
                    <span className="ml-4 text-cream-300 text-sm">
                      {step.number === '03' ? 'terminal' : 'Cortexfile.yml'}
                    </span>
                  </div>
                  {/* Code Content */}
                  <pre className="p-6 text-cream-100 text-sm font-mono overflow-x-auto">
                    <code>{step.code}</code>
                  </pre>
                </div>
              </div>
            </div>
          ))}
        </div>
      </div>
    </section>
  )
}
