import { Configuration, FrontendApi } from '@ory/kratos-client'
import getConfig from 'next/config'

const { publicRuntimeConfig } = getConfig()

const sdk = new FrontendApi(
  new Configuration({
    basePath: publicRuntimeConfig.kratosPublicEndpoint,
    baseOptions: { withCredentials: true },
  }),
)

export default sdk

export const sdkServer = new FrontendApi(
  new Configuration({ basePath: publicRuntimeConfig.kratosInternalEndpoint }),
)
