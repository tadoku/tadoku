import { NextPageContext } from 'next'
import { Request, Response } from 'express'

export interface ExpressNextContext extends NextPageContext {
  req?: Request
  res?: Response
}

export type Mapper<A, B> = (rawData: A) => B
