export default ({ env }) => ({
  auth: {
    secret: env("ADMIN_JWT_SECRET", "thums-up-admin-secret"),
  },
  apiToken: {
    salt: env("API_TOKEN_SALT", "thums-up-api-token-salt"),
  },
  transfer: {
    token: {
      salt: env("TRANSFER_TOKEN_SALT", "thums-up-transfer-token-salt"),
    },
  },
  flags: {
    nps: false,
    promoteEE: false,
  },
});
