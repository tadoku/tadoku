import Image from 'next/image'
import TadokuLogo from './logo.svg'
import TadokuLogoLight from './logo-light.svg'

interface LogoProps {
  scale?: number
  priority?: boolean
}

export const Logo = ({ scale = 1, priority }: LogoProps) => (
  <Image
    src={TadokuLogo}
    alt="Tadoku"
    height={29 * scale}
    width={158 * scale}
    priority={priority}
  />
)

export const LogoInverted = ({ scale = 1, priority }: LogoProps) => (
  <Image
    src={TadokuLogoLight}
    alt="Tadoku"
    height={29 * scale}
    width={158 * scale}
    priority={priority}
  />
)
