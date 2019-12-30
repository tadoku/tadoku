export const validateEmail = (email: string): boolean =>
  email != '' && !!email.match(/.+@.+/)

export const validatePassword = (password: string): boolean =>
  password != '' && password.length >= 6

export const validateDisplayName = (name: string): boolean =>
  /^([\p{Alphabetic}\p{Mark}\p{Decimal_Number}\p{Connector_Punctuation}\p{Join_Control}_-]{2,18})$/u.exec(
    name,
  ) !== null
