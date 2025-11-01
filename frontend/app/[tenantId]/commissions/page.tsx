'use client'

import { useEffect, useState } from 'react'
import { useParams } from 'next/navigation'
import Link from 'next/link'
import { ProtectedRoute } from '@/components/ProtectedRoute'
import { UserMenu } from '@/components/UserMenu'
import { TenantSwitcher } from '@/components/TenantSwitcher'
import { useAuth } from '@/contexts/AuthContext'
import { Dialog } from '@/components/Dialog'

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
  approvedAt: string | null
  paidAt: string | null
  customer: {
    id: string
    firstName: string | null
    lastName: string | null
    email: string
  }
}

interface Affiliate {
  id: string
  firstName: string
  lastName: string
  email: string
}

function CommissionsContent() {
  const params = useParams()
  const tenantId = params.tenantId as string
  const { user } = useAuth()

  const [commissions, setCommissions] = useState<Commission[]>([])
  const [affiliates, setAffiliates] = useState<Affiliate[]>([])
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)
  const [statusFilter, setStatusFilter] = useState<string>('')
  const [affiliateIdFilter, setAffiliateIdFilter] = useState<string>('')

  // Dialog management
  const [dialog, setDialog] = useState<{
    isOpen: boolean
    type: 'info' | 'warning' | 'error' | 'success' | 'confirm'
    title: string
    message?: string
    children?: React.ReactNode
    onConfirm?: () => void
    confirmText?: string
    cancelText?: string
    showCancel?: boolean
  }>({
    isOpen: false,
    type: 'info',
    title: '',
  })

  const closeDialog = () => {
    setDialog({ ...dialog, isOpen: false })
  }

  const showAlert = (title: string, message: string, type: 'info' | 'error' | 'success' | 'warning' = 'info') => {
    setDialog({
      isOpen: true,
      type,
      title,
      message,
      showCancel: false,
      confirmText: 'OK',
    })
  }

  const showConfirm = (
    title: string,
    message: string,
    onConfirm: () => void,
    type: 'confirm' | 'warning' = 'confirm'
  ) => {
    setDialog({
      isOpen: true,
      type,
      title,
      message,
      onConfirm,
      confirmText: 'Confirm',
      cancelText: 'Cancel',
      showCancel: true,
    })
  }

  const apiUrl = process.env.NEXT_PUBLIC_API_URL || 'http://localhost:8081'

  // Fetch affiliates for dropdown
  useEffect(() => {
    async function fetchAffiliates() {
      if (!user) return

      try {
        const idToken = await user.getIdToken()
        const response = await fetch(`${apiUrl}/api/v1/${tenantId}/affiliates`, {
          headers: {
            'Authorization': `Bearer ${idToken}`,
            'Content-Type': 'application/json',
          },
        })

        if (response.ok) {
          const data = await response.json()
          setAffiliates(data || [])
        }
      } catch (err) {
        console.error('Failed to fetch affiliates:', err)
      }
    }

    fetchAffiliates()
  }, [tenantId, user, apiUrl])

  useEffect(() => {
    async function fetchCommissions() {
      if (!user) {
        setLoading(false)
        return
      }

      try {
        const idToken = await user.getIdToken()

        // Build URL with optional filters
        let url = `${apiUrl}/api/v1/${tenantId}/commissions?`
        const params = new URLSearchParams()

        if (affiliateIdFilter) {
          params.append('affiliateId', affiliateIdFilter)
        }
        if (statusFilter) {
          params.append('status', statusFilter)
        }

        url += params.toString()

        const response = await fetch(url, {
          headers: {
            'Authorization': `Bearer ${idToken}`,
            'Content-Type': 'application/json',
          },
        })

        if (!response.ok) {
          throw new Error('Failed to fetch commissions')
        }

        const data = await response.json()
        setCommissions(data || [])
      } catch (err) {
        setError(err instanceof Error ? err.message : 'An error occurred')
      } finally {
        setLoading(false)
      }
    }

    fetchCommissions()
  }, [tenantId, user, statusFilter, affiliateIdFilter, apiUrl])

  const approveCommission = async (commissionId: string) => {
    if (!user) return

    showConfirm(
      'Approve Commission',
      'Are you sure you want to approve this commission?',
      async () => {
        try {
          const idToken = await user.getIdToken()
          const response = await fetch(
            `${apiUrl}/api/v1/${tenantId}/commissions/${commissionId}/approve`,
            {
              method: 'PUT',
              headers: {
                'Authorization': `Bearer ${idToken}`,
                'Content-Type': 'application/json',
              },
            }
          )

          if (!response.ok) throw new Error('Failed to approve commission')

          const updatedCommission = await response.json()

          // Update local state
          setCommissions(
            commissions.map((c) => (c.id === commissionId ? updatedCommission : c))
          )
          showAlert('Success', 'Commission approved successfully', 'success')
        } catch (err) {
          showAlert('Approval Failed', err instanceof Error ? err.message : 'Failed to approve commission', 'error')
        }
      },
      'confirm'
    )
  }

  const markCommissionPaid = async (commissionId: string) => {
    if (!user) return

    showConfirm(
      'Mark Commission as Paid',
      'Are you sure you want to mark this commission as PAID?',
      async () => {
        try {
          const idToken = await user.getIdToken()
          const response = await fetch(
            `${apiUrl}/api/v1/${tenantId}/commissions/${commissionId}/mark-paid`,
            {
              method: 'PUT',
              headers: {
                'Authorization': `Bearer ${idToken}`,
                'Content-Type': 'application/json',
              },
            }
          )

          if (!response.ok) throw new Error('Failed to mark commission as paid')

          const updatedCommission = await response.json()

          // Update local state
          setCommissions(
            commissions.map((c) => (c.id === commissionId ? updatedCommission : c))
          )

          showAlert('Success', 'Commission marked as paid', 'success')
        } catch (err) {
          showAlert('Failed', err instanceof Error ? err.message : 'Failed to mark commission as paid', 'error')
        }
      },
      'confirm'
    )
  }

  const [cancelModalOpen, setCancelModalOpen] = useState(false)
  const [commissionToCancel, setCommissionToCancel] = useState<string | null>(null)
  const [cancelReason, setCancelReason] = useState('')

  const openCancelModal = (commissionId: string) => {
    setCommissionToCancel(commissionId)
    setCancelReason('')
    setCancelModalOpen(true)
  }

  const closeCancelModal = () => {
    setCancelModalOpen(false)
    setCommissionToCancel(null)
    setCancelReason('')
  }

  const cancelCommission = async () => {
    if (!user || !commissionToCancel || !cancelReason.trim()) {
      showAlert('Validation Error', 'Cancellation reason is required', 'warning')
      return
    }

    try {
      const idToken = await user.getIdToken()
      const response = await fetch(
        `${apiUrl}/api/v1/${tenantId}/commissions/${commissionToCancel}/cancel`,
        {
          method: 'PUT',
          headers: {
            'Authorization': `Bearer ${idToken}`,
            'Content-Type': 'application/json',
          },
          body: JSON.stringify({ reason: cancelReason }),
        }
      )

      if (!response.ok) throw new Error('Failed to cancel commission')

      const updatedCommission = await response.json()

      // Update local state
      setCommissions(
        commissions.map((c) => (c.id === commissionToCancel ? updatedCommission : c))
      )

      closeCancelModal()
      showAlert('Success', 'Commission cancelled successfully', 'success')
    } catch (err) {
      showAlert('Cancellation Failed', err instanceof Error ? err.message : 'Failed to cancel commission', 'error')
    }
  }

  const getTotalCommissions = () => {
    return commissions.reduce((sum, c) => sum + c.commissionAmount, 0)
  }

  const getCommissionsByStatus = (status: string) => {
    return commissions.filter((c) => c.status === status).length
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
              <span className="text-xs sm:text-sm font-semibold text-green-600">
                Commissions
              </span>
              <TenantSwitcher />
              <UserMenu />
            </div>
          </div>
        </div>
      </nav>

      <main className="max-w-7xl mx-auto py-6 sm:px-6 lg:px-8">
        <div className="px-4 py-6 sm:px-0">
          <div className="mb-6">
            <h2 className="text-2xl font-bold text-gray-900">Commissions</h2>
            <p className="mt-1 text-sm text-gray-500">
              View and manage affiliate commissions
            </p>
          </div>

          {/* Summary Stats */}
          {commissions.length > 0 && (
            <div className="grid grid-cols-1 gap-5 sm:grid-cols-4 mb-6">
              <div className="bg-white overflow-hidden shadow rounded-lg">
                <div className="p-5">
                  <div className="flex items-center">
                    <div className="flex-1">
                      <dt className="text-sm font-medium text-gray-500 truncate">
                        Total Commissions
                      </dt>
                      <dd className="mt-1 text-2xl font-semibold text-gray-900">
                        ${getTotalCommissions().toFixed(2)}
                      </dd>
                    </div>
                  </div>
                </div>
              </div>
              <div className="bg-white overflow-hidden shadow rounded-lg">
                <div className="p-5">
                  <div className="flex items-center">
                    <div className="flex-1">
                      <dt className="text-sm font-medium text-gray-500 truncate">Pending</dt>
                      <dd className="mt-1 text-2xl font-semibold text-yellow-600">
                        {getCommissionsByStatus('PENDING')}
                      </dd>
                    </div>
                  </div>
                </div>
              </div>
              <div className="bg-white overflow-hidden shadow rounded-lg">
                <div className="p-5">
                  <div className="flex items-center">
                    <div className="flex-1">
                      <dt className="text-sm font-medium text-gray-500 truncate">Approved</dt>
                      <dd className="mt-1 text-2xl font-semibold text-blue-600">
                        {getCommissionsByStatus('APPROVED')}
                      </dd>
                    </div>
                  </div>
                </div>
              </div>
              <div className="bg-white overflow-hidden shadow rounded-lg">
                <div className="p-5">
                  <div className="flex items-center">
                    <div className="flex-1">
                      <dt className="text-sm font-medium text-gray-500 truncate">Paid</dt>
                      <dd className="mt-1 text-2xl font-semibold text-green-600">
                        {getCommissionsByStatus('PAID')}
                      </dd>
                    </div>
                  </div>
                </div>
              </div>
            </div>
          )}

          {/* Filters */}
          <div className="bg-white shadow sm:rounded-lg p-4 mb-6">
            <div className="grid grid-cols-1 gap-4 sm:grid-cols-2">
              <div>
                <label className="block text-sm font-medium text-gray-700 mb-1">
                  Affiliate
                </label>
                <select
                  value={affiliateIdFilter}
                  onChange={(e) => setAffiliateIdFilter(e.target.value)}
                  className="w-full px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-blue-500 focus:border-blue-500"
                >
                  <option value="">All Affiliates</option>
                  {affiliates.map((affiliate) => (
                    <option key={affiliate.id} value={affiliate.id}>
                      {affiliate.firstName} {affiliate.lastName} ({affiliate.email})
                    </option>
                  ))}
                </select>
              </div>
              <div>
                <label className="block text-sm font-medium text-gray-700 mb-1">Status</label>
                <select
                  value={statusFilter}
                  onChange={(e) => setStatusFilter(e.target.value)}
                  className="w-full px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-blue-500 focus:border-blue-500"
                >
                  <option value="">All Statuses</option>
                  <option value="PENDING">Pending</option>
                  <option value="APPROVED">Approved</option>
                  <option value="PAID">Paid</option>
                  <option value="CANCELLED">Cancelled</option>
                </select>
              </div>
            </div>
          </div>

          {loading && (
            <div className="text-center py-12">
              <div className="inline-block animate-spin rounded-full h-8 w-8 border-b-2 border-gray-900"></div>
              <p className="mt-2 text-sm text-gray-500">Loading commissions...</p>
            </div>
          )}

          {error && (
            <div className="bg-red-50 border border-red-200 rounded-md p-4">
              <p className="text-sm text-red-800">Error: {error}</p>
            </div>
          )}

          {!loading && !error && commissions.length === 0 && (
            <div className="bg-white shadow overflow-hidden sm:rounded-lg p-6 text-center">
              <p className="text-gray-500">No commissions found</p>
            </div>
          )}

          {!loading && !error && commissions.length > 0 && (
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
                      Discount
                    </th>
                    <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                      Net Amount
                    </th>
                    <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                      Commission
                    </th>
                    <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                      Status
                    </th>
                    <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                      Date
                    </th>
                    <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                      Actions
                    </th>
                  </tr>
                </thead>
                <tbody className="bg-white divide-y divide-gray-200">
                  {commissions.map((commission) => (
                    <tr key={commission.id} className="hover:bg-gray-50">
                      <td className="px-6 py-4 whitespace-nowrap">
                        <div className="text-sm font-medium text-gray-900">
                          {commission.customer
                            ? `${commission.customer.firstName || ''} ${commission.customer.lastName || ''}`.trim() || commission.customer.email
                            : 'N/A'}
                        </div>
                        {commission.customer && (
                          <div className="text-sm text-gray-500">{commission.customer.email}</div>
                        )}
                      </td>
                      <td className="px-6 py-4 whitespace-nowrap text-sm text-gray-900">
                        ${commission.orderAmount.toFixed(2)}
                      </td>
                      <td className="px-6 py-4 whitespace-nowrap text-sm text-gray-500">
                        ${commission.discountAmount.toFixed(2)}
                      </td>
                      <td className="px-6 py-4 whitespace-nowrap text-sm text-gray-900">
                        ${commission.netAmount.toFixed(2)}
                      </td>
                      <td className="px-6 py-4 whitespace-nowrap">
                        <div className="text-sm font-medium text-gray-900">
                          ${commission.commissionAmount.toFixed(2)}
                        </div>
                        <div className="text-sm text-gray-500">
                          {commission.commissionRate.toFixed(1)}%
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
                      <td className="px-6 py-4 whitespace-nowrap text-sm">
                        <div className="flex items-center gap-2">
                          {commission.status === 'PENDING' && (
                            <>
                              <button
                                onClick={() => approveCommission(commission.id)}
                                className="text-blue-600 hover:text-blue-900 font-medium"
                              >
                                Approve
                              </button>
                              <button
                                onClick={() => openCancelModal(commission.id)}
                                className="text-red-600 hover:text-red-900 font-medium"
                              >
                                Cancel
                              </button>
                            </>
                          )}
                          {commission.status === 'APPROVED' && (
                            <>
                              <button
                                onClick={() => markCommissionPaid(commission.id)}
                                className="text-green-600 hover:text-green-900 font-medium"
                              >
                                Mark Paid
                              </button>
                              <button
                                onClick={() => openCancelModal(commission.id)}
                                className="text-red-600 hover:text-red-900 font-medium"
                              >
                                Cancel
                              </button>
                            </>
                          )}
                          {(commission.status === 'PAID' || commission.status === 'CANCELLED') && (
                            <span className="text-gray-400">No actions</span>
                          )}
                        </div>
                      </td>
                    </tr>
                  ))}
                </tbody>
              </table>
            </div>
          )}

          {!loading && !error && commissions.length > 0 && (
            <div className="mt-6">
              <p className="text-sm text-gray-500">
                Showing <span className="font-semibold">{commissions.length}</span> commissions
              </p>
            </div>
          )}
        </div>
      </main>

      {/* Cancel Commission Modal */}
      {cancelModalOpen && (
        <div className="fixed inset-0 bg-gray-600 bg-opacity-50 overflow-y-auto h-full w-full z-50">
          <div className="relative top-20 mx-auto p-5 border w-96 shadow-lg rounded-md bg-white">
            <div className="mt-3">
              <h3 className="text-lg font-medium text-gray-900 mb-4">Cancel Commission</h3>
              <div className="mb-4">
                <label className="block text-sm font-medium text-gray-700 mb-2">
                  Cancellation Reason *
                </label>
                <textarea
                  value={cancelReason}
                  onChange={(e) => setCancelReason(e.target.value)}
                  rows={4}
                  className="w-full px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-blue-500 focus:border-blue-500"
                  placeholder="Enter reason for cancellation..."
                />
              </div>
              <div className="flex gap-3">
                <button
                  onClick={cancelCommission}
                  className="flex-1 bg-red-600 text-white px-4 py-2 rounded-md hover:bg-red-700"
                >
                  Cancel Commission
                </button>
                <button
                  onClick={closeCancelModal}
                  className="flex-1 bg-gray-200 text-gray-700 px-4 py-2 rounded-md hover:bg-gray-300"
                >
                  Close
                </button>
              </div>
            </div>
          </div>
        </div>
      )}

      {/* Dialog Component */}
      <Dialog
        isOpen={dialog.isOpen}
        onClose={closeDialog}
        onConfirm={dialog.onConfirm}
        title={dialog.title}
        message={dialog.message}
        type={dialog.type}
        confirmText={dialog.confirmText}
        cancelText={dialog.cancelText}
        showCancel={dialog.showCancel}
      >
        {dialog.children}
      </Dialog>
    </div>
  )
}

export default function CommissionsPage() {
  return (
    <ProtectedRoute>
      <CommissionsContent />
    </ProtectedRoute>
  )
}
