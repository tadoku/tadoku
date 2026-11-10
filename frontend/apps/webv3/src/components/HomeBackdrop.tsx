/** A seeded generator, so the server and the client draw the same pattern. */
function random(seed: number) {
  return () => {
    seed = (seed + 0x6d2b79f5) | 0
    let t = Math.imul(seed ^ (seed >>> 15), 1 | seed)
    t = (t + Math.imul(t ^ (t >>> 7), 61 | t)) ^ t
    return ((t ^ (t >>> 14)) >>> 0) / 4294967296
  }
}

const heat = ['fill-heat-0', 'fill-heat-1', 'fill-heat-2', 'fill-heat-3', 'fill-heat-4']
const size = 44
const gap = 8
const weeks = 53

/** A decorative year of activity squares, busier toward the present. It does not show real activity. */
const cells = (() => {
  const next = random(6)
  const squares = []
  for (let week = 0; week < weeks; week++) {
    for (let day = 0; day < 7; day++) {
      const busy = 0.25 + (week / weeks) * 0.5
      const level = next() > busy ? 0 : Math.min(4, 1 + Math.floor(next() * next() * 4.6))
      squares.push({ key: `${week}-${day}`, x: week * (size + gap), y: day * (size + gap), className: heat[level] })
    }
  }
  return squares
})()

/**
 * The guest homepage's backdrop: an activity grid behind the top of the page that leaves the hero's text clear. It
 * sits at the top of the document under a negative z-index, so the application frame must isolate its stacking.
 */
export function HomeBackdrop() {
  return (
    <div
      aria-hidden="true"
      className={[
        'pointer-events-none absolute inset-x-0 top-0 -z-1 overflow-hidden select-none',
        '[mask-image:var(--effect-hero-backdrop-mask)] mask-intersect',
      ].join(' ')}
    >
      <svg
        className="mt-20 block h-home-backdrop-grid w-full opacity-35"
        viewBox={`0 0 ${weeks * (size + gap)} ${7 * (size + gap)}`}
        preserveAspectRatio="xMaxYMin slice"
      >
        {cells.map(({ key, ...cell }) => <rect key={key} {...cell} width={size} height={size} rx={4} />)}
      </svg>
    </div>
  )
}
