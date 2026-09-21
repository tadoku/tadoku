// Faker v10 ships TypeScript 5-only types; this shim keeps the styleguide on 4.9.
export const faker: {
  number: {
    int: (options: { min: number; max: number }) => number
  }
}
