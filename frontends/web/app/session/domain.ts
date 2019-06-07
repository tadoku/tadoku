export const validateEmail = (email: string): boolean =>
  email != '' && !!email.match(/.+@.+/)

export const validatePassword = (password: string): boolean =>
  password != '' && password.length >= 6

export const validateDisplayName = (name: string): boolean =>
  name != '' && name.length >= 2
