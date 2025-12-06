module.exports = () => ({
  connection: {
    client: 'postgres',
    connection: {
      host: 'postgres_strapi',
      port: 5432,
      database: 'strapi-cms',
      user: 'prashantpal',
      password: 'password123',
    },
  },
});
