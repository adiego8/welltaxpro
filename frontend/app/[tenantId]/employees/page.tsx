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

function EmployeesContent() {
  const params = useParams()
  const tenantId = params.tenantId as string
  const { user } = useAuth()

  const [employees, setEmployees] = useState<Employee[]>([])
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)
  const [includeInactive, setIncludeInactive] = useState(false)

  useEffect(() => {
    async function fetchEmployees() {
      if (!user) {
        setLoading(false)
        return
      }

      try {
        const apiUrl = process.env.NEXT_PUBLIC_API_URL || 'http://localhost:8081'
        const idToken = await user.getIdToken()

        const url = includeInactive
          ? `${apiUrl}/api/v1/employees?includeInactive=true`
          : `${apiUrl}/api/v1/employees`

        const response = await fetch(url, {
          headers: {
            'Authorization': `Bearer ${idToken}`,
            'Content-Type': 'application/json',
          },
        })

        if (!response.ok) {
          throw new Error('Failed to fetch employees')
        }

        const data = await response.json()
        setEmployees(data || [])
      } catch (err) {
        setError(err instanceof Error ? err.message : 'An error occurred')
      } finally {
        setLoading(false)
      }
    }

    fetchEmployees()
  }, [user, includeInactive])

  return (
    <div className="min-h-screen bg-gradient-to-br from-slate-50 via-blue-50 to-slate-50">
      <nav className="bg-white/80 backdrop-blur-sm border-b border-gray-200">
        <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
          <div className="flex justify-between h-16 items-center">
            <div className="flex-shrink-0">
              <Link href="/">
                <img src="/logo.png" alt="WellTaxPro" className="h-12 sm:h-16 cursor-pointer" />
              </Link>
            </div>
            <div className="flex items-center gap-2 sm:gap-4">
              <Link
                href="/"
                className="text-xs sm:text-sm text-gray-600 hover:text-gray-900"
              >
                Home
              </Link>
              <span className="text-gray-300">|</span>
              <span className="text-xs sm:text-sm font-semibold text-orange-600">
                Employees
              </span>
              <div className="hidden lg:block text-xs text-gray-500 border-l pl-4 ml-4">
                <span className="font-mono">{tenantId}</span>
              </div>
              <UserMenu />
            </div>
          </div>
        </div>
      </nav>

      <main className="max-w-7xl mx-auto py-6 sm:px-6 lg:px-8">
        <div className="px-4 py-6 sm:px-0">
          <div className="mb-6">
            <h2 className="text-2xl font-bold text-gray-900">Employees</h2>
            <p className="mt-1 text-sm text-gray-500">
              Manage system employees and their access levels
            </p>
          </div>

          <div className="mb-4 flex items-center gap-4">
            <label className="flex items-center">
              <input
                type="checkbox"
                checked={includeInactive}
                onChange={(e) => setIncludeInactive(e.target.checked)}
                className="h-4 w-4 text-blue-600 focus:ring-blue-500 border-gray-300 rounded"
              />
              <span className="ml-2 text-sm text-gray-700">Show inactive employees</span>
            </label>
          </div>

          {loading && (
            <div className="text-center py-12">
              <div className="inline-block animate-spin rounded-full h-8 w-8 border-b-2 border-gray-900"></div>
              <p className="mt-2 text-sm text-gray-500">Loading employees...</p>
            </div>
          )}

          {error && (
            <div className="bg-red-50 border border-red-200 rounded-md p-4">
              <p className="text-sm text-red-800">Error: {error}</p>
            </div>
          )}

          {!loading && !error && employees.length === 0 && (
            <div className="bg-white shadow overflow-hidden sm:rounded-lg p-6 text-center">
              <p className="text-gray-500">No employees found</p>
            </div>
          )}

          {!loading && !error && employees.length > 0 && (
            <div className="bg-white shadow overflow-hidden sm:rounded-lg">
              <table className="min-w-full divide-y divide-gray-200">
                <thead className="bg-gray-50">
                  <tr>
                    <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                      Name
                    </th>
                    <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                      Email
                    </th>
                    <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                      Role
                    </th>
                    <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                      Status
                    </th>
                    <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                      Created
                    </th>
                  </tr>
                </thead>
                <tbody className="bg-white divide-y divide-gray-200">
                  {employees.map((employee) => (
                    <tr
                      key={employee.id}
                      className="hover:bg-gray-100 cursor-pointer"
                      onClick={() => window.location.href = `/${tenantId}/employees/${employee.id}`}
                    >
                      <td className="px-6 py-4 whitespace-nowrap">
                        <div className="text-sm font-medium text-blue-600">
                          {employee.firstName && employee.lastName
                            ? `${employee.firstName} ${employee.lastName}`
                            : 'Not set'}
                        </div>
                      </td>
                      <td className="px-6 py-4 whitespace-nowrap">
                        <div className="text-sm text-gray-900">{employee.email}</div>
                      </td>
                      <td className="px-6 py-4 whitespace-nowrap">
                        <span className="px-2 inline-flex text-xs leading-5 font-semibold rounded-full bg-blue-100 text-blue-800 capitalize">
                          {employee.role}
                        </span>
                      </td>
                      <td className="px-6 py-4 whitespace-nowrap">
                        <span
                          className={`px-2 inline-flex text-xs leading-5 font-semibold rounded-full ${
                            employee.isActive
                              ? 'bg-green-100 text-green-800'
                              : 'bg-red-100 text-red-800'
                          }`}
                        >
                          {employee.isActive ? 'Active' : 'Inactive'}
                        </span>
                      </td>
                      <td className="px-6 py-4 whitespace-nowrap text-sm text-gray-500">
                        {new Date(employee.createdAt).toLocaleDateString()}
                      </td>
                    </tr>
                  ))}
                </tbody>
              </table>
            </div>
          )}

          <div className="mt-6">
            <p className="text-sm text-gray-500">
              Total employees: <span className="font-semibold">{employees.length}</span>
            </p>
          </div>
        </div>
      </main>
    </div>
  )
}

export default function EmployeesPage() {
  return (
    <ProtectedRoute>
      <EmployeesContent />
    </ProtectedRoute>
  )
}
