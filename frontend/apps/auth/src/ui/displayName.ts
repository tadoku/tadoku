export const DISPLAY_NAME_MAX_LENGTH = 32

export const inputMaxLength = (name: string): number | undefined =>
  name === 'traits.display_name' ? DISPLAY_NAME_MAX_LENGTH : undefined
