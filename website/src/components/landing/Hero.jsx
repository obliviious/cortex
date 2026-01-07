import { FiArrowRight } from 'react-icons/fi'
import { Link } from 'react-router-dom'

export default function Hero() {
  return (
    <section className="py-16 md:py-24 bg-pattern">
      <div className="container-custom">
        <div className="grid grid-cols-1 lg:grid-cols-2 gap-12 items-center">
          {/* Left Content */}
          <div>
            {/* Decorative Quote */}
            <div className="text-coral-500 text-6xl font-display leading-none mb-2">&ldquo;</div>

            {/* Main Headline */}
            <h1 className="font-display text-5xl md:text-6xl lg:text-7xl text-dark-400 leading-tight mb-6">
              Orchestrate AI Agents, without the headaches.
            </h1>

            {/* Subtext */}
            <p className="text-dark-100 text-lg md:text-xl mb-8 max-w-lg">
              A powerful CLI tool to run AI workflows defined in YAML.
              Simple setup, parallel execution, no surprises.
            </p>

            {/* CTAs */}
            <div className="flex flex-wrap items-center gap-4">
              <a href="#get-started" className="btn-primary">
                Try for free
              </a>
              <Link to="/docs" className="btn-secondary flex items-center gap-2">
                How it works <FiArrowRight className="w-4 h-4" />
              </Link>
            </div>
          </div>

          {/* Right Content - Terminal Mockups */}
          <div className="relative">
            {/* Background Terminal (White) */}
            <div className="absolute top-0 right-0 w-[90%] bg-white rounded-3xl shadow-lg p-6 transform translate-x-4 -translate-y-4">
              {/* Terminal Header */}
              <div className="flex items-center gap-2 mb-4">
                <div className="w-3 h-3 rounded-full bg-red-400"></div>
                <div className="w-3 h-3 rounded-full bg-yellow-400"></div>
                <div className="w-3 h-3 rounded-full bg-green-400"></div>
                <span className="ml-4 text-dark-100 text-sm">Terminal</span>
              </div>

              {/* Terminal Content */}
              <div className="font-mono text-sm text-dark-300 space-y-2">
                <p className="text-dark-100">$ cortex init</p>
                <p className="text-green-600">Created Cortexfile.yml</p>
                <p className="text-dark-100 mt-4">$ cortex run</p>
                <p className="text-coral-500">Running 3 tasks...</p>
              </div>
            </div>

            {/* Foreground Terminal (Coral) */}
            <div className="relative z-10 w-[85%] bg-coral-500 rounded-3xl shadow-xl p-6 mt-20 ml-auto">
              {/* Terminal Header */}
              <div className="flex items-center gap-2 mb-4">
                <div className="w-3 h-3 rounded-full bg-coral-700"></div>
                <div className="w-3 h-3 rounded-full bg-coral-600"></div>
                <div className="w-3 h-3 rounded-full bg-coral-400"></div>
                <span className="ml-4 text-white/70 text-sm">Output</span>
              </div>

              {/* Terminal Content */}
              <div className="font-mono text-sm text-white space-y-3">
                <div className="flex items-center gap-2">
                  <span className="text-white/60">[1/3]</span>
                  <span>analyze</span>
                  <span className="ml-auto bg-white/20 px-2 py-0.5 rounded text-xs">done</span>
                </div>
                <div className="flex items-center gap-2">
                  <span className="text-white/60">[2/3]</span>
                  <span>review</span>
                  <span className="ml-auto bg-white/20 px-2 py-0.5 rounded text-xs">done</span>
                </div>
                <div className="flex items-center gap-2">
                  <span className="text-white/60">[3/3]</span>
                  <span>implement</span>
                  <span className="ml-auto bg-white/20 px-2 py-0.5 rounded text-xs">running</span>
                </div>
                <div className="pt-4 border-t border-white/20 mt-4">
                  <p className="text-white/80 text-xs uppercase tracking-wide">Claude Code</p>
                  <p className="text-white font-medium">architect</p>
                </div>
              </div>
            </div>
          </div>
        </div>
      </div>
    </section>
  )
}
