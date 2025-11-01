'use client'

import { useEffect, useState } from 'react'
import { useParams } from 'next/navigation'
import Link from 'next/link'
import { ProtectedRoute } from '@/components/ProtectedRoute'
import { UserMenu } from '@/components/UserMenu'
import { TenantSwitcher } from '@/components/TenantSwitcher'
import { useAuth } from '@/contexts/AuthContext'

interface FilingStatus {
  id: string
  filingId: string
  latestStep: number
  isCompleted: boolean
  status: string
}

interface Filing {
  id: string
  year: number
  userId: string
  maritalStatus: string | null
  income: number | null
  createdAt: string
  updatedAt: string | null
  status: FilingStatus | null
}

interface Client {
  id: string
  firstName: string | null
  lastName: string | null
  email: string
}

interface ClientComprehensive {
  client: Client
  filings: Filing[]
}

function FilingsContent() {
  const params = useParams()
  const tenantId = params.tenantId as string
  const { user } = useAuth()

  const [allClientsData, setAllClientsData] = useState<ClientComprehensive[]>([])
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)
  const [statusFilter, setStatusFilter] = useState<string>('')
  const [yearFilter, setYearFilter] = useState<string>('')

  // Fetch all filings once
  useEffect(() => {
    async function fetchFilings() {
      if (!user) {
        setLoading(false)
        return
      }

      try {
        const apiUrl = process.env.NEXT_PUBLIC_API_URL || 'http://localhost:8081'
        const idToken = await user.getIdToken()

        // Fetch with large limit to get all filings (can be improved with progressive loading later)
        const url = `${apiUrl}/api/v1/${tenantId}/filings?limit=1000&offset=0`

        const response = await fetch(url, {
          headers: {
            'Authorization': `Bearer ${idToken}`,
            'Content-Type': 'application/json',
          },
        })

        if (!response.ok) {
          throw new Error('Failed to fetch filings')
        }

        const data = await response.json()
        setAllClientsData(data || [])
      } catch (err) {
        setError(err instanceof Error ? err.message : 'An error occurred')
      } finally {
        setLoading(false)
      }
    }

    fetchFilings()
  }, [tenantId, user])

  // Filter data client-side based on current filter selections
  const filteredClientsData = allClientsData.map(clientData => {
    // Filter filings for this client
    const filteredFilings = clientData.filings?.filter(filing => {
      let matches = true

      // Apply status filter (case-insensitive comparison)
      if (statusFilter && filing.status) {
        matches = matches && filing.status.status.toLowerCase() === statusFilter.toLowerCase()
      }

      // Apply year filter
      if (yearFilter) {
        matches = matches && filing.year === parseInt(yearFilter)
      }

      return matches
    }) || []

    // Return client with filtered filings
    return {
      ...clientData,
      filings: filteredFilings
    }
  }).filter(clientData => clientData.filings.length > 0) // Only include clients with matching filings

  // Calculate total filings from filtered clients
  const totalFilings = filteredClientsData.reduce((sum, client) => sum + (client.filings?.length || 0), 0)

  const getStatusBadge = (status: string) => {
    const statusColors: Record<string, string> = {
      'pending': 'bg-yellow-100 text-yellow-800',
      'in_progress': 'bg-blue-100 text-blue-800',
      'completed': 'bg-green-100 text-green-800',
      'submitted': 'bg-purple-100 text-purple-800',
      'rejected': 'bg-red-100 text-red-800',
    }
    return statusColors[status] || 'bg-gray-100 text-gray-800'
  }

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
              <Link
                href={`/${tenantId}/clients`}
                className="text-xs sm:text-sm text-gray-600 hover:text-gray-900"
              >
                Clients
              </Link>
              <span className="text-gray-300">|</span>
              <span className="text-xs sm:text-sm font-semibold text-blue-600">
                Filings
              </span>
              <TenantSwitcher />
              <UserMenu />
            </div>
          </div>
        </div>
      </nav>

      <main className="max-w-7xl mx-auto py-8 sm:px-6 lg:px-8">
        <div className="px-4 sm:px-0">
          <div className="mb-8">
            <h2 className="text-2xl sm:text-3xl font-bold text-gray-900">Tax Filings</h2>
            <p className="mt-1 text-sm text-gray-600">
              View and filter all tax filings across clients
            </p>
          </div>

          {/* Filters */}
          <div className="bg-white rounded-xl shadow-sm border border-gray-200 p-6 mb-6">
            <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
              <div>
                <label htmlFor="status" className="block text-sm font-medium text-gray-700 mb-2">
                  Status
                </label>
                <select
                  id="status"
                  value={statusFilter}
                  onChange={(e) => setStatusFilter(e.target.value)}
                  className="w-full px-4 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-blue-500 focus:border-blue-500"
                >
                  <option value="">All Statuses</option>
                  <option value="pending">Pending</option>
                  <option value="in_progress">In Progress</option>
                  <option value="completed">Completed</option>
                  <option value="submitted">Submitted</option>
                  <option value="rejected">Rejected</option>
                </select>
              </div>
              <div>
                <label htmlFor="year" className="block text-sm font-medium text-gray-700 mb-2">
                  Year
                </label>
                <select
                  id="year"
                  value={yearFilter}
                  onChange={(e) => setYearFilter(e.target.value)}
                  className="w-full px-4 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-blue-500 focus:border-blue-500"
                >
                  <option value="">All Years</option>
                  <option value="2024">2024</option>
                  <option value="2023">2023</option>
                  <option value="2022">2022</option>
                  <option value="2021">2021</option>
                  <option value="2020">2020</option>
                </select>
              </div>
            </div>
          </div>

          {loading && (
            <div className="bg-white rounded-xl shadow-sm border border-gray-200 p-12">
              <div className="text-center">
                <div className="inline-block animate-spin rounded-full h-10 w-10 border-4 border-blue-200 border-t-blue-600"></div>
                <p className="mt-4 text-sm font-medium text-gray-600">Loading filings...</p>
              </div>
            </div>
          )}

          {error && (
            <div className="bg-white rounded-xl shadow-sm border border-red-200 p-6">
              <div className="flex items-start">
                <div className="flex-shrink-0">
                  <svg className="h-6 w-6 text-red-600" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                    <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M12 8v4m0 4h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z" />
                  </svg>
                </div>
                <div className="ml-3">
                  <h3 className="text-sm font-medium text-red-800">Error loading filings</h3>
                  <p className="mt-1 text-sm text-red-700">{error}</p>
                </div>
              </div>
            </div>
          )}

          {!loading && !error && totalFilings === 0 && (
            <div className="bg-white rounded-xl shadow-sm border border-gray-200 p-12">
              <div className="text-center">
                <div className="mx-auto h-16 w-16 rounded-full bg-blue-100 flex items-center justify-center">
                  <svg className="h-8 w-8 text-blue-600" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                    <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M9 12h6m-6 4h6m2 5H7a2 2 0 01-2-2V5a2 2 0 012-2h5.586a1 1 0 01.707.293l5.414 5.414a1 1 0 01.293.707V19a2 2 0 01-2 2z" />
                  </svg>
                </div>
                <h3 className="mt-4 text-lg font-semibold text-gray-900">No filings found</h3>
                <p className="mt-2 text-sm text-gray-600 max-w-sm mx-auto">
                  {statusFilter || yearFilter ? 'No filings match your filter criteria. Try adjusting your filters.' : 'No filings found in the system.'}
                </p>
              </div>
            </div>
          )}

          {!loading && !error && totalFilings > 0 && (
            <>
              {/* Stats Card */}
              <div className="bg-white rounded-xl shadow-sm border border-gray-200 p-6 mb-6">
                <div className="flex items-center justify-between">
                  <div>
                    <p className="text-sm font-medium text-gray-600">Total Filings</p>
                    <p className="text-3xl font-bold text-gray-900 mt-1">{totalFilings}</p>
                  </div>
                  <div className="h-12 w-12 rounded-full bg-blue-100 flex items-center justify-center">
                    <svg className="h-6 w-6 text-blue-600" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                      <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M9 12h6m-6 4h6m2 5H7a2 2 0 01-2-2V5a2 2 0 012-2h5.586a1 1 0 01.707.293l5.414 5.414a1 1 0 01.293.707V19a2 2 0 01-2 2z" />
                    </svg>
                  </div>
                </div>
              </div>

              {/* Filings List */}
              <div className="space-y-4">
                {filteredClientsData.map((clientData) => (
                  clientData.filings && clientData.filings.map((filing) => (
                    <Link
                      key={filing.id}
                      href={`/${tenantId}/clients/${clientData.client.id}`}
                      className="block bg-white rounded-xl shadow-sm border border-gray-200 hover:shadow-md hover:border-blue-300 transition-all"
                    >
                      <div className="p-6">
                        <div className="flex items-start justify-between">
                          <div className="flex-1">
                            <div className="flex items-center gap-3 mb-2">
                              <h3 className="text-lg font-semibold text-gray-900">
                                {clientData.client.firstName} {clientData.client.lastName}
                              </h3>
                              <span className={`px-3 py-1 rounded-full text-xs font-medium ${filing.status ? getStatusBadge(filing.status.status) : 'bg-gray-100 text-gray-800'}`}>
                                {filing.status?.status || 'Unknown'}
                              </span>
                            </div>
                            <p className="text-sm text-gray-600">{clientData.client.email}</p>
                            <div className="mt-3 flex items-center gap-4 text-sm text-gray-500">
                              <span className="flex items-center gap-1">
                                <svg className="w-4 h-4" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                                  <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M8 7V3m8 4V3m-9 8h10M5 21h14a2 2 0 002-2V7a2 2 0 00-2-2H5a2 2 0 00-2 2v12a2 2 0 002 2z" />
                                </svg>
                                Tax Year: {filing.year}
                              </span>
                              {filing.maritalStatus && (
                                <span>Status: {filing.maritalStatus}</span>
                              )}
                              {filing.income && (
                                <span>Income: ${filing.income.toLocaleString()}</span>
                              )}
                            </div>
                          </div>
                          <div className="ml-4 flex-shrink-0">
                            <svg className="w-5 h-5 text-gray-400" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                              <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M9 5l7 7-7 7" />
                            </svg>
                          </div>
                        </div>
                      </div>
                    </Link>
                  ))
                ))}
              </div>
            </>
          )}
        </div>
      </main>
    </div>
  )
}

export default function FilingsPage() {
  return (
    <ProtectedRoute>
      <FilingsContent />
    </ProtectedRoute>
  )
}
