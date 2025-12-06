const path = require('path');
const fs = require('fs');

module.exports = ({ env }) => {
  const serviceAccountPath = path.resolve(__dirname, '../../gcp-service-account.json');
  const serviceAccount = fs.existsSync(serviceAccountPath) 
    ? JSON.parse(fs.readFileSync(serviceAccountPath, 'utf8'))
    : undefined;

  return {
    email: false,
    "api-tokens": false,
    webhooks: false,
    "audit-logs": false,
    upload: {
      config: {
        provider: '@strapi-community/strapi-provider-upload-google-cloud-storage',
        providerOptions: {
          bucketName: env('GCS_BUCKET_NAME', 'coca-cola-notifications-3a790.appspot.com'),
          baseUrl: `https://storage.googleapis.com/${env('GCS_BUCKET_NAME', 'coca-cola-notifications-3a790.appspot.com')}`,
          basePath: 'thums-up/strapi',
          publicFiles: true,
          uniform: false,
          serviceAccount: serviceAccount,
        },
      },
    },
  };
};

