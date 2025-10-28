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
}

interface AffiliateStats {
  totalClicks: number
  totalConversions: number
  conversionRate: number
  totalCommissionsEarned: number
  pendingCommissions: number
  approvedCommissions: number
  paidCommissions: number
  totalOrders: number
  totalRevenue: number
}

interface Commission {
  id: string
  orderAmount: number
  discountAmount: number
  netAmount: number
  commissionRate: number
  commissionAmount: number
  status: string
  createdAt: string
  customer: {
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

export default function AffiliateDashboardPage() {
  const params = useParams()
  const searchParams = useSearchParams()
  const tenantId = params.tenantId as string
  const affiliateId = params.affiliateId as string
  const token = searchParams.get('token')

  const [data, setData] = useState<DashboardData | null>(null)
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)

  useEffect(() => {
    async function fetchDashboard() {
      if (!token) {
        setError('No access token provided')
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
          throw new Error('Failed to fetch dashboard data')
        }

        const dashboardData = await response.json()
        setData(dashboardData)
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
          <div className="inline-block animate-spin rounded-full h-12 w-12 border-b-2 border-blue-600"></div>
          <p className="mt-4 text-sm text-gray-500">Loading dashboard...</p>
        </div>
      </div>
    )
  }

  if (error || !data) {
    return (
      <div className="min-h-screen bg-gray-50 flex items-center justify-center">
        <div className="max-w-md w-full">
          <div className="bg-white shadow-lg rounded-lg p-8 text-center">
            <div className="mx-auto flex items-center justify-center h-12 w-12 rounded-full bg-red-100 mb-4">
              <svg
                className="h-6 w-6 text-red-600"
                fill="none"
                viewBox="0 0 24 24"
                stroke="currentColor"
              >
                <path
                  strokeLinecap="round"
                  strokeLinejoin="round"
                  strokeWidth={2}
                  d="M12 9v2m0 4h.01m-6.938 4h13.856c1.54 0 2.502-1.667 1.732-3L13.732 4c-.77-1.333-2.694-1.333-3.464 0L3.34 16c-.77 1.333.192 3 1.732 3z"
                />
              </svg>
            </div>
            <h3 className="text-lg font-medium text-gray-900 mb-2">Access Denied</h3>
            <p className="text-sm text-gray-500">{error || 'Unable to load dashboard'}</p>
          </div>
        </div>
      </div>
    )
  }

  const { affiliate, stats, commissions } = data

  return (
    <div className="min-h-screen bg-gray-50">
      {/* Header */}
      <div className="bg-white shadow">
        <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-6">
          <div className="md:flex md:items-center md:justify-between">
            <div className="flex-1 min-w-0">
              <h1 className="text-3xl font-bold text-gray-900">
                {affiliate.firstName} {affiliate.lastName}
              </h1>
              <p className="mt-1 text-sm text-gray-500">Affiliate Dashboard</p>
            </div>
            <div className="mt-4 flex md:mt-0 md:ml-4">
              <span
                className={`px-4 py-2 rounded-full text-sm font-semibold ${
                  affiliate.isActive
                    ? 'bg-green-100 text-green-800'
                    : 'bg-red-100 text-red-800'
                }`}
              >
                {affiliate.isActive ? 'Active' : 'Inactive'}
              </span>
            </div>
          </div>
        </div>
      </div>

      <main className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-8">
        {/* Performance Stats */}
        <div className="mb-8">
          <h2 className="text-lg font-semibold text-gray-900 mb-4">Performance Overview</h2>
          <div className="grid grid-cols-1 gap-5 sm:grid-cols-2 lg:grid-cols-4">
            <div className="bg-white overflow-hidden shadow rounded-lg">
              <div className="p-5">
                <div className="flex items-center">
                  <div className="flex-shrink-0">
                    <svg
                      className="h-6 w-6 text-gray-400"
                      fill="none"
                      viewBox="0 0 24 24"
                      stroke="currentColor"
                    >
                      <path
                        strokeLinecap="round"
                        strokeLinejoin="round"
                        strokeWidth={2}
                        d="M15 15l-2 5L9 9l11 4-5 2zm0 0l5 5M7.188 2.239l.777 2.897M5.136 7.965l-2.898-.777M13.95 4.05l-2.122 2.122m-5.657 5.656l-2.12 2.122"
                      />
                    </svg>
                  </div>
                  <div className="ml-5 w-0 flex-1">
                    <dt className="text-sm font-medium text-gray-500 truncate">Total Clicks</dt>
                    <dd className="text-2xl font-semibold text-gray-900">{stats.totalClicks}</dd>
                  </div>
                </div>
              </div>
            </div>

            <div className="bg-white overflow-hidden shadow rounded-lg">
              <div className="p-5">
                <div className="flex items-center">
                  <div className="flex-shrink-0">
                    <svg
                      className="h-6 w-6 text-gray-400"
                      fill="none"
                      viewBox="0 0 24 24"
                      stroke="currentColor"
                    >
                      <path
                        strokeLinecap="round"
                        strokeLinejoin="round"
                        strokeWidth={2}
                        d="M9 12l2 2 4-4m6 2a9 9 0 11-18 0 9 9 0 0118 0z"
                      />
                    </svg>
                  </div>
                  <div className="ml-5 w-0 flex-1">
                    <dt className="text-sm font-medium text-gray-500 truncate">Conversions</dt>
                    <dd className="text-2xl font-semibold text-gray-900">
                      {stats.totalConversions}
                    </dd>
                  </div>
                </div>
              </div>
            </div>

            <div className="bg-white overflow-hidden shadow rounded-lg">
              <div className="p-5">
                <div className="flex items-center">
                  <div className="flex-shrink-0">
                    <svg
                      className="h-6 w-6 text-gray-400"
                      fill="none"
                      viewBox="0 0 24 24"
                      stroke="currentColor"
                    >
                      <path
                        strokeLinecap="round"
                        strokeLinejoin="round"
                        strokeWidth={2}
                        d="M13 7h8m0 0v8m0-8l-8 8-4-4-6 6"
                      />
                    </svg>
                  </div>
                  <div className="ml-5 w-0 flex-1">
                    <dt className="text-sm font-medium text-gray-500 truncate">
                      Conversion Rate
                    </dt>
                    <dd className="text-2xl font-semibold text-gray-900">
                      {stats.conversionRate.toFixed(2)}%
                    </dd>
                  </div>
                </div>
              </div>
            </div>

            <div className="bg-white overflow-hidden shadow rounded-lg">
              <div className="p-5">
                <div className="flex items-center">
                  <div className="flex-shrink-0">
                    <svg
                      className="h-6 w-6 text-gray-400"
                      fill="none"
                      viewBox="0 0 24 24"
                      stroke="currentColor"
                    >
                      <path
                        strokeLinecap="round"
                        strokeLinejoin="round"
                        strokeWidth={2}
                        d="M12 8c-1.657 0-3 .895-3 2s1.343 2 3 2 3 .895 3 2-1.343 2-3 2m0-8c1.11 0 2.08.402 2.599 1M12 8V7m0 1v8m0 0v1m0-1c-1.11 0-2.08-.402-2.599-1M21 12a9 9 0 11-18 0 9 9 0 0118 0z"
                      />
                    </svg>
                  </div>
                  <div className="ml-5 w-0 flex-1">
                    <dt className="text-sm font-medium text-gray-500 truncate">Total Revenue</dt>
                    <dd className="text-2xl font-semibold text-gray-900">
                      ${stats.totalRevenue.toFixed(2)}
                    </dd>
                  </div>
                </div>
              </div>
            </div>
          </div>
        </div>

        {/* Earnings Breakdown */}
        <div className="mb-8">
          <h2 className="text-lg font-semibold text-gray-900 mb-4">Earnings Breakdown</h2>
          <div className="grid grid-cols-1 gap-5 sm:grid-cols-4">
            <div className="bg-white overflow-hidden shadow rounded-lg">
              <div className="p-5">
                <dt className="text-sm font-medium text-gray-500 truncate">Total Earned</dt>
                <dd className="mt-1 text-3xl font-semibold text-green-600">
                  ${stats.totalCommissionsEarned.toFixed(2)}
                </dd>
              </div>
            </div>

            <div className="bg-white overflow-hidden shadow rounded-lg">
              <div className="p-5">
                <dt className="text-sm font-medium text-gray-500 truncate">Pending</dt>
                <dd className="mt-1 text-3xl font-semibold text-yellow-600">
                  ${stats.pendingCommissions.toFixed(2)}
                </dd>
              </div>
            </div>

            <div className="bg-white overflow-hidden shadow rounded-lg">
              <div className="p-5">
                <dt className="text-sm font-medium text-gray-500 truncate">Approved</dt>
                <dd className="mt-1 text-3xl font-semibold text-blue-600">
                  ${stats.approvedCommissions.toFixed(2)}
                </dd>
              </div>
            </div>

            <div className="bg-white overflow-hidden shadow rounded-lg">
              <div className="p-5">
                <dt className="text-sm font-medium text-gray-500 truncate">Paid</dt>
                <dd className="mt-1 text-3xl font-semibold text-green-600">
                  ${stats.paidCommissions.toFixed(2)}
                </dd>
              </div>
            </div>
          </div>
        </div>

        {/* Recent Commissions */}
        <div className="mb-8">
          <h2 className="text-lg font-semibold text-gray-900 mb-4">Recent Commissions</h2>
          {commissions.length === 0 ? (
            <div className="bg-white shadow overflow-hidden sm:rounded-lg p-6 text-center">
              <p className="text-gray-500">No commissions yet</p>
            </div>
          ) : (
            <div className="bg-white shadow overflow-hidden sm:rounded-lg">
              <table className="min-w-full divide-y divide-gray-200">
                <thead className="bg-gray-50">
                  <tr>
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
                    <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                      Date
                    </th>
                  </tr>
                </thead>
                <tbody className="bg-white divide-y divide-gray-200">
                  {commissions.map((commission) => (
                    <tr key={commission.id}>
                      <td className="px-6 py-4 whitespace-nowrap">
                        <div className="text-sm font-medium text-gray-900">
                          {commission.customer.firstName} {commission.customer.lastName}
                        </div>
                        <div className="text-sm text-gray-500">{commission.customer.email}</div>
                      </td>
                      <td className="px-6 py-4 whitespace-nowrap text-sm text-gray-900">
                        ${commission.orderAmount.toFixed(2)}
                      </td>
                      <td className="px-6 py-4 whitespace-nowrap">
                        <div className="text-sm font-medium text-gray-900">
                          ${commission.commissionAmount.toFixed(2)}
                        </div>
                        <div className="text-sm text-gray-500">
                          {commission.commissionRate.toFixed(1)}% rate
                        </div>
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
                      <td className="px-6 py-4 whitespace-nowrap text-sm text-gray-500">
                        {new Date(commission.createdAt).toLocaleDateString()}
                      </td>
                    </tr>
                  ))}
                </tbody>
              </table>
            </div>
          )}
        </div>

        {/* Payout Info */}
        <div className="bg-blue-50 border border-blue-200 rounded-lg p-6">
          <h3 className="text-lg font-semibold text-blue-900 mb-2">Payout Information</h3>
          <div className="grid grid-cols-1 md:grid-cols-2 gap-4 text-sm">
            <div>
              <span className="font-medium text-blue-900">Payout Method:</span>
              <span className="ml-2 text-blue-700">{affiliate.payoutMethod}</span>
            </div>
            <div>
              <span className="font-medium text-blue-900">Payout Threshold:</span>
              <span className="ml-2 text-blue-700">${affiliate.payoutThreshold.toFixed(2)}</span>
            </div>
            <div>
              <span className="font-medium text-blue-900">Commission Rate:</span>
              <span className="ml-2 text-blue-700">
                {affiliate.defaultCommissionRate.toFixed(1)}%
              </span>
            </div>
          </div>
        </div>
      </main>

      {/* Footer */}
      <footer className="bg-white border-t border-gray-200 mt-12">
        <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-6">
          <p className="text-center text-sm text-gray-500">
            Affiliate Dashboard - For support, contact your program administrator
          </p>
        </div>
      </footer>
    </div>
  )
}
