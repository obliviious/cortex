export default function Stats() {
  const stats = [
    { value: '5+', label: 'Agent Tools Supported' },
    { value: '100%', label: 'Parallel Execution' },
    { value: 'YAML', label: 'Simple Configuration' },
  ]

  return (
    <section className="py-8">
      <div className="container-custom">
        <div className="bg-coral-500 rounded-3xl py-12 px-8">
          <div className="grid grid-cols-1 md:grid-cols-3 gap-8 text-center">
            {stats.map((stat, index) => (
              <div key={index}>
                <div className="font-display text-5xl md:text-6xl text-dark-400 mb-2">
                  {stat.value}
                </div>
                <div className="text-dark-400/80 font-medium">
                  {stat.label}
                </div>
              </div>
            ))}
          </div>
        </div>
      </div>
    </section>
  )
}
