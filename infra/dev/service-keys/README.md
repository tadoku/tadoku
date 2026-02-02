# Development Service Keys

These keys are for **local development only**. They are intentionally committed to the repository to make local testing easier.

## Files

| File | Purpose |
|------|---------|
| `immersion-api.key` | Private key for immersion-api to sign tokens |
| `immersion-api.pub` | Public key for other services to verify immersion-api tokens |

## Usage

**immersion-api** (caller): Set `SERVICE_PRIVATE_KEY_PATH` to the `.key` file.

**profile-api** (receiver): Set `SERVICE_PUBLIC_KEYS_DIR` to this directory.

## Generating new keys

To add a new service:

```bash
# Generate private key
openssl ecparam -name prime256v1 -genkey -noout -out <service-name>.key

# Extract public key
openssl ec -in <service-name>.key -pubout -out <service-name>.pub
```

## Production

Production keys should be generated separately and stored in a secrets manager (e.g., Kubernetes secrets). Never use these development keys in production.
