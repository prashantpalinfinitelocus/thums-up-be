# Thums Up Backend - Sequence Diagrams

## Complete System Flow Diagrams

```mermaid
sequenceDiagram
    actor User as User (Mobile/Web App)
    participant Frontend as Frontend
    participant Middleware as Gin Middleware
    participant Handler as Handler Layer
    participant Service as Service Layer
    participant TxnMgr as Transaction Manager
    participant Repository as Repository (GORM)
    participant DB as PostgreSQL
    participant Worker as Worker Pool
    participant Infobip as Infobip API
    participant Firebase as Firebase FCM
    participant GCS as Google Cloud Storage

    Note over User, GCS: 1. OTP Onboarding Flow - Send OTP

    User->>Frontend: Enter phone number
    Frontend->>Middleware: POST /api/v1/auth/send-otp
    Note right of Frontend: {phone_number: "9876543210"}
    
    Middleware->>Middleware: CORS + Logger + Recovery
    Middleware->>Handler: Route to AuthHandler.SendOTP()
    
    Handler->>Handler: c.ShouldBindJSON(&SendOTPRequest)
    Handler->>Handler: Validate: required, 10 digits, numeric
    
    alt Validation fails
        Handler-->>Frontend: 400 Bad Request
        Note right of Handler: {success: false, error: "Validation failed", details: {...}}
        Frontend-->>User: Show validation errors
    end
    
    Handler->>Service: authService.SendOTP(ctx, phoneNumber)
    
    Service->>Service: Validate phone format
    Service->>Service: Check rate limit (3/hour)
    
    alt Rate limit exceeded
        Service-->>Handler: AppError{429, "Too many requests"}
        Handler-->>Frontend: 429 Too Many Requests
        Frontend-->>User: "Please try after 60 seconds"
    end
    
    Service->>Service: Generate 6-digit random OTP
    Service->>Service: Set expiry: time.Now() + 5 minutes
    
    Service->>Repository: otpRepo.Create(OTPLog)
    Repository->>DB: INSERT INTO otp_logs
    Note right of DB: phone_number, otp, expires_at, is_verified=false, attempts=0
    DB-->>Repository: OTP record created
    Repository-->>Service: Success
    
    Service->>Worker: Queue Task: SendOTPTask
    Note right of Worker: {phone_number, otp, template}
    
    Worker->>Infobip: POST /sms/2/text/advanced
    Note right of Infobip: {"from": "ThumsUp", "to": phone, "text": "Your OTP: ..."}
    Infobip-->>Worker: 200 OK {messageId, status}
    Worker->>Worker: Log delivery status
    
    Service-->>Handler: nil (success)
    Handler->>Handler: Wrap in SuccessResponse
    Handler-->>Frontend: 200 OK
    Note right of Handler: {success: true, message: "OTP sent successfully", data: null}
    Frontend-->>User: "OTP sent to your phone"

    Note over User, GCS: 2. OTP Verification Flow

    User->>Frontend: Enter OTP code
    Frontend->>Middleware: POST /api/v1/auth/verify-otp
    Note right of Frontend: {phone_number: "9876543210", otp: "123456"}
    
    Middleware->>Handler: Route to AuthHandler.VerifyOTP()
    Handler->>Handler: c.ShouldBindJSON(&VerifyOTPRequest)
    Handler->>Handler: Validate: phone (10 digits), otp (6 digits)
    
    Handler->>Service: authService.VerifyOTP(ctx, phone, otp)
    
    Service->>Repository: otpRepo.FindLatestNonVerified(phone)
    Repository->>DB: SELECT * FROM otp_logs WHERE phone_number=? AND is_verified=false ORDER BY created_at DESC LIMIT 1
    DB-->>Repository: Latest OTP record
    Repository-->>Service: OTPLog entity
    
    Service->>Service: Check if OTP expired (expires_at < now)
    alt OTP expired
        Service-->>Handler: AppError{401, "OTP expired"}
        Handler-->>Frontend: 401 Unauthorized
        Frontend-->>User: "OTP has expired, request new one"
    end
    
    Service->>Service: Compare OTP values
    alt OTP mismatch
        Service->>Repository: IncrementAttempts(otpID)
        Repository->>DB: UPDATE otp_logs SET attempts = attempts + 1
        
        Service->>Service: Check attempts >= 3
        alt Max attempts exceeded
            Service->>Repository: MarkOTPInvalid(otpID)
            Service-->>Handler: AppError{401, "Max attempts exceeded"}
        else
            Service-->>Handler: AppError{401, "Invalid OTP"}
        end
        Handler-->>Frontend: 401 Unauthorized
        Frontend-->>User: "Invalid OTP, try again"
    end
    
    Service->>Repository: userRepo.FindByPhoneNumber(phone)
    Repository->>DB: SELECT * FROM users WHERE phone_number=? AND is_active=true
    DB-->>Repository: User record or nil
    Repository-->>Service: User entity or error
    
    alt User not found
        Service-->>Handler: AppError{400, "User not found. Please complete signup"}
        Handler-->>Frontend: 400 Bad Request
        Frontend-->>User: Redirect to signup screen
    end
    
    Service->>Repository: otpRepo.MarkAsVerified(otpID)
    Repository->>DB: UPDATE otp_logs SET is_verified=true, verified_at=NOW()
    DB-->>Repository: Success
    
    Service->>Service: Generate JWT Access Token
    Note right of Service: Claims: {user_id, phone, exp: now+1h, iat: now}
    Service->>Service: Sign with JWT secret
    
    Service->>Service: Generate JWT Refresh Token
    Note right of Service: Claims: {token_id, user_id, exp: now+30d, iat: now}
    
    Service->>Repository: refreshTokenRepo.Create(RefreshToken)
    Repository->>DB: INSERT INTO refresh_tokens
    Note right of DB: user_id, token, expires_at, is_revoked=false
    DB-->>Repository: Token record created
    
    Service->>Service: Build TokenResponse DTO
    Service-->>Handler: TokenResponse{access_token, refresh_token, expires_in, user_id, name, phone}
    
    Handler->>Handler: Wrap in SuccessResponse
    Handler-->>Frontend: 200 OK
    Note right of Handler: {success: true, data: {access_token, refresh_token, ...}}
    Frontend->>Frontend: Store tokens (localStorage/secure storage)
    Frontend-->>User: Navigate to home screen

    Note over User, GCS: 3. User Signup Flow

    User->>Frontend: Fill signup form
    Frontend->>Middleware: POST /api/v1/auth/signup
    Note right of Frontend: {phone_number, name, email?, referral_code?, device_token?}
    
    Middleware->>Handler: Route to AuthHandler.SignUp()
    Handler->>Handler: c.ShouldBindJSON(&SignUpRequest)
    Handler->>Handler: Validate: phone (required, 10 digits), name (required), email (optional, valid format)
    
    Handler->>Service: authService.SignUp(ctx, SignUpRequest)
    
    Service->>TxnMgr: WithTransaction(ctx, func(tx))
    TxnMgr->>TxnMgr: db.Begin()
    
    Service->>Repository: userRepo.FindByPhoneNumber(tx, phone)
    Repository->>DB: SELECT * FROM users WHERE phone_number=?
    
    alt User already exists
        TxnMgr->>TxnMgr: tx.Rollback()
        Service-->>Handler: AppError{400, "User already exists"}
        Handler-->>Frontend: 400 Bad Request
        Frontend-->>User: "Phone number already registered"
    end
    
    alt Referral code provided
        Service->>Repository: userRepo.FindByReferralCode(tx, referralCode)
        Repository->>DB: SELECT * FROM users WHERE referral_code=? AND is_active=true
        
        alt Invalid referral code
            TxnMgr->>TxnMgr: tx.Rollback()
            Service-->>Handler: AppError{400, "Invalid referral code"}
            Handler-->>Frontend: 400 Bad Request
            Frontend-->>User: "Referral code not found"
        end
    end
    
    Service->>Service: Generate unique referral code (8 chars alphanumeric)
    Service->>Service: Create User entity
    Note right of Service: id: UUID, phone, name, email, referral_code, referred_by, device_token, is_verified=false, is_active=true
    
    Service->>Repository: userRepo.Create(tx, user)
    Repository->>DB: INSERT INTO users
    DB-->>Repository: User created with ID
    Repository-->>Service: User entity
    
    Service->>Service: Generate JWT tokens (access + refresh)
    Service->>Repository: refreshTokenRepo.Create(tx, refreshToken)
    Repository->>DB: INSERT INTO refresh_tokens
    DB-->>Repository: Success
    
    TxnMgr->>TxnMgr: tx.Commit()
    TxnMgr-->>Service: Transaction committed
    
    Service->>Service: Build TokenResponse DTO
    Service-->>Handler: TokenResponse
    
    Handler->>Handler: Wrap in SuccessResponse
    Handler-->>Frontend: 201 Created
    Note right of Handler: {success: true, data: {access_token, refresh_token, ...}}
    Frontend->>Frontend: Store tokens
    Frontend-->>User: Welcome screen

    Note over User, GCS: 4. Refresh Token Flow

    Frontend->>Frontend: Detect access token expiry
    Frontend->>Middleware: POST /api/v1/auth/refresh
    Note right of Frontend: {refresh_token: "eyJhbG..."}
    
    Middleware->>Handler: Route to AuthHandler.RefreshToken()
    Handler->>Handler: c.ShouldBindJSON(&RefreshTokenRequest)
    
    Handler->>Service: authService.RefreshToken(ctx, refreshToken)
    
    Service->>Service: Validate JWT signature & decode
    alt Invalid signature
        Service-->>Handler: AppError{401, "Invalid refresh token"}
        Handler-->>Frontend: 401 Unauthorized
        Frontend-->>User: Redirect to login
    end
    
    Service->>Repository: refreshTokenRepo.FindByToken(refreshToken)
    Repository->>DB: SELECT * FROM refresh_tokens WHERE token=?
    DB-->>Repository: RefreshToken record
    Repository-->>Service: RefreshToken entity
    
    Service->>Service: Check is_revoked flag
    alt Token revoked
        Service-->>Handler: AppError{401, "Token has been revoked"}
        Handler-->>Frontend: 401 Unauthorized
        Frontend-->>User: Redirect to login
    end
    
    Service->>Service: Check expires_at < now
    alt Token expired
        Service-->>Handler: AppError{401, "Refresh token expired"}
        Handler-->>Frontend: 401 Unauthorized
        Frontend-->>User: Redirect to login
    end
    
    Service->>Repository: userRepo.FindByID(userID)
    Repository->>DB: SELECT * FROM users WHERE id=? AND is_active=true
    DB-->>Repository: User entity
    
    Service->>TxnMgr: WithTransaction(ctx, func(tx))
    TxnMgr->>TxnMgr: db.Begin()
    
    Service->>Repository: refreshTokenRepo.RevokeToken(tx, oldTokenID)
    Repository->>DB: UPDATE refresh_tokens SET is_revoked=true WHERE id=?
    
    Service->>Service: Generate new JWT tokens
    Service->>Repository: refreshTokenRepo.Create(tx, newRefreshToken)
    Repository->>DB: INSERT INTO refresh_tokens
    
    TxnMgr->>TxnMgr: tx.Commit()
    
    Service-->>Handler: TokenResponse (new tokens)
    Handler-->>Frontend: 200 OK
    Frontend->>Frontend: Update stored tokens
    Frontend-->>User: Continue session

    Note over User, GCS: 5. Get User Profile Flow

    User->>Frontend: Open profile page
    Frontend->>Middleware: GET /api/v1/profile
    Note right of Frontend: Authorization: Bearer <access_token>
    
    Middleware->>Middleware: CORS + Logger + Recovery
    Middleware->>Middleware: AuthMiddleware.Handle()
    
    Middleware->>Middleware: Extract Authorization header
    Middleware->>Middleware: Parse "Bearer <token>"
    
    alt Missing/Invalid token
        Middleware-->>Frontend: 401 Unauthorized
        Frontend-->>User: Redirect to login
    end
    
    Middleware->>Middleware: Validate JWT signature
    Middleware->>Middleware: Check expiry (exp claim)
    
    Middleware->>Repository: userRepo.FindByID(userID from token)
    Repository->>DB: SELECT * FROM users WHERE id=? AND is_active=true AND deleted_at IS NULL
    DB-->>Repository: User entity
    
    Middleware->>Middleware: c.Set("user", userEntity)
    Middleware->>Middleware: c.Set("user_id", userEntity.ID)
    Middleware->>Handler: c.Next() → ProfileHandler.GetProfile()
    
    Handler->>Handler: user := c.Get("user")
    Handler->>Handler: Cast to *entities.User
    
    Handler->>Service: userService.GetUser(ctx, userID)
    Service->>Repository: userRepo.FindByID(userID)
    Repository->>DB: SELECT * FROM users WHERE id=?
    DB-->>Repository: User with all fields
    Repository-->>Service: User entity
    
    Service-->>Handler: User entity
    Handler->>Handler: Map to ProfileResponseDTO
    Note right of Handler: {user: {id, phone_number, name, email, is_active, is_verified, referral_code, referred_by, created_at, updated_at}}
    
    Handler-->>Frontend: 200 OK {ProfileResponseDTO}
    Frontend-->>User: Display profile information

    Note over User, GCS: 6. Update User Profile Flow

    User->>Frontend: Edit name/email and save
    Frontend->>Middleware: PATCH /api/v1/profile
    Note right of Frontend: Authorization: Bearer <token><br/>{name?: "Updated Name", email?: "new@email.com"}
    
    Middleware->>Middleware: AuthMiddleware validates token
    Middleware->>Middleware: Load user, set in context
    Middleware->>Handler: ProfileHandler.UpdateProfile()
    
    Handler->>Handler: user := c.Get("user")
    Handler->>Handler: c.ShouldBindJSON(&UpdateProfileRequestDTO)
    Handler->>Handler: Validate: name (optional), email (optional, valid format)
    
    Handler->>Service: userService.UpdateUser(ctx, userID, UpdateProfileRequestDTO)
    
    Service->>TxnMgr: WithTransaction(ctx, func(tx))
    TxnMgr->>TxnMgr: db.Begin()
    
    Service->>Repository: userRepo.FindByID(tx, userID)
    Repository->>DB: SELECT * FROM users WHERE id=?
    
    alt User not found
        TxnMgr->>TxnMgr: tx.Rollback()
        Service-->>Handler: Error "User not found"
        Handler-->>Frontend: 404 Not Found
    end
    
    alt Email provided and changed
        Service->>Repository: userRepo.FindByEmail(tx, newEmail)
        Repository->>DB: SELECT * FROM users WHERE email=? AND id != ?
        
        alt Email already in use
            TxnMgr->>TxnMgr: tx.Rollback()
            Service-->>Handler: Error "Email already in use"
            Handler-->>Frontend: 400 Bad Request
            Frontend-->>User: "Email is already registered"
        end
    end
    
    Service->>Service: Update user fields (name, email)
    Service->>Repository: userRepo.Update(tx, user)
    Repository->>DB: UPDATE users SET name=?, email=?, updated_at=NOW() WHERE id=?
    DB-->>Repository: Updated user
    
    TxnMgr->>TxnMgr: tx.Commit()
    
    Service-->>Handler: Updated user entity
    Handler-->>Frontend: 200 OK {updated user data}
    Frontend-->>User: "Profile updated successfully"

    Note over User, GCS: 7. Address Management Flow - Get Addresses

    User->>Frontend: View delivery addresses
    Frontend->>Middleware: GET /api/v1/profile/address
    Note right of Frontend: Authorization: Bearer <token>
    
    Middleware->>Middleware: AuthMiddleware validates & loads user
    Middleware->>Handler: AddressHandler.GetAddresses()
    
    Handler->>Handler: user := c.Get("user")
    Handler->>Service: userService.GetUserAddresses(ctx, userID)
    
    Service->>Repository: addressRepo.FindByUserID(userID)
    Repository->>DB: SELECT a.*, s.name as state, c.name as city FROM address a JOIN state s JOIN city c WHERE user_id=? AND is_deleted=false AND is_active=true
    DB-->>Repository: Address list with joined data
    
    Service->>Service: Map to AddressResponseDTO[]
    Service-->>Handler: Address list
    Handler-->>Frontend: 200 OK [AddressResponseDTO, ...]
    Frontend-->>User: Display address list

    Note over User, GCS: 8. Address Management Flow - Add Address

    User->>Frontend: Fill new address form
    Frontend->>Middleware: POST /api/v1/profile/address
    Note right of Frontend: Authorization: Bearer <token><br/>{address1, address2?, pincode, state, city, nearest_landmark?, shipping_mobile?, is_default}
    
    Middleware->>Middleware: AuthMiddleware validates
    Middleware->>Handler: AddressHandler.AddAddress()
    
    Handler->>Handler: user := c.Get("user")
    Handler->>Handler: c.ShouldBindJSON(&AddressRequestDTO)
    Handler->>Handler: Validate: address1 (required), pincode (required, 6 digits), state (required), city (required)
    
    Handler->>Service: userService.AddUserAddress(ctx, userID, AddressRequestDTO)
    
    Service->>TxnMgr: WithTransaction(ctx, func(tx))
    TxnMgr->>TxnMgr: db.Begin()
    
    Service->>Repository: userRepo.FindByID(tx, userID)
    Repository->>DB: SELECT * FROM users WHERE id=?
    
    alt User not found
        TxnMgr->>TxnMgr: tx.Rollback()
        Service-->>Handler: Error "user not found"
        Handler-->>Frontend: 404 Not Found
    end
    
    Service->>Repository: stateRepo.FindByName(tx, stateName)
    Repository->>DB: SELECT * FROM state WHERE name=? AND is_active=true AND is_deleted=false
    
    alt State not found
        TxnMgr->>TxnMgr: tx.Rollback()
        Service-->>Handler: Error "Invalid state: ... not found"
        Handler-->>Frontend: 400 Bad Request
        Frontend-->>User: "State not found"
    end
    
    Service->>Repository: cityRepo.FindByNameAndState(tx, cityName, stateID)
    Repository->>DB: SELECT * FROM city WHERE name=? AND state_id=? AND is_active=true AND is_deleted=false
    
    alt City not found or mismatch
        TxnMgr->>TxnMgr: tx.Rollback()
        Service-->>Handler: Error "Invalid city: ... not found in state ..."
        Handler-->>Frontend: 400 Bad Request
        Frontend-->>User: "City not found in selected state"
    end
    
    Service->>Repository: pinCodeRepo.FindByPincodeAndCity(tx, pincode, cityID)
    Repository->>DB: SELECT * FROM pin_code WHERE pincode=? AND city_id=? AND is_active=true AND is_deleted=false
    
    alt Pincode not found
        TxnMgr->>TxnMgr: tx.Rollback()
        Service-->>Handler: Error "Invalid pincode: ... not found in city ..."
        Handler-->>Frontend: 400 Bad Request
        Frontend-->>User: "Pincode not found"
    end
    
    Service->>Service: Check is_deliverable flag
    alt Not deliverable
        TxnMgr->>TxnMgr: tx.Rollback()
        Service-->>Handler: Error "Pincode ... is not deliverable"
        Handler-->>Frontend: 400 Bad Request
        Frontend-->>User: "Sorry, we don't deliver to this location yet"
    end
    
    alt is_default = true
        Service->>Repository: addressRepo.UnsetAllDefaults(tx, userID)
        Repository->>DB: UPDATE address SET is_default=false WHERE user_id=?
    end
    
    Service->>Service: Create Address entity
    Note right of Service: user_id, address1, address2, pincode, pin_code_id, city_id, state_id, nearest_landmark, shipping_mobile, is_default, is_active=true, is_deleted=false, created_by=userID
    
    Service->>Repository: addressRepo.Create(tx, address)
    Repository->>DB: INSERT INTO address
    DB-->>Repository: Address created with ID
    
    TxnMgr->>TxnMgr: tx.Commit()
    
    Service->>Service: Map to AddressResponseDTO
    Service-->>Handler: AddressResponseDTO
    Handler-->>Frontend: 201 Created {AddressResponseDTO}
    Frontend-->>User: "Address added successfully"

    Note over User, GCS: 9. Address Management Flow - Update Address

    User->>Frontend: Edit existing address
    Frontend->>Middleware: PUT /api/v1/profile/address/:addressId
    Note right of Frontend: Authorization: Bearer <token><br/>Path: addressId=123<br/>Body: {updated address fields}
    
    Middleware->>Middleware: AuthMiddleware validates
    Middleware->>Handler: AddressHandler.UpdateAddress()
    
    Handler->>Handler: addressID := c.Param("addressId")
    Handler->>Handler: c.ShouldBindJSON(&AddressRequestDTO)
    Handler->>Service: userService.UpdateUserAddress(ctx, userID, addressID, dto)
    
    Service->>TxnMgr: WithTransaction(ctx, func(tx))
    TxnMgr->>TxnMgr: db.Begin()
    
    Service->>Repository: addressRepo.FindByID(tx, addressID)
    Repository->>DB: SELECT * FROM address WHERE id=? AND is_deleted=false
    
    alt Address not found
        TxnMgr->>TxnMgr: tx.Rollback()
        Service-->>Handler: Error "address not found"
        Handler-->>Frontend: 404 Not Found
    end
    
    Service->>Service: Check address.user_id == userID
    alt Ownership mismatch
        TxnMgr->>TxnMgr: tx.Rollback()
        Service-->>Handler: Error "address does not belong to user"
        Handler-->>Frontend: 403 Forbidden
        Frontend-->>User: "Access denied"
    end
    
    Service->>Service: Validate new state/city/pincode (same logic as Add)
    Service->>Service: Update address fields
    Service->>Repository: addressRepo.Update(tx, address)
    Repository->>DB: UPDATE address SET ... , last_modified_by=?, last_modified_on=NOW() WHERE id=?
    
    TxnMgr->>TxnMgr: tx.Commit()
    Service-->>Handler: Updated AddressResponseDTO
    Handler-->>Frontend: 200 OK {AddressResponseDTO}
    Frontend-->>User: "Address updated"

    Note over User, GCS: 10. Address Management Flow - Delete Address

    User->>Frontend: Click delete on address
    Frontend->>Middleware: DELETE /api/v1/profile/address/:addressId
    Note right of Frontend: Authorization: Bearer <token>
    
    Middleware->>Middleware: AuthMiddleware validates
    Middleware->>Handler: AddressHandler.DeleteAddress()
    
    Handler->>Handler: addressID := c.Param("addressId")
    Handler->>Service: userService.DeleteUserAddress(ctx, userID, addressID)
    
    Service->>TxnMgr: WithTransaction(ctx, func(tx))
    TxnMgr->>TxnMgr: db.Begin()
    
    Service->>Repository: addressRepo.FindByID(tx, addressID)
    Repository->>DB: SELECT * FROM address WHERE id=? AND is_deleted=false
    
    alt Ownership check fails
        Service-->>Handler: Error "address does not belong to user"
        Handler-->>Frontend: 403 Forbidden
    end
    
    Service->>Repository: addressRepo.SoftDelete(tx, addressID)
    Repository->>DB: UPDATE address SET is_deleted=true, is_active=false WHERE id=?
    DB-->>Repository: Success
    
    TxnMgr->>TxnMgr: tx.Commit()
    Service-->>Handler: nil (success)
    Handler-->>Frontend: 200 OK {message: "Address deleted successfully"}
    Frontend-->>User: Remove from list

    Note over User, GCS: 11. Thunder Seat Contest - Get Active Questions

    User->>Frontend: Open Thunder Seat page
    Frontend->>Middleware: GET /api/v1/questions/active
    Note right of Frontend: No auth required (public endpoint)
    
    Middleware->>Handler: QuestionHandler.GetActiveQuestions()
    Handler->>Service: questionService.GetActiveQuestions(ctx)
    
    Service->>Repository: questionRepo.FindActive()
    Repository->>DB: SELECT * FROM question_master WHERE is_active=true AND is_deleted=false ORDER BY created_on DESC
    DB-->>Repository: Question list
    
    Service->>Service: Map to QuestionResponse[]
    Service-->>Handler: Question list
    Handler->>Handler: Wrap in SuccessResponse
    Handler-->>Frontend: 200 OK {success: true, data: [questions]}
    Frontend-->>User: Display active questions

    Note over User, GCS: 12. Thunder Seat Contest - Get Current Week

    Frontend->>Middleware: GET /api/v1/thunder-seat/current-week
    Note right of Frontend: No auth required
    
    Middleware->>Handler: ThunderSeatHandler.GetCurrentWeek()
    Handler->>Service: thunderSeatService.GetCurrentWeek(ctx)
    
    Service->>Service: Calculate week info from campaign start
    Note right of Service: campaignStart = 2024-01-01<br/>weekNumber = int((now - campaignStart) / 7 days) + 1<br/>startDate = campaignStart + (weekNumber-1)*7 days<br/>endDate = startDate + 6 days + 23:59:59
    
    Service->>Service: Build CurrentWeekResponse
    Service-->>Handler: CurrentWeekResponse{week_number, start_date, end_date}
    Handler->>Handler: Wrap in SuccessResponse
    Handler-->>Frontend: 200 OK {success: true, data: {...}}
    Frontend-->>User: Display "Week 3: Jan 15 - Jan 21"

    Note over User, GCS: 13. Thunder Seat Contest - Submit Answer

    User->>Frontend: Write answer and submit
    Frontend->>Middleware: POST /api/v1/thunder-seat
    Note right of Frontend: Authorization: Bearer <token><br/>{week_number: 3, question_id: 1, answer: "My answer..."}
    
    Middleware->>Middleware: AuthMiddleware validates token
    Middleware->>Middleware: Load user entity
    Middleware->>Middleware: c.Set("user", userEntity)
    Middleware->>Handler: ThunderSeatHandler.SubmitAnswer()
    
    Handler->>Handler: user := c.Get("user").(*entities.User)
    Handler->>Handler: c.ShouldBindJSON(&ThunderSeatSubmitRequest)
    Handler->>Handler: Validate: week_number (required), question_id (required), answer (required, max 1000 chars)
    
    alt Validation fails
        Handler-->>Frontend: 400 Bad Request {success: false, error: "Validation failed", details: {...}}
        Frontend-->>User: Show field errors
    end
    
    Handler->>Service: thunderSeatService.SubmitAnswer(ctx, request, userID)
    
    Service->>TxnMgr: WithTransaction(ctx, func(tx))
    TxnMgr->>TxnMgr: db.Begin()
    
    Service->>Repository: questionRepo.FindByID(tx, questionID)
    Repository->>DB: SELECT * FROM question_master WHERE id=? AND is_active=true AND is_deleted=false
    
    alt Question not found or inactive
        TxnMgr->>TxnMgr: tx.Rollback()
        Service-->>Handler: AppError{400, "Question not found or inactive"}
        Handler-->>Frontend: 400 Bad Request
        Frontend-->>User: "Question is no longer active"
    end
    
    Service->>Service: Validate week_number
    Service->>Service: currentWeek = calculateCurrentWeek()
    
    alt week_number > currentWeek
        TxnMgr->>TxnMgr: tx.Rollback()
        Service-->>Handler: AppError{400, "Contest for this week has not started"}
        Handler-->>Frontend: 400 Bad Request
    end
    
    Service->>Repository: thunderSeatRepo.CheckDuplicate(tx, userID, questionID, weekNumber)
    Repository->>DB: SELECT COUNT(*) FROM thunder_seat WHERE user_id=? AND question_id=? AND week_number=?
    DB-->>Repository: Count = 1 or 0
    
    alt Duplicate submission
        TxnMgr->>TxnMgr: tx.Rollback()
        Service-->>Handler: AppError{400, "You have already submitted for this question this week"}
        Handler-->>Frontend: 400 Bad Request
        Frontend-->>User: "Already submitted"
    end
    
    Service->>Service: Create ThunderSeat entity
    Note right of Service: user_id, question_id, week_number, answer, created_by=userID, created_on=NOW()
    
    Service->>Repository: thunderSeatRepo.Create(tx, submission)
    Repository->>DB: INSERT INTO thunder_seat (user_id, question_id, week_number, answer, created_by, created_on) VALUES (...)
    DB-->>Repository: Submission created with ID
    
    TxnMgr->>TxnMgr: tx.Commit()
    TxnMgr-->>Service: Transaction committed
    
    Service->>Service: Map to ThunderSeatResponse
    Service-->>Handler: ThunderSeatResponse
    Handler->>Handler: Wrap in SuccessResponse
    Handler-->>Frontend: 201 Created {success: true, data: {...}, message: "Answer submitted successfully"}
    Frontend-->>User: "Submission successful! Good luck!"

    Note over User, GCS: 14. Thunder Seat Contest - Get User Submissions

    User->>Frontend: View my submissions
    Frontend->>Middleware: GET /api/v1/thunder-seat/submissions
    Note right of Frontend: Authorization: Bearer <token>
    
    Middleware->>Middleware: AuthMiddleware validates
    Middleware->>Handler: ThunderSeatHandler.GetUserSubmissions()
    
    Handler->>Handler: userID := c.Get("user_id").(string)
    Handler->>Service: thunderSeatService.GetUserSubmissions(ctx, userID)
    
    Service->>Repository: thunderSeatRepo.FindByUserID(userID)
    Repository->>DB: SELECT * FROM thunder_seat WHERE user_id=? ORDER BY created_on DESC
    DB-->>Repository: Submission list
    
    Service->>Service: Map to ThunderSeatResponse[]
    Service-->>Handler: Submission list
    Handler->>Handler: Wrap in SuccessResponse
    Handler-->>Frontend: 200 OK {success: true, data: [...]}
    Frontend-->>User: Display submission history

    Note over User, GCS: 15. Winner Management - Get All Winners (Paginated)

    User->>Frontend: View winners page
    Frontend->>Middleware: GET /api/v1/winners?limit=20&offset=0
    Note right of Frontend: No auth required
    
    Middleware->>Handler: WinnerHandler.GetAllWinners()
    Handler->>Handler: Bind query params: AllWinnersRequest{limit, offset}
    Handler->>Handler: Validate: limit (1-100), offset (>=0)
    
    Handler->>Service: winnerService.GetAllWinners(ctx, limit, offset)
    
    Service->>Repository: winnerRepo.FindAllPaginated(limit, offset)
    Repository->>DB: SELECT * FROM thunder_seat_winner ORDER BY created_on DESC LIMIT ? OFFSET ?
    DB-->>Repository: Winner records
    
    Service->>Repository: winnerRepo.CountTotal()
    Repository->>DB: SELECT COUNT(*) FROM thunder_seat_winner
    DB-->>Repository: total = 95
    
    Service->>Service: Calculate pagination metadata
    Note right of Service: totalPages = ceil(95 / 20) = 5<br/>currentPage = (0 / 20) + 1 = 1
    
    Service->>Service: Map to WinnerResponse[]
    Service-->>Handler: (winners, total)
    
    Handler->>Handler: Build PaginatedResponse
    Handler-->>Frontend: 200 OK
    Note right of Handler: {success: true, data: [...], meta: {page: 1, page_size: 20, total_pages: 5, total_count: 95}}
    Frontend-->>User: Display winner list with pagination

    Note over User, GCS: 16. Winner Management - Get Winners by Week

    Frontend->>Middleware: GET /api/v1/winners/week/3
    Middleware->>Handler: WinnerHandler.GetWinnersByWeek()
    
    Handler->>Handler: weekNumber := c.Param("weekNumber")
    Handler->>Handler: Convert to int, validate
    
    Handler->>Service: winnerService.GetWinnersByWeek(ctx, weekNumber)
    
    Service->>Repository: winnerRepo.FindByWeek(weekNumber)
    Repository->>DB: SELECT * FROM thunder_seat_winner WHERE week_number=? ORDER BY created_on ASC
    DB-->>Repository: Week 3 winners
    
    Service-->>Handler: WinnerResponse[]
    Handler->>Handler: Wrap in SuccessResponse
    Handler-->>Frontend: 200 OK {success: true, data: [...]}
    Frontend-->>User: Display Week 3 winners

    Note over User, GCS: 17. Admin - Select Winners (Random Selection)

    actor Admin as Admin User
    Admin->>Frontend: Access admin panel
    Frontend->>Middleware: POST /api/v1/admin/winners/select
    Note right of Frontend: X-API-Key: <admin_api_key><br/>{week_number: 3, number_of_winners: 10}
    
    Middleware->>Middleware: APIKeyMiddleware.Handle()
    Middleware->>Middleware: apiKey := c.GetHeader("X-API-Key")
    Middleware->>Middleware: Compare with config.XAPIKey
    
    alt Invalid/Missing API Key
        Middleware-->>Frontend: 401 Unauthorized {success: false, error: "Invalid or missing API key"}
        Frontend-->>Admin: Access denied
    end
    
    Middleware->>Handler: WinnerHandler.SelectWinners()
    
    Handler->>Handler: c.ShouldBindJSON(&SelectWinnersRequest)
    Handler->>Handler: Validate: week_number (required), number_of_winners (required, min 1)
    
    Handler->>Service: winnerService.SelectWinners(ctx, request)
    
    Service->>Service: Calculate current week
    Service->>Service: Validate week_number < currentWeek
    
    alt Future or current week
        Service-->>Handler: AppError{400, "Can only select winners for past weeks"}
        Handler-->>Frontend: 400 Bad Request
        Frontend-->>Admin: "Week 3 is still ongoing"
    end
    
    Service->>Repository: winnerRepo.CheckExisting(weekNumber)
    Repository->>DB: SELECT COUNT(*) FROM thunder_seat_winner WHERE week_number=?
    
    alt Winners already selected
        Service-->>Handler: AppError{400, "Winners already selected for week 3"}
        Handler-->>Frontend: 400 Bad Request
        Frontend-->>Admin: "Winners already announced"
    end
    
    Service->>Repository: thunderSeatRepo.FindByWeek(weekNumber)
    Repository->>DB: SELECT * FROM thunder_seat WHERE week_number=?
    DB-->>Repository: All submissions for week
    
    alt No submissions
        Service-->>Handler: AppError{400, "No submissions found for week 3"}
        Handler-->>Frontend: 400 Bad Request
    end
    
    Service->>Service: submissionCount = len(submissions)
    Service->>Service: if number_of_winners > submissionCount: adjust to submissionCount
    
    Service->>Service: Random selection algorithm
    Note right of Service: 1. Shuffle submissions<br/>2. Pick first N submissions<br/>3. selectedSubmissions[:number_of_winners]
    
    Service->>TxnMgr: WithTransaction(ctx, func(tx))
    TxnMgr->>TxnMgr: db.Begin()
    
    loop For each selected submission
        Service->>Service: Create ThunderSeatWinner entity
        Note right of Service: user_id, thunder_seat_id, week_number, created_by="SYSTEM", created_on=NOW()
        
        Service->>Repository: winnerRepo.Create(tx, winner)
        Repository->>DB: INSERT INTO thunder_seat_winner
        DB-->>Repository: Winner record created
    end
    
    TxnMgr->>TxnMgr: tx.Commit()
    TxnMgr-->>Service: All winners inserted
    
    loop For each winner
        Service->>Worker: Queue Task: NotifyWinner
        Note right of Worker: {user_id, week_number, notification_type: "winner_announcement"}
        
        Worker->>Repository: userRepo.FindByID(user_id)
        Repository->>DB: SELECT * FROM users WHERE id=?
        DB-->>Repository: User entity with device_token
        
        alt device_token exists
            Worker->>Firebase: Send FCM push notification
            Note right of Firebase: {to: device_token, notification: {title: "Congratulations!", body: "You won Thunder Seat Week 3!"}, data: {type: "winner", week: 3}}
            Firebase-->>Worker: Message sent successfully
        end
        
        Worker->>Infobip: Send SMS/WhatsApp message
        Note right of Infobip: "Congratulations! You're a Thunder Seat winner for Week 3. Check the app for details."
        Infobip-->>Worker: Delivered
        
        Worker->>Worker: Log notification status
    end
    
    Service->>Service: Map winners to WinnerResponse[]
    Service-->>Handler: Winner list
    Handler->>Handler: Wrap in SuccessResponse
    Handler-->>Frontend: 201 Created {success: true, data: [...], message: "Winners selected successfully"}
    Frontend-->>Admin: "10 winners selected and notified"

    Note over User, GCS: 18. Future Enhancement - Profile Picture Upload (GCS Integration)

    User->>Frontend: Upload profile picture
    Frontend->>Middleware: POST /api/v1/profile/upload-picture
    Note right of Frontend: Authorization: Bearer <token><br/>Content-Type: multipart/form-data<br/>File: image.jpg
    
    Middleware->>Middleware: AuthMiddleware validates
    Middleware->>Handler: ProfileHandler.UploadProfilePicture()
    
    Handler->>Handler: Parse multipart form
    Handler->>Handler: Validate: file size (<5MB), type (jpg/png)
    
    Handler->>Service: userService.UploadProfilePicture(ctx, userID, fileBytes, fileName)
    
    Service->>Service: Generate unique file name
    Note right of Service: fileName = "profiles/{userID}_{timestamp}.jpg"
    
    Service->>GCS: gcsService.Upload(ctx, fileName, fileBytes)
    GCS->>GCS: Validate credentials (service account)
    GCS->>GCS: Open bucket: thums-up-assets
    GCS->>GCS: Write object: profiles/user123_1234567890.jpg
    GCS-->>Service: Upload successful
    
    Service->>GCS: gcsService.GenerateSignedURL(fileName, 7*24*time.Hour)
    Note right of GCS: Signed URL valid for 7 days, allows public read access without auth
    GCS-->>Service: signedURL
    
    Service->>TxnMgr: WithTransaction(ctx, func(tx))
    TxnMgr->>TxnMgr: db.Begin()
    
    Service->>Repository: userRepo.UpdateProfilePicture(tx, userID, signedURL)
    Repository->>DB: UPDATE users SET profile_picture_url=?, updated_at=NOW() WHERE id=?
    DB-->>Repository: Updated
    
    TxnMgr->>TxnMgr: tx.Commit()
    
    Service-->>Handler: {asset_url: signedURL}
    Handler-->>Frontend: 200 OK {success: true, data: {asset_url: "https://storage.googleapis.com/..."}}
    Frontend->>Frontend: Update UI with new picture
    Frontend-->>User: Profile picture updated
```

## Key Architectural Highlights

### 1. **Middleware Chain**
- **CORS**: Handles cross-origin requests
- **Logger**: Logs all requests with structured fields
- **Recovery**: Panic recovery to prevent server crashes
- **ErrorHandler**: Global error formatting
- **AuthMiddleware**: JWT validation + user context injection
- **APIKeyMiddleware**: Admin endpoint protection

### 2. **Error Handling**
- Custom `AppError` type with HTTP status codes
- Consistent error response format: `{success: false, error: string, details?: object}`
- Validation errors formatted and returned with field-level details

### 3. **Transaction Management**
- All write operations wrapped in transactions
- Auto-rollback on errors
- ACID guarantees for multi-step operations

### 4. **Authentication Flow**
- JWT-based with access + refresh tokens
- Access token: 1 hour expiry
- Refresh token: 30 days expiry, stored in DB
- Token rotation on refresh

### 5. **Authorization**
- **Public endpoints**: Questions, Winners (no auth)
- **User endpoints**: Profile, Address, Thunder Seat (JWT auth)
- **Admin endpoints**: Winner selection (X-API-Key header)

### 6. **Async Processing**
- Worker pool for background tasks
- OTP sending via Infobip
- Push notifications via Firebase
- Non-blocking notification delivery

### 7. **Data Validation**
- Request binding with validation tags
- Business logic validation in service layer
- Database constraints (unique, foreign keys)

### 8. **Soft Delete Pattern**
- `is_deleted` flag on entities
- Records retained for audit/recovery
- Queries filter out deleted records



