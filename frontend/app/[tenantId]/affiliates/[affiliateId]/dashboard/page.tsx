'use client'

import { useEffect, useState } from 'react'
import { useParams, useSearchParams } from 'next/navigation'

interface Affiliate {
  id: string
  firstName: string
  lastName: string
  email: string
  phone: string | null
  defaultCommissionRate: number
  payoutMethod: string
  payoutThreshold: number
  isActive: boolean
  createdAt: string
}

interface AffiliateStats {
  affiliateId: string
  totalClicks: number
  totalConversions: number
  conversionRate: number
  totalCommissionsEarned: number
  pendingCommissions: number
  approvedCommissions: number
  paidCommissions: number
  cancelledCommissions: number
  totalOrders: number
  totalRevenue: number
}

interface Commission {
  id: string
  affiliateId: string
  orderAmount: number
  discountAmount: number
  netAmount: number
  commissionRate: number
  commissionAmount: number
  status: string
  createdAt: string
  customer?: {
    id: string
    firstName: string | null
    lastName: string | null
    email: string
  }
}

interface DashboardData {
  affiliate: Affiliate
  stats: AffiliateStats
  commissions: Commission[]
}

export default function PublicAffiliateDashboard() {
  const params = useParams()
  const searchParams = useSearchParams()
  const tenantId = params.tenantId as string
  const affiliateId = params.affiliateId as string
  const token = searchParams.get('token')

  const [dashboard, setDashboard] = useState<DashboardData | null>(null)
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)

  useEffect(() => {
    async function fetchDashboard() {
      if (!token) {
        setError('Access token is required')
        setLoading(false)
        return
      }

      try {
        const apiUrl = process.env.NEXT_PUBLIC_API_URL || 'http://localhost:8081'
        const response = await fetch(
          `${apiUrl}/api/v1/${tenantId}/affiliates/${affiliateId}/dashboard?token=${token}`
        )

        if (!response.ok) {
          if (response.status === 401) {
            throw new Error('Invalid or expired access token')
          }
          throw new Error('Failed to load dashboard')
        }

        const data = await response.json()
        setDashboard(data)
      } catch (err) {
        setError(err instanceof Error ? err.message : 'An error occurred')
      } finally {
        setLoading(false)
      }
    }

    fetchDashboard()
  }, [tenantId, affiliateId, token])

  if (loading) {
    return (
      <div className="min-h-screen bg-gray-50 flex items-center justify-center">
        <div className="text-center">
          <div className="inline-block animate-spin rounded-full h-8 w-8 border-b-2 border-blue-600"></div>
          <p className="mt-2 text-sm text-gray-500">Loading dashboard...</p>
        </div>
      </div>
    )
  }

  if (error || !dashboard) {
    return (
      <div className="min-h-screen bg-gray-50 flex items-center justify-center">
        <div className="max-w-md mx-auto text-center">
          <div className="bg-red-50 border border-red-200 rounded-lg p-6">
            <div className="text-red-600 mb-2">
              <svg className="w-12 h-12 mx-auto" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M12 9v2m0 4h.01m-6.938 4h13.856c1.54 0 2.502-1.667 1.732-3L13.732 4c-.77-1.333-2.694-1.333-3.464 0L3.34 16c-.77 1.333.192 3 1.732 3z" />
              </svg>
            </div>
            <h3 className="text-lg font-medium text-gray-900 mb-2">Access Denied</h3>
            <p className="text-sm text-red-800">{error || 'Unable to load dashboard'}</p>
            <p className="text-xs text-gray-600 mt-4">
              Please contact support if you believe this is an error.
            </p>
          </div>
        </div>
      </div>
    )
  }

  const { affiliate, stats, commissions } = dashboard

  return (
    <div className="min-h-screen bg-gray-50">
      {/* Header */}
      <nav className="bg-white shadow-sm">
        <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
          <div className="flex justify-between h-16 items-center">
            <div className="flex-shrink-0">
              <h1 className="text-2xl font-bold text-gray-900">Affiliate Dashboard</h1>
            </div>
            <div className="text-sm text-gray-500">
              {tenantId}
            </div>
          </div>
        </div>
      </nav>

      <main className="max-w-7xl mx-auto py-6 sm:px-6 lg:px-8">
        <div className="px-4 py-6 sm:px-0">
          {/* Welcome Section */}
          <div className="mb-6">
            <h2 className="text-3xl font-bold text-gray-900">
              Welcome, {affiliate.firstName} {affiliate.lastName}!
            </h2>
            <p className="mt-1 text-sm text-gray-500">{affiliate.email}</p>
          </div>

          {/* Stats Cards */}
          <div className="grid grid-cols-1 gap-5 sm:grid-cols-2 lg:grid-cols-4 mb-6">
            {/* Total Earnings */}
            <div className="bg-white overflow-hidden shadow rounded-lg">
              <div className="p-5">
                <div className="flex items-center">
                  <div className="flex-shrink-0">
                    <svg className="h-6 w-6 text-green-600" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                      <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M12 8c-1.657 0-3 .895-3 2s1.343 2 3 2 3 .895 3 2-1.343 2-3 2m0-8c1.11 0 2.08.402 2.599 1M12 8V7m0 1v8m0 0v1m0-1c-1.11 0-2.08-.402-2.599-1M21 12a9 9 0 11-18 0 9 9 0 0118 0z" />
                    </svg>
                  </div>
                  <div className="ml-5 w-0 flex-1">
                    <dl>
                      <dt className="text-sm font-medium text-gray-500 truncate">
                        Total Earnings
                      </dt>
                      <dd className="text-2xl font-semibold text-gray-900">
                        ${stats.totalCommissionsEarned.toFixed(2)}
                      </dd>
                    </dl>
                  </div>
                </div>
              </div>
            </div>

            {/* Pending Commissions */}
            <div className="bg-white overflow-hidden shadow rounded-lg">
              <div className="p-5">
                <div className="flex items-center">
                  <div className="flex-shrink-0">
                    <svg className="h-6 w-6 text-yellow-600" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                      <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M12 8v4l3 3m6-3a9 9 0 11-18 0 9 9 0 0118 0z" />
                    </svg>
                  </div>
                  <div className="ml-5 w-0 flex-1">
                    <dl>
                      <dt className="text-sm font-medium text-gray-500 truncate">
                        Pending
                      </dt>
                      <dd className="text-2xl font-semibold text-gray-900">
                        ${stats.pendingCommissions.toFixed(2)}
                      </dd>
                    </dl>
                  </div>
                </div>
              </div>
            </div>

            {/* Paid Commissions */}
            <div className="bg-white overflow-hidden shadow rounded-lg">
              <div className="p-5">
                <div className="flex items-center">
                  <div className="flex-shrink-0">
                    <svg className="h-6 w-6 text-blue-600" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                      <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M9 12l2 2 4-4m6 2a9 9 0 11-18 0 9 9 0 0118 0z" />
                    </svg>
                  </div>
                  <div className="ml-5 w-0 flex-1">
                    <dl>
                      <dt className="text-sm font-medium text-gray-500 truncate">
                        Paid Out
                      </dt>
                      <dd className="text-2xl font-semibold text-gray-900">
                        ${stats.paidCommissions.toFixed(2)}
                      </dd>
                    </dl>
                  </div>
                </div>
              </div>
            </div>

            {/* Conversion Rate */}
            <div className="bg-white overflow-hidden shadow rounded-lg">
              <div className="p-5">
                <div className="flex items-center">
                  <div className="flex-shrink-0">
                    <svg className="h-6 w-6 text-purple-600" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                      <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M13 7h8m0 0v8m0-8l-8 8-4-4-6 6" />
                    </svg>
                  </div>
                  <div className="ml-5 w-0 flex-1">
                    <dl>
                      <dt className="text-sm font-medium text-gray-500 truncate">
                        Conversion Rate
                      </dt>
                      <dd className="text-2xl font-semibold text-gray-900">
                        {stats.conversionRate.toFixed(1)}%
                      </dd>
                    </dl>
                  </div>
                </div>
              </div>
            </div>
          </div>

          {/* Additional Stats */}
          <div className="grid grid-cols-1 gap-5 sm:grid-cols-3 mb-6">
            <div className="bg-white overflow-hidden shadow rounded-lg p-5">
              <dt className="text-sm font-medium text-gray-500 truncate">Total Orders</dt>
              <dd className="mt-1 text-3xl font-semibold text-gray-900">{stats.totalOrders}</dd>
            </div>
            <div className="bg-white overflow-hidden shadow rounded-lg p-5">
              <dt className="text-sm font-medium text-gray-500 truncate">Total Revenue Generated</dt>
              <dd className="mt-1 text-3xl font-semibold text-gray-900">${stats.totalRevenue.toFixed(2)}</dd>
            </div>
            <div className="bg-white overflow-hidden shadow rounded-lg p-5">
              <dt className="text-sm font-medium text-gray-500 truncate">Commission Rate</dt>
              <dd className="mt-1 text-3xl font-semibold text-gray-900">{affiliate.defaultCommissionRate.toFixed(1)}%</dd>
            </div>
          </div>

          {/* Recent Commissions */}
          <div className="bg-white shadow overflow-hidden sm:rounded-lg">
            <div className="px-4 py-5 sm:px-6">
              <h3 className="text-lg leading-6 font-medium text-gray-900">
                Recent Commissions
              </h3>
              <p className="mt-1 text-sm text-gray-500">
                Your latest commission earnings
              </p>
            </div>
            <div className="border-t border-gray-200">
              {!commissions || commissions.length === 0 ? (
                <div className="px-4 py-5 text-center text-sm text-gray-500">
                  No commissions yet
                </div>
              ) : (
                <table className="min-w-full divide-y divide-gray-200">
                  <thead className="bg-gray-50">
                    <tr>
                      <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                        Date
                      </th>
                      <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                        Customer
                      </th>
                      <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                        Order Amount
                      </th>
                      <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                        Your Commission
                      </th>
                      <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                        Status
                      </th>
                    </tr>
                  </thead>
                  <tbody className="bg-white divide-y divide-gray-200">
                    {commissions.map((commission) => (
                      <tr key={commission.id}>
                        <td className="px-6 py-4 whitespace-nowrap text-sm text-gray-500">
                          {new Date(commission.createdAt).toLocaleDateString()}
                        </td>
                        <td className="px-6 py-4 whitespace-nowrap text-sm text-gray-900">
                          {commission.customer
                            ? `${commission.customer.firstName || ''} ${commission.customer.lastName || ''}`.trim() || commission.customer.email
                            : 'N/A'}
                        </td>
                        <td className="px-6 py-4 whitespace-nowrap text-sm text-gray-900">
                          ${commission.orderAmount.toFixed(2)}
                        </td>
                        <td className="px-6 py-4 whitespace-nowrap text-sm font-medium text-green-600">
                          ${commission.commissionAmount.toFixed(2)}
                          <span className="text-gray-500 ml-1">
                            ({commission.commissionRate.toFixed(1)}%)
                          </span>
                        </td>
                        <td className="px-6 py-4 whitespace-nowrap">
                          <span
                            className={`px-2 inline-flex text-xs leading-5 font-semibold rounded-full ${
                              commission.status === 'PAID'
                                ? 'bg-green-100 text-green-800'
                                : commission.status === 'APPROVED'
                                ? 'bg-blue-100 text-blue-800'
                                : commission.status === 'PENDING'
                                ? 'bg-yellow-100 text-yellow-800'
                                : 'bg-red-100 text-red-800'
                            }`}
                          >
                            {commission.status}
                          </span>
                        </td>
                      </tr>
                    ))}
                  </tbody>
                </table>
              )}
            </div>
          </div>

          {/* Footer */}
          <div className="mt-8 text-center text-sm text-gray-500">
            <p>Questions? Contact the WellTaxPro Team</p>
          </div>
        </div>
      </main>
    </div>
  )
}
