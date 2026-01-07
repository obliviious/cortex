import Navbar from '../components/layout/Navbar'
import Footer from '../components/layout/Footer'
import Hero from '../components/landing/Hero'
import Stats from '../components/landing/Stats'
import Features from '../components/landing/Features'
import HowItWorks from '../components/landing/HowItWorks'
import CTA from '../components/landing/CTA'

export default function Landing() {
  return (
    <div className="min-h-screen">
      <Navbar />
      <main>
        <Hero />
        <Stats />
        <Features />
        <HowItWorks />
        <CTA />
      </main>
      <Footer />
    </div>
  )
}
