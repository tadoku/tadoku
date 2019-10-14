import { NextPageContext } from 'next'

export type Mapper<A, B> = (rawData: A) => B
