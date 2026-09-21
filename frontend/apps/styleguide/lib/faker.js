// Runtime import only; TypeScript uses the sibling .d.ts so 4.9 never parses faker v10 types.
import { faker as fakerRuntime } from '@faker-js/faker'

export const faker = fakerRuntime
