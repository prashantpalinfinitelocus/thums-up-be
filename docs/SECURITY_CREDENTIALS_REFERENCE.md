# Security Assessment - Credentials Reference

**⚠️ SECURITY WARNING:** This document contains placeholder credentials for security testing. Actual production credentials should be provided separately through secure channels.

---

## Test Environment Credentials

### Application Access

#### Development Environment
- **Base URL:** `http://localhost:8080`
- **Health Check:** `http://localhost:8080/health`
- **Swagger UI:** `http://localhost:8080/swagger/index.html`

#### Staging Environment
- **Base URL:** `https://<staging-service-url>.run.app`
- **API Base Path:** `/backend/api/v1`

#### Production Environment
- **Base URL:** `https://tccc-tja-test-cloudrun-backend-<hash>-<region>.a.run.app`
- **API Base Path:** `/backend/api/v1`

---

## Test User Accounts

### Test User 1
```json
{
  "phone_number": "9876543210",
  "name": "Security Test User 1",
  "email": "security.test1@example.com",
  "password": "N/A (OTP-based auth)"
}
```

**To Create:**
```bash
POST /backend/api/v1/auth/send-otp
{
  "phone_number": "9876543210"
}

POST /backend/api/v1/auth/signup
{
  "phone_number": "9876543210",
  "name": "Security Test User 1",
  "email": "security.test1@example.com"
}
```

### Test User 2
```json
{
  "phone_number": "9876543211",
  "name": "Security Test User 2",
  "email": "security.test2@example.com"
}
```

### Admin Test Account
- **API Key:** `[To be provided by development team]`
- **Header:** `X-API-Key: <admin-api-key>`
- **Usage:** For testing admin endpoints like `/admin/winners/select`

---

## Database Credentials (Development Only)

⚠️ **Note:** These are for local development/testing only. Production credentials should be provided separately.

### Development Database
```bash
DB_HOST=localhost
DB_PORT=5432
DB_USER=postgres
DB_PASSWORD=[local-dev-password]
DB_NAME=thums_up_db
DB_SSL_MODE=disable
```

### Docker Database (if using docker-compose)
```bash
DB_HOST=postgres
DB_PORT=5432
DB_USER=postgres
DB_PASSWORD=[docker-password]
DB_NAME=thums_up_db
DB_SSL_MODE=disable
```

---

## JWT Test Tokens

### Sample Access Token Structure
```json
{
  "user_id": "550e8400-e29b-41d4-a716-446655440000",
  "phone_number": "9876543210",
  "exp": 1705843200,
  "iat": 1705839600
}
```

### How to Obtain Test Tokens

1. **Via Signup:**
   ```bash
   POST /backend/api/v1/auth/signup
   Response includes: access_token, refresh_token
   ```

2. **Via OTP Verification:**
   ```bash
   POST /backend/api/v1/auth/send-otp
   POST /backend/api/v1/auth/verify-otp
   Response includes: access_token, refresh_token
   ```

3. **Via Refresh:**
   ```bash
   POST /backend/api/v1/auth/refresh
   Body: { "refresh_token": "<token>" }
   ```

---

## Environment Variables Template

### Required for Security Testing

```bash
# Application
APP_ENV=development
APP_PORT=8080
ALLOWED_ORIGINS=*
SWAGGER_HOST=localhost:8080

# Database
DB_HOST=localhost
DB_PORT=5432
DB_USER=postgres
DB_PASSWORD=[provided-separately]
DB_NAME=thums_up_db
DB_SSL_MODE=disable

# JWT
JWT_SECRET_KEY=[provided-separately-min-32-chars]
JWT_ACCESS_TOKEN_EXPIRY=3600
JWT_REFRESH_TOKEN_EXPIRY=2592000

# Infobip (SMS/WhatsApp)
INFOBIP_BASE_URL=https://api.infobip.com
INFOBIP_API_KEY=[provided-separately]
INFOBIP_WA_NUMBER=[provided-separately]

# Google Cloud Platform
GCP_BUCKET_NAME=[provided-separately]
GCP_PROJECT_ID=ai0016084-tja-thunderzone-test
GCP_URL=https://storage.googleapis.com
GOOGLE_PUBSUB_PROJECT_ID=[provided-separately]
GOOGLE_PUBSUB_SUBSCRIPTION_ID=[provided-separately]
GOOGLE_PUBSUB_TOPIC_ID=[provided-separately]
FIREBASE_SERVICE_KEY_PATH=[provided-separately]

# Admin
X_API_KEY=[provided-separately]
```

---

## API Key Configuration

### Admin API Key
- **Header Name:** `X-API-Key`
- **Usage:** Required for admin endpoints
- **Example:**
  ```bash
  curl -X POST https://<api-url>/backend/api/v1/admin/winners/select \
    -H "X-API-Key: <admin-key>" \
    -H "Content-Type: application/json" \
    -d '{"week_number": 1, "number_of_winners": 10}'
  ```

### JWT Bearer Token
- **Header Name:** `Authorization`
- **Format:** `Bearer <access_token>`
- **Example:**
  ```bash
  curl -X GET https://<api-url>/backend/api/v1/profile \
    -H "Authorization: Bearer <jwt-token>" \
    -H "Content-Type: application/json"
  ```

---

## Third-Party Service Credentials

### Infobip
- **Base URL:** `https://api.infobip.com` (or as configured)
- **Authentication:** API Key in `Authorization: App <key>` header
- **Note:** Out of scope for direct testing

### Firebase
- **Service Account Key:** JSON file path configured via `FIREBASE_SERVICE_KEY_PATH`
- **Note:** Out of scope for direct testing

### Google Cloud Storage
- **Bucket Name:** Configured via `GCP_BUCKET_NAME`
- **Authentication:** Service account (Workload Identity in Cloud Run)
- **Note:** Direct bucket access out of scope; signed URLs in scope

---

## Test Data Setup

### Creating Test Users

1. **Send OTP:**
   ```bash
   POST /backend/api/v1/auth/send-otp
   {
     "phone_number": "9876543210"
   }
   ```

2. **Sign Up (if new user):**
   ```bash
   POST /backend/api/v1/auth/signup
   {
     "phone_number": "9876543210",
     "name": "Test User",
     "email": "test@example.com"
   }
   ```

3. **Verify OTP (if existing user):**
   ```bash
   POST /backend/api/v1/auth/verify-otp
   {
     "phone_number": "9876543210",
     "otp": "123456"
   }
   ```

### Creating Test Data

#### Contest Week
```bash
POST /backend/api/v1/contest-weeks
Authorization: Bearer <token>
{
  "start_date": "2024-01-15T00:00:00Z",
  "end_date": "2024-01-21T23:59:59Z",
  "winner_count": 10
}
```

#### Question
```bash
POST /backend/api/v1/questions
Authorization: Bearer <token>
{
  "question_text": "What makes Thums Up unique?",
  "language_id": 1
}
```

#### Address
```bash
POST /backend/api/v1/profile/address
Authorization: Bearer <token>
{
  "address1": "123 Test Street",
  "address2": "Apt 4B",
  "pincode": 400001,
  "state": "Maharashtra",
  "city": "Mumbai",
  "is_default": true
}
```

---

## Security Testing Scenarios

### Authentication Testing
1. **Valid OTP Flow:**
   - Send OTP → Verify with correct OTP → Get tokens

2. **Invalid OTP Flow:**
   - Send OTP → Verify with wrong OTP → Check error handling

3. **Expired OTP Flow:**
   - Send OTP → Wait 5+ minutes → Verify → Check expiry handling

4. **Rate Limiting:**
   - Send 4+ OTP requests in 1 hour → Check rate limit enforcement

### Authorization Testing
1. **Valid Token:**
   - Use valid JWT → Access protected endpoint → Should succeed

2. **Expired Token:**
   - Use expired JWT → Access protected endpoint → Should fail

3. **Invalid Token:**
   - Use malformed JWT → Access protected endpoint → Should fail

4. **Missing Token:**
   - No Authorization header → Access protected endpoint → Should fail

5. **Admin Endpoint:**
   - Use valid API key → Access admin endpoint → Should succeed
   - Use invalid API key → Access admin endpoint → Should fail

### Input Validation Testing
1. **Phone Number:**
   - Valid: 10 digits
   - Invalid: 9 digits, 11 digits, non-numeric, special characters

2. **Email:**
   - Valid: proper email format
   - Invalid: missing @, invalid domain, special characters

3. **File Upload:**
   - Valid: image files (jpg, png)
   - Invalid: executable files, oversized files, wrong MIME type

---

## Credential Rotation Policy

### Production Credentials
- **JWT Secret:** Rotate every 90 days
- **API Keys:** Rotate every 180 days
- **Database Passwords:** Rotate every 90 days
- **Service Account Keys:** Rotate every 180 days

### Test Credentials
- **Test Users:** Can be created/deleted as needed
- **Test API Keys:** Separate from production keys

---

## Secure Credential Storage

### Development
- Use `.env` file (not committed to git)
- Use `.env.example` for template (no actual secrets)

### Staging/Production
- Use Google Secret Manager
- Reference secrets in Cloud Run via `--set-secrets` flag
- Never commit secrets to version control

### Example Secret Manager Usage
```bash
# In cloudbuild-main.yaml
--set-secrets=JWT_SECRET_KEY=JWT_SECRET_KEY:latest,DB_PASSWORD=DB_PASSWORD:latest
```

---

## Contact for Credentials

### Development Team
- **Contact:** [Development Team Contact]
- **For:** Test environment credentials, API keys

### Security Team
- **Contact:** [Security Team Contact]
- **For:** Production credentials, security testing access

### Infrastructure Team
- **Contact:** [Infrastructure Team Contact]
- **For:** GCP service account keys, database access

---

**⚠️ IMPORTANT:** 
- Never commit actual credentials to version control
- Use environment variables or secret management systems
- Rotate credentials regularly
- Use different credentials for dev/staging/production
- Report any exposed credentials immediately

---

*This document should be kept secure and updated as credentials are rotated.*

