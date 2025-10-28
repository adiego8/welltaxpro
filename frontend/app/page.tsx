'use client';

import { useEffect, useState } from 'react';
import Link from 'next/link';
import { useRouter } from 'next/navigation';
import { useAuth } from '@/contexts/AuthContext';
import { LoginPage } from '@/components/LoginPage';
import { UserMenu } from '@/components/UserMenu';
import { EmployeeRegistrationLoading } from '@/components/EmployeeRegistrationLoading';
import { EnvironmentCheck } from '@/components/EnvironmentCheck';

interface TenantAccess {
  tenantId: string;
  tenantName: string;
  role: string;
  isActive: boolean;
}

export default function Home() {
  const router = useRouter();
  const { user, employee, loading, employeeLoading } = useAuth();
  const [tenants, setTenants] = useState<TenantAccess[]>([]);
  const [tenantsLoading, setTenantsLoading] = useState(true);

  // Fetch employee's tenant access
  useEffect(() => {
    async function fetchTenants() {
      if (!user || !employee) {
        setTenantsLoading(false);
        return;
      }

      try {
        const apiUrl = process.env.NEXT_PUBLIC_API_URL || 'http://localhost:8081';
        const idToken = await user.getIdToken();

        const response = await fetch(`${apiUrl}/api/v1/employees/me/tenants`, {
          headers: {
            'Authorization': `Bearer ${idToken}`,
            'Content-Type': 'application/json',
          },
        });

        if (!response.ok) {
          throw new Error('Failed to fetch tenants');
        }

        const data = await response.json();
        const activeTenants = (data || []).filter((t: TenantAccess) => t.isActive);
        setTenants(activeTenants);

        // If no tenants and user is admin, redirect to tenant setup
        if (activeTenants.length === 0 && employee.role === 'admin') {
          router.push('/admin/tenants/new');
        }
      } catch (err) {
        console.error('Failed to fetch tenants:', err);
      } finally {
        setTenantsLoading(false);
      }
    }

    fetchTenants();
  }, [user, employee, router]);

  if (loading) {
    return (
      <div className="min-h-screen bg-gray-50 flex items-center justify-center">
        <div className="text-center">
          <div className="w-8 h-8 border-2 border-gray-300 border-t-blue-600 rounded-full animate-spin mx-auto mb-4" />
          <p className="text-gray-600">Loading...</p>
        </div>
      </div>
    );
  }

  // Show a setup message if Firebase is not configured
  if (typeof window !== 'undefined' && (!process.env.NEXT_PUBLIC_FIREBASE_API_KEY || !process.env.NEXT_PUBLIC_FIREBASE_PROJECT_ID)) {
    return (
      <div className="min-h-screen bg-gray-50 flex items-center justify-center">
        <div className="max-w-md mx-auto text-center p-6 bg-white rounded-lg shadow-lg">
          <h2 className="text-2xl font-bold text-gray-900 mb-4">Setup Required</h2>
          <p className="text-gray-600 mb-4">
            Please configure your Firebase environment variables to enable authentication.
          </p>
          <p className="text-sm text-gray-500">
            See FIREBASE_AUTH_SETUP.md for detailed instructions.
          </p>
        </div>
      </div>
    );
  }

  if (!user) {
    return <LoginPage />;
  }

  // Show employee registration loading if user is signed in but employee is being created
  if (user && employeeLoading) {
    return <EmployeeRegistrationLoading />;
  }

  // Show loading while fetching tenants
  if (tenantsLoading) {
    return (
      <div className="min-h-screen bg-gray-50 flex items-center justify-center">
        <div className="text-center">
          <div className="w-8 h-8 border-2 border-gray-300 border-t-blue-600 rounded-full animate-spin mx-auto mb-4" />
          <p className="text-gray-600">Loading tenant access...</p>
        </div>
      </div>
    );
  }

  // If no tenants and not admin, show message
  if (tenants.length === 0 && employee?.role !== 'admin') {
    return (
      <div className="min-h-screen bg-gray-50 flex items-center justify-center">
        <div className="max-w-md mx-auto text-center p-6 bg-white rounded-lg shadow-lg">
          <h2 className="text-2xl font-bold text-gray-900 mb-4">No Tenant Access</h2>
          <p className="text-gray-600 mb-4">
            You don't have access to any tenants yet. Please contact your administrator.
          </p>
        </div>
      </div>
    );
  }

  // Use first tenant as default (or show tenant selector if multiple)
  const defaultTenant = tenants[0];

  return (
    <div className="min-h-screen bg-gradient-to-br from-slate-50 via-blue-50 to-indigo-50">
      {/* Clean Navigation */}
      <nav className="bg-white/80 backdrop-blur-sm border-b border-gray-200">
        <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
          <div className="flex justify-between h-16 items-center">
            <div className="flex-shrink-0">
              <Link href="/">
                <img src="/logo.png" alt="WellTaxPro" className="h-12 sm:h-16 cursor-pointer" />
              </Link>
            </div>
            <div className="flex items-center gap-4">
              <Link
                href="/"
                className="text-sm font-semibold text-blue-600 hover:text-blue-700"
              >
                Home
              </Link>
              {employee?.role === 'admin' && (
                <Link
                  href="/admin/tenants"
                  className="text-sm text-gray-600 hover:text-gray-900"
                >
                  Account Management
                </Link>
              )}
              <UserMenu />
            </div>
          </div>
        </div>
      </nav>

      <main className="max-w-7xl mx-auto py-8 sm:px-6 lg:px-8">
        <div className="px-4 sm:px-0">
          {/* Hero Section */}
          <div className="text-center mb-12">
            <div className="inline-flex items-center justify-center w-20 h-20 rounded-full bg-gradient-to-br from-blue-500 to-indigo-600 text-white text-2xl font-bold mb-4">
              {employee?.firstName?.charAt(0)}{employee?.lastName?.charAt(0)}
            </div>
            <h1 className="text-4xl sm:text-5xl font-bold text-gray-900 mb-3">
              Welcome back, {employee?.firstName || user.displayName?.split(' ')[0] || 'there'}!
            </h1>
            <p className="text-lg text-gray-600">
              {employee?.role && (
                <span className="inline-flex items-center px-3 py-1 rounded-full text-sm font-medium bg-blue-100 text-blue-800 capitalize">
                  {employee.role}
                </span>
              )}
            </p>
          </div>

          {/* Tenant Selector (if multiple) */}
          {tenants.length > 1 && (
            <div className="mb-8 bg-white rounded-xl shadow-sm border border-gray-200 p-6">
              <h3 className="text-lg font-semibold text-gray-900 mb-4">Your Accounts</h3>
              <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 gap-4">
                {tenants.map((tenant) => (
                  <Link
                    key={tenant.tenantId}
                    href={`/${tenant.tenantId}/clients`}
                    className="block p-4 border-2 border-gray-200 rounded-lg hover:border-blue-500 hover:shadow-md transition-all"
                  >
                    <h4 className="font-semibold text-gray-900">{tenant.tenantName}</h4>
                    <p className="text-sm text-gray-600 mt-1 capitalize">{tenant.role}</p>
                  </Link>
                ))}
              </div>
            </div>
          )}

          {/* Quick Actions - Navigation Cards */}
          {defaultTenant && (
            <div>
              <h2 className="text-2xl font-bold text-gray-900 mb-6">Quick Access</h2>
              <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 gap-6">
                {/* Clients Card */}
                <Link
                  href={`/${defaultTenant.tenantId}/clients`}
                  className="group bg-white rounded-xl shadow-sm border border-gray-200 hover:shadow-lg hover:border-blue-300 transition-all overflow-hidden"
                >
                  <div className="p-6">
                    <div className="flex items-center justify-between mb-4">
                      <div className="w-12 h-12 rounded-lg bg-blue-100 flex items-center justify-center group-hover:bg-blue-200 transition-colors">
                        <svg className="w-6 h-6 text-blue-600" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                          <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M17 20h5v-2a3 3 0 00-5.356-1.857M17 20H7m10 0v-2c0-.656-.126-1.283-.356-1.857M7 20H2v-2a3 3 0 015.356-1.857M7 20v-2c0-.656.126-1.283.356-1.857m0 0a5.002 5.002 0 019.288 0M15 7a3 3 0 11-6 0 3 3 0 016 0zm6 3a2 2 0 11-4 0 2 2 0 014 0zM7 10a2 2 0 11-4 0 2 2 0 014 0z" />
                        </svg>
                      </div>
                      <svg className="w-5 h-5 text-gray-400 group-hover:text-blue-600 group-hover:translate-x-1 transition-all" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                        <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M9 5l7 7-7 7" />
                      </svg>
                    </div>
                    <h3 className="text-xl font-semibold text-gray-900 group-hover:text-blue-600 transition-colors">Clients</h3>
                    <p className="text-sm text-gray-600 mt-2">Manage tax clients and filings</p>
                  </div>
                </Link>

                {/* Affiliates Card */}
                <Link
                  href={`/${defaultTenant.tenantId}/affiliates`}
                  className="group bg-white rounded-xl shadow-sm border border-gray-200 hover:shadow-lg hover:border-purple-300 transition-all overflow-hidden"
                >
                  <div className="p-6">
                    <div className="flex items-center justify-between mb-4">
                      <div className="w-12 h-12 rounded-lg bg-purple-100 flex items-center justify-center group-hover:bg-purple-200 transition-colors">
                        <svg className="w-6 h-6 text-purple-600" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                          <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M12 4.354a4 4 0 110 5.292M15 21H3v-1a6 6 0 0112 0v1zm0 0h6v-1a6 6 0 00-9-5.197M13 7a4 4 0 11-8 0 4 4 0 018 0z" />
                        </svg>
                      </div>
                      <svg className="w-5 h-5 text-gray-400 group-hover:text-purple-600 group-hover:translate-x-1 transition-all" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                        <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M9 5l7 7-7 7" />
                      </svg>
                    </div>
                    <h3 className="text-xl font-semibold text-gray-900 group-hover:text-purple-600 transition-colors">Affiliates</h3>
                    <p className="text-sm text-gray-600 mt-2">Manage affiliate partners</p>
                  </div>
                </Link>

                {/* Commissions Card */}
                <Link
                  href={`/${defaultTenant.tenantId}/commissions`}
                  className="group bg-white rounded-xl shadow-sm border border-gray-200 hover:shadow-lg hover:border-green-300 transition-all overflow-hidden"
                >
                  <div className="p-6">
                    <div className="flex items-center justify-between mb-4">
                      <div className="w-12 h-12 rounded-lg bg-green-100 flex items-center justify-center group-hover:bg-green-200 transition-colors">
                        <svg className="w-6 h-6 text-green-600" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                          <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M12 8c-1.657 0-3 .895-3 2s1.343 2 3 2 3 .895 3 2-1.343 2-3 2m0-8c1.11 0 2.08.402 2.599 1M12 8V7m0 1v8m0 0v1m0-1c-1.11 0-2.08-.402-2.599-1M21 12a9 9 0 11-18 0 9 9 0 0118 0z" />
                        </svg>
                      </div>
                      <svg className="w-5 h-5 text-gray-400 group-hover:text-green-600 group-hover:translate-x-1 transition-all" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                        <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M9 5l7 7-7 7" />
                      </svg>
                    </div>
                    <h3 className="text-xl font-semibold text-gray-900 group-hover:text-green-600 transition-colors">Commissions</h3>
                    <p className="text-sm text-gray-600 mt-2">Track commission payments</p>
                  </div>
                </Link>

                {/* Employees Card - Admin Only */}
                {employee?.role === 'admin' && (
                  <Link
                    href={`/${defaultTenant.tenantId}/employees`}
                    className="group bg-white rounded-xl shadow-sm border border-gray-200 hover:shadow-lg hover:border-orange-300 transition-all overflow-hidden"
                  >
                    <div className="p-6">
                      <div className="flex items-center justify-between mb-4">
                        <div className="w-12 h-12 rounded-lg bg-orange-100 flex items-center justify-center group-hover:bg-orange-200 transition-colors">
                          <svg className="w-6 h-6 text-orange-600" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                            <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M21 13.255A23.931 23.931 0 0112 15c-3.183 0-6.22-.62-9-1.745M16 6V4a2 2 0 00-2-2h-4a2 2 0 00-2 2v2m4 6h.01M5 20h14a2 2 0 002-2V8a2 2 0 00-2-2H5a2 2 0 00-2 2v10a2 2 0 002 2z" />
                          </svg>
                        </div>
                        <svg className="w-5 h-5 text-gray-400 group-hover:text-orange-600 group-hover:translate-x-1 transition-all" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                          <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M9 5l7 7-7 7" />
                        </svg>
                      </div>
                      <h3 className="text-xl font-semibold text-gray-900 group-hover:text-orange-600 transition-colors">Employees</h3>
                      <p className="text-sm text-gray-600 mt-2">Manage team members</p>
                    </div>
                  </Link>
                )}

                {/* Account Management - Admin Only */}
                {employee?.role === 'admin' && (
                  <Link
                    href="/admin/tenants"
                    className="group bg-gradient-to-br from-gray-900 to-gray-800 rounded-xl shadow-sm border border-gray-700 hover:shadow-lg transition-all overflow-hidden"
                  >
                    <div className="p-6">
                      <div className="flex items-center justify-between mb-4">
                        <div className="w-12 h-12 rounded-lg bg-white/10 flex items-center justify-center group-hover:bg-white/20 transition-colors">
                          <svg className="w-6 h-6 text-white" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                            <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M10.325 4.317c.426-1.756 2.924-1.756 3.35 0a1.724 1.724 0 002.573 1.066c1.543-.94 3.31.826 2.37 2.37a1.724 1.724 0 001.065 2.572c1.756.426 1.756 2.924 0 3.35a1.724 1.724 0 00-1.066 2.573c.94 1.543-.826 3.31-2.37 2.37a1.724 1.724 0 00-2.572 1.065c-.426 1.756-2.924 1.756-3.35 0a1.724 1.724 0 00-2.573-1.066c-1.543.94-3.31-.826-2.37-2.37a1.724 1.724 0 00-1.065-2.572c-1.756-.426-1.756-2.924 0-3.35a1.724 1.724 0 001.066-2.573c-.94-1.543.826-3.31 2.37-2.37.996.608 2.296.07 2.572-1.065z" />
                            <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M15 12a3 3 0 11-6 0 3 3 0 016 0z" />
                          </svg>
                        </div>
                        <svg className="w-5 h-5 text-white/60 group-hover:text-white group-hover:translate-x-1 transition-all" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                          <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M9 5l7 7-7 7" />
                        </svg>
                      </div>
                      <h3 className="text-xl font-semibold text-white">Settings</h3>
                      <p className="text-sm text-gray-300 mt-2">Manage accounts & system</p>
                    </div>
                  </Link>
                )}
              </div>
            </div>
          )}
        </div>
      </main>
      <EnvironmentCheck />
    </div>
  );
}
