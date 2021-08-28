export interface Service {
  internalHost: string
  endpoint: string
}

const createService = (suffix: string): Service => ({
  internalHost: `${process.env.API_ROOT}/${suffix}`,
  endpoint: `/api/${suffix}`,
})

const services: { [key: string]: Service } = {
  tadokuContest: createService('reading-contest'),
  identity: createService('identity'),
  blog: createService('blog'),
}

export const getService = (service: string): Service => {
  if (service in services) {
    return services[service]
  }

  throw new Error(`service not found with name ${service}`)
}
