import { Configuration, V0alpha2Api, V0alpha2ApiInterface } from '@ory/client'

const sdk: V0alpha2ApiInterface = new V0alpha2Api(
  new Configuration({ basePath: 'http://account.langlog.be/kratos' }),
) as unknown as V0alpha2ApiInterface

export default sdk

export const sdkServer: V0alpha2ApiInterface = new V0alpha2Api(
  new Configuration({ basePath: 'http://kratos-public' }),
) as unknown as V0alpha2ApiInterface
