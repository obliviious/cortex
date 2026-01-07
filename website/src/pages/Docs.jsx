import { Link } from 'react-router-dom'
import { FiArrowLeft } from 'react-icons/fi'
import DocsSidebar from '../components/docs/DocsSidebar'
import DocsContent from '../components/docs/DocsContent'

export default function Docs() {
  return (
    <div className="min-h-screen bg-cream-200">
      {/* Header */}
      <header className="border-b border-cream-400 bg-cream-200 sticky top-0 z-50">
        <div className="container-custom py-4 flex items-center justify-between">
          <div className="flex items-center gap-6">
            <Link to="/" className="flex items-center gap-2">
              <img src="/cortex-logo.svg" alt="Cortex" className="w-8 h-8" />
              <span className="font-display text-2xl text-dark-400">Cortex</span>
            </Link>
            <span className="text-cream-400">|</span>
            <span className="text-dark-100 font-medium">Documentation</span>
          </div>
          <Link
            to="/"
            className="flex items-center gap-2 text-dark-100 hover:text-coral-500 transition-colors"
          >
            <FiArrowLeft className="w-4 h-4" />
            Back to Home
          </Link>
        </div>
      </header>

      {/* Main Content */}
      <div className="container-custom py-12">
        <div className="flex gap-12">
          {/* Sidebar */}
          <DocsSidebar />

          {/* Content */}
          <DocsContent />
        </div>
      </div>

      {/* Footer */}
      <footer className="border-t border-cream-400 py-8">
        <div className="container-custom text-center text-dark-100 text-sm">
          <p>MIT License. Open source and free to use.</p>
        </div>
      </footer>
    </div>
  )
}
