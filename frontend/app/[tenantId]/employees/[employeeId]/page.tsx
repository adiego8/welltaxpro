'use client'

import { useEffect, useState } from 'react'
import { useParams } from 'next/navigation'
import Link from 'next/link'
import { ProtectedRoute } from '@/components/ProtectedRoute'
import { UserMenu } from '@/components/UserMenu'
import { useAuth } from '@/contexts/AuthContext'

interface Employee {
  id: string
  firebaseUid: string
  email: string
  firstName: string | null
  lastName: string | null
  role: string
  isActive: boolean
  createdAt: string
  updatedAt: string | null
}

function EmployeeDetailContent() {
  const params = useParams()
  const tenantId = params.tenantId as string
  const employeeId = params.employeeId as string
  const { user } = useAuth()

  const [employee, setEmployee] = useState<Employee | null>(null)
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)

  const apiUrl = process.env.NEXT_PUBLIC_API_URL || 'http://localhost:8081'

  useEffect(() => {
    async function fetchData() {
      if (!user) {
        setLoading(false)
        return
      }

      try {
        const idToken = await user.getIdToken()
        const headers = {
          'Authorization': `Bearer ${idToken}`,
          'Content-Type': 'application/json',
        }

        // Fetch employee details
        const employeeResponse = await fetch(
          `${apiUrl}/api/v1/employees/${employeeId}`,
          { headers }
        )
        if (!employeeResponse.ok) throw new Error('Failed to fetch employee')
        const employeeData = await employeeResponse.json()
        setEmployee(employeeData)
      } catch (err) {
        setError(err instanceof Error ? err.message : 'An error occurred')
      } finally {
        setLoading(false)
      }
    }

    fetchData()
  }, [tenantId, employeeId, user, apiUrl])

  if (loading) {
    return (
      <div className="min-h-screen bg-gray-50 flex items-center justify-center">
        <div className="text-center">
          <div className="inline-block animate-spin rounded-full h-8 w-8 border-b-2 border-gray-900"></div>
          <p className="mt-2 text-sm text-gray-500">Loading employee...</p>
        </div>
      </div>
    )
  }

  if (error || !employee) {
    return (
      <div className="min-h-screen bg-gray-50 flex items-center justify-center">
        <div className="bg-red-50 border border-red-200 rounded-md p-4">
          <p className="text-sm text-red-800">Error: {error || 'Employee not found'}</p>
        </div>
      </div>
    )
  }

  return (
    <div className="min-h-screen bg-gray-50">
      <nav className="bg-white shadow-sm">
        <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
          <div className="flex justify-between h-16 items-center">
            <div className="flex-shrink-0">
              <Link href="/">
                <img src="/logo.png" alt="WellTaxPro" className="h-16 cursor-pointer" />
              </Link>
            </div>
            <div className="flex items-center gap-4">
              <Link
                href={`/${tenantId}/employees`}
                className="text-sm text-gray-600 hover:text-gray-900"
              >
                ← Back to Employees
              </Link>
              <div className="text-sm text-gray-500 border-l pl-4 ml-4">
                Account: <span className="font-semibold">{tenantId}</span>
              </div>
              <UserMenu />
            </div>
          </div>
        </div>
      </nav>

      <main className="max-w-7xl mx-auto py-6 sm:px-6 lg:px-8">
        <div className="px-4 py-6 sm:px-0">
          {/* Employee Header */}
          <div className="mb-6">
            <h2 className="text-2xl font-bold text-gray-900">
              {employee.firstName && employee.lastName
                ? `${employee.firstName} ${employee.lastName}`
                : employee.email}
            </h2>
            <p className="mt-1 text-sm text-gray-500">{employee.email}</p>
          </div>

          {/* Employee Details */}
          <div className="bg-white shadow overflow-hidden sm:rounded-lg mb-6">
            <div className="px-4 py-5 sm:px-6 flex justify-between items-center">
              <h3 className="text-lg leading-6 font-medium text-gray-900">
                Employee Details
              </h3>
              <div className="flex items-center gap-2">
                <span
                  className={`px-3 py-1 inline-flex text-sm leading-5 font-semibold rounded-full ${
                    employee.isActive
                      ? 'bg-green-100 text-green-800'
                      : 'bg-red-100 text-red-800'
                  }`}
                >
                  {employee.isActive ? 'Active' : 'Inactive'}
                </span>
                <span className="px-3 py-1 inline-flex text-sm leading-5 font-semibold rounded-full bg-blue-100 text-blue-800 capitalize">
                  {employee.role}
                </span>
              </div>
            </div>
            <div className="border-t border-gray-200">
              <dl>
                <div className="bg-gray-50 px-4 py-5 sm:grid sm:grid-cols-3 sm:gap-4 sm:px-6">
                  <dt className="text-sm font-medium text-gray-500">Full Name</dt>
                  <dd className="mt-1 text-sm text-gray-900 sm:mt-0 sm:col-span-2">
                    {employee.firstName && employee.lastName
                      ? `${employee.firstName} ${employee.lastName}`
                      : 'Not set'}
                  </dd>
                </div>
                <div className="bg-white px-4 py-5 sm:grid sm:grid-cols-3 sm:gap-4 sm:px-6">
                  <dt className="text-sm font-medium text-gray-500">Email</dt>
                  <dd className="mt-1 text-sm text-gray-900 sm:mt-0 sm:col-span-2">
                    {employee.email}
                  </dd>
                </div>
                <div className="bg-gray-50 px-4 py-5 sm:grid sm:grid-cols-3 sm:gap-4 sm:px-6">
                  <dt className="text-sm font-medium text-gray-500">Role</dt>
                  <dd className="mt-1 text-sm text-gray-900 sm:mt-0 sm:col-span-2 capitalize">
                    {employee.role}
                  </dd>
                </div>
                <div className="bg-white px-4 py-5 sm:grid sm:grid-cols-3 sm:gap-4 sm:px-6">
                  <dt className="text-sm font-medium text-gray-500">Firebase UID</dt>
                  <dd className="mt-1 text-sm text-gray-900 sm:mt-0 sm:col-span-2 font-mono text-xs">
                    {employee.firebaseUid}
                  </dd>
                </div>
                <div className="bg-gray-50 px-4 py-5 sm:grid sm:grid-cols-3 sm:gap-4 sm:px-6">
                  <dt className="text-sm font-medium text-gray-500">Created</dt>
                  <dd className="mt-1 text-sm text-gray-900 sm:mt-0 sm:col-span-2">
                    {new Date(employee.createdAt).toLocaleString()}
                  </dd>
                </div>
                {employee.updatedAt && (
                  <div className="bg-white px-4 py-5 sm:grid sm:grid-cols-3 sm:gap-4 sm:px-6">
                    <dt className="text-sm font-medium text-gray-500">Last Updated</dt>
                    <dd className="mt-1 text-sm text-gray-900 sm:mt-0 sm:col-span-2">
                      {new Date(employee.updatedAt).toLocaleString()}
                    </dd>
                  </div>
                )}
              </dl>
            </div>
          </div>

          {/* Tenant Access */}
          <div className="bg-white shadow overflow-hidden sm:rounded-lg mb-6">
            <div className="px-4 py-5 sm:px-6">
              <h3 className="text-lg leading-6 font-medium text-gray-900">
                Tenant Access
              </h3>
              <p className="mt-1 text-sm text-gray-500">
                Tenants this employee has access to
              </p>
            </div>
            <div className="border-t border-gray-200 px-4 py-5">
              <div className="text-center text-sm text-gray-500">
                <p>Tenant access management coming soon</p>
                <p className="mt-2 text-xs">
                  Currently: This employee has access to tenant <span className="font-semibold">{tenantId}</span>
                </p>
              </div>
            </div>
          </div>

          {/* Activity Log */}
          <div className="bg-white shadow overflow-hidden sm:rounded-lg">
            <div className="px-4 py-5 sm:px-6">
              <h3 className="text-lg leading-6 font-medium text-gray-900">
                Activity Log
              </h3>
              <p className="mt-1 text-sm text-gray-500">
                Recent activity and audit trail
              </p>
            </div>
            <div className="border-t border-gray-200 px-4 py-5 text-center text-sm text-gray-500">
              Activity log coming soon
            </div>
          </div>
        </div>
      </main>
    </div>
  )
}

export default function EmployeeDetailPage() {
  return (
    <ProtectedRoute>
      <EmployeeDetailContent />
    </ProtectedRoute>
  )
}
