import Image from 'next/image'
import TadokuLogo from './logo.svg'
import TadokuLogoLight from './logo-light.svg'

interface LogoProps {
  scale?: number
}

export const Logo = ({ scale = 1 }: LogoProps) => (
  <Image
    src={TadokuLogo}
    alt="Tadoku"
    height={29 * scale}
    width={158 * scale}
  />
)

export const LogoInverted = ({ scale = 1 }: LogoProps) => (
  <Image
    src={TadokuLogoLight}
    alt="Tadoku"
    height={29 * scale}
    width={158 * scale}
  />
)
