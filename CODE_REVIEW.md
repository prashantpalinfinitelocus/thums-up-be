# 🔍 Comprehensive Go Backend Code Review

**Branch:** `staging`  
**Reviewed By:** Senior Principal Go Engineer  
**Date:** December 12, 2025

---

## Executive Summary

| Category | Critical | High | Medium | Low |
|----------|----------|------|--------|-----|
| **Concurrency & Safety** | 3 | 2 | 1 | - |
| **Error Handling** | 4 | 3 | 2 | - |
| **Database & Performance** | 2 | 4 | 3 | - |
| **Architecture** | 1 | 2 | 4 | - |
| **Total** | **10** | **11** | **10** | - |

---

## 🚨 Critical Issues (Potential Panics/Bugs)

### 1. [handlers/notify_me_handler.go:60] Goroutine Leak with Cancelled Context

```go
go h.notificationService.PublishNotifyMeMessage(c.Request.Context(), req.PhoneNumber, req.Email)
```

**Risk:** The goroutine uses `c.Request.Context()` which is cancelled when the HTTP response is sent. The PubSub publish (`result.Get(ctx)`) will fail with context cancellation, and notifications will be silently lost.

**Fix:**
```go
bgCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
go func() {
    defer cancel()
    if err := h.notificationService.PublishNotifyMeMessage(bgCtx, req.PhoneNumber, req.Email); err != nil {
        log.WithError(err).Error("Failed to publish notify me message")
    }
}()
```

---

### 2. [handlers/thunder_seat_handler.go:27-28] Unchecked Context.Get - Potential Panic

```go
user, _ := c.Get("user")
userID := user.(*entities.User).ID  // PANIC if user is nil!
```

**Risk:** If middleware fails or is misconfigured, `user` will be `nil`, causing a nil pointer dereference panic.

**Fix:**
```go
user, exists := c.Get("user")
if !exists {
    c.JSON(http.StatusUnauthorized, dtos.ErrorResponse{Success: false, Error: "User not authenticated"})
    return
}
userEntity, ok := user.(*entities.User)
if !ok || userEntity == nil {
    c.JSON(http.StatusInternalServerError, dtos.ErrorResponse{Success: false, Error: "Invalid user context"})
    return
}
userID := userEntity.ID
```

---

### 3. [handlers/profile_handler.go:31] Unchecked Type Assertion - Potential Panic

```go
userEntity := user.(*entities.User)
```

**Risk:** Same as above - no safety check on type assertion.

---

### 4. [middlewares/auth_middleware.go:71] uuid.MustParse - Potential Panic

```go
user, err := userRepo.FindById(context.Background(), db, uuid.MustParse(claims.UserID))
```

**Risk:** `MustParse` panics on invalid UUID. If JWT token is tampered with, this causes a server crash.

**Fix:**
```go
userUUID, err := uuid.Parse(claims.UserID)
if err != nil {
    c.JSON(http.StatusUnauthorized, gin.H{"success": false, "error": "Invalid user ID in token"})
    c.Abort()
    return
}
user, err := userRepo.FindById(c.Request.Context(), db, userUUID)
```

---

### 5. [services/user_service.go] uuid.MustParse Throughout Service Layer

Multiple occurrences of `uuid.MustParse(userID)` that can panic:
- Line 61, 83, 137, 242, 363, 516

**Risk:** If any caller passes an invalid UUID string, the server panics.

---

### 6. [services/notify_me_service.go:40-41] Silently Ignored Error + Debug Statement

```go
existing, _ := s.notifyMeRepo.FindByPhoneNumber(ctx, s.txnManager.GetDB(), req.PhoneNumber)
fmt.Println("existing", existing)  // Debug in production!
```

**Risk:** 
1. Database errors are silently ignored, potentially causing duplicate subscriptions
2. Debug print statement leaks sensitive data to stdout

---

### 7. [services/auth_service.go:127] Same Pattern - Ignored Errors

```go
existing, _ := s.userRepo.FindByPhoneNumber(ctx, s.txnManager.GetDB(), req.PhoneNumber)
```

**Risk:** Database connection failures are silently ignored.

---

### 8. [services/thunder_seat_service.go:50] Same Pattern

```go
existing, _ := s.thunderSeatRepo.CheckUserSubmission(ctx, s.txnManager.GetDB(), userID, req.QuestionID)
```

---

### 9. [cmd/subscriber/main.go:51] context.Background() Never Cancelled

```go
ctx := context.Background()
// ...
go func() {
    if err := subscriber.Subscribe(ctx, cfg.PubSubConfig.SubscriptionID, messageHandler); err != nil {
        log.Fatalf("Failed to start subscriber: %v", err)
    }
}()
```

**Risk:** The context is never cancelled on shutdown, so graceful shutdown doesn't work properly. The `log.Fatalf` in a goroutine also exits the entire process without cleanup.

**Fix:**
```go
ctx, cancel := context.WithCancel(context.Background())
defer cancel()

// In signal handler:
quit := make(chan os.Signal, 1)
signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
<-quit
cancel() // Cancel context before cleanup
```

---

### 10. [services/notify_me_service.go:83] Incorrect Error Comparison

```go
if err == gorm.ErrRecordNotFound {
```

**Risk:** Direct equality comparison fails when errors are wrapped. Should use `errors.Is()`.

---

## ⚠️ High Severity Issues

### 1. [Multiple Handlers] Type Assertion Instead of errors.As

Throughout handlers, the pattern:
```go
if appErr, ok := err.(*errors.AppError); ok {
```

**Issue:** Will not work with wrapped errors. Per Go 1.13+, use `errors.As`.

**Affected Files:**
- `auth_handler.go`: Lines 38, 74, 109, 144
- `notify_me_handler.go`: Lines 44, 89, 113
- `question_handler.go`: Lines 48, 73
- `thunder_seat_handler.go`: Lines 43, 77, 101
- `winner_handler.go`: Lines 40, 75, 110

---

### 2. [errors/errors.go] AppError Missing Unwrap Method

```go
type AppError struct {
    StatusCode int
    Message    string
    Err        error
}
```

**Issue:** Without `Unwrap()`, `errors.Is` and `errors.As` cannot traverse the error chain.

**Fix:**
```go
func (e *AppError) Unwrap() error {
    return e.Err
}
```

---

### 3. [handlers/address_handler.go] Error Matching via strings.Contains

```go
switch {
case strings.Contains(err.Error(), "state"):
case strings.Contains(err.Error(), "city"):
// ...
}
```

**Risk:** Fragile error handling. If error message wording changes, logic breaks. Use sentinel errors or typed errors instead.

---

### 4. [handlers/profile_handler.go:67-74] Same Issue

```go
if err.Error() == "user not found" {
if err.Error() == "email already in use" {
```

---

### 5. [utils/utils.go:16, 35] Deprecated rand.Seed

```go
rand.Seed(time.Now().UnixNano())
```

**Issue:** `rand.Seed` is deprecated in Go 1.20+. Also, calling it multiple times per function is wasteful and creates predictable seeds.

**Fix:**
```go
// At package level, use crypto/rand for security-sensitive operations
import "crypto/rand"
import "math/big"

func GenerateOTP(length int) string {
    const digits = "0123456789"
    otp := make([]byte, length)
    for i := range otp {
        n, _ := rand.Int(rand.Reader, big.NewInt(int64(len(digits))))
        otp[i] = digits[n.Int64()]
    }
    return string(otp)
}
```

---

## ⚡ Database & Performance Issues

### 1. [repository/notify_me_repository.go:35-40] Unbounded Query - OOM Risk

```go
func (r *notifyMeRepository) FindUnnotified(ctx context.Context, db *gorm.DB) ([]entities.NotifyMe, error) {
    var records []entities.NotifyMe
    if err := db.WithContext(ctx).Where("is_notified = ?", false).Find(&records).Error; err != nil {
```

**Risk:** No `LIMIT`. If there are millions of unnotified records, this causes OOM.

**Fix:** Add pagination:
```go
func (r *notifyMeRepository) FindUnnotified(ctx context.Context, db *gorm.DB, limit, offset int) ([]entities.NotifyMe, error)
```

---

### 2. [repository/question_repository.go:26-31] Same Issue

```go
func (r *questionRepository) FindActiveQuestions(...)
```

No pagination - could return unlimited results.

---

### 3. [services/user_service.go:304-311] N+1 Update Loop

```go
for _, addr := range existingAddresses {
    if err := s.addressRepo.UpdateFields(ctx, tx, addr.ID, map[string]interface{}{
        "is_default": false,
    }); err != nil {
```

**Risk:** One UPDATE per address. Should be batch update.

**Fix:**
```go
if err := tx.Model(&entities.Address{}).
    Where("user_id = ? AND is_default = ? AND is_deleted = ?", userID, true, false).
    Update("is_default", false).Error; err != nil {
    return err
}
```

---

### 4. [services/winner_service.go:60-74] N+1 Insert Loop

```go
for _, entry := range randomEntries {
    winner := &entities.ThunderSeatWinner{...}
    if err := s.winnerRepo.Create(ctx, tx, winner); err != nil {
```

**Fix:** Use batch insert:
```go
if err := tx.Create(&winners).Error; err != nil {
    return err
}
```

---

### 5. [services/user_service.go] Read-Only Operations in Transactions

```go
func (s *userService) GetUser(ctx context.Context, userID string) (*entities.User, error) {
    tx, err := s.txnManager.StartTxn()
    // ...
    user, err := s.userRepo.FindById(ctx, tx, uuid.MustParse(userID))
```

**Issue:** Read-only operations wrapped in transactions add overhead.

---

### 6. [entities/notify_me.go:15] Missing Index for Queried Field

```go
IsNotified  bool      `gorm:"default:false" json:"is_notified"`
```

**Issue:** `FindUnnotified` queries on `is_notified = false` but no index exists.

**Fix:**
```go
IsNotified  bool      `gorm:"default:false;index:idx_notify_me_is_notified" json:"is_notified"`
```

---

### 7. [services/notify_me_service.go:39-49] TOCTOU Race Condition

```go
existing, _ := s.notifyMeRepo.FindByPhoneNumber(...)
if existing != nil {
    return existing, false, nil
}
// Gap here - another request could insert!
s.notifyMeRepo.Create(ctx, tx, notifyMe)
```

**Risk:** Race condition between check and create. Can cause duplicate records if two requests arrive simultaneously.

**Fix:** Use `ON CONFLICT DO NOTHING` (upsert) or wrap in transaction with `SELECT FOR UPDATE`.

---

## 📐 Architecture & Design Issues

### 1. [services/*] Interface Definition at Implementation Site

All interfaces are defined in the services package where they're implemented:
```go
// services/auth_service.go
type AuthService interface { ... }
type authService struct { ... }
```

**Issue:** Go best practice is to define interfaces at the consumer site ("accept interfaces, return structs"). This creates tight coupling.

**Recommendation:** Define interfaces in handlers package or a separate `contracts/` package.

---

### 2. [cmd/server.go:29-35] Package-Level Variables

```go
var (
    db             *gorm.DB
    firebaseClient *vendors.FirebaseClient
    infobipClient  *vendors.InfobipClient
    pubsubClient   interface{}
    gcsService     utils.GCSService
)
```

**Issue:** Global mutable state makes testing difficult and creates hidden dependencies.

**Fix:** Use dependency injection container or pass dependencies explicitly.

---

### 3. [middlewares/auth_middleware.go:71] Uses context.Background()

```go
user, err := userRepo.FindById(context.Background(), db, uuid.MustParse(claims.UserID))
```

**Issue:** Should use `c.Request.Context()` to respect request cancellation and timeouts.

---

### 4. [repository/location_repository.go:30-40] Unused db Field

```go
type stateRepository struct {
    db *gorm.DB  // Never used!
}

func NewStateRepository() StateRepository {
    return &stateRepository{}  // db is nil
}
```

**Issue:** The `db` field is never set or used - all methods receive `db` as a parameter anyway.

---

### 5. [dtos/profile_dto.go:42] Leaking Entity in DTO

```go
type ProfileResponseDTO struct {
    User entities.User `json:"user"`  // Exposes internal entity!
}
```

**Issue:** DTOs should not expose internal entities. This couples API response to database schema.

---

### 6. [config/config.go:74] Ignored Error from godotenv

```go
godotenv.Load()  // Error ignored
```

**Fix:**
```go
if err := godotenv.Load(); err != nil {
    log.Warn("No .env file found, using environment variables")
}
```

---

## 🔒 Security Issues

### 1. [middlewares/cors_middleware.go:9] Allow All Origins

```go
c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
```

**Risk:** In production, this allows any website to make authenticated requests.

---

### 2. [utils/utils.go] OTP Generation Uses math/rand

**Risk:** `math/rand` is not cryptographically secure. OTPs are predictable.

---

### 3. [vendors/infobip.go:66] Response Body Read Error Ignored

```go
body, _ := io.ReadAll(resp.Body)
```

---

## 📋 Summary by File

| File | Issues Found |
|------|--------------|
| `handlers/notify_me_handler.go` | Goroutine leak, type assertion |
| `handlers/thunder_seat_handler.go` | Unchecked context.Get |
| `handlers/profile_handler.go` | Unchecked type assertion, string error matching |
| `handlers/address_handler.go` | String error matching |
| `services/notify_me_service.go` | Ignored error, debug print, TOCTOU race, wrong error comparison |
| `services/auth_service.go` | Ignored errors, deprecated rand |
| `services/user_service.go` | Multiple uuid.MustParse, N+1 queries, unnecessary transactions |
| `services/winner_service.go` | N+1 inserts |
| `services/thunder_seat_service.go` | Ignored error |
| `middlewares/auth_middleware.go` | uuid.MustParse panic, wrong context |
| `repository/notify_me_repository.go` | Unbounded query, no affected rows check |
| `repository/location_repository.go` | Unused struct field |
| `utils/utils.go` | Deprecated rand.Seed, insecure OTP |
| `errors/errors.go` | Missing Unwrap |
| `cmd/subscriber/main.go` | Context never cancelled |
| `cmd/server.go` | Global variables |

---

## 📌 Priority Action Items

### P0 (Fix Immediately)
- [ ] Fix goroutine context leak in `notify_me_handler.go`
- [ ] Add nil checks for all `c.Get("user")` calls
- [ ] Replace `uuid.MustParse` with `uuid.Parse` + error handling
- [ ] Remove debug `fmt.Println` statements
- [ ] Fix ignored errors in check-then-insert patterns

### P1 (Fix This Sprint)
- [ ] Add `Unwrap()` to `AppError`
- [ ] Use `errors.As` instead of type assertion
- [ ] Use crypto/rand for OTP generation
- [ ] Add pagination to unbounded queries
- [ ] Fix TOCTOU race conditions with proper locking/upsert

### P2 (Technical Debt)
- [ ] Refactor interfaces to consumer side
- [ ] Remove global variables, use DI
- [ ] Replace string error matching with typed errors
- [ ] Add missing database indexes
- [ ] Batch N+1 queries

