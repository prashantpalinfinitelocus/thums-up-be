/**
 * home-page router
 */

import { factories } from "@strapi/strapi";

export default factories.createCoreRouter("api::content.content", {
  config: {
    findOne: {
      auth: false,
    },
    find: {
      auth: false,
    },
  },
});
