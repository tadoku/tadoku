export type Mapper<A, B> = (rawData: A) => B

export interface Mappers<Raw, Original> {
  toRaw: (original: Original) => Raw
  fromRaw: (raw: Raw) => Original
}

export interface MappersWithOptional<Raw, Original>
  extends Mappers<Raw, Original> {
  optional: {
    toRaw: (original: Original | undefined) => Raw | undefined
    fromRaw: (raw: Raw | undefined) => Original | undefined
  }
}
