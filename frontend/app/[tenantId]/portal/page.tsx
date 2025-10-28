'use client';

import { useEffect } from 'react';
import { useRouter, useParams, useSearchParams } from 'next/navigation';

export default function PortalPage() {
  const router = useRouter();
  const params = useParams();
  const searchParams = useSearchParams();

  const tenantId = params.tenantId as string;
  const token = searchParams.get('token');

  useEffect(() => {
    // Check if user already has a valid session token
    const sessionToken = sessionStorage.getItem('portalToken');
    const tokenExpiry = sessionStorage.getItem('tokenExpiry');

    if (sessionToken && tokenExpiry && Date.now() < parseInt(tokenExpiry)) {
      // User has valid session, redirect to dashboard
      router.push(`/${tenantId}/portal/dashboard`);
      return;
    }

    // If magic link token is present, redirect to verification page
    if (token) {
      router.push(`/${tenantId}/portal/verify?token=${token}`);
      return;
    }

    // No token at all - show error
    // This will be handled by the return statement below
  }, [token, tenantId, router]);

  // Show error only if no token and not redirecting
  if (!token) {
    return (
      <div className="min-h-screen bg-gradient-to-br from-blue-50 to-indigo-100 flex items-center justify-center p-4">
        <div className="bg-white rounded-2xl shadow-xl p-8 max-w-md w-full">
          <div className="text-center">
            <div className="mx-auto flex items-center justify-center h-12 w-12 rounded-full bg-red-100 mb-4">
              <svg className="h-6 w-6 text-red-600" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M6 18L18 6M6 6l12 12" />
              </svg>
            </div>
            <h2 className="text-xl font-semibold text-gray-900 mb-2">Invalid Access Link</h2>
            <p className="text-gray-600 mb-6">
              Missing access token. Please check your email for the correct portal link.
            </p>
            <p className="text-sm text-gray-500">
              If you continue to have issues, please contact your tax preparer for assistance.
            </p>
          </div>
        </div>
      </div>
    )
  }

  // Show loading while redirecting
  return (
    <div className="min-h-screen bg-gradient-to-br from-blue-50 to-indigo-100 flex items-center justify-center p-4">
      <div className="bg-white rounded-2xl shadow-xl p-8 max-w-md w-full">
        <div className="text-center">
          <div className="animate-spin rounded-full h-12 w-12 border-b-2 border-blue-600 mx-auto"></div>
          <p className="mt-4 text-gray-600">Redirecting...</p>
        </div>
      </div>
    </div>
  );
}
