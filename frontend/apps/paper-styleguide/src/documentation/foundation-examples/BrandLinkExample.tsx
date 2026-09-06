import lightWordmark from 'paper-ui/assets/brand/wordmark-accent.svg?no-inline'
import darkWordmark from 'paper-ui/assets/brand/wordmark-reversed.svg?no-inline'

export default function BrandLink({
  theme = 'light',
}: {
  theme?: 'light' | 'dark'
}) {
  return (
    <a href="/" aria-label="Tadoku home">
      <img
        src={theme === 'dark' ? darkWordmark : lightWordmark}
        width={158}
        height={29}
        alt=""
      />
    </a>
  )
}
