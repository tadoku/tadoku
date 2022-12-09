import { Configuration, V0alpha2Api, V0alpha2ApiInterface } from '@ory/client'

const sdk: V0alpha2ApiInterface = new V0alpha2Api(
  new Configuration({ basePath: '/kratos' }),
) as unknown as V0alpha2ApiInterface

export default sdk
