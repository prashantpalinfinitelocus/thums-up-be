package docs

const SwaggerInfo = `{
    "swagger": "2.0",
    "info": {
        "title": "Thums Up Backend API",
        "description": "API documentation for Thums Up Backend Service",
        "version": "1.0"
    },
    "basePath": "/api/v1",
    "paths": {
        "/profile": {
            "get": {
                "tags": ["Profile"],
                "summary": "Get user profile",
                "description": "Retrieve authenticated user's profile information including avatar",
                "security": [{"Bearer": []}],
                "responses": {
                    "200": {
                        "description": "User profile retrieved successfully",
                        "schema": {"$ref": "#/definitions/ProfileResponse"}
                    },
                    "401": {"description": "Unauthorized"},
                    "404": {"description": "User not found"},
                    "500": {"description": "Internal server error"}
                }
            },
            "patch": {
                "tags": ["Profile"],
                "summary": "Update user profile",
                "description": "Update authenticated user's profile information including name, email, and avatar",
                "security": [{"Bearer": []}],
                "parameters": [{
                    "name": "body",
                    "in": "body",
                    "required": true,
                    "schema": {"$ref": "#/definitions/UpdateProfileRequest"}
                }],
                "responses": {
                    "200": {
                        "description": "Profile updated successfully",
                        "schema": {"$ref": "#/definitions/User"}
                    },
                    "400": {"description": "Invalid request body or email already in use"},
                    "401": {"description": "Unauthorized"},
                    "404": {"description": "User not found"},
                    "500": {"description": "Internal server error"}
                }
            }
        },
        "/profile/address": {
            "get": {
                "tags": ["Address"],
                "summary": "Get user addresses",
                "description": "Retrieve all addresses for authenticated user",
                "security": [{"Bearer": []}],
                "responses": {
                    "200": {
                        "description": "Addresses retrieved successfully",
                        "schema": {
                            "type": "array",
                            "items": {"$ref": "#/definitions/AddressResponse"}
                        }
                    },
                    "401": {"description": "Unauthorized"},
                    "500": {"description": "Internal server error"}
                }
            },
            "post": {
                "tags": ["Address"],
                "summary": "Add user address",
                "description": "Add a new address for authenticated user",
                "security": [{"Bearer": []}],
                "parameters": [{
                    "name": "body",
                    "in": "body",
                    "required": true,
                    "schema": {"$ref": "#/definitions/AddressRequest"}
                }],
                "responses": {
                    "201": {
                        "description": "Address created successfully",
                        "schema": {"$ref": "#/definitions/AddressResponse"}
                    },
                    "400": {"description": "Invalid request body or pincode not deliverable"},
                    "401": {"description": "Unauthorized"},
                    "500": {"description": "Internal server error"}
                }
            }
        },
        "/profile/address/{addressId}": {
            "put": {
                "tags": ["Address"],
                "summary": "Update user address",
                "description": "Update an existing address for authenticated user",
                "security": [{"Bearer": []}],
                "parameters": [
                    {
                        "name": "addressId",
                        "in": "path",
                        "required": true,
                        "type": "integer"
                    },
                    {
                        "name": "body",
                        "in": "body",
                        "required": true,
                        "schema": {"$ref": "#/definitions/AddressRequest"}
                    }
                ],
                "responses": {
                    "200": {
                        "description": "Address updated successfully",
                        "schema": {"$ref": "#/definitions/AddressResponse"}
                    },
                    "400": {"description": "Invalid request or address does not belong to user"},
                    "401": {"description": "Unauthorized"},
                    "404": {"description": "Address not found"},
                    "500": {"description": "Internal server error"}
                }
            },
            "delete": {
                "tags": ["Address"],
                "summary": "Delete user address",
                "description": "Delete an address for authenticated user",
                "security": [{"Bearer": []}],
                "parameters": [{
                    "name": "addressId",
                    "in": "path",
                    "required": true,
                    "type": "integer"
                }],
                "responses": {
                    "200": {"description": "Address deleted successfully"},
                    "400": {"description": "Invalid address ID or address does not belong to user"},
                    "401": {"description": "Unauthorized"},
                    "404": {"description": "Address not found"},
                    "500": {"description": "Internal server error"}
                }
            }
        },
        "/avatars": {
            "get": {
                "tags": ["Avatar"],
                "summary": "Get all avatars",
                "description": "Retrieve all avatars, optionally filtered by published status",
                "parameters": [{
                    "name": "is_published",
                    "in": "query",
                    "type": "boolean",
                    "required": false,
                    "description": "Filter by published status"
                }],
                "responses": {
                    "200": {
                        "description": "Avatars retrieved successfully",
                        "schema": {
                            "type": "object",
                            "properties": {
                                "avatars": {
                                    "type": "array",
                                    "items": {"$ref": "#/definitions/AvatarResponse"}
                                }
                            }
                        }
                    },
                    "400": {"description": "Invalid query parameter"},
                    "500": {"description": "Internal server error"}
                }
            },
            "post": {
                "tags": ["Avatar"],
                "summary": "Create avatar",
                "description": "Create a new avatar (requires authentication)",
                "security": [{"Bearer": []}],
                "parameters": [{
                    "name": "body",
                    "in": "body",
                    "required": true,
                    "schema": {"$ref": "#/definitions/CreateAvatarRequest"}
                }],
                "responses": {
                    "201": {
                        "description": "Avatar created successfully",
                        "schema": {"$ref": "#/definitions/AvatarResponse"}
                    },
                    "400": {"description": "Invalid request body"},
                    "401": {"description": "Unauthorized"},
                    "500": {"description": "Internal server error"}
                }
            }
        },
        "/avatars/{avatarId}": {
            "get": {
                "tags": ["Avatar"],
                "summary": "Get avatar by ID",
                "description": "Retrieve a specific avatar by ID",
                "parameters": [{
                    "name": "avatarId",
                    "in": "path",
                    "required": true,
                    "type": "integer"
                }],
                "responses": {
                    "200": {
                        "description": "Avatar retrieved successfully",
                        "schema": {"$ref": "#/definitions/AvatarResponse"}
                    },
                    "400": {"description": "Invalid avatar ID"},
                    "404": {"description": "Avatar not found"},
                    "500": {"description": "Internal server error"}
                }
            }
        }
    },
    "definitions": {
        "User": {
            "type": "object",
            "properties": {
                "id": {"type": "string"},
                "phone_number": {"type": "string"},
                "name": {"type": "string"},
                "email": {"type": "string"},
                "avatar_id": {"type": "integer"},
                "is_active": {"type": "boolean"},
                "is_verified": {"type": "boolean"},
                "referral_code": {"type": "string"},
                "referred_by": {"type": "string"},
                "device_token": {"type": "string"},
                "created_at": {"type": "string", "format": "date-time"},
                "updated_at": {"type": "string", "format": "date-time"}
            }
        },
        "UpdateProfileRequest": {
            "type": "object",
            "properties": {
                "name": {"type": "string"},
                "email": {"type": "string"},
                "avatar_id": {"type": "integer"}
            }
        },
        "UserProfile": {
            "type": "object",
            "properties": {
                "id": {"type": "string"},
                "phone_number": {"type": "string"},
                "name": {"type": "string"},
                "email": {"type": "string"},
                "avatar_image": {"type": "string", "description": "Public URL of the avatar image from GCS"},
                "is_active": {"type": "boolean"},
                "is_verified": {"type": "boolean"},
                "referral_code": {"type": "string"},
                "referred_by": {"type": "string"},
                "created_at": {"type": "string", "format": "date-time"},
                "updated_at": {"type": "string", "format": "date-time"}
            }
        },
        "ProfileResponse": {
            "type": "object",
            "properties": {
                "user": {"$ref": "#/definitions/UserProfile"}
            }
        },
        "AddressRequest": {
            "type": "object",
            "required": ["address1", "pincode", "state", "city"],
            "properties": {
                "address1": {"type": "string"},
                "address2": {"type": "string"},
                "pincode": {"type": "integer"},
                "state": {"type": "string"},
                "city": {"type": "string"},
                "nearest_landmark": {"type": "string"},
                "shipping_mobile": {"type": "string"},
                "is_default": {"type": "boolean"}
            }
        },
        "AddressResponse": {
            "type": "object",
            "properties": {
                "id": {"type": "integer"},
                "address1": {"type": "string"},
                "address2": {"type": "string"},
                "pincode": {"type": "integer"},
                "state": {"type": "string"},
                "city": {"type": "string"},
                "nearest_landmark": {"type": "string"},
                "shipping_mobile": {"type": "string"},
                "is_default": {"type": "boolean"},
                "is_active": {"type": "boolean"},
                "created_on": {"type": "string", "format": "date-time"},
                "last_modified_on": {"type": "string", "format": "date-time"}
            }
        },
        "CreateAvatarRequest": {
            "type": "object",
            "required": ["name", "image_key"],
            "properties": {
                "name": {"type": "string"},
                "image_key": {"type": "string", "description": "GCS object key/path for the uploaded avatar image"},
                "is_published": {"type": "boolean"}
            }
        },
        "AvatarResponse": {
            "type": "object",
            "properties": {
                "id": {"type": "integer"},
                "name": {"type": "string"},
                "image_url": {"type": "string", "description": "Public URL of the avatar image from GCS"},
                "is_published": {"type": "boolean"},
                "published_by": {"type": "string"},
                "published_on": {"type": "string", "format": "date-time"},
                "is_active": {"type": "boolean"},
                "created_on": {"type": "string", "format": "date-time"},
                "last_modified_on": {"type": "string", "format": "date-time"}
            }
        }
    },
    "securityDefinitions": {
        "Bearer": {
            "type": "apiKey",
            "name": "Authorization",
            "in": "header",
            "description": "Enter your bearer token in the format: Bearer {token}"
        }
    }
}`
