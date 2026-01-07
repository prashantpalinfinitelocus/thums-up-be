# Security Assessment Documentation

**Document Version:** 1.0  
**Last Updated:** January 2025  
**Prepared For:** Security Team Assessment  
**Application:** Thums Up Backend API

---

## Table of Contents

1. [Application Overview](#application-overview)
2. [Application URLs and Domains](#application-urls-and-domains)
3. [In-Scope URLs](#in-scope-urls)
4. [Out-of-Scope URLs](#out-of-scope-urls)
5. [Credentials and Access](#credentials-and-access)
6. [Workflow Documentation](#workflow-documentation)
7. [Sample Data](#sample-data)
8. [CodeQL Scan Results](#codeql-scan-results)
9. [Dependabot Scan Results](#dependabot-scan-results)
10. [Code Scanning Results](#code-scanning-results)
11. [Secret Scanning Results](#secret-scanning-results)
12. [DAST/StackHawk Scan Results](#daststackhawk-scan-results)
13. [Environment Configuration](#environment-configuration)
14. [API Endpoints Reference](#api-endpoints-reference)
15. [Third-Party Integrations](#third-party-integrations)

---

## Application Overview

**Application Name:** Thums Up Backend API  
**Technology Stack:** Go 1.23.8, Gin Framework, PostgreSQL, Google Cloud Platform  
**Architecture:** RESTful API with Clean Architecture pattern  
**Primary Function:** Backend service for gamified engagement platform with Thunder Seat contests

### Key Features
- User authentication via OTP (SMS/WhatsApp)
- Profile and address management
- Contest week management
- Question submission and answer management
- Thunder Seat contest participation
- Winner selection and KYC management
- File uploads (avatars, documents) via Google Cloud Storage

---

## Application URLs and Domains

### Production Environment
- **Base URL:** `https://tccc-tja-test-cloudrun-backend-<hash>-<region>.a.run.app`
- **API Base Path:** `/backend/api/v1`
- **Health Check:** `https://<base-url>/health`
- **Swagger Documentation:** `https://<base-url>/swagger/index.html`

### Staging Environment
- **Base URL:** `https://<staging-service-name>-<hash>-<region>.a.run.app`
- **API Base Path:** `/backend/api/v1`

### Development Environment
- **Base URL:** `http://localhost:8080`
- **API Base Path:** `/backend/api/v1`
- **Health Check:** `http://localhost:8080/health`
- **Swagger Documentation:** `http://localhost:8080/swagger/index.html`

### GCP Project Details
- **Project ID:** `ai0016084-tja-thunderzone-test`
- **Region:** `asia-south1`
- **Service Name:** `tccc-tja-test-cloudrun-backend`
- **Container Registry:** `${_REGION}-docker.pkg.dev/${_PROJECT_ID}/${_REPO_NAME}/${_MAIN_SERVICE}`

---

## In-Scope URLs

The following URLs and endpoints are **IN SCOPE** for security assessment:

### API Endpoints (All under `/backend/api/v1`)

#### Authentication Endpoints
- `POST /backend/api/v1/auth/send-otp` - Send OTP to phone number
- `POST /backend/api/v1/auth/verify-otp` - Verify OTP and get tokens
- `POST /backend/api/v1/auth/signup` - Register new user
- `POST /backend/api/v1/auth/refresh` - Refresh access token
- `GET /backend/api/v1/auth/login-count` - Get user login count (authenticated)

#### Profile Management
- `GET /backend/api/v1/profile` - Get user profile (authenticated)
- `PATCH /backend/api/v1/profile` - Update user profile (authenticated)
- `POST /backend/api/v1/profile/address` - Add address (authenticated)
- `GET /backend/api/v1/profile/address` - Get user addresses (authenticated)
- `PUT /backend/api/v1/profile/address/:addressId` - Update address (authenticated)
- `DELETE /backend/api/v1/profile/address/:addressId` - Delete address (authenticated)
- `GET /backend/api/v1/profile/questions` - Get user questions (authenticated)
- `POST /backend/api/v1/profile/questions/text` - Get question by ID (authenticated)
- `POST /backend/api/v1/profile/questions` - Answer questions (authenticated)
- `POST /backend/api/v1/profile/questions/create` - Create questions (authenticated)

#### Question Management
- `GET /backend/api/v1/questions/active` - Get active questions (public)
- `POST /backend/api/v1/questions` - Submit question (authenticated)

#### Thunder Seat Contest
- `GET /backend/api/v1/thunder-seat/current-week` - Get current contest week (public)
- `GET /backend/api/v1/thunder-seat/submissions` - Get user submissions (authenticated)
- `POST /backend/api/v1/thunder-seat` - Submit answer (authenticated)

#### Contest Week Management
- `GET /backend/api/v1/contest-weeks` - Get all contest weeks (public)
- `GET /backend/api/v1/contest-weeks/active` - Get active contest week (public)
- `GET /backend/api/v1/contest-weeks/:weekNumber` - Get contest week by number (public)
- `POST /backend/api/v1/contest-weeks` - Create contest week (authenticated)
- `POST /backend/api/v1/contest-weeks/activate` - Activate contest week (authenticated)

#### Winner Management
- `GET /backend/api/v1/winners` - Get all winners with pagination (public)
- `GET /backend/api/v1/winners/week/:weekNumber` - Get winners by week (public)
- `GET /backend/api/v1/winners/status` - Check winner status (authenticated)
- `POST /backend/api/v1/winners/mark-viewed` - Mark banner as viewed (authenticated)
- `POST /backend/api/v1/winners/kyc` - Submit winner KYC (authenticated)

#### Avatar Management
- `GET /backend/api/v1/avatars` - Get all avatars (authenticated)
- `GET /backend/api/v1/avatars/:avatarId` - Get avatar by ID (authenticated)
- `POST /backend/api/v1/avatars` - Create avatar (authenticated)

#### State Management
- `GET /backend/api/v1/states` - Get all states (public)

#### Website Status
- `GET /backend/api/v1/website-status` - Get website status (public)

#### Admin Endpoints
- `POST /backend/api/v1/admin/winners/select` - Select winners for week (API key required)

#### System Endpoints
- `GET /health` - Health check endpoint (public)
- `GET /swagger/*` - API documentation (public)

### Google Cloud Storage
- **Bucket:** Configured via `GCP_BUCKET_NAME` environment variable
- **Base URL:** Configured via `GCP_URL` environment variable
- **Access:** Via signed URLs generated by the application

---

## Out-of-Scope URLs

The following URLs and services are **OUT OF SCOPE** for security assessment:

### Third-Party Services

1. **Infobip API**
   - **Base URL:** Configured via `INFOBIP_BASE_URL` environment variable
   - **Endpoints:**
     - `POST /sms/2/text/advanced` - SMS/WhatsApp sending
   - **Reason:** Third-party service, not owned by the application

2. **Firebase Cloud Messaging (FCM)**
   - **Service:** `firebase.google.com/go/v4`
   - **Endpoints:** All Firebase API endpoints
   - **Reason:** Third-party Google service for push notifications

3. **Google Cloud Pub/Sub**
   - **Service:** `cloud.google.com/go/pubsub`
   - **Endpoints:** All Pub/Sub API endpoints
   - **Reason:** Third-party Google service for message queuing

4. **Google Cloud Storage (Direct Access)**
   - **Service:** `storage.googleapis.com`
   - **Direct bucket URLs:** Out of scope
   - **Note:** Only signed URLs generated by the application are in scope

5. **PostgreSQL Database**
   - **Direct database connections:** Out of scope
   - **Note:** Only API endpoints that interact with the database are in scope

6. **Strapi CMS**
   - **URL:** `http://localhost:1338/strapi/admin` (development)
   - **Reason:** Separate CMS application, not part of main backend API

7. **External Monitoring Services**
   - Prometheus metrics endpoints (if exposed separately)
   - Cloud Monitoring dashboards

### Infrastructure Services
- Google Cloud Run service management endpoints
- Google Cloud Build endpoints
- Google Secret Manager API
- Google IAM API
- Container registry endpoints

---

## Credentials and Access

### Environment Variables Required

The following environment variables are required for the application to run. **These should be provided by the security team for testing purposes:**

#### Application Configuration
```bash
APP_ENV=production|staging|development
APP_PORT=8080
ALLOWED_ORIGINS=https://thumsup.com,https://www.thumsup.com
SWAGGER_HOST=<hostname>:8080
```

#### Database Configuration
```bash
DB_HOST=<postgres-host>
DB_PORT=5432
DB_USER=<database-user>
DB_PASSWORD=<database-password>
DB_NAME=<database-name>
DB_SSL_MODE=require|disable|verify-full
DATABASE_SSL=true|false
DATABASE_SSL_REJECT_UNAUTHORIZED=true|false
```

#### JWT Configuration
```bash
JWT_SECRET_KEY=<32-character-minimum-secret-key>
JWT_ACCESS_TOKEN_EXPIRY=3600
JWT_REFRESH_TOKEN_EXPIRY=2592000
```

#### Infobip Configuration (SMS/WhatsApp)
```bash
INFOBIP_BASE_URL=https://api.infobip.com
INFOBIP_API_KEY=<infobip-api-key>
INFOBIP_WA_NUMBER=<whatsapp-number>
```

#### Google Cloud Platform Configuration
```bash
GCP_BUCKET_NAME=<gcs-bucket-name>
GCP_PROJECT_ID=ai0016084-tja-thunderzone-test
GCP_URL=https://storage.googleapis.com
GOOGLE_PUBSUB_PROJECT_ID=<pubsub-project-id>
GOOGLE_PUBSUB_SUBSCRIPTION_ID=<subscription-id>
GOOGLE_PUBSUB_TOPIC_ID=<topic-id>
FIREBASE_SERVICE_KEY_PATH=<path-to-firebase-key.json>
```

#### Admin API Key
```bash
X_API_KEY=<admin-api-key-for-admin-endpoints>
```

### Service Account Credentials

For GCP services, the application uses:
- **Service Account:** `tccc-tja-test-cloudrun-sa@ai0016084-tja-thunderzone-test.iam.gserviceaccount.com`
- **Authentication:** Workload Identity (no explicit credentials file needed in Cloud Run)

### Test Credentials (Development Only)

⚠️ **Note:** These are for development/testing only. Production credentials should be provided separately.

#### Database (Development)
- **Host:** `localhost` or `postgres` (Docker)
- **Port:** `5432`
- **User:** `postgres` (or as configured)
- **Password:** As per local setup
- **Database:** `thums_up_db`

#### Test User Accounts
Sample test users can be created via the signup endpoint:
```json
POST /backend/api/v1/auth/signup
{
  "phone_number": "9876543210",
  "name": "Test User",
  "email": "test@example.com"
}
```

#### Admin API Key (Test)
- **Header:** `X-API-Key: <test-admin-key>`
- **Note:** Should be provided by development team

### Access Methods

1. **API Access:**
   - Public endpoints: No authentication required
   - User endpoints: Bearer token (JWT) in `Authorization` header
   - Admin endpoints: `X-API-Key` header

2. **Database Access:**
   - Direct access: Out of scope
   - Via API: All endpoints are in scope

3. **GCS Access:**
   - Direct bucket access: Out of scope
   - Signed URLs: In scope (generated by application)

---

## Workflow Documentation

### User Registration and Authentication Flow

1. **Send OTP**
   ```
   POST /backend/api/v1/auth/send-otp
   Body: { "phone_number": "9876543210" }
   ```
   - Validates phone number (10 digits)
   - Generates 6-digit OTP
   - Stores OTP in database with 5-minute expiry
   - Sends OTP via Infobip (SMS/WhatsApp)
   - Rate limited: 3 requests per phone per hour

2. **Verify OTP (Existing User)**
   ```
   POST /backend/api/v1/auth/verify-otp
   Body: { "phone_number": "9876543210", "otp": "123456" }
   ```
   - Validates OTP and expiry
   - Checks max attempts (3)
   - Finds user by phone number
   - Generates JWT access token (1 hour expiry)
   - Generates refresh token (30 days expiry)
   - Returns tokens and user info

3. **Sign Up (New User)**
   ```
   POST /backend/api/v1/auth/signup
   Body: { "phone_number": "9876543210", "name": "John Doe", "email": "john@example.com" }
   ```
   - Validates input
   - Checks for duplicate phone number
   - Validates referral code (if provided)
   - Creates user record
   - Generates unique referral code
   - Returns JWT tokens

4. **Refresh Token**
   ```
   POST /backend/api/v1/auth/refresh
   Body: { "refresh_token": "<token>" }
   ```
   - Validates refresh token
   - Checks if revoked or expired
   - Generates new access and refresh tokens
   - Revokes old refresh token

### Contest Participation Flow

1. **Get Active Contest Week**
   ```
   GET /backend/api/v1/contest-weeks/active
   ```
   - Returns current active contest week with date range

2. **Get Active Questions**
   ```
   GET /backend/api/v1/questions/active
   ```
   - Returns all active questions for current week

3. **Submit Answer**
   ```
   POST /backend/api/v1/thunder-seat
   Authorization: Bearer <access_token>
   Body: { "question_id": 1, "answer": "My answer text" }
   ```
   - Validates user authentication
   - Validates question exists and is active
   - Checks if submission is within active week date range
   - Prevents duplicate submissions (user + question + week)
   - Creates submission record

4. **View Submissions**
   ```
   GET /backend/api/v1/thunder-seat/submissions
   Authorization: Bearer <access_token>
   ```
   - Returns all user's submissions

### Winner Selection Flow (Admin)

1. **Select Winners**
   ```
   POST /backend/api/v1/admin/winners/select
   X-API-Key: <admin-key>
   Body: { "week_number": 3, "number_of_winners": 10 }
   ```
   - Validates API key
   - Validates week is past week (not current/future)
   - Checks if winners already selected
   - Retrieves all submissions for week
   - Randomly selects N winners
   - Creates winner records
   - Sends push notifications to winners (async)

2. **View Winners**
   ```
   GET /backend/api/v1/winners?limit=20&offset=0
   GET /backend/api/v1/winners/week/3
   ```
   - Returns paginated list of winners
   - Can filter by week number

### Profile Management Flow

1. **Get Profile**
   ```
   GET /backend/api/v1/profile
   Authorization: Bearer <access_token>
   ```

2. **Update Profile**
   ```
   PATCH /backend/api/v1/profile
   Authorization: Bearer <access_token>
   Body: { "name": "Updated Name", "email": "new@example.com" }
   ```

3. **Manage Addresses**
   - Add: `POST /backend/api/v1/profile/address`
   - List: `GET /backend/api/v1/profile/address`
   - Update: `PUT /backend/api/v1/profile/address/:addressId`
   - Delete: `DELETE /backend/api/v1/profile/address/:addressId`

### File Upload Flow

1. **Upload Avatar**
   ```
   POST /backend/api/v1/avatars
   Authorization: Bearer <access_token>
   Content-Type: multipart/form-data
   ```
   - Validates file type and size
   - Uploads to Google Cloud Storage
   - Generates signed URL (7 days expiry)
   - Stores metadata in database

---

## Sample Data

### Sample User Data

#### Test User 1
```json
{
  "phone_number": "9876543210",
  "name": "John Doe",
  "email": "john.doe@example.com",
  "referral_code": "ABC12345"
}
```

#### Test User 2
```json
{
  "phone_number": "9876543211",
  "name": "Jane Smith",
  "email": "jane.smith@example.com",
  "referral_code": "XYZ67890"
}
```

### Sample OTP Request
```json
POST /backend/api/v1/auth/send-otp
{
  "phone_number": "9876543210"
}
```

### Sample Address Data
```json
{
  "address1": "123 Main Street",
  "address2": "Apartment 4B",
  "pincode": 400001,
  "state": "Maharashtra",
  "city": "Mumbai",
  "nearest_landmark": "Near City Mall",
  "shipping_mobile": "9876543210",
  "is_default": true
}
```

### Sample Question Submission
```json
POST /backend/api/v1/thunder-seat
Authorization: Bearer <access_token>
{
  "question_id": 1,
  "answer": "Thums Up's unique toofani taste sets it apart from other beverages!"
}
```

### Sample Contest Week
```json
{
  "week_number": 1,
  "start_date": "2024-01-15T00:00:00Z",
  "end_date": "2024-01-21T23:59:59Z",
  "is_active": true,
  "winner_count": 10
}
```

### Sample Winner Selection Request
```json
POST /backend/api/v1/admin/winners/select
X-API-Key: <admin-key>
{
  "week_number": 1,
  "number_of_winners": 10
}
```

### Sample JWT Token Structure
```json
{
  "user_id": "550e8400-e29b-41d4-a716-446655440000",
  "phone_number": "9876543210",
  "exp": 1705843200,
  "iat": 1705839600
}
```

---

## CodeQL Scan Results

### Scan Configuration
- **Tool:** GitHub CodeQL
- **Language:** Go
- **Scan Date:** [To be updated after scan]
- **Repository:** [Repository URL]

### Scan Results Summary
- **Total Issues Found:** [Pending]
- **Critical:** [Pending]
- **High:** [Pending]
- **Medium:** [Pending]
- **Low:** [Pending]

### Detailed Results
[Results will be added after CodeQL scan is performed]

**Reference:** https://github.com/github/codeql-action

### Remediation Status
- [ ] Critical issues resolved
- [ ] High issues resolved
- [ ] Medium issues reviewed
- [ ] Low issues documented

---

## Dependabot Scan Results

### Scan Configuration
- **Tool:** GitHub Dependabot
- **Scan Date:** [To be updated after scan]
- **Package Manager:** Go Modules

### Dependency Summary
- **Total Dependencies:** 23 direct, ~150+ transitive
- **Vulnerable Dependencies:** [Pending]
- **Outdated Dependencies:** [Pending]

### Key Dependencies
```
cloud.google.com/go/pubsub v1.49.0
firebase.google.com/go/v4 v4.16.1
github.com/gin-gonic/gin v1.10.1
github.com/golang-jwt/jwt/v5 v5.2.2
gorm.io/gorm v1.30.0
gorm.io/driver/postgres v1.5.11
```

### Vulnerability Details
[Results will be added after Dependabot scan is performed]

**Reference:** https://docs.github.com/en/code-security/dependabot

### Remediation Plan
- [ ] Review all high/critical vulnerabilities
- [ ] Update vulnerable dependencies
- [ ] Test after updates
- [ ] Document breaking changes

---

## Code Scanning Results

### Scan Configuration
- **Tool:** GitHub Advanced Security / CodeQL
- **Scan Date:** [To be updated after scan]
- **Repository:** [Repository URL]

### Scan Results Summary
- **Total Findings:** [Pending]
- **Critical:** [Pending]
- **High:** [Pending]
- **Medium:** [Pending]
- **Low/Info:** [Pending]

### Security Categories Analyzed
- SQL Injection
- Cross-Site Scripting (XSS)
- Authentication/Authorization Issues
- Insecure Deserialization
- Insecure Direct Object References
- Security Misconfiguration
- Sensitive Data Exposure
- Insufficient Logging & Monitoring

### Detailed Findings
[Results will be added after code scanning is performed]

**Reference:** https://docs.github.com/en/code-security/code-scanning

### Remediation Status
- [ ] All critical issues addressed
- [ ] High priority issues reviewed
- [ ] Medium priority issues documented
- [ ] Security team approval obtained

---

## Secret Scanning Results

### Scan Configuration
- **Tool:** GitHub Secret Scanning / GitGuardian
- **Scan Date:** [To be updated after scan]
- **Repository:** [Repository URL]

### Scan Results Summary
- **Secrets Detected:** [Pending]
- **High Risk:** [Pending]
- **Medium Risk:** [Pending]
- **Low Risk:** [Pending]

### Secret Types Scanned
- API Keys
- Database Credentials
- JWT Secrets
- OAuth Tokens
- Private Keys
- Service Account Keys
- Cloud Provider Credentials

### Detailed Findings
[Results will be added after secret scanning is performed]

**Note:** The following files are known to contain configuration but should not contain actual secrets:
- `config/config.go` - Configuration structure (no actual secrets)
- `.env.example` - Example environment variables (no actual secrets)
- `cloudbuild-main.yaml` - References to Secret Manager (no actual secrets)

### Remediation Actions
- [ ] All exposed secrets rotated
- [ ] Secrets moved to Secret Manager
- [ ] Git history cleaned (if needed)
- [ ] Access logs reviewed

**Reference:** https://docs.github.com/en/code-security/secret-scanning

---

## DAST/StackHawk Scan Results

### Scan Configuration
- **Tool:** StackHawk DAST Scanner
- **Scan Date:** [To be updated after scan]
- **Target URL:** [Application URL]
- **Scan Type:** Full Application Scan
- **Authentication:** JWT Bearer Token

### Scan Results Summary
- **Total Vulnerabilities:** [Pending]
- **Critical:** [Pending]
- **High:** [Pending]
- **Medium:** [Pending]
- **Low:** [Pending]
- **Info:** [Pending]

### Scan Coverage
- **Endpoints Tested:** [Pending]
- **Authentication Flows Tested:** [Pending]
- **Authorization Tests:** [Pending]
- **Input Validation Tests:** [Pending]

### Vulnerability Categories
- OWASP Top 10 (2021)
- API Security
- Authentication & Session Management
- Authorization & Access Control
- Input Validation
- Error Handling
- Cryptography
- Security Headers

### Detailed Findings
[Results will be added after StackHawk scan is performed]

**StackHawk Dashboard:** https://auth.stackhawk.com/login

### Remediation Status
- [ ] Critical vulnerabilities addressed
- [ ] High priority issues fixed
- [ ] Medium priority issues reviewed
- [ ] Security headers configured
- [ ] Rate limiting verified

### Scan Configuration File
```yaml
# stackhawk.yml (example)
app:
  applicationId: thums-up-backend
  env: Production
  host: https://<application-url>
  openApiSpec: ./docs/swagger.yaml
authentication:
  type: Bearer
  token: <test-jwt-token>
```

---

## Environment Configuration

### Development Environment
```bash
APP_ENV=development
APP_PORT=8080
ALLOWED_ORIGINS=*
SWAGGER_HOST=localhost:8080
```

### Staging Environment
```bash
APP_ENV=staging
APP_PORT=8080
ALLOWED_ORIGINS=https://staging.thumsup.com
SWAGGER_HOST=<staging-host>:8080
```

### Production Environment
```bash
APP_ENV=production
APP_PORT=8080
ALLOWED_ORIGINS=https://thumsup.com,https://www.thumsup.com
SWAGGER_HOST=<production-host>:8080
```

### Security Headers Configuration
The application should implement the following security headers:
- `X-Content-Type-Options: nosniff`
- `X-Frame-Options: DENY`
- `X-XSS-Protection: 1; mode=block`
- `Strict-Transport-Security: max-age=31536000; includeSubDomains`
- `Content-Security-Policy: default-src 'self'`

---

## API Endpoints Reference

### Complete Endpoint List

#### Public Endpoints (No Authentication)
- `GET /health`
- `GET /swagger/*`
- `POST /backend/api/v1/auth/send-otp`
- `POST /backend/api/v1/auth/verify-otp`
- `POST /backend/api/v1/auth/signup`
- `POST /backend/api/v1/auth/refresh`
- `GET /backend/api/v1/questions/active`
- `GET /backend/api/v1/thunder-seat/current-week`
- `GET /backend/api/v1/contest-weeks`
- `GET /backend/api/v1/contest-weeks/active`
- `GET /backend/api/v1/contest-weeks/:weekNumber`
- `GET /backend/api/v1/winners`
- `GET /backend/api/v1/winners/week/:weekNumber`
- `GET /backend/api/v1/states`
- `GET /backend/api/v1/website-status`

#### Authenticated Endpoints (Bearer Token Required)
- `GET /backend/api/v1/auth/login-count`
- `GET /backend/api/v1/profile`
- `PATCH /backend/api/v1/profile`
- `POST /backend/api/v1/profile/address`
- `GET /backend/api/v1/profile/address`
- `PUT /backend/api/v1/profile/address/:addressId`
- `DELETE /backend/api/v1/profile/address/:addressId`
- `GET /backend/api/v1/profile/questions`
- `POST /backend/api/v1/profile/questions/text`
- `POST /backend/api/v1/profile/questions`
- `POST /backend/api/v1/profile/questions/create`
- `POST /backend/api/v1/questions`
- `GET /backend/api/v1/thunder-seat/submissions`
- `POST /backend/api/v1/thunder-seat`
- `POST /backend/api/v1/contest-weeks`
- `POST /backend/api/v1/contest-weeks/activate`
- `GET /backend/api/v1/avatars`
- `GET /backend/api/v1/avatars/:avatarId`
- `POST /backend/api/v1/avatars`
- `GET /backend/api/v1/winners/status`
- `POST /backend/api/v1/winners/mark-viewed`
- `POST /backend/api/v1/winners/kyc`

#### Admin Endpoints (X-API-Key Required)
- `POST /backend/api/v1/admin/winners/select`

---

## Third-Party Integrations

### Infobip (SMS/WhatsApp)
- **Purpose:** OTP delivery
- **Base URL:** Configurable via `INFOBIP_BASE_URL`
- **Authentication:** API Key in `Authorization: App <key>` header
- **Endpoint:** `POST /sms/2/text/advanced`
- **Out of Scope:** Direct API testing

### Firebase Cloud Messaging
- **Purpose:** Push notifications
- **Service:** Google Firebase
- **Authentication:** Service account key
- **Out of Scope:** Direct API testing

### Google Cloud Storage
- **Purpose:** File storage (avatars, documents)
- **Bucket:** Configurable via `GCP_BUCKET_NAME`
- **Access:** Signed URLs (in scope)
- **Direct Access:** Out of scope

### Google Cloud Pub/Sub
- **Purpose:** Message queuing
- **Service:** Google Cloud Pub/Sub
- **Out of Scope:** Direct API testing

### PostgreSQL Database
- **Purpose:** Data persistence
- **Direct Access:** Out of scope
- **Via API:** All endpoints in scope

---

## Security Testing Checklist

### Authentication & Authorization
- [ ] OTP generation and validation
- [ ] JWT token generation and validation
- [ ] Token expiry handling
- [ ] Refresh token mechanism
- [ ] API key validation for admin endpoints
- [ ] User authorization checks
- [ ] Session management

### Input Validation
- [ ] Phone number validation
- [ ] Email validation
- [ ] File upload validation
- [ ] SQL injection prevention
- [ ] XSS prevention
- [ ] Command injection prevention
- [ ] Path traversal prevention

### Data Protection
- [ ] Sensitive data encryption
- [ ] Password/OTP hashing
- [ ] PII protection
- [ ] Data in transit (HTTPS)
- [ ] Data at rest encryption

### API Security
- [ ] Rate limiting
- [ ] CORS configuration
- [ ] Security headers
- [ ] Error message handling
- [ ] API versioning

### Infrastructure Security
- [ ] Container security
- [ ] Secret management
- [ ] Network security
- [ ] Logging and monitoring

---

## Contact Information

### Development Team
- **Repository:** [GitHub Repository URL]
- **Team:** Thums Up Backend Team

### Security Team
- **Contact:** [Security Team Contact]
- **Reference:** https://wiki.coke.com/confluence/display/infomix/DevSecOps+Security+Offering

---

## Appendix

### A. Swagger/OpenAPI Specification
- **Location:** `/docs/swagger.yaml` and `/docs/swagger.json`
- **Access:** `GET /swagger/index.html`

### B. Architecture Documentation
- **Location:** `/docs/HLD_LLD.md`
- **Contains:** System architecture, database schema, API flows

### C. Deployment Documentation
- **Cloud Build:** `cloudbuild-main.yaml`, `cloudbuild-prod-main.yaml`
- **Docker:** `Dockerfile.dev`, `Dockerfile.subscriber`

### D. Database Migrations
- **Location:** `/migrations/`
- **Format:** SQL migration files

---

**Document Status:** Ready for Security Assessment  
**Next Steps:** 
1. Perform CodeQL scan
2. Run Dependabot scan
3. Execute Code scanning
4. Run Secret scanning
5. Perform DAST/StackHawk scan
6. Update this document with results

---

*This document should be updated with actual scan results before final submission to the security team.*

