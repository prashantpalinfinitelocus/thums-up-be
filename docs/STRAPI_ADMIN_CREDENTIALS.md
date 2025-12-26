# Strapi Admin Credentials Guide

## Admin Access URL
**Admin Panel:** `http://localhost:1338/strapi/admin`

## Database Connection Details
- **Host:** `postgres_strapi` (from Docker) or `localhost:5438` (from host)
- **Database:** `strapi-cms`
- **Username:** `prashantpal`
- **Password:** `password123`

---

## Method 1: Retrieve Existing Admin Credentials

### Using the provided script:
```bash
./scripts/get-strapi-admin-credentials.sh
```

### Manual database query:
```bash
# If Docker is running:
docker exec thums_up_postgres_strapi psql -U prashantpal -d strapi-cms -c "SELECT email, firstname, lastname, is_active FROM strapi_admin_users;"

# Or direct connection:
PGPASSWORD=password123 psql -h localhost -p 5438 -U prashantpal -d strapi-cms -c "SELECT email, firstname, lastname, is_active FROM strapi_admin_users;"
```

**Note:** Passwords are hashed (bcrypt) and cannot be retrieved from the database. You'll only see the email address.

---

## Method 2: Reset Admin Password

If you know the admin email but forgot the password:

```bash
cd strapi
npm run strapi admin:reset-user-password --email=your-email@example.com
```

This will prompt you to set a new password.

---

## Method 3: Create New Admin User (First Time Setup)

### Option A: Via Strapi Admin UI (First Run)
1. Start Strapi: `docker-compose up strapi` or `cd strapi && npm run dev`
2. Navigate to: `http://localhost:1338/strapi/admin`
3. Fill in the registration form:
   - First name
   - Last name
   - Email
   - Password

### Option B: Via Strapi CLI
```bash
cd strapi
npm run strapi admin:create-user
```

### Option C: Via Database (Advanced)
If you need to create an admin user directly in the database, you'll need to:
1. Hash the password using bcrypt
2. Insert into `strapi_admin_users` table
3. Assign appropriate roles in `strapi_admin_users_roles_links`

**Recommended:** Use Option A or B instead.

---

## Method 4: Check if Admin User Exists

```bash
# Check via Docker
docker exec thums_up_postgres_strapi psql -U prashantpal -d strapi-cms -c "SELECT COUNT(*) FROM strapi_admin_users;"

# If count is 0, you need to create an admin user (Method 3)
# If count > 0, use Method 1 to see the email, then Method 2 to reset password
```

---

## Quick Start Checklist

1. ✅ Start Docker: `docker-compose up -d`
2. ✅ Wait for containers to be healthy
3. ✅ Check if admin exists: `./scripts/get-strapi-admin-credentials.sh`
4. ✅ If no admin exists, go to `http://localhost:1338/strapi/admin` to create one
5. ✅ If admin exists but password forgotten, reset it using Method 2

---

## Troubleshooting

### "Cannot connect to database"
- Ensure Docker is running: `docker ps`
- Check if postgres_strapi container is running: `docker ps | grep postgres_strapi`
- Start containers: `docker-compose up -d postgres_strapi`

### "No admin users found"
- This is normal for first-time setup
- Navigate to `http://localhost:1338/strapi/admin` to create the first admin user

### "Forgot admin email"
- Query the database using Method 1 to see all admin emails
- Or check with your team members who have access

---

## Security Notes

⚠️ **Important:** 
- Never commit admin credentials to version control
- Change default passwords in production
- Use environment variables for sensitive data
- Regularly rotate JWT secrets and API tokens








