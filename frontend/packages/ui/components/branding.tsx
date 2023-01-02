import Image from 'next/image'
import TadokuLogo from './logo.svg'
import TadokuLogoLight from './logo-light.svg'

interface LogoProps {
  scale?: number
  priority?: boolean
  className?: string
}

export const Logo = ({ scale = 1, priority, className }: LogoProps) => (
  <Image
    src={TadokuLogo}
    alt="Tadoku"
    height={29 * scale}
    width={158 * scale}
    priority={priority}
    className={className}
  />
)

export const LogoInverted = ({ scale = 1, priority, className }: LogoProps) => (
  <Image
    src={TadokuLogoLight}
    alt="Tadoku"
    height={29 * scale}
    width={158 * scale}
    priority={priority}
    className={className}
  />
)
