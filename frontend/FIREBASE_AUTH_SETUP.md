# WellTaxPro Frontend - Firebase Authentication & Employee Registration Setup

This guide will help you set up Google Sign-In with Firebase authentication and automatic employee registration in your WellTaxPro Next.js application.

## Firebase Setup

1. **Create a Firebase Project**
   - Go to [Firebase Console](https://console.firebase.google.com/)
   - Click "Create a project" or select an existing project
   - Enable Google Analytics (optional)

2. **Enable Authentication**
   - In the Firebase Console, go to "Authentication" > "Sign-in method"
   - Enable "Google" as a sign-in provider
   - Add your domain to authorized domains

3. **Get Configuration Values**
   - Go to Project Settings > General
   - In "Your apps" section, click the web icon (`</>`) to create a web app
   - Copy the configuration object

4. **Configure OAuth Consent Screen**
   - Go to Google Cloud Console
   - Navigate to APIs & Services > OAuth consent screen
   - Configure your app information

## Backend API Setup

Ensure your backend API has the Employee endpoints running at `/api/v1/employees`. The frontend expects:

- `POST /api/v1/employees` - Create new employee after Google signup
- `GET /api/v1/employees/me` - Get current employee info (requires Firebase auth)
- `PUT /api/v1/employees/me` - Update employee info (requires Firebase auth)

## Environment Setup

1. **Create Environment File**
   ```bash
   cp .env.local.example .env.local
   ```

2. **Update Environment Variables**
   Edit `.env.local` with your Firebase configuration:
   ```env
   NEXT_PUBLIC_FIREBASE_API_KEY=your_api_key_here
   NEXT_PUBLIC_FIREBASE_AUTH_DOMAIN=your_project_id.firebaseapp.com
   NEXT_PUBLIC_FIREBASE_PROJECT_ID=your_project_id
   NEXT_PUBLIC_FIREBASE_STORAGE_BUCKET=your_project_id.appspot.com
   NEXT_PUBLIC_FIREBASE_MESSAGING_SENDER_ID=your_sender_id
   NEXT_PUBLIC_FIREBASE_APP_ID=your_app_id
   NEXT_PUBLIC_API_URL=http://localhost:8081
   ```

## Components Overview

### AuthContext (`contexts/AuthContext.tsx`)
- Manages authentication state across the app
- Provides login/logout functions
- Handles user session persistence
- **NEW**: Automatically registers employees after Google OAuth
- **NEW**: Manages employee data state

### EmployeeService (`services/employeeService.ts`)
- **NEW**: Handles API calls to the employee endpoints
- Creates, fetches, and updates employee records
- Provides TypeScript interfaces for employee data

### GoogleSignInButton (`components/GoogleSignInButton.tsx`)
- Reusable Google Sign-In button component
- Includes loading states and error handling
- Styled with Tailwind CSS

### UserMenu (`components/UserMenu.tsx`)
- Displays user profile information
- **NEW**: Shows employee role and name from database
- Dropdown menu with sign-out option
- Shows user avatar and name

### LoginPage (`components/LoginPage.tsx`)
- Full-page login interface
- Branded login experience
- Responsive design

### ProtectedRoute (`components/ProtectedRoute.tsx`)
- Wrapper component for protected pages
- Redirects unauthenticated users to login
- Shows loading states

### EmployeeRegistrationLoading (`components/EmployeeRegistrationLoading.tsx`)
- **NEW**: Loading screen shown during employee registration
- Appears briefly after Google OAuth while creating employee record

## Authentication & Registration Flow

### Complete User Journey
1. User visits the app
2. If not authenticated, sees the login page
3. Clicks "Sign in with Google"
4. Completes Google OAuth flow
5. **NEW**: System automatically checks if employee exists in database
6. **NEW**: If employee doesn't exist, creates new employee record
7. **NEW**: Shows registration loading screen during employee creation
8. User is redirected to main app with full session (Firebase + Employee data)

### Employee Registration Process
```typescript
// Automatic flow after Google OAuth
const authFlow = async (firebaseUser) => {
  try {
    // Try to fetch existing employee
    const employee = await employeeService.getCurrentEmployee(idToken);
    setEmployee(employee);
  } catch (error) {
    // If employee doesn't exist, create one
    const newEmployee = await employeeService.createEmployee({
      firebaseUid: firebaseUser.uid,
      email: firebaseUser.email,
      firstName: firebaseUser.displayName?.split(' ')[0],
      lastName: firebaseUser.displayName?.split(' ').slice(1).join(' '),
      role: 'accountant'
    });
    setEmployee(newEmployee.employee);
  }
};
```

## Usage

### Protecting Pages
Wrap any page that requires authentication:

```tsx
import { ProtectedRoute } from '@/components/ProtectedRoute';

export default function MyProtectedPage() {
  return (
    <ProtectedRoute>
      <YourPageContent />
    </ProtectedRoute>
  );
}
```

### Using Auth Context
Access user data and employee information:

```tsx
import { useAuth } from '@/contexts/AuthContext';

function MyComponent() {
  const { user, employee, loading, employeeLoading, signOut } = useAuth();
  
  if (loading) return <div>Loading authentication...</div>;
  if (employeeLoading) return <div>Setting up your account...</div>;
  
  return (
    <div>
      <p>Welcome, {employee?.firstName || user?.displayName}</p>
      <p>Role: {employee?.role}</p>
      <button onClick={signOut}>Sign Out</button>
    </div>
  );
}
```

### Employee Data Structure
```typescript
interface Employee {
  id: string;
  firebaseUid: string;
  email: string;
  firstName: string | null;
  lastName: string | null;
  role: string; // 'admin', 'accountant', 'support'
  isActive: boolean;
  createdAt: string;
  updatedAt: string;
}
```

## Development

1. **Install Dependencies**
   ```bash
   npm install
   ```

2. **Run Development Server**
   ```bash
   npm run dev
   ```

3. **Build for Production**
   ```bash
   npm run build
   npm start
   ```

## Security Notes

- All Firebase configuration is stored in environment variables
- Public keys (like API keys) are safe to expose in client-side code
- Firebase handles OAuth security and token management
- Employee API endpoints use Firebase ID token validation
- User sessions persist across browser refreshes
- Employee data is automatically synced with Firebase authentication

## Troubleshooting

### Common Issues

1. **"Firebase: Firebase App named '[DEFAULT]' already exists"**
   - Check if Firebase is initialized multiple times
   - Ensure proper conditional initialization in `lib/firebase.ts`

2. **Google Sign-In popup blocked**
   - Check browser popup blocker settings
   - Ensure your domain is added to Firebase authorized domains

3. **OAuth configuration error**
   - Verify OAuth consent screen is configured
   - Check that redirect URLs match your domain

4. **Environment variables not loading**
   - Ensure `.env.local` exists and is properly formatted
   - Restart the development server after changes
   - Variables must start with `NEXT_PUBLIC_` for client-side access

5. **Employee creation fails**
   - Check that backend API is running and accessible
   - Verify `NEXT_PUBLIC_API_URL` is correct
   - Check browser network tab for API errors
   - Ensure backend employee endpoints are implemented

6. **Employee data not showing**
   - Check browser console for API errors
   - Verify Firebase ID token is being sent correctly
   - Check backend logs for authentication issues

## API Integration

### Employee Endpoints Expected
```typescript
// POST /api/v1/employees - Create employee (public)
{
  "firebaseUid": "string",
  "email": "string", 
  "firstName": "string?",
  "lastName": "string?",
  "role": "accountant" | "admin" | "support"
}

// GET /api/v1/employees/me - Get current employee (authenticated)
// Headers: Authorization: Bearer <firebase_id_token>

// PUT /api/v1/employees/me - Update employee (authenticated) 
// Headers: Authorization: Bearer <firebase_id_token>
{
  "firstName": "string?",
  "lastName": "string?"
}
```

## Next Steps

- Configure user roles and permissions in backend
- Add custom claims for tenant-based access
- Implement password reset functionality
- Add social login providers (Facebook, Twitter, etc.)
- Set up Firebase Security Rules for data access
- Add employee role-based UI restrictions
- Implement employee profile editing functionality
