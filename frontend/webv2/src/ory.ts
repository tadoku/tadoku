import { Configuration, V0alpha2Api, V0alpha2ApiInterface } from '@ory/client'
import getConfig from 'next/config'

const { publicRuntimeConfig } = getConfig()

const sdk: V0alpha2ApiInterface = new V0alpha2Api(
  new Configuration({ basePath: publicRuntimeConfig.kratosPublicEndpoint }),
) as unknown as V0alpha2ApiInterface

export default sdk

export const sdkServer: V0alpha2ApiInterface = new V0alpha2Api(
  new Configuration({ basePath: publicRuntimeConfig.kratosInternalEndpoint }),
) as unknown as V0alpha2ApiInterface
