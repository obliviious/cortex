import { useState } from 'react'
import { Link } from 'react-router-dom'
import { HiMenuAlt3, HiX } from 'react-icons/hi'
import { FiGithub } from 'react-icons/fi'

export default function Navbar() {
  const [isOpen, setIsOpen] = useState(false)

  return (
    <nav className="py-4 relative z-50">
      <div className="container-custom">
        <div className="flex items-center justify-between">
          {/* Logo */}
          <Link to="/" className="flex items-center gap-2">
            <img src="/cortex-logo.svg" alt="Cortex" className="w-8 h-8" />
            <span className="font-display text-2xl text-dark-400">Cortex</span>
          </Link>

          {/* Desktop Navigation */}
          <div className="hidden md:flex items-center gap-8">
            <a href="#features" className="text-dark-100 hover:text-dark-400 transition-colors">
              Features
            </a>
            <a href="#how-it-works" className="text-dark-100 hover:text-dark-400 transition-colors">
              How it works
            </a>
            <Link to="/docs" className="text-dark-100 hover:text-dark-400 transition-colors">
              Docs
            </Link>
            <a
              href="https://github.com/obliviious/cortex"
              target="_blank"
              rel="noopener noreferrer"
              className="text-dark-100 hover:text-dark-400 transition-colors"
            >
              <FiGithub className="w-5 h-5" />
            </a>
          </div>

          {/* CTA Button */}
          <div className="hidden md:block">
            <a
              href="#get-started"
              className="btn-primary"
            >
              Get Started
            </a>
          </div>

          {/* Mobile Menu Button */}
          <button
            className="md:hidden p-2"
            onClick={() => setIsOpen(!isOpen)}
          >
            {isOpen ? <HiX className="w-6 h-6" /> : <HiMenuAlt3 className="w-6 h-6" />}
          </button>
        </div>

        {/* Mobile Navigation */}
        {isOpen && (
          <div className="md:hidden absolute top-full left-0 right-0 bg-cream-200 border-t border-cream-400 py-4">
            <div className="container-custom flex flex-col gap-4">
              <a
                href="#features"
                className="text-dark-100 hover:text-dark-400 transition-colors py-2"
                onClick={() => setIsOpen(false)}
              >
                Features
              </a>
              <a
                href="#how-it-works"
                className="text-dark-100 hover:text-dark-400 transition-colors py-2"
                onClick={() => setIsOpen(false)}
              >
                How it works
              </a>
              <Link
                to="/docs"
                className="text-dark-100 hover:text-dark-400 transition-colors py-2"
                onClick={() => setIsOpen(false)}
              >
                Docs
              </Link>
              <a
                href="https://github.com/obliviious/cortex"
                target="_blank"
                rel="noopener noreferrer"
                className="text-dark-100 hover:text-dark-400 transition-colors py-2 flex items-center gap-2"
              >
                <FiGithub className="w-5 h-5" /> GitHub
              </a>
              <a href="#get-started" className="btn-primary w-fit" onClick={() => setIsOpen(false)}>
                Get Started
              </a>
            </div>
          </div>
        )}
      </div>
    </nav>
  )
}
