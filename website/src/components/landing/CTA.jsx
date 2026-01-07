import { FiTerminal, FiCopy, FiCheck } from 'react-icons/fi'
import { useState } from 'react'

export default function CTA() {
  const [copied, setCopied] = useState(false)
  const installCommand = 'npm install -g @insien/cortex-cli'

  const copyToClipboard = () => {
    navigator.clipboard.writeText(installCommand)
    setCopied(true)
    setTimeout(() => setCopied(false), 2000)
  }

  return (
    <section id="get-started" className="py-20">
      <div className="container-custom">
        <div className="bg-dark-400 rounded-3xl p-12 md:p-16 text-center">
          {/* Icon */}
          <div className="w-16 h-16 bg-coral-500 rounded-2xl flex items-center justify-center mx-auto mb-8">
            <FiTerminal className="w-8 h-8 text-white" />
          </div>

          {/* Headline */}
          <h2 className="font-display text-4xl md:text-5xl text-white mb-4">
            Ready to orchestrate?
          </h2>
          <p className="text-cream-300 text-lg max-w-xl mx-auto mb-8">
            Install Cortex in seconds and start running AI workflows.
          </p>

          {/* Install Command */}
          <div className="bg-dark-300 rounded-xl p-4 max-w-lg mx-auto mb-8 flex items-center justify-between gap-4">
            <code className="text-coral-400 font-mono text-sm md:text-base truncate">
              {installCommand}
            </code>
            <button
              onClick={copyToClipboard}
              className="flex-shrink-0 p-2 hover:bg-dark-200 rounded-lg transition-colors"
              title="Copy to clipboard"
            >
              {copied ? (
                <FiCheck className="w-5 h-5 text-green-400" />
              ) : (
                <FiCopy className="w-5 h-5 text-cream-300" />
              )}
            </button>
          </div>

          {/* Alternative Installs */}
          <p className="text-cream-400 text-sm mb-6">Or install with:</p>
          <div className="flex flex-wrap justify-center gap-4 text-sm">
            <code className="bg-dark-300 px-4 py-2 rounded-lg text-cream-300">
              brew install cortex
            </code>
            <code className="bg-dark-300 px-4 py-2 rounded-lg text-cream-300">
              go install github.com/obliviious/cortex@latest
            </code>
          </div>
        </div>
      </div>
    </section>
  )
}
