import { Link } from 'react-router-dom'
import { FiGithub, FiTerminal } from 'react-icons/fi'

export default function Footer() {
  return (
    <footer className="py-12 border-t border-cream-400">
      <div className="container-custom">
        <div className="grid grid-cols-1 md:grid-cols-4 gap-8">
          {/* Brand */}
          <div className="md:col-span-2">
            <Link to="/" className="flex items-center gap-2 mb-4">
              <img src="/cortex-logo.svg" alt="Cortex" className="w-8 h-8" />
              <span className="font-display text-2xl text-dark-400">Cortex</span>
            </Link>
            <p className="text-dark-100 max-w-md">
              A powerful CLI tool that orchestrates AI agent workflows defined in YAML.
              Run multiple agents in parallel, chain outputs, and automate complex tasks.
            </p>
          </div>

          {/* Links */}
          <div>
            <h4 className="font-semibold text-dark-400 mb-4">Product</h4>
            <ul className="space-y-2">
              <li>
                <a href="#features" className="text-dark-100 hover:text-coral-500 transition-colors">
                  Features
                </a>
              </li>
              <li>
                <Link to="/docs" className="text-dark-100 hover:text-coral-500 transition-colors">
                  Documentation
                </Link>
              </li>
              <li>
                <a href="#get-started" className="text-dark-100 hover:text-coral-500 transition-colors">
                  Get Started
                </a>
              </li>
            </ul>
          </div>

          {/* Resources */}
          <div>
            <h4 className="font-semibold text-dark-400 mb-4">Resources</h4>
            <ul className="space-y-2">
              <li>
                <a
                  href="https://github.com/obliviious/cortex"
                  target="_blank"
                  rel="noopener noreferrer"
                  className="text-dark-100 hover:text-coral-500 transition-colors flex items-center gap-2"
                >
                  <FiGithub className="w-4 h-4" /> GitHub
                </a>
              </li>
              <li>
                <a
                  href="https://www.npmjs.com/package/@insien/cortex-cli"
                  target="_blank"
                  rel="noopener noreferrer"
                  className="text-dark-100 hover:text-coral-500 transition-colors flex items-center gap-2"
                >
                  <FiTerminal className="w-4 h-4" /> NPM Package
                </a>
              </li>
            </ul>
          </div>
        </div>

        {/* Copyright */}
        <div className="mt-12 pt-8 border-t border-cream-400 flex flex-col md:flex-row justify-between items-center gap-4">
          <p className="text-dark-100 text-sm">
            MIT License. Open source and free to use.
          </p>
          <p className="text-dark-100 text-sm">
            Built with React + Tailwind CSS
          </p>
        </div>
      </div>
    </footer>
  )
}
