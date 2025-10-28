# Employee API Documentation

## Overview
The Employee API provides endpoints for managing employee records in the WellTaxPro system. This is particularly useful when users sign up with Google OAuth and need to be registered in the local system.

## Endpoints

### 1. Create Employee
**POST** `/api/v1/employees`

Creates a new employee record when a user signs up with Google.

**Request Body:**
```json
{
  "firebaseUid": "firebase_user_uid_here",
  "email": "user@example.com",
  "firstName": "John",
  "lastName": "Doe",
  "role": "accountant"
}
```

**Response (201 Created):**
```json
{
  "success": true,
  "message": "Employee created successfully",
  "employee": {
    "id": "123e4567-e89b-12d3-a456-426614174000",
    "firebaseUid": "firebase_user_uid_here",
    "email": "user@example.com",
    "firstName": "John",
    "lastName": "Doe",
    "role": "accountant",
    "isActive": true,
    "createdAt": "2025-10-05T17:00:00Z",
    "updatedAt": "2025-10-05T17:00:00Z"
  }
}
```

**Response (200 OK - if employee already exists):**
```json
{
  "success": true,
  "message": "Employee already exists",
  "employee": {
    "id": "123e4567-e89b-12d3-a456-426614174000",
    "firebaseUid": "firebase_user_uid_here",
    "email": "user@example.com",
    "firstName": "John",
    "lastName": "Doe",
    "role": "accountant",
    "isActive": true,
    "createdAt": "2025-10-05T17:00:00Z",
    "updatedAt": "2025-10-05T17:00:00Z"
  }
}
```

**Fields:**
- `firebaseUid` (required): The Firebase UID from Google OAuth
- `email` (required): User's email address
- `firstName` (optional): User's first name
- `lastName` (optional): User's last name
- `role` (optional): Employee role - defaults to "accountant". Valid values: "admin", "accountant", "support"
- `tenantIds` (optional): Array of tenant IDs this employee should have access to

### 2. Get Current Employee's Tenant Access
**GET** `/api/v1/employees/me/tenants`

Returns the list of tenants the current authenticated employee has access to.

**Headers:**
```
Authorization: Bearer <firebase_id_token>
```

**Response (200 OK):**
```json
[
  {
    "tenantId": "tenant-123",
    "tenantName": "ABC Accounting Firm",
    "role": "accountant",
    "isActive": true
  },
  {
    "tenantId": "tenant-456", 
    "tenantName": "XYZ Tax Services",
    "role": "admin",
    "isActive": true
  }
]
```

### 3. Get Current Employee
**GET** `/api/v1/employees/me`

Returns the current authenticated employee's information.

**Headers:**
```
Authorization: Bearer <firebase_id_token>
```

**Response (200 OK):**
```json
{
  "id": "123e4567-e89b-12d3-a456-426614174000",
  "firebaseUid": "firebase_user_uid_here",
  "email": "user@example.com",
  "firstName": "John",
  "lastName": "Doe",
  "role": "accountant",
  "isActive": true,
  "createdAt": "2025-10-05T17:00:00Z",
  "updatedAt": "2025-10-05T17:00:00Z"
}
```

### 4. Update Current Employee
**PUT** `/api/v1/employees/me`

Updates the current authenticated employee's information.

**Headers:**
```
Authorization: Bearer <firebase_id_token>
```

**Request Body:**
```json
{
  "firstName": "Jane",
  "lastName": "Smith"
}
```

**Response (200 OK):**
```json
{
  "id": "123e4567-e89b-12d3-a456-426614174000",
  "firebaseUid": "firebase_user_uid_here",
  "email": "user@example.com",
  "firstName": "Jane",
  "lastName": "Smith",
  "role": "accountant",
  "isActive": true,
  "createdAt": "2025-10-05T17:00:00Z",
  "updatedAt": "2025-10-05T17:00:00Z"
}
```

### 5. Assign Employee to Tenant (Admin Only)
**POST** `/api/v1/employees/{employeeId}/tenants`

Assigns an employee access to a specific tenant. Only admin users can perform this action.

**Headers:**
```
Authorization: Bearer <firebase_id_token>
```

**Request Body:**
```json
{
  "employeeId": "123e4567-e89b-12d3-a456-426614174000",
  "tenantId": "tenant-123",
  "role": "accountant"
}
```

**Response (201 Created):**
```json
{
  "success": true,
  "message": "Employee assigned to tenant successfully"
}
```

### 6. Remove Employee from Tenant (Admin Only)
**DELETE** `/api/v1/employees/{employeeId}/tenants/{tenantId}`

Removes an employee's access to a specific tenant. Only admin users can perform this action.

**Headers:**
```
Authorization: Bearer <firebase_id_token>
```

**Response (200 OK):**
```json
{
  "success": true,
  "message": "Employee removed from tenant successfully"
}
```

## Integration Flow

### Google OAuth Signup Flow
1. User signs up with Google OAuth on your frontend
2. Frontend receives Firebase ID token
3. Frontend calls `POST /api/v1/employees` with the Firebase UID and user info
4. Backend creates employee record in local database
5. User can now access protected endpoints using their Firebase ID token

### Example Frontend Integration (JavaScript)
```javascript
// After Google OAuth signup
async function createEmployeeAfterSignup(firebaseUser) {
  const response = await fetch('/api/v1/employees', {
    method: 'POST',
    headers: {
      'Content-Type': 'application/json',
    },
    body: JSON.stringify({
      firebaseUid: firebaseUser.uid,
      email: firebaseUser.email,
      firstName: firebaseUser.displayName?.split(' ')[0],
      lastName: firebaseUser.displayName?.split(' ')[1],
      role: 'accountant'
    })
  });

  const result = await response.json();
  
  if (result.success) {
    console.log('Employee created:', result.employee);
    // Redirect to dashboard or show success message
  } else {
    console.error('Failed to create employee');
  }
}

// To get current employee info
async function getCurrentEmployee(idToken) {
  const response = await fetch('/api/v1/employees/me', {
    headers: {
      'Authorization': `Bearer ${idToken}`
    }
  });

  const employee = await response.json();
  return employee;
}
```

## Error Responses

**400 Bad Request:**
```json
{
  "error": "Firebase UID is required"
}
```

**401 Unauthorized:**
```json
{
  "error": "Unauthorized: Invalid token"
}
```

**500 Internal Server Error:**
```json
{
  "error": "Failed to create employee"
}
```

## Security Notes

- The `POST /api/v1/employees` endpoint is public to allow user registration
- The `GET` and `PUT` endpoints require Firebase authentication
- Employee roles are validated server-side
- All authenticated endpoints use Firebase ID token validation
