export default () => {
  return async (ctx, next) => {
    if (ctx.request.url.startsWith('/strapi')) {
      ctx.request.url = ctx.request.url.replace(/^\/strapi/, '') || '/';
    }
    await next();
  };
};
