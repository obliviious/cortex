import { BrowserRouter as Router, Routes, Route } from 'react-router-dom'
import { Analytics } from '@vercel/analytics/react'
import Landing from './pages/Landing'
import Docs from './pages/Docs'

function App() {
  return (
    <>
      <Router>
        <Routes>
          <Route path="/" element={<Landing />} />
          <Route path="/docs" element={<Docs />} />
          <Route path="/docs/:section" element={<Docs />} />
        </Routes>
      </Router>
      <Analytics />
    </>
  )
}

export default App
