import Image from 'next/image'

interface LogoProps {
  scale?: number
}

export const Logo = ({ scale = 1 }: LogoProps) => (
  <Image
    src="/img/logo.svg"
    alt="Tadoku"
    height={29 * scale}
    width={158 * scale}
  />
)

export const LogoInverted = ({ scale = 1 }: LogoProps) => (
  <Image
    src="/img/logo-light.svg"
    alt="Tadoku"
    height={29 * scale}
    width={158 * scale}
  />
)
