'use client'

import { useEffect, useState } from 'react'
import { useParams } from 'next/navigation'
import Link from 'next/link'
import { ProtectedRoute } from '@/components/ProtectedRoute'
import { UserMenu } from '@/components/UserMenu'
import { useAuth } from '@/contexts/AuthContext'
import { Dialog } from '@/components/Dialog'

interface Affiliate {
  id: string
  firstName: string
  lastName: string
  email: string
  phone: string | null
  defaultCommissionRate: number
  stripeConnectAccountId: string | null
  payoutMethod: string
  payoutThreshold: number
  isActive: boolean
  createdAt: string
  updatedAt: string | null
}

interface AffiliateToken {
  id: string
  affiliateId: string
  expiresAt: string | null
  lastUsedAt: string | null
  isActive: boolean
  notes: string | null
  createdAt: string
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
  customer: {
    id: string
    firstName: string | null
    lastName: string | null
    email: string
  }
}

interface DiscountCode {
  id: string
  code: string
  description: string | null
  discountType: string
  discountValue: number
  maxUses: number | null
  currentUses: number
  validFrom: string | null
  validUntil: string | null
  isActive: boolean
  isAffiliateCode: boolean
  affiliateId: string
  commissionRate: number | null
  createdAt: string
  updatedAt: string | null
}

function AffiliateDetailContent() {
  const params = useParams()
  const tenantId = params.tenantId as string
  const affiliateId = params.affiliateId as string
  const { user } = useAuth()

  const [affiliate, setAffiliate] = useState<Affiliate | null>(null)
  const [tokens, setTokens] = useState<AffiliateToken[]>([])
  const [commissions, setCommissions] = useState<Commission[]>([])
  const [discountCodes, setDiscountCodes] = useState<DiscountCode[]>([])
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)
  const [showTokenModal, setShowTokenModal] = useState(false)
  const [generatedToken, setGeneratedToken] = useState<string | null>(null)
  const [showCreateCodeModal, setShowCreateCodeModal] = useState(false)
  const [showEditCodeModal, setShowEditCodeModal] = useState(false)
  const [editingCode, setEditingCode] = useState<DiscountCode | null>(null)
  const [codeFormData, setCodeFormData] = useState({
    code: '',
    description: '',
    discountType: 'PERCENTAGE',
    discountValue: '',
    maxUses: '',
    validFrom: '',
    validUntil: '',
    commissionRate: '',
  })
  const [codeSubmitError, setCodeSubmitError] = useState<string | null>(null)
  const [codeSubmitting, setCodeSubmitting] = useState(false)

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

        // Fetch affiliate details
        const affiliateResponse = await fetch(
          `${apiUrl}/api/v1/${tenantId}/affiliates/${affiliateId}`,
          { headers }
        )
        if (!affiliateResponse.ok) throw new Error('Failed to fetch affiliate')
        const affiliateData = await affiliateResponse.json()
        setAffiliate(affiliateData)

        // Fetch tokens
        const tokensResponse = await fetch(
          `${apiUrl}/api/v1/${tenantId}/affiliates/${affiliateId}/tokens`,
          { headers }
        )
        if (tokensResponse.ok) {
          const tokensData = await tokensResponse.json()
          setTokens(tokensData || [])
        }

        // Fetch commissions
        const commissionsResponse = await fetch(
          `${apiUrl}/api/v1/${tenantId}/commissions?affiliateId=${affiliateId}&limit=20`,
          { headers }
        )
        if (commissionsResponse.ok) {
          const commissionsData = await commissionsResponse.json()
          setCommissions(commissionsData || [])
        }

        // Fetch discount codes
        const codesResponse = await fetch(
          `${apiUrl}/api/v1/${tenantId}/discount-codes?affiliateId=${affiliateId}`,
          { headers }
        )
        if (codesResponse.ok) {
          const codesData = await codesResponse.json()
          setDiscountCodes(codesData || [])
        }
      } catch (err) {
        setError(err instanceof Error ? err.message : 'An error occurred')
      } finally {
        setLoading(false)
      }
    }

    fetchData()
  }, [tenantId, affiliateId, user, apiUrl])

  const generateToken = async () => {
    if (!user) return

    try {
      const idToken = await user.getIdToken()
      const response = await fetch(
        `${apiUrl}/api/v1/${tenantId}/affiliates/${affiliateId}/generate-token`,
        {
          method: 'POST',
          headers: {
            'Authorization': `Bearer ${idToken}`,
            'Content-Type': 'application/json',
          },
          body: JSON.stringify({}),
        }
      )

      if (!response.ok) throw new Error('Failed to generate token')

      const data = await response.json()
      setGeneratedToken(data.token)
      setShowTokenModal(true)

      // Refresh tokens list
      const tokensResponse = await fetch(
        `${apiUrl}/api/v1/${tenantId}/affiliates/${affiliateId}/tokens`,
        {
          headers: {
            'Authorization': `Bearer ${idToken}`,
            'Content-Type': 'application/json',
          },
        }
      )
      if (tokensResponse.ok) {
        const tokensData = await tokensResponse.json()
        setTokens(tokensData || [])
      }
    } catch (err) {
      showAlert('Token Generation Failed', err instanceof Error ? err.message : 'Failed to generate token', 'error')
    }
  }

  const revokeToken = async (tokenId: string) => {
    if (!user) return

    showConfirm(
      'Revoke Token',
      'Are you sure you want to revoke this token? This action cannot be undone.',
      async () => {
        try {
          const idToken = await user.getIdToken()
          const response = await fetch(
            `${apiUrl}/api/v1/${tenantId}/affiliates/${affiliateId}/tokens/${tokenId}`,
            {
              method: 'DELETE',
              headers: {
                'Authorization': `Bearer ${idToken}`,
                'Content-Type': 'application/json',
              },
            }
          )

          if (!response.ok) throw new Error('Failed to revoke token')

          // Refresh tokens list
          setTokens(tokens.filter((t) => t.id !== tokenId))
          showAlert('Success', 'Token revoked successfully', 'success')
        } catch (err) {
          showAlert('Revocation Failed', err instanceof Error ? err.message : 'Failed to revoke token', 'error')
        }
      },
      'warning'
    )
  }

  const copyAccessLink = () => {
    if (!generatedToken) return
    const link = `${window.location.origin}/${tenantId}/affiliates/${affiliateId}/dashboard?token=${generatedToken}`
    navigator.clipboard.writeText(link)
    showAlert('Success', 'Access link copied to clipboard!', 'success')
  }

  const createDiscountCode = async (e: React.FormEvent) => {
    e.preventDefault()
    if (!user) return

    setCodeSubmitError(null)
    setCodeSubmitting(true)

    try {
      const idToken = await user.getIdToken()

      const payload: {
        code: string;
        discountType: string;
        discountValue: number;
        affiliateId: string;
        description?: string;
        maxUses?: number;
        validFrom?: string;
        validUntil?: string;
        commissionRate?: number;
      } = {
        code: codeFormData.code.toUpperCase(),
        discountType: codeFormData.discountType,
        discountValue: parseFloat(codeFormData.discountValue),
        affiliateId: affiliateId,
      }

      if (codeFormData.description) payload.description = codeFormData.description
      if (codeFormData.maxUses) payload.maxUses = parseInt(codeFormData.maxUses)
      if (codeFormData.validFrom) payload.validFrom = codeFormData.validFrom + ' 00:00:00'
      if (codeFormData.validUntil) payload.validUntil = codeFormData.validUntil + ' 23:59:59'
      if (codeFormData.commissionRate) payload.commissionRate = parseFloat(codeFormData.commissionRate)

      const response = await fetch(`${apiUrl}/api/v1/${tenantId}/discount-codes`, {
        method: 'POST',
        headers: {
          'Authorization': `Bearer ${idToken}`,
          'Content-Type': 'application/json',
        },
        body: JSON.stringify(payload),
      })

      if (!response.ok) {
        const errorText = await response.text()
        throw new Error(errorText || 'Failed to create discount code')
      }

      const newCode = await response.json()
      setDiscountCodes([newCode, ...discountCodes])
      setShowCreateCodeModal(false)
      setCodeFormData({
        code: '',
        description: '',
        discountType: 'PERCENTAGE',
        discountValue: '',
        maxUses: '',
        validFrom: '',
        validUntil: '',
        commissionRate: '',
      })
    } catch (err) {
      setCodeSubmitError(err instanceof Error ? err.message : 'Failed to create discount code')
    } finally {
      setCodeSubmitting(false)
    }
  }

  const openEditCodeModal = (code: DiscountCode) => {
    setEditingCode(code)
    setCodeFormData({
      code: code.code,
      description: code.description || '',
      discountType: code.discountType,
      discountValue: code.discountValue.toString(),
      maxUses: code.maxUses?.toString() || '',
      validFrom: code.validFrom ? code.validFrom.split(' ')[0] : '',
      validUntil: code.validUntil ? code.validUntil.split(' ')[0] : '',
      commissionRate: code.commissionRate?.toString() || '',
    })
    setShowEditCodeModal(true)
  }

  const updateDiscountCode = async (e: React.FormEvent) => {
    e.preventDefault()
    if (!user || !editingCode) return

    setCodeSubmitError(null)
    setCodeSubmitting(true)

    try {
      const idToken = await user.getIdToken()

      const payload: {
        code: string;
        discountType: string;
        discountValue: number;
        isActive: boolean;
        description?: string;
        maxUses?: number;
        validFrom?: string;
        validUntil?: string;
        commissionRate?: number;
      } = {
        code: codeFormData.code.toUpperCase(),
        discountType: codeFormData.discountType,
        discountValue: parseFloat(codeFormData.discountValue),
        isActive: editingCode.isActive,
      }

      if (codeFormData.description) payload.description = codeFormData.description
      if (codeFormData.maxUses) payload.maxUses = parseInt(codeFormData.maxUses)
      if (codeFormData.validFrom) payload.validFrom = codeFormData.validFrom + ' 00:00:00'
      if (codeFormData.validUntil) payload.validUntil = codeFormData.validUntil + ' 23:59:59'
      if (codeFormData.commissionRate) payload.commissionRate = parseFloat(codeFormData.commissionRate)

      const response = await fetch(
        `${apiUrl}/api/v1/${tenantId}/discount-codes/${editingCode.id}`,
        {
          method: 'PUT',
          headers: {
            'Authorization': `Bearer ${idToken}`,
            'Content-Type': 'application/json',
          },
          body: JSON.stringify(payload),
        }
      )

      if (!response.ok) {
        const errorText = await response.text()
        throw new Error(errorText || 'Failed to update discount code')
      }

      const updatedCode = await response.json()
      setDiscountCodes(discountCodes.map((c) => (c.id === updatedCode.id ? updatedCode : c)))
      setShowEditCodeModal(false)
      setEditingCode(null)
      setCodeFormData({
        code: '',
        description: '',
        discountType: 'PERCENTAGE',
        discountValue: '',
        maxUses: '',
        validFrom: '',
        validUntil: '',
        commissionRate: '',
      })
    } catch (err) {
      setCodeSubmitError(err instanceof Error ? err.message : 'Failed to update discount code')
    } finally {
      setCodeSubmitting(false)
    }
  }

  const deactivateDiscountCode = async (codeId: string) => {
    if (!user) return

    showConfirm(
      'Deactivate Discount Code',
      'Are you sure you want to deactivate this discount code?',
      async () => {
        try {
          const idToken = await user.getIdToken()
          const response = await fetch(
            `${apiUrl}/api/v1/${tenantId}/discount-codes/${codeId}/deactivate`,
            {
              method: 'PUT',
              headers: {
                'Authorization': `Bearer ${idToken}`,
                'Content-Type': 'application/json',
              },
            }
          )

          if (!response.ok) throw new Error('Failed to deactivate discount code')

          // Update the code in state
          setDiscountCodes(
            discountCodes.map((c) => (c.id === codeId ? { ...c, isActive: false } : c))
          )
          showAlert('Success', 'Discount code deactivated successfully', 'success')
        } catch (err) {
          showAlert('Deactivation Failed', err instanceof Error ? err.message : 'Failed to deactivate discount code', 'error')
        }
      },
      'warning'
    )
  }

  if (loading) {
    return (
      <div className="min-h-screen bg-gradient-to-br from-slate-50 via-purple-50 to-slate-50 flex items-center justify-center">
        <div className="bg-white rounded-xl shadow-sm border border-gray-200 p-12">
          <div className="text-center">
            <div className="inline-block animate-spin rounded-full h-10 w-10 border-4 border-purple-200 border-t-purple-600"></div>
            <p className="mt-4 text-sm font-medium text-gray-600">Loading affiliate...</p>
          </div>
        </div>
      </div>
    )
  }

  if (error || !affiliate) {
    return (
      <div className="min-h-screen bg-gradient-to-br from-slate-50 via-purple-50 to-slate-50 flex items-center justify-center">
        <div className="bg-white rounded-xl shadow-sm border border-red-200 p-6">
          <div className="flex items-start">
            <div className="flex-shrink-0">
              <svg className="h-6 w-6 text-red-600" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M12 8v4m0 4h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z" />
              </svg>
            </div>
            <div className="ml-3">
              <h3 className="text-sm font-medium text-red-800">Error loading affiliate</h3>
              <p className="mt-1 text-sm text-red-700">{error || 'Affiliate not found'}</p>
            </div>
          </div>
        </div>
      </div>
    )
  }

  return (
    <div className="min-h-screen bg-gradient-to-br from-slate-50 via-purple-50 to-slate-50">
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
                href={`/${tenantId}/affiliates`}
                className="text-xs sm:text-sm text-gray-600 hover:text-gray-900"
              >
                Affiliates
              </Link>
              <span className="text-gray-300">|</span>
              <span className="text-xs sm:text-sm font-semibold text-purple-600">
                Affiliate Details
              </span>
              <div className="hidden lg:block text-xs text-gray-500 border-l pl-4 ml-4">
                <span className="font-mono">{tenantId}</span>
              </div>
              <UserMenu />
            </div>
          </div>
        </div>
      </nav>

      <main className="max-w-7xl mx-auto py-8 sm:px-6 lg:px-8">
        <div className="px-4 sm:px-0">
          {/* Affiliate Header */}
          <div className="mb-8">
            <div className="flex items-center gap-4 mb-2">
              <div className="w-16 h-16 rounded-full bg-gradient-to-br from-purple-500 to-purple-600 flex items-center justify-center text-white text-xl font-bold">
                {affiliate.firstName?.charAt(0)}{affiliate.lastName?.charAt(0)}
              </div>
              <div>
                <h2 className="text-2xl sm:text-3xl font-bold text-gray-900">
                  {affiliate.firstName} {affiliate.lastName}
                </h2>
                <p className="mt-1 text-sm text-gray-600">{affiliate.email}</p>
              </div>
            </div>
          </div>

          {/* Affiliate Details */}
          <div className="bg-white shadow-sm border border-gray-200 overflow-hidden rounded-xl mb-6">
            <div className="px-4 py-5 sm:px-6 flex justify-between items-center">
              <h3 className="text-lg leading-6 font-medium text-gray-900">
                Affiliate Details
              </h3>
              <span
                className={`px-3 py-1 inline-flex text-sm leading-5 font-semibold rounded-full ${
                  affiliate.isActive
                    ? 'bg-green-100 text-green-800'
                    : 'bg-red-100 text-red-800'
                }`}
              >
                {affiliate.isActive ? 'Active' : 'Inactive'}
              </span>
            </div>
            <div className="border-t border-gray-200">
              <dl>
                <div className="bg-gray-50 px-4 py-5 sm:grid sm:grid-cols-3 sm:gap-4 sm:px-6">
                  <dt className="text-sm font-medium text-gray-500">Full Name</dt>
                  <dd className="mt-1 text-sm text-gray-900 sm:mt-0 sm:col-span-2">
                    {affiliate.firstName} {affiliate.lastName}
                  </dd>
                </div>
                <div className="bg-white px-4 py-5 sm:grid sm:grid-cols-3 sm:gap-4 sm:px-6">
                  <dt className="text-sm font-medium text-gray-500">Email</dt>
                  <dd className="mt-1 text-sm text-gray-900 sm:mt-0 sm:col-span-2">
                    {affiliate.email}
                  </dd>
                </div>
                <div className="bg-gray-50 px-4 py-5 sm:grid sm:grid-cols-3 sm:gap-4 sm:px-6">
                  <dt className="text-sm font-medium text-gray-500">Phone</dt>
                  <dd className="mt-1 text-sm text-gray-900 sm:mt-0 sm:col-span-2">
                    {affiliate.phone || '-'}
                  </dd>
                </div>
                <div className="bg-white px-4 py-5 sm:grid sm:grid-cols-3 sm:gap-4 sm:px-6">
                  <dt className="text-sm font-medium text-gray-500">Default Commission Rate</dt>
                  <dd className="mt-1 text-sm text-gray-900 sm:mt-0 sm:col-span-2">
                    {affiliate.defaultCommissionRate.toFixed(1)}%
                  </dd>
                </div>
                <div className="bg-gray-50 px-4 py-5 sm:grid sm:grid-cols-3 sm:gap-4 sm:px-6">
                  <dt className="text-sm font-medium text-gray-500">Payout Method</dt>
                  <dd className="mt-1 text-sm text-gray-900 sm:mt-0 sm:col-span-2">
                    {affiliate.payoutMethod}
                  </dd>
                </div>
                <div className="bg-white px-4 py-5 sm:grid sm:grid-cols-3 sm:gap-4 sm:px-6">
                  <dt className="text-sm font-medium text-gray-500">Payout Threshold</dt>
                  <dd className="mt-1 text-sm text-gray-900 sm:mt-0 sm:col-span-2">
                    ${affiliate.payoutThreshold.toFixed(2)}
                  </dd>
                </div>
                <div className="bg-gray-50 px-4 py-5 sm:grid sm:grid-cols-3 sm:gap-4 sm:px-6">
                  <dt className="text-sm font-medium text-gray-500">Created</dt>
                  <dd className="mt-1 text-sm text-gray-900 sm:mt-0 sm:col-span-2">
                    {new Date(affiliate.createdAt).toLocaleString()}
                  </dd>
                </div>
              </dl>
            </div>
          </div>

          {/* Access Tokens */}
          <div className="bg-white shadow-sm border border-gray-200 overflow-hidden rounded-xl mb-6">
            <div className="px-4 py-5 sm:px-6 flex justify-between items-center">
              <div>
                <h3 className="text-lg leading-6 font-medium text-gray-900">
                  Access Tokens
                </h3>
                <p className="mt-1 text-sm text-gray-500">
                  Generate secure links for affiliate dashboard access
                </p>
              </div>
              <button
                onClick={generateToken}
                className="px-4 py-2 bg-blue-600 text-white text-sm font-medium rounded-md hover:bg-blue-700"
              >
                Generate Token
              </button>
            </div>
            <div className="border-t border-gray-200">
              {tokens.length === 0 ? (
                <div className="px-4 py-5 text-center text-sm text-gray-500">
                  No tokens generated yet
                </div>
              ) : (
                <table className="min-w-full divide-y divide-gray-200">
                  <thead className="bg-gray-50">
                    <tr>
                      <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase">
                        Created
                      </th>
                      <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase">
                        Last Used
                      </th>
                      <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase">
                        Expires
                      </th>
                      <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase">
                        Status
                      </th>
                      <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase">
                        Actions
                      </th>
                    </tr>
                  </thead>
                  <tbody className="bg-white divide-y divide-gray-200">
                    {tokens.map((token) => (
                      <tr key={token.id}>
                        <td className="px-6 py-4 whitespace-nowrap text-sm text-gray-900">
                          {new Date(token.createdAt).toLocaleDateString()}
                        </td>
                        <td className="px-6 py-4 whitespace-nowrap text-sm text-gray-500">
                          {token.lastUsedAt
                            ? new Date(token.lastUsedAt).toLocaleDateString()
                            : 'Never'}
                        </td>
                        <td className="px-6 py-4 whitespace-nowrap text-sm text-gray-500">
                          {token.expiresAt
                            ? new Date(token.expiresAt).toLocaleDateString()
                            : 'Never'}
                        </td>
                        <td className="px-6 py-4 whitespace-nowrap">
                          <span
                            className={`px-2 inline-flex text-xs leading-5 font-semibold rounded-full ${
                              token.isActive
                                ? 'bg-green-100 text-green-800'
                                : 'bg-red-100 text-red-800'
                            }`}
                          >
                            {token.isActive ? 'Active' : 'Revoked'}
                          </span>
                        </td>
                        <td className="px-6 py-4 whitespace-nowrap text-sm">
                          {token.isActive && (
                            <button
                              onClick={() => revokeToken(token.id)}
                              className="text-red-600 hover:text-red-900"
                            >
                              Revoke
                            </button>
                          )}
                        </td>
                      </tr>
                    ))}
                  </tbody>
                </table>
              )}
            </div>
          </div>

          {/* Discount Codes */}
          <div className="bg-white shadow-sm border border-gray-200 overflow-hidden rounded-xl mb-6">
            <div className="px-4 py-5 sm:px-6 flex justify-between items-center">
              <div>
                <h3 className="text-lg leading-6 font-medium text-gray-900">
                  Discount Codes
                </h3>
                <p className="mt-1 text-sm text-gray-500">
                  Manage discount codes and track customer usage
                </p>
              </div>
              <button
                onClick={() => setShowCreateCodeModal(true)}
                className="px-4 py-2 bg-blue-600 text-white text-sm font-medium rounded-md hover:bg-blue-700"
              >
                Create Code
              </button>
            </div>
            <div className="border-t border-gray-200">
              {discountCodes.length === 0 ? (
                <div className="px-4 py-5 text-center text-sm text-gray-500">
                  No discount codes created yet
                </div>
              ) : (
                <table className="min-w-full divide-y divide-gray-200">
                  <thead className="bg-gray-50">
                    <tr>
                      <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase">
                        Code
                      </th>
                      <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase">
                        Discount
                      </th>
                      <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase">
                        Commission
                      </th>
                      <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase">
                        Usage
                      </th>
                      <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase">
                        Status
                      </th>
                      <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase">
                        Actions
                      </th>
                    </tr>
                  </thead>
                  <tbody className="bg-white divide-y divide-gray-200">
                    {discountCodes.map((code) => (
                      <tr key={code.id}>
                        <td className="px-6 py-4 whitespace-nowrap">
                          <div className="text-sm font-medium text-gray-900">{code.code}</div>
                          {code.description && (
                            <div className="text-xs text-gray-500">{code.description}</div>
                          )}
                        </td>
                        <td className="px-6 py-4 whitespace-nowrap text-sm text-gray-900">
                          {code.discountType === 'PERCENTAGE'
                            ? `${code.discountValue}% off`
                            : `$${code.discountValue} off`}
                        </td>
                        <td className="px-6 py-4 whitespace-nowrap text-sm text-gray-900">
                          {code.commissionRate ? `${code.commissionRate.toFixed(1)}%` : '-'}
                        </td>
                        <td className="px-6 py-4 whitespace-nowrap text-sm text-gray-500">
                          {code.currentUses}
                          {code.maxUses ? ` / ${code.maxUses}` : ' (unlimited)'}
                        </td>
                        <td className="px-6 py-4 whitespace-nowrap">
                          <span
                            className={`px-2 inline-flex text-xs leading-5 font-semibold rounded-full ${
                              code.isActive
                                ? 'bg-green-100 text-green-800'
                                : 'bg-red-100 text-red-800'
                            }`}
                          >
                            {code.isActive ? 'Active' : 'Inactive'}
                          </span>
                        </td>
                        <td className="px-6 py-4 whitespace-nowrap text-sm space-x-2">
                          <button
                            onClick={() => openEditCodeModal(code)}
                            className="text-blue-600 hover:text-blue-900"
                          >
                            Edit
                          </button>
                          {code.isActive && (
                            <button
                              onClick={() => deactivateDiscountCode(code.id)}
                              className="text-red-600 hover:text-red-900"
                            >
                              Deactivate
                            </button>
                          )}
                        </td>
                      </tr>
                    ))}
                  </tbody>
                </table>
              )}
            </div>
          </div>

          {/* Recent Commissions */}
          <div className="bg-white shadow-sm border border-gray-200 overflow-hidden rounded-xl">
            <div className="px-4 py-5 sm:px-6">
              <h3 className="text-lg leading-6 font-medium text-gray-900">
                Recent Commissions
              </h3>
              <p className="mt-1 text-sm text-gray-500">Last 20 commissions</p>
            </div>
            <div className="border-t border-gray-200">
              {commissions.length === 0 ? (
                <div className="px-4 py-5 text-center text-sm text-gray-500">
                  No commissions yet
                </div>
              ) : (
                <table className="min-w-full divide-y divide-gray-200">
                  <thead className="bg-gray-50">
                    <tr>
                      <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase">
                        Customer
                      </th>
                      <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase">
                        Order Amount
                      </th>
                      <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase">
                        Commission
                      </th>
                      <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase">
                        Status
                      </th>
                      <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase">
                        Date
                      </th>
                    </tr>
                  </thead>
                  <tbody className="bg-white divide-y divide-gray-200">
                    {commissions.map((commission) => (
                      <tr key={commission.id}>
                        <td className="px-6 py-4 whitespace-nowrap text-sm text-gray-900">
                          {commission.customer.firstName} {commission.customer.lastName}
                        </td>
                        <td className="px-6 py-4 whitespace-nowrap text-sm text-gray-900">
                          ${commission.orderAmount.toFixed(2)}
                        </td>
                        <td className="px-6 py-4 whitespace-nowrap text-sm font-medium text-gray-900">
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
                        <td className="px-6 py-4 whitespace-nowrap text-sm text-gray-500">
                          {new Date(commission.createdAt).toLocaleDateString()}
                        </td>
                      </tr>
                    ))}
                  </tbody>
                </table>
              )}
            </div>
          </div>
        </div>
      </main>

      {/* Token Generated Modal */}
      {showTokenModal && generatedToken && (
        <div className="fixed z-10 inset-0 overflow-y-auto">
          <div className="flex items-center justify-center min-h-screen px-4">
            <div className="fixed inset-0 bg-gray-500 bg-opacity-75 transition-opacity"></div>
            <div className="relative bg-white rounded-lg max-w-2xl w-full p-6">
              <h3 className="text-lg font-medium text-gray-900 mb-4">
                Access Token Generated
              </h3>
              <p className="text-sm text-gray-500 mb-4">
                Share this link with the affiliate to access their dashboard. This is the only time the full link will be displayed.
              </p>
              <div className="bg-gray-50 p-4 rounded-md mb-2">
                <p className="text-xs text-gray-600 mb-1">Dashboard URL:</p>
                <p className="text-xs font-mono break-all text-blue-600">
                  {`${window.location.origin}/${tenantId}/affiliates/${affiliateId}/dashboard?token=${generatedToken}`}
                </p>
              </div>
              <div className="bg-yellow-50 border border-yellow-200 rounded-md p-3 mb-4">
                <p className="text-xs text-yellow-800">
                  <strong>Important:</strong> Keep this link secure. Anyone with this link can access the affiliate&apos;s dashboard.
                </p>
              </div>
              <div className="flex gap-2">
                <button
                  onClick={copyAccessLink}
                  className="flex-1 px-4 py-2 bg-blue-600 text-white text-sm font-medium rounded-md hover:bg-blue-700"
                >
                  Copy Dashboard Link
                </button>
                <button
                  onClick={() => {
                    setShowTokenModal(false)
                    setGeneratedToken(null)
                  }}
                  className="flex-1 px-4 py-2 bg-gray-200 text-gray-700 text-sm font-medium rounded-md hover:bg-gray-300"
                >
                  Close
                </button>
              </div>
            </div>
          </div>
        </div>
      )}

      {/* Create Discount Code Modal */}
      {showCreateCodeModal && (
        <div className="fixed z-10 inset-0 overflow-y-auto">
          <div className="flex items-center justify-center min-h-screen px-4">
            <div
              className="fixed inset-0 bg-gray-500 bg-opacity-75 transition-opacity"
              onClick={() => setShowCreateCodeModal(false)}
            ></div>
            <div className="relative bg-white rounded-lg max-w-lg w-full p-6">
              <h3 className="text-lg font-medium text-gray-900 mb-4">Create Discount Code</h3>

              {codeSubmitError && (
                <div className="mb-4 bg-red-50 border border-red-200 rounded-md p-3">
                  <p className="text-sm text-red-800">{codeSubmitError}</p>
                </div>
              )}

              <form onSubmit={createDiscountCode}>
                <div className="space-y-4">
                  <div>
                    <label className="block text-sm font-medium text-gray-700 mb-1">
                      Code <span className="text-red-500">*</span>
                    </label>
                    <input
                      type="text"
                      required
                      value={codeFormData.code}
                      onChange={(e) =>
                        setCodeFormData({ ...codeFormData, code: e.target.value.toUpperCase() })
                      }
                      placeholder="e.g., SAVE15"
                      className="w-full px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-blue-500 focus:border-blue-500"
                    />
                  </div>

                  <div>
                    <label className="block text-sm font-medium text-gray-700 mb-1">
                      Description
                    </label>
                    <input
                      type="text"
                      value={codeFormData.description}
                      onChange={(e) =>
                        setCodeFormData({ ...codeFormData, description: e.target.value })
                      }
                      placeholder="e.g., Summer promotion"
                      className="w-full px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-blue-500 focus:border-blue-500"
                    />
                  </div>

                  <div className="grid grid-cols-2 gap-4">
                    <div>
                      <label className="block text-sm font-medium text-gray-700 mb-1">
                        Discount Type <span className="text-red-500">*</span>
                      </label>
                      <select
                        value={codeFormData.discountType}
                        onChange={(e) =>
                          setCodeFormData({ ...codeFormData, discountType: e.target.value })
                        }
                        className="w-full px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-blue-500 focus:border-blue-500"
                      >
                        <option value="PERCENTAGE">Percentage</option>
                        <option value="FIXED_AMOUNT">Fixed Amount</option>
                      </select>
                    </div>
                    <div>
                      <label className="block text-sm font-medium text-gray-700 mb-1">
                        Discount Value <span className="text-red-500">*</span>
                      </label>
                      <input
                        type="number"
                        step="0.01"
                        min="0"
                        required
                        value={codeFormData.discountValue}
                        onChange={(e) =>
                          setCodeFormData({ ...codeFormData, discountValue: e.target.value })
                        }
                        placeholder={codeFormData.discountType === 'PERCENTAGE' ? '15' : '50'}
                        className="w-full px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-blue-500 focus:border-blue-500"
                      />
                    </div>
                  </div>

                  <div>
                    <label className="block text-sm font-medium text-gray-700 mb-1">
                      Commission Rate (%)
                    </label>
                    <input
                      type="number"
                      step="0.1"
                      min="0"
                      max="100"
                      value={codeFormData.commissionRate}
                      onChange={(e) =>
                        setCodeFormData({ ...codeFormData, commissionRate: e.target.value })
                      }
                      placeholder={`Default: ${affiliate?.defaultCommissionRate || 0}%`}
                      className="w-full px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-blue-500 focus:border-blue-500"
                    />
                    <p className="mt-1 text-xs text-gray-500">
                      Leave empty to use affiliate&apos;s default commission rate
                    </p>
                  </div>

                  <div>
                    <label className="block text-sm font-medium text-gray-700 mb-1">
                      Max Uses
                    </label>
                    <input
                      type="number"
                      min="1"
                      value={codeFormData.maxUses}
                      onChange={(e) =>
                        setCodeFormData({ ...codeFormData, maxUses: e.target.value })
                      }
                      placeholder="Leave empty for unlimited"
                      className="w-full px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-blue-500 focus:border-blue-500"
                    />
                  </div>

                  <div className="grid grid-cols-2 gap-4">
                    <div>
                      <label className="block text-sm font-medium text-gray-700 mb-1">
                        Valid From
                      </label>
                      <input
                        type="date"
                        value={codeFormData.validFrom}
                        onChange={(e) =>
                          setCodeFormData({ ...codeFormData, validFrom: e.target.value })
                        }
                        className="w-full px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-blue-500 focus:border-blue-500"
                      />
                    </div>
                    <div>
                      <label className="block text-sm font-medium text-gray-700 mb-1">
                        Valid Until
                      </label>
                      <input
                        type="date"
                        value={codeFormData.validUntil}
                        onChange={(e) =>
                          setCodeFormData({ ...codeFormData, validUntil: e.target.value })
                        }
                        className="w-full px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-blue-500 focus:border-blue-500"
                      />
                    </div>
                  </div>
                </div>

                <div className="mt-6 flex gap-3">
                  <button
                    type="submit"
                    disabled={codeSubmitting}
                    className="flex-1 px-4 py-2 bg-blue-600 text-white text-sm font-medium rounded-md hover:bg-blue-700 disabled:opacity-50 disabled:cursor-not-allowed"
                  >
                    {codeSubmitting ? 'Creating...' : 'Create Code'}
                  </button>
                  <button
                    type="button"
                    onClick={() => setShowCreateCodeModal(false)}
                    disabled={codeSubmitting}
                    className="flex-1 px-4 py-2 bg-gray-200 text-gray-700 text-sm font-medium rounded-md hover:bg-gray-300 disabled:opacity-50"
                  >
                    Cancel
                  </button>
                </div>
              </form>
            </div>
          </div>
        </div>
      )}

      {/* Edit Discount Code Modal */}
      {showEditCodeModal && editingCode && (
        <div className="fixed z-10 inset-0 overflow-y-auto">
          <div className="flex items-center justify-center min-h-screen px-4">
            <div
              className="fixed inset-0 bg-gray-500 bg-opacity-75 transition-opacity"
              onClick={() => setShowEditCodeModal(false)}
            ></div>
            <div className="relative bg-white rounded-lg max-w-lg w-full p-6">
              <h3 className="text-lg font-medium text-gray-900 mb-4">Edit Discount Code</h3>

              {codeSubmitError && (
                <div className="mb-4 bg-red-50 border border-red-200 rounded-md p-3">
                  <p className="text-sm text-red-800">{codeSubmitError}</p>
                </div>
              )}

              <form onSubmit={updateDiscountCode}>
                <div className="space-y-4">
                  <div>
                    <label className="block text-sm font-medium text-gray-700 mb-1">
                      Code <span className="text-red-500">*</span>
                    </label>
                    <input
                      type="text"
                      required
                      value={codeFormData.code}
                      onChange={(e) =>
                        setCodeFormData({ ...codeFormData, code: e.target.value.toUpperCase() })
                      }
                      className="w-full px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-blue-500 focus:border-blue-500"
                    />
                  </div>

                  <div>
                    <label className="block text-sm font-medium text-gray-700 mb-1">
                      Description
                    </label>
                    <input
                      type="text"
                      value={codeFormData.description}
                      onChange={(e) =>
                        setCodeFormData({ ...codeFormData, description: e.target.value })
                      }
                      className="w-full px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-blue-500 focus:border-blue-500"
                    />
                  </div>

                  <div className="grid grid-cols-2 gap-4">
                    <div>
                      <label className="block text-sm font-medium text-gray-700 mb-1">
                        Discount Type <span className="text-red-500">*</span>
                      </label>
                      <select
                        value={codeFormData.discountType}
                        onChange={(e) =>
                          setCodeFormData({ ...codeFormData, discountType: e.target.value })
                        }
                        className="w-full px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-blue-500 focus:border-blue-500"
                      >
                        <option value="PERCENTAGE">Percentage</option>
                        <option value="FIXED_AMOUNT">Fixed Amount</option>
                      </select>
                    </div>
                    <div>
                      <label className="block text-sm font-medium text-gray-700 mb-1">
                        Discount Value <span className="text-red-500">*</span>
                      </label>
                      <input
                        type="number"
                        step="0.01"
                        min="0"
                        required
                        value={codeFormData.discountValue}
                        onChange={(e) =>
                          setCodeFormData({ ...codeFormData, discountValue: e.target.value })
                        }
                        className="w-full px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-blue-500 focus:border-blue-500"
                      />
                    </div>
                  </div>

                  <div>
                    <label className="block text-sm font-medium text-gray-700 mb-1">
                      Commission Rate (%)
                    </label>
                    <input
                      type="number"
                      step="0.1"
                      min="0"
                      max="100"
                      value={codeFormData.commissionRate}
                      onChange={(e) =>
                        setCodeFormData({ ...codeFormData, commissionRate: e.target.value })
                      }
                      className="w-full px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-blue-500 focus:border-blue-500"
                    />
                  </div>

                  <div>
                    <label className="block text-sm font-medium text-gray-700 mb-1">
                      Max Uses
                    </label>
                    <input
                      type="number"
                      min="1"
                      value={codeFormData.maxUses}
                      onChange={(e) =>
                        setCodeFormData({ ...codeFormData, maxUses: e.target.value })
                      }
                      className="w-full px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-blue-500 focus:border-blue-500"
                    />
                  </div>

                  <div className="grid grid-cols-2 gap-4">
                    <div>
                      <label className="block text-sm font-medium text-gray-700 mb-1">
                        Valid From
                      </label>
                      <input
                        type="date"
                        value={codeFormData.validFrom}
                        onChange={(e) =>
                          setCodeFormData({ ...codeFormData, validFrom: e.target.value })
                        }
                        className="w-full px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-blue-500 focus:border-blue-500"
                      />
                    </div>
                    <div>
                      <label className="block text-sm font-medium text-gray-700 mb-1">
                        Valid Until
                      </label>
                      <input
                        type="date"
                        value={codeFormData.validUntil}
                        onChange={(e) =>
                          setCodeFormData({ ...codeFormData, validUntil: e.target.value })
                        }
                        className="w-full px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-blue-500 focus:border-blue-500"
                      />
                    </div>
                  </div>

                  <div className="bg-gray-50 p-3 rounded-md">
                    <p className="text-sm text-gray-700">
                      <strong>Current Uses:</strong> {editingCode.currentUses}
                      {editingCode.maxUses && ` / ${editingCode.maxUses}`}
                    </p>
                    <p className="text-sm text-gray-700 mt-1">
                      <strong>Status:</strong>{' '}
                      <span
                        className={
                          editingCode.isActive ? 'text-green-600' : 'text-red-600'
                        }
                      >
                        {editingCode.isActive ? 'Active' : 'Inactive'}
                      </span>
                    </p>
                  </div>
                </div>

                <div className="mt-6 flex gap-3">
                  <button
                    type="submit"
                    disabled={codeSubmitting}
                    className="flex-1 px-4 py-2 bg-blue-600 text-white text-sm font-medium rounded-md hover:bg-blue-700 disabled:opacity-50 disabled:cursor-not-allowed"
                  >
                    {codeSubmitting ? 'Updating...' : 'Update Code'}
                  </button>
                  <button
                    type="button"
                    onClick={() => {
                      setShowEditCodeModal(false)
                      setEditingCode(null)
                    }}
                    disabled={codeSubmitting}
                    className="flex-1 px-4 py-2 bg-gray-200 text-gray-700 text-sm font-medium rounded-md hover:bg-gray-300 disabled:opacity-50"
                  >
                    Cancel
                  </button>
                </div>
              </form>
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

export default function AffiliateDetailPage() {
  return (
    <ProtectedRoute>
      <AffiliateDetailContent />
    </ProtectedRoute>
  )
}
