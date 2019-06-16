import { NextContext } from 'next'
import * as Express from 'express'

export interface ExpressNextContext extends NextContext {
  req?: Express.Request
  res?: Express.Response
}

export type RawMapper<RawDataType, DataType> = (
  rawData: RawDataType,
) => DataType
