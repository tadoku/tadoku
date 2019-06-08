const Constants = {
  colors: {
    primary: '#6969FF', //6320EE
    primaryWithAlpha: (alpha: number) => `rgba(105, 105, 255, ${alpha})`,
    secondary: '#1B264F',
    dark: '#211A1D', // 211A1D
    light: '#FFFFFF',
    lightGray: 'rgba(0, 0, 0, 0.08)',
    destructive: '#FC4A49', //8B1E3F
    lightDestructive: 'rgba(252, 74, 73, 0.1)', //8B1E3F
    destructiveWithAlpha: (alpha: number) => `rgba(252, 74, 73, ${alpha})`,
  },
  maxWidth: '1200px',
}

export default Constants
