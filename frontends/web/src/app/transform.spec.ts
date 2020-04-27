import {
  createMappers,
  createSerializer,
  createCollectionSerializer,
  optionalizeSerializer,
} from './transform'

interface Raw {
  a: string
  b: string
}

interface Target {
  a: string
  b: Date
}

const toRaw = (target: Target): Raw => ({
  a: target.a,
  b: target.b.toISOString(),
})

const fromRaw = (raw: Raw): Target => ({
  a: raw.a,
  b: new Date(raw.b),
})

describe('createMappers', () => {
  it('should be able to map between raw/target back and forth', () => {
    const mappers = createMappers({ toRaw, fromRaw })
    const raw = { a: 'foobar', b: '2020-04-27T04:53:54.893Z' }
    expect(mappers.toRaw(mappers.fromRaw(raw))).toMatchObject(raw)
  })
})

describe('createSerializer', () => {
  it('should be able to serialize and deserialize back and forth', () => {
    const mappers = createMappers({ toRaw, fromRaw })
    const serializer = createSerializer(mappers)
    const target = { a: 'foobar', b: new Date() }
    expect(serializer.deserialize(serializer.serialize(target))).toMatchObject(
      target,
    )
  })
})

describe('createCollectionSerializer', () => {
  it('should be able to serialize and deserialize a collection back and forth', () => {
    const mappers = createMappers({ toRaw, fromRaw })
    const serializer = createCollectionSerializer(mappers)
    const target = [
      { a: 'foo', b: new Date() },
      { a: 'bar', b: new Date() },
    ]
    expect(serializer.deserialize(serializer.serialize(target))).toMatchObject(
      target,
    )
  })
})

describe('optionalizeSerializer', () => {
  it('should be able to handle undefined values correctly', () => {
    const mappers = createMappers({ toRaw, fromRaw })
    const serializer = optionalizeSerializer(createSerializer(mappers))
    const target = { a: 'foobar', b: new Date() }

    expect(serializer.deserialize(serializer.serialize(target))).toMatchObject(
      target,
    )
    expect(serializer.deserialize(serializer.serialize(undefined))).toEqual(
      undefined,
    )
  })
})
