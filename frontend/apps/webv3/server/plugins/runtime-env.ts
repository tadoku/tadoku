import { accessSync, constants, statSync } from 'node:fs'
import { isIP } from 'node:net'
import { definePlugin } from 'nitro'
import { runtimeOrigins } from '../../src/lib/urls.server'

/** Checks every runtime variable before the server listens, so a bad value stops the container instead of failing pages. */
function runtimeEnvErrors() {
  const errors: string[] = []
  try {
    runtimeOrigins()
  } catch (error) {
    errors.push((error as Error).message)
  }

  const { PORT, HOST, SERVER_SHUTDOWN_TIMEOUT, NODE_EXTRA_CA_CERTS } = process.env
  if (PORT !== undefined && !(/^\d{1,5}$/.test(PORT) && Number(PORT) >= 1 && Number(PORT) <= 65535)) {
    errors.push('PORT must be a port number from 1 to 65535')
  }
  if (HOST !== undefined && HOST !== 'localhost' && !isIP(HOST)) {
    errors.push('HOST must be an IP address or localhost')
  }
  if (SERVER_SHUTDOWN_TIMEOUT !== undefined && !/^[1-9]\d{0,3}$/.test(SERVER_SHUTDOWN_TIMEOUT)) {
    errors.push('SERVER_SHUTDOWN_TIMEOUT must be a whole number of seconds')
  }
  if (NODE_EXTRA_CA_CERTS) {
    let readableFile = false
    try {
      accessSync(NODE_EXTRA_CA_CERTS, constants.R_OK)
      readableFile = statSync(NODE_EXTRA_CA_CERTS).isFile()
    } catch {
      // A missing or unreadable path is reported below.
    }
    if (!readableFile) errors.push('NODE_EXTRA_CA_CERTS must name a readable file')
  }
  return errors
}

export default definePlugin(() => {
  const errors = runtimeEnvErrors()
  if (errors.length === 0) return

  process.stderr.write(`Webv3 cannot start: ${errors.join('; ')}\n`)
  process.exit(1)
})
