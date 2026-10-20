// Static checks of the rendered base, the dev-cli branch configuration and the worker's DevCLI metadata.
const fs = require('node:fs')
const path = require('node:path')
const Y = require('../../../../frontend/node_modules/yaml')

function checkConfiguration({ root, report }, docs) {
  const webv3 = docs.find(d => d.kind === 'Deployment' && d.metadata.name === 'frontend-webv3')
  if (!webv3 || webv3.metadata.namespace !== 'tdk-dev-frontend-webv3') throw new Error('Missing always-on Webv3 Deployment')
  const webv3Pod = webv3.spec.template.spec
  const webv3Container = webv3Pod.containers[0]
  if (webv3Container.command || webv3Container.readinessProbe?.httpGet?.path !== '/healthz' || webv3Container.livenessProbe?.httpGet?.path !== '/healthz') throw new Error('Webv3 must run its native Nitro server with API-independent health probes')
  if (webv3Pod.automountServiceAccountToken !== false || !webv3Container.securityContext?.readOnlyRootFilesystem || webv3Pod.securityContext?.runAsUser !== 1000) throw new Error('Webv3 must run without API credentials under its published restricted identity')
  const webv3Config = docs.find(d => d.kind === 'ConfigMap' && d.metadata.namespace === webv3.metadata.namespace && d.metadata.name === 'tadoku-cli-webv3')
  if (webv3Config?.data.TADOKU_WEB_URL !== 'https://tadoku.dev.lab' || webv3Config?.data.TADOKU_API_URL !== 'https://tadoku.dev.lab/api/internal') throw new Error('Webv3 runtime must use development product and API URLs')
  const webv3Route = docs.find(d => d.kind === 'HTTPRoute' && d.metadata.name === 'tadoku-cli-webv3-base')
  const webv3Backend = docs.find(d => d.kind === 'Backend' && d.metadata.name === 'tadoku-cli-webv3')
  const webv3Ingress = docs.find(d => d.kind === 'Ingress' && d.metadata.name === 'tadoku-webv3')
  if (webv3Route?.spec.hostnames.join(',') !== 'next.tadoku.dev.lab' || webv3Backend?.spec.endpoints[0].fqdn.hostname !== 'frontend-webv3.tdk-dev-frontend-webv3.svc.cluster.local' || webv3Route?.spec.rules[0].backendRefs[0].name !== webv3Backend.metadata.name || webv3Ingress?.spec.tls[0].hosts.join(',') !== 'next.tadoku.dev.lab') throw new Error('Webv3 requires canonical gateway routing and browser TLS')
  const proxyRoute = docs.find(d => d.kind === 'HTTPRoute' && d.metadata.name === 'tadoku-cli-proxy-base')
  if (!proxyRoute?.spec.hostnames.includes('next.tadoku.dev.lab')) throw new Error('Webv3 browser API requests must use Oathkeeper')
  const api = docs.find(d => d.kind === 'Deployment' && d.metadata.namespace === 'tdk-dev-tadoku-api' && d.metadata.name === 'tadoku-api')
  if (!api) throw new Error('Missing base Tadoku API Deployment')
  if (docs.some(d => d.kind === 'Deployment' && d.metadata.name === 'immersion-api')) throw new Error('Legacy Immersion Deployment remains in dev base')
  const envValue = (container, name) => container.env.find(e => e.name === name)?.value
  const worker = docs.find(d => d.kind === 'Deployment' && d.metadata.namespace === 'tdk-dev-tadoku-api' && d.metadata.name === 'tadoku-worker')
  if (!worker || worker.spec.replicas !== 1 || worker.spec.strategy?.type !== 'Recreate') throw new Error('Base worker must be a single Recreate Deployment')
  const workerPod = worker.spec.template.spec
  const workerContainer = workerPod.containers[0]
  const apiContainer = api.spec.template.spec.containers[0]
  const branchApi = Y.parse(fs.readFileSync(path.join(root, '.dev/tadoku-api.yaml'), 'utf8')).spec.containers[0]
  const branchWorker = Y.parse(fs.readFileSync(path.join(root, '.dev/tadoku-worker.yaml'), 'utf8')).spec.containers[0]
  if (envValue(workerContainer, 'WORKER_POSTGRES_DATABASE') !== 'tadoku' || envValue(workerContainer, 'WORKER_LEADERBOARD_CACHE_PREFIX') !== undefined || envValue(apiContainer, 'API_LEADERBOARD_CACHE_PREFIX') !== undefined) throw new Error('Base worker must use the base database and both base processes must use tenant-derived cache without a private-database prefix')
  if (envValue(branchWorker, 'WORKER_POSTGRES_DATABASE') !== '${DEV_VAR_DATABASE}' || envValue(branchWorker, 'WORKER_BRANCH') !== '${DEV_VAR_TENANT}' || envValue(branchWorker, 'WORKER_LEADERBOARD_CACHE_PREFIX') !== undefined) throw new Error('Branch worker must use the resolved shared database and exact tenant without a cache prefix')
  if (envValue(branchApi, 'API_POSTGRES_DATABASE') !== '${DEV_VAR_DATABASE}' || envValue(branchApi, 'API_BRANCH') !== '${DEV_VAR_TENANT}' || envValue(branchApi, 'API_LEADERBOARD_CACHE_PREFIX') !== undefined) throw new Error('Branch API must use the resolved shared database and exact tenant without a cache prefix')
  const identityEnvironment = {
    API_KRATOS_PUBLIC_URL: 'http://kratos-public.tdk-dev-kratos',
    API_IDENTITY_COOKIE_DOMAIN: 'tadoku.dev.lab',
    API_IDENTITY_TRUSTED_PROXY_CIDRS: '10.53.0.0/24',
  }
  for (const [name, expected] of Object.entries(identityEnvironment)) {
    if (envValue(apiContainer, name) !== expected || envValue(branchApi, name) !== expected) throw new Error(`Base and branch identity configuration must set ${name}=${expected}`)
  }
  report.identityEnvironment = identityEnvironment
  const devConfig = Y.parse(fs.readFileSync(path.join(root, '.dev/config.yaml'), 'utf8'))
  const taskNames = new Set(devConfig.tasks.map(task => task.name))
  const hooks = [devConfig.hooks?.beforeUp, devConfig.hooks?.afterDown, ...Object.values(devConfig.hooks?.deployables || {}).flatMap(deployable => [deployable.beforeStart, deployable.afterStop])].flat().filter(Boolean)
  if (devConfig.kubeContext !== 'homelab-talos-dev' || devConfig.variables?.DATABASE !== 'tadoku' || devConfig.variables?.TENANT !== 'tadoku/${DEV_ROUTE}') throw new Error('Normal branch provisioning must resolve a tenant on the Talos base database')
  if (!hooks.length || hooks.some(name => !taskNames.has(name)) || taskNames.has('seed')) throw new Error('Tenant lifecycle hooks must name declared tasks and replace the private-database seed task')
  if (devConfig.hooks.beforeUp.join(',') !== 'tenant' || devConfig.hooks.afterDown.join(',') !== 'tenant-teardown' || devConfig.hooks.deployables['tadoku-worker']?.beforeStart?.join(',') !== 'worker-override-set' || devConfig.hooks.deployables['tadoku-worker']?.afterStop?.join(',') !== 'worker-override-clear') throw new Error('Tenant and worker ownership hooks must bracket the correct lifecycle operations')
  const migrationProfile = devConfig.profiles?.find(profile => profile.name === 'isolated-database')
  if (migrationProfile?.whenChanged?.join(',') !== 'services/tadoku-api/migrations/' || migrationProfile.variables?.DATABASE !== 'tadoku-${DEV_ROUTE}' || migrationProfile.hooks?.beforeUp?.join(',') !== 'migrate,tenant' || migrationProfile.hooks?.afterDown !== undefined) throw new Error('Migration branches must automatically select an isolated database, migrate before tenant provisioning and inherit tenant teardown')
  for (const profile of devConfig.profiles) {
    if (Object.keys(profile.variables || {}).some(name => !(name in devConfig.variables))) throw new Error(`Profile ${profile.name} names undeclared variables`)
    const profileHooks = [profile.hooks?.beforeUp, profile.hooks?.afterDown, ...Object.values(profile.hooks?.deployables || {}).flatMap(deployable => [deployable.beforeStart, deployable.afterStop])].flat().filter(Boolean)
    if (profileHooks.some(name => !taskNames.has(name))) throw new Error(`Profile ${profile.name} names undeclared lifecycle tasks`)
  }
  const migrationTask = devConfig.tasks.find(task => task.name === 'migrate')
  if (migrationTask?.namespace !== 'tdk-dev-data' || migrationTask.dependencies?.join(',') !== 'database' || migrationTask.target !== 'tadoku-dev-db.tdk-dev-data/tadoku-${DEV_ROUTE}') throw new Error('Branch migration must run with its database dependency and owner Secret in the data namespace')
  const branchMigration = Y.parse(fs.readFileSync(path.join(root, migrationTask.manifest), 'utf8'))
  for (const container of [...branchMigration.spec.template.spec.containers, ...branchMigration.spec.template.spec.initContainers]) {
    const databaseName = container.name === 'migrate' ? 'POSTGRES_DATABASE' : 'PGDATABASE'
    const userName = container.name === 'migrate' ? 'POSTGRES_USER' : 'PGUSER'
    const passwordName = container.name === 'migrate' ? 'POSTGRES_PASSWORD' : 'PGPASSWORD'
    if (envValue(container, databaseName) !== 'tadoku-${DEV_ROUTE}' || [userName, passwordName].some(name => container.env.find(env => env.name === name)?.valueFrom?.secretKeyRef?.name !== 'tadoku-owner.tadoku-dev-db.credentials.postgresql.acid.zalan.do')) throw new Error('Branch migrations and grants must use the literal branch database with owner credentials, never the base or a resolved database variable')
  }
  report.migrationProfile = { name: migrationProfile.name, whenChanged: migrationProfile.whenChanged, database: migrationProfile.variables.DATABASE, beforeUp: migrationProfile.hooks.beforeUp, afterDown: devConfig.hooks.afterDown, namespace: migrationTask.namespace }
  for (const name of ['tenant', 'tenant-teardown', 'worker-override-set', 'worker-override-clear']) {
    const task = devConfig.tasks.find(task => task.name === name)
    if (!task || task.namespace !== 'tdk-dev-data' || task.dependencies?.length || task.target !== 'tadoku-dev-db.tdk-dev-data/tadoku-${DEV_ROUTE}') throw new Error(`Lifecycle task ${name} must serialize by route without a database dependency`)
    const job = Y.parse(fs.readFileSync(path.join(root, task.manifest), 'utf8'))
    const pod = job.spec.template.spec
    const container = pod.containers.find(container => container.name === task.container)
    if (job.spec.backoffLimit !== 0 || job.spec.activeDeadlineSeconds !== 240 || job.spec.ttlSecondsAfterFinished !== 86400 || pod.automountServiceAccountToken !== false || job.spec.template.metadata.labels['app.kubernetes.io/name'] !== 'tadoku-tenant') throw new Error(`Lifecycle task ${name} must be bounded and identified for provider network policy`)
    const security = container.securityContext
    if (!security?.runAsNonRoot || security.runAsUser !== 65532 || security.allowPrivilegeEscalation !== false || security.seccompProfile?.type !== 'RuntimeDefault' || !security.capabilities?.drop?.includes('ALL')) throw new Error(`Lifecycle task ${name} must satisfy restricted Pod security`)
    if (envValue(container, 'PGHOST') !== 'tadoku-dev-db.tdk-dev-data' || envValue(container, 'PGDATABASE') !== '${DEV_VAR_DATABASE}' || envValue(container, 'TENANT_FLIPT_ENVIRONMENT') !== 'test') throw new Error(`Lifecycle task ${name} must use the resolved database and isolated Flipt environment`)
  }
  const lifecycleIngress = docs.find(doc => doc.kind === 'NetworkPolicy' && doc.metadata.namespace === 'tdk-dev-flipt' && doc.metadata.name === 'allow-tenant-lifecycle-ingress')
  if (!lifecycleIngress?.spec.ingress.some(rule => rule.ports?.some(port => port.port === 8080 && port.protocol === 'TCP') && rule.from?.some(source => source.namespaceSelector?.matchLabels?.['kubernetes.io/metadata.name'] === 'tdk-dev-data' && source.podSelector?.matchLabels?.['app.kubernetes.io/name'] === 'tadoku-tenant'))) throw new Error('Only lifecycle-labelled data Jobs must receive direct Flipt management access')
  if (workerPod.automountServiceAccountToken !== false || workerContainer.image === apiContainer.image || !workerContainer.image.startsWith('ghcr.io/tadoku/tadoku/tadoku-worker:latest')) throw new Error('Worker image or private Pod configuration is invalid')
  if (docs.some(d => d.kind === 'Service' && d.spec?.selector?.app === 'tadoku-worker') || docs.some(d => d.kind === 'HTTPRoute' && JSON.stringify(d.spec).includes('tadoku-worker'))) throw new Error('Worker must have no Service or public route')
  const resources = workerContainer.resources
  if (!resources?.requests?.cpu || !resources.requests.memory || !resources.limits?.cpu || !resources.limits.memory) throw new Error('Worker must declare CPU and memory requests and limits')
  const emailBindings = container => container.env.filter(e => e.name === 'WORKER_EMAIL_ENABLED')
  if ([workerContainer, branchWorker].some(container => emailBindings(container).length !== 1 || emailBindings(container)[0].value !== 'false')) throw new Error('Base and branch workers must bind WORKER_EMAIL_ENABLED=false explicitly')
  const kratosConfig = docs.find(d => d.kind === 'ConfigMap' && d.metadata.namespace === 'tdk-dev-kratos' && d.metadata.name.startsWith('kratos-config'))
  const courier = Y.parse(kratosConfig?.data?.['kratos.yaml'] || '{}').courier || {}
  const courierURL = new URL(courier.smtp?.connection_uri || 'invalid:')
  if (Object.keys(courier).join(',') !== 'smtp' || courierURL.protocol !== 'smtp:' || courierURL.host !== 'mailhog.tdk-dev-data:1025') throw new Error('Kratos must deliver only through SMTP to MailHog')
  report.leaderboardWorkers = { asyncJobs: 'tadoku-worker', baseDatabase: 'tadoku', branchDatabase: envValue(branchWorker, 'WORKER_POSTGRES_DATABASE'), branchTenant: envValue(branchWorker, 'WORKER_BRANCH'), baseKeyPattern: 'tenant:tadoku/prod:leaderboard:*', branchKeyPattern: 'tenant:tadoku/${DEV_ROUTE}:leaderboard:*', image: workerContainer.image }
}

function checkWorkerMetadata({ run }) {
  run('bazel', ['build', '//services/tadoku-api:dev', '//services/tadoku-api:worker_dev'])
  const bazelBin = run('bazel', ['info', 'bazel-bin'])
  const apiMetadata = JSON.parse(fs.readFileSync(path.join(bazelBin, 'services/tadoku-api/dev.dev.json')))
  const workerMetadata = JSON.parse(fs.readFileSync(path.join(bazelBin, 'services/tadoku-api/worker_dev.dev.json')))
  if (apiMetadata.selectionGroup !== 'tadoku-api' || workerMetadata.selectionGroup !== 'tadoku-api' || workerMetadata.kind !== 'worker') throw new Error('Branch API and worker are not selected as one group')
  if (workerMetadata.imageTarget !== '//services/tadoku-api/cmd/tadoku-worker:cli_image' || workerMetadata.pushTarget !== '//services/tadoku-api/cmd/tadoku-worker:cli_push' || workerMetadata.workloadTemplate !== '.dev/tadoku-worker.yaml') throw new Error('Worker is not an independent DevCLI image/deployable')
  if (['publicPath', 'internalHost', 'publicProxy', 'baseService', 'servicePort'].some(key => key in workerMetadata)) throw new Error('Worker metadata must not request a route or Service')
}

module.exports = { checkConfiguration, checkWorkerMetadata }
