'use client'

import { useEffect, useState } from 'react'
import { useParams } from 'next/navigation'
import Link from 'next/link'
import { ProtectedRoute } from '@/components/ProtectedRoute'
import { UserMenu } from '@/components/UserMenu'
import { TenantSwitcher } from '@/components/TenantSwitcher'
import { useAuth } from '@/contexts/AuthContext'
import { Dialog } from '@/components/Dialog'

interface ClientData {
  client: {
    id: string
    firstName: string | null
    middleName: string | null
    lastName: string | null
    email: string
    phone: string | null
    dob: string | null
    ssn: string | null
    address1: string | null
    address2: string | null
    city: string | null
    state: string | null
    zipcode: number | null
    role: string
    createdAt: string
  }
  spouse: {
    id: string
    userId: string
    firstName: string
    middleName: string | null
    lastName: string
    email: string | null
    phone: string | null
    dob: string
    ssn: string
    isDeath: boolean
    deathDate: string | null
    createdAt: string
  } | null
  dependents: Array<{
    id: string
    userId: string
    firstName: string
    middleName: string | null
    lastName: string
    dob: string
    ssn: string
    relationship: string
    timeWithApplicant: string
    exclusiveClaim: boolean
    documents?: string[]
    createdAt: string
    updatedAt: string | null
  }>
  filings: Array<{
    id: string
    year: number
    userId: string
    maritalStatus: string | null
    spouseId: string | null
    sourceOfIncome: string[]
    deductions: string[]
    income: number | null
    marketplaceInsurance: boolean | null
    createdAt: string
    updatedAt: string | null
    status?: {
      id: string
      filingId: string
      latestStep: number
      isCompleted: boolean
      status: string
    }
    documents?: Array<{
      id: string
      name: string
      type: string
      createdAt: string
    }>
    properties?: Array<{
      id: string
      address1: string
      address2: string | null
      city: string
      state: string
      zipcode: string
      purchasePrice: number
      closingCost: number
      purchaseDate: string
      rents: number | null
      royalties: number | null
      expenses?: Array<{
        id: string
        name: string
        amount: number
      }>
    }>
    iraContributions?: Array<{
      id: string
      accountType: string
      amount: number
    }>
    charities?: Array<{
      id: string
      name: string
      contribution: number
    }>
    childcares?: Array<{
      id: string
      name: string
      amount: number
      taxId: string
      address1: string
      address2: string | null
      city: string
      state: string
      zipcode: string
    }>
    payments?: Array<{
      id: string
      amount: number
      originalAmount: number | null
      discountAmount: number | null
      discountCode: string | null
      status: string
      createdAt: string
      items?: Array<{
        id: string
        name: string
        quantity: number
        unitAmount: number
      }>
    }>
    discounts?: Array<{
      id: string
      code: string | null
      originalAmount: number
      discountAmount: number
      finalAmount: number
      appliedAt: string
    }>
  }>
}

function ClientDetailContent() {
  const params = useParams()
  const tenantId = params.tenantId as string
  const clientId = params.clientId as string
  const { user } = useAuth()

  const [clientData, setClientData] = useState<ClientData | null>(null)
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)
  const [expandedFilings, setExpandedFilings] = useState<Set<string>>(new Set())

  // Document management state
  const [uploadingDoc, setUploadingDoc] = useState<string | null>(null) // filingId currently uploading to
  const [selectedFile, setSelectedFile] = useState<File | null>(null)
  const [selectedDocType, setSelectedDocType] = useState<string>('W2')

  // Filing status management
  const [completingFiling, setCompletingFiling] = useState<string | null>(null) // filingId being marked as completed

  // Signature management
  const [signatureModalOpen, setSignatureModalOpen] = useState(false)
  const [selectedFilingForSignature, setSelectedFilingForSignature] = useState<string | null>(null)
  const [sendingSignature, setSendingSignature] = useState(false)

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
    message: string | React.ReactNode,
    onConfirm: () => void,
    type: 'confirm' | 'warning' = 'confirm'
  ) => {
    setDialog({
      isOpen: true,
      type,
      title,
      message: typeof message === 'string' ? message : undefined,
      children: typeof message !== 'string' ? message : undefined,
      onConfirm,
      confirmText: 'Confirm',
      cancelText: 'Cancel',
      showCancel: true,
    })
  }

  const toggleFiling = (filingId: string) => {
    setExpandedFilings(prev => {
      const next = new Set(prev)
      if (next.has(filingId)) {
        next.delete(filingId)
      } else {
        next.add(filingId)
      }
      return next
    })
  }

  // Document management functions
  const handleUploadDocument = async (filingId: string) => {
    if (!selectedFile || !user) {
      showAlert('No File Selected', 'Please select a file to upload', 'warning')
      return
    }

    setUploadingDoc(filingId)
    try {
      const apiUrl = process.env.NEXT_PUBLIC_API_URL || 'http://localhost:8081'
      const idToken = await user.getIdToken()

      const formData = new FormData()
      formData.append('file', selectedFile)
      formData.append('type', selectedDocType)
      formData.append('userId', clientData?.client.id || '')

      const response = await fetch(
        `${apiUrl}/api/v1/${tenantId}/filings/${filingId}/documents`,
        {
          method: 'POST',
          headers: {
            'Authorization': `Bearer ${idToken}`,
          },
          body: formData,
        }
      )

      if (!response.ok) {
        throw new Error('Failed to upload document')
      }

      const newDoc = await response.json()

      // Update client data with new document
      setClientData(prev => {
        if (!prev) return prev
        return {
          ...prev,
          filings: prev.filings.map(f =>
            f.id === filingId
              ? { ...f, documents: [...(f.documents || []), newDoc] }
              : f
          )
        }
      })

      setSelectedFile(null)
      showAlert('Success', 'Document uploaded successfully', 'success')
    } catch (err) {
      showAlert('Upload Failed', err instanceof Error ? err.message : 'Failed to upload document', 'error')
    } finally {
      setUploadingDoc(null)
    }
  }

  const handleDownloadDocument = async (documentId: string) => {
    if (!user) return

    try {
      const apiUrl = process.env.NEXT_PUBLIC_API_URL || 'http://localhost:8081'
      const idToken = await user.getIdToken()

      const response = await fetch(
        `${apiUrl}/api/v1/${tenantId}/documents/${documentId}/download`,
        {
          headers: {
            'Authorization': `Bearer ${idToken}`,
          },
        }
      )

      if (!response.ok) {
        throw new Error('Failed to get download URL')
      }

      const data = await response.json()
      window.open(data.url, '_blank')
    } catch (err) {
      showAlert('Download Failed', err instanceof Error ? err.message : 'Failed to download document', 'error')
    }
  }

  const handleDeleteDocument = async (filingId: string, documentId: string, documentName: string) => {
    if (!user) return

    showConfirm(
      'Delete Document',
      `Are you sure you want to delete "${documentName}"? This action cannot be undone.`,
      async () => {
        try {
          const apiUrl = process.env.NEXT_PUBLIC_API_URL || 'http://localhost:8081'
          const idToken = await user.getIdToken()

          const response = await fetch(
            `${apiUrl}/api/v1/${tenantId}/documents/${documentId}`,
            {
              method: 'DELETE',
              headers: {
                'Authorization': `Bearer ${idToken}`,
              },
            }
          )

          if (!response.ok) {
            throw new Error('Failed to delete document')
          }

          // Update client data by removing deleted document
          setClientData(prev => {
            if (!prev) return prev
            return {
              ...prev,
              filings: prev.filings.map(f =>
                f.id === filingId
                  ? { ...f, documents: (f.documents || []).filter(d => d.id !== documentId) }
                  : f
              )
            }
          })

          showAlert('Success', 'Document deleted successfully', 'success')
        } catch (err) {
          showAlert('Delete Failed', err instanceof Error ? err.message : 'Failed to delete document', 'error')
        }
      },
      'warning'
    )
  }

  // Send signature request
  const handleSendSignatureRequest = async (signatureData: {
    pdfPath: string;
    taxPayerEmail: string;
    taxPayerName: string;
    taxPayerSsn: string;
    spouseName: string;
    spouseEmail: string;
    grossIncome: number;
    totalTax: number;
    taxWithHeld: number;
    refund: number;
    owed: number;
    spouseSignature: boolean;
  }) => {
    if (!user || !clientData || !selectedFilingForSignature) return

    setSendingSignature(true)
    try {
      const apiUrl = process.env.NEXT_PUBLIC_API_URL || 'http://localhost:8081'
      const idToken = await user.getIdToken()

      const response = await fetch(
        `${apiUrl}/api/v1/${tenantId}/signature/send`,
        {
          method: 'POST',
          headers: {
            'Authorization': `Bearer ${idToken}`,
            'Content-Type': 'application/json',
          },
          body: JSON.stringify(signatureData),
        }
      )

      if (!response.ok) {
        throw new Error('Failed to send signature request')
      }

      await response.json()
      showAlert('Success', 'Signature request sent successfully', 'success')
      setSignatureModalOpen(false)
      setSelectedFilingForSignature(null)
    } catch (err) {
      showAlert('Failed', err instanceof Error ? err.message : 'Failed to send signature request', 'error')
    } finally {
      setSendingSignature(false)
    }
  }

  // Mark filing as completed
  const handleMarkFilingCompleted = async (filingId: string) => {
    if (!user) return

    const checklistContent = (
      <div className="mt-3">
        <p className="text-sm text-gray-700 mb-3">Before marking this filing as completed, please confirm:</p>
        <ul className="space-y-2 text-sm text-gray-600">
          <li className="flex items-start">
            <svg className="h-5 w-5 text-green-500 mr-2 flex-shrink-0" fill="none" viewBox="0 0 24 24" stroke="currentColor">
              <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M5 13l4 4L19 7" />
            </svg>
            <span>Filing documents are attached</span>
          </li>
          <li className="flex items-start">
            <svg className="h-5 w-5 text-green-500 mr-2 flex-shrink-0" fill="none" viewBox="0 0 24 24" stroke="currentColor">
              <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M5 13l4 4L19 7" />
            </svg>
            <span>Signature document is attached</span>
          </li>
          <li className="flex items-start">
            <svg className="h-5 w-5 text-green-500 mr-2 flex-shrink-0" fill="none" viewBox="0 0 24 24" stroke="currentColor">
              <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M5 13l4 4L19 7" />
            </svg>
            <span>Payment has been completed</span>
          </li>
        </ul>
        <p className="text-sm text-gray-700 mt-4 font-medium">Are you sure you want to mark this filing as COMPLETED?</p>
      </div>
    )

    showConfirm(
      'Complete Filing',
      checklistContent,
      async () => {
        setCompletingFiling(filingId)
        try {
          const apiUrl = process.env.NEXT_PUBLIC_API_URL || 'http://localhost:8081'
          const idToken = await user.getIdToken()

          const response = await fetch(
            `${apiUrl}/api/v1/${tenantId}/filings/${filingId}/complete`,
            {
              method: 'PUT',
              headers: {
                'Authorization': `Bearer ${idToken}`,
              },
            }
          )

          if (!response.ok) {
            throw new Error('Failed to mark filing as completed')
          }

          // Update client data to reflect new status
          setClientData(prev => {
            if (!prev) return prev
            return {
              ...prev,
              filings: prev.filings.map(f =>
                f.id === filingId && f.status
                  ? { ...f, status: { ...f.status, isCompleted: true, status: 'COMPLETED' } }
                  : f
              )
            }
          })

          showAlert('Success', 'Filing marked as completed successfully', 'success')
        } catch (err) {
          showAlert('Failed', err instanceof Error ? err.message : 'Failed to mark filing as completed', 'error')
        } finally {
          setCompletingFiling(null)
        }
      },
      'confirm'
    )
  }

  useEffect(() => {
    async function fetchClientData() {
      if (!user) {
        setLoading(false)
        return
      }

      try {
        const apiUrl = process.env.NEXT_PUBLIC_API_URL || 'http://localhost:8081'
        const idToken = await user.getIdToken()

        const response = await fetch(`${apiUrl}/api/v1/${tenantId}/clients/${clientId}/comprehensive`, {
          headers: {
            'Authorization': `Bearer ${idToken}`,
            'Content-Type': 'application/json',
          },
        })

        if (!response.ok) {
          throw new Error('Failed to fetch client data')
        }

        const data = await response.json()
        setClientData(data)
      } catch (err) {
        setError(err instanceof Error ? err.message : 'An error occurred')
      } finally {
        setLoading(false)
      }
    }

    fetchClientData()
  }, [tenantId, clientId, user])

  if (loading) {
    return (
      <div className="min-h-screen bg-gray-50 flex items-center justify-center">
        <div className="inline-block animate-spin rounded-full h-8 w-8 border-b-2 border-gray-900"></div>
      </div>
    )
  }

  if (error || !clientData) {
    return (
      <div className="min-h-screen bg-gray-50 p-6">
        <div className="bg-red-50 border border-red-200 rounded-md p-4">
          <p className="text-sm text-red-800">Error: {error || 'Client not found'}</p>
        </div>
      </div>
    )
  }

  const { client, spouse, dependents, filings } = clientData

  return (
    <div className="min-h-screen bg-gradient-to-br from-slate-50 via-blue-50 to-slate-50">
      {/* Navigation */}
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
                Client Details
              </span>
              <TenantSwitcher />
              <UserMenu />
            </div>
          </div>
        </div>
      </nav>

      <main className="max-w-7xl mx-auto py-6 sm:px-6 lg:px-8">
        <div className="px-4 py-6 sm:px-0">
          {/* Client Header */}
          <div className="bg-white shadow overflow-hidden sm:rounded-lg mb-6">
            <div className="px-4 py-5 sm:px-6">
              <h3 className="text-base sm:text-lg leading-6 font-medium text-gray-900">Client Information</h3>
              <p className="mt-1 text-xs sm:text-sm text-gray-500">Primary taxpayer details</p>
            </div>
            <div className="border-t border-gray-200">
              <dl>
                <div className="bg-gray-50 px-4 py-5 sm:grid sm:grid-cols-3 sm:gap-4 sm:px-6">
                  <dt className="text-sm font-bold text-gray-600">Full Name</dt>
                  <dd className="mt-1 text-sm text-gray-900 sm:mt-0 sm:col-span-2">
                    {client.firstName} {client.middleName && `${client.middleName} `}{client.lastName}
                  </dd>
                </div>
                <div className="bg-white px-4 py-5 sm:grid sm:grid-cols-3 sm:gap-4 sm:px-6">
                  <dt className="text-sm font-bold text-gray-600">Email</dt>
                  <dd className="mt-1 text-sm text-gray-900 sm:mt-0 sm:col-span-2">{client.email}</dd>
                </div>
                <div className="bg-gray-50 px-4 py-5 sm:grid sm:grid-cols-3 sm:gap-4 sm:px-6">
                  <dt className="text-sm font-bold text-gray-600">Phone</dt>
                  <dd className="mt-1 text-sm text-gray-900 sm:mt-0 sm:col-span-2">{client.phone || '-'}</dd>
                </div>
                <div className="bg-white px-4 py-5 sm:grid sm:grid-cols-3 sm:gap-4 sm:px-6">
                  <dt className="text-sm font-bold text-gray-600">Date of Birth</dt>
                  <dd className="mt-1 text-sm text-gray-900 sm:mt-0 sm:col-span-2">
                    {client.dob ? new Date(client.dob).toLocaleDateString() : '-'}
                  </dd>
                </div>
                <div className="bg-gray-50 px-4 py-5 sm:grid sm:grid-cols-3 sm:gap-4 sm:px-6">
                  <dt className="text-sm font-bold text-gray-600">SSN</dt>
                  <dd className="mt-1 text-sm text-gray-900 sm:mt-0 sm:col-span-2">{client.ssn || '-'}</dd>
                </div>
                <div className="bg-white px-4 py-5 sm:grid sm:grid-cols-3 sm:gap-4 sm:px-6">
                  <dt className="text-sm font-bold text-gray-600">Address Line 1</dt>
                  <dd className="mt-1 text-sm text-gray-900 sm:mt-0 sm:col-span-2">{client.address1 || '-'}</dd>
                </div>
                {client.address2 && (
                  <div className="bg-gray-50 px-4 py-5 sm:grid sm:grid-cols-3 sm:gap-4 sm:px-6">
                    <dt className="text-sm font-bold text-gray-600">Address Line 2</dt>
                    <dd className="mt-1 text-sm text-gray-900 sm:mt-0 sm:col-span-2">{client.address2}</dd>
                  </div>
                )}
                <div className="bg-gray-50 px-4 py-5 sm:grid sm:grid-cols-3 sm:gap-4 sm:px-6">
                  <dt className="text-sm font-bold text-gray-600">City</dt>
                  <dd className="mt-1 text-sm text-gray-900 sm:mt-0 sm:col-span-2">{client.city || '-'}</dd>
                </div>
                <div className="bg-white px-4 py-5 sm:grid sm:grid-cols-3 sm:gap-4 sm:px-6">
                  <dt className="text-sm font-bold text-gray-600">State</dt>
                  <dd className="mt-1 text-sm text-gray-900 sm:mt-0 sm:col-span-2">{client.state || '-'}</dd>
                </div>
                <div className="bg-gray-50 px-4 py-5 sm:grid sm:grid-cols-3 sm:gap-4 sm:px-6">
                  <dt className="text-sm font-bold text-gray-600">Zip Code</dt>
                  <dd className="mt-1 text-sm text-gray-900 sm:mt-0 sm:col-span-2">{client.zipcode || '-'}</dd>
                </div>
                <div className="bg-white px-4 py-5 sm:grid sm:grid-cols-3 sm:gap-4 sm:px-6">
                  <dt className="text-sm font-bold text-gray-600">Role</dt>
                  <dd className="mt-1 text-sm text-gray-900 sm:mt-0 sm:col-span-2">{client.role}</dd>
                </div>
                <div className="bg-gray-50 px-4 py-5 sm:grid sm:grid-cols-3 sm:gap-4 sm:px-6">
                  <dt className="text-sm font-bold text-gray-600">Record Created</dt>
                  <dd className="mt-1 text-sm text-gray-900 sm:mt-0 sm:col-span-2">
                    {new Date(client.createdAt).toLocaleDateString()}
                  </dd>
                </div>
              </dl>
            </div>
          </div>

          {/* Spouse */}
          {spouse && (
            <div className="bg-white shadow overflow-hidden sm:rounded-lg mb-6">
              <div className="px-4 py-5 sm:px-6">
                <h3 className="text-base sm:text-lg leading-6 font-medium text-gray-900">Spouse Information</h3>
                <p className="mt-1 text-xs sm:text-sm text-gray-500">Spouse details</p>
              </div>
              <div className="border-t border-gray-200">
                <dl>
                  <div className="bg-gray-50 px-4 py-5 sm:grid sm:grid-cols-3 sm:gap-4 sm:px-6">
                    <dt className="text-sm font-bold text-gray-600">Full Name</dt>
                    <dd className="mt-1 text-sm text-gray-900 sm:mt-0 sm:col-span-2">
                      {spouse.firstName} {spouse.middleName && `${spouse.middleName} `}{spouse.lastName}
                    </dd>
                  </div>
                  <div className="bg-white px-4 py-5 sm:grid sm:grid-cols-3 sm:gap-4 sm:px-6">
                    <dt className="text-sm font-bold text-gray-600">Email</dt>
                    <dd className="mt-1 text-sm text-gray-900 sm:mt-0 sm:col-span-2">{spouse.email || '-'}</dd>
                  </div>
                  <div className="bg-gray-50 px-4 py-5 sm:grid sm:grid-cols-3 sm:gap-4 sm:px-6">
                    <dt className="text-sm font-bold text-gray-600">Phone</dt>
                    <dd className="mt-1 text-sm text-gray-900 sm:mt-0 sm:col-span-2">{spouse.phone || '-'}</dd>
                  </div>
                  <div className="bg-white px-4 py-5 sm:grid sm:grid-cols-3 sm:gap-4 sm:px-6">
                    <dt className="text-sm font-bold text-gray-600">Date of Birth</dt>
                    <dd className="mt-1 text-sm text-gray-900 sm:mt-0 sm:col-span-2">{new Date(spouse.dob).toLocaleDateString()}</dd>
                  </div>
                  <div className="bg-gray-50 px-4 py-5 sm:grid sm:grid-cols-3 sm:gap-4 sm:px-6">
                    <dt className="text-sm font-bold text-gray-600">SSN</dt>
                    <dd className="mt-1 text-sm text-gray-900 sm:mt-0 sm:col-span-2">{spouse.ssn}</dd>
                  </div>
                  {spouse.isDeath && (
                    <>
                      <div className="bg-white px-4 py-5 sm:grid sm:grid-cols-3 sm:gap-4 sm:px-6">
                        <dt className="text-sm font-bold text-gray-600">Status</dt>
                        <dd className="mt-1 text-sm text-gray-900 sm:mt-0 sm:col-span-2">Deceased</dd>
                      </div>
                      {spouse.deathDate && (
                        <div className="bg-gray-50 px-4 py-5 sm:grid sm:grid-cols-3 sm:gap-4 sm:px-6">
                          <dt className="text-sm font-bold text-gray-600">Date of Death</dt>
                          <dd className="mt-1 text-sm text-gray-900 sm:mt-0 sm:col-span-2">{new Date(spouse.deathDate).toLocaleDateString()}</dd>
                        </div>
                      )}
                    </>
                  )}
                  <div className="bg-white px-4 py-5 sm:grid sm:grid-cols-3 sm:gap-4 sm:px-6">
                    <dt className="text-sm font-bold text-gray-600">Record Created</dt>
                    <dd className="mt-1 text-sm text-gray-900 sm:mt-0 sm:col-span-2">{new Date(spouse.createdAt).toLocaleDateString()}</dd>
                  </div>
                </dl>
              </div>
            </div>
          )}

          {/* Dependents */}
          {dependents && dependents.length > 0 && (
            <div className="bg-white shadow overflow-hidden sm:rounded-lg mb-6">
              <div className="px-4 py-5 sm:px-6">
                <h3 className="text-base sm:text-lg leading-6 font-medium text-gray-900">Dependents ({dependents.length})</h3>
                <p className="mt-1 text-xs sm:text-sm text-gray-500">Dependent information</p>
              </div>
              <div className="border-t border-gray-200 divide-y divide-gray-200">
                {dependents.map((dep, idx) => (
                  <div key={dep.id}>
                    <div className="px-4 py-3 bg-gray-50">
                      <h4 className="text-sm font-semibold text-gray-900">
                        Dependent {idx + 1}
                      </h4>
                    </div>
                    <dl>
                      <div className="bg-white px-4 py-5 sm:grid sm:grid-cols-3 sm:gap-4 sm:px-6">
                        <dt className="text-sm font-bold text-gray-600">Full Name</dt>
                        <dd className="mt-1 text-sm text-gray-900 sm:mt-0 sm:col-span-2">
                          {dep.firstName} {dep.middleName && `${dep.middleName} `}{dep.lastName}
                        </dd>
                      </div>
                      <div className="bg-gray-50 px-4 py-5 sm:grid sm:grid-cols-3 sm:gap-4 sm:px-6">
                        <dt className="text-sm font-bold text-gray-600">Relationship</dt>
                        <dd className="mt-1 text-sm text-gray-900 sm:mt-0 sm:col-span-2">{dep.relationship}</dd>
                      </div>
                      <div className="bg-white px-4 py-5 sm:grid sm:grid-cols-3 sm:gap-4 sm:px-6">
                        <dt className="text-sm font-bold text-gray-600">Date of Birth</dt>
                        <dd className="mt-1 text-sm text-gray-900 sm:mt-0 sm:col-span-2">{new Date(dep.dob).toLocaleDateString()}</dd>
                      </div>
                      <div className="bg-gray-50 px-4 py-5 sm:grid sm:grid-cols-3 sm:gap-4 sm:px-6">
                        <dt className="text-sm font-bold text-gray-600">SSN</dt>
                        <dd className="mt-1 text-sm text-gray-900 sm:mt-0 sm:col-span-2">{dep.ssn}</dd>
                      </div>
                      <div className="bg-white px-4 py-5 sm:grid sm:grid-cols-3 sm:gap-4 sm:px-6">
                        <dt className="text-sm font-bold text-gray-600">Time With Applicant</dt>
                        <dd className="mt-1 text-sm text-gray-900 sm:mt-0 sm:col-span-2">{dep.timeWithApplicant}</dd>
                      </div>
                      <div className="bg-gray-50 px-4 py-5 sm:grid sm:grid-cols-3 sm:gap-4 sm:px-6">
                        <dt className="text-sm font-bold text-gray-600">Exclusive Claim</dt>
                        <dd className="mt-1 text-sm text-gray-900 sm:mt-0 sm:col-span-2">{dep.exclusiveClaim ? 'Yes' : 'No'}</dd>
                      </div>
                      <div className="bg-white px-4 py-5 sm:grid sm:grid-cols-3 sm:gap-4 sm:px-6">
                        <dt className="text-sm font-bold text-gray-600">Record Created</dt>
                        <dd className="mt-1 text-sm text-gray-900 sm:mt-0 sm:col-span-2">{new Date(dep.createdAt).toLocaleDateString()}</dd>
                      </div>
                      {dep.updatedAt && (
                        <div className="bg-gray-50 px-4 py-5 sm:grid sm:grid-cols-3 sm:gap-4 sm:px-6">
                          <dt className="text-sm font-bold text-gray-600">Last Updated</dt>
                          <dd className="mt-1 text-sm text-gray-900 sm:mt-0 sm:col-span-2">{new Date(dep.updatedAt).toLocaleDateString()}</dd>
                        </div>
                      )}
                      {dep.documents && dep.documents.length > 0 && (
                        <div className="bg-white px-4 py-5 sm:grid sm:grid-cols-3 sm:gap-4 sm:px-6 border-t-2 border-blue-100">
                          <dt className="text-sm font-bold text-gray-600">Required Documentation</dt>
                          <dd className="mt-1 text-sm text-gray-900 sm:mt-0 sm:col-span-2">
                            <ul className="list-disc list-inside space-y-1">
                              {dep.documents.map((doc, docIdx) => (
                                <li key={docIdx} className="text-gray-700">
                                  {doc === 'healthcareProvider' ? 'Healthcare provider statement' :
                                   doc === 'socialServiceRecords' ? 'Social service records' :
                                   doc === 'medicalRecords' ? 'Medical records' :
                                   doc === 'schoolRecords' ? 'School records' :
                                   doc === 'childcareRecords' ? 'Childcare records' :
                                   doc === 'birthCertificate' ? 'Birth certificate' :
                                   doc === 'adoptionPapers' ? 'Adoption papers' :
                                   doc === 'courtOrder' ? 'Court order' :
                                   doc}
                                </li>
                              ))}
                            </ul>
                          </dd>
                        </div>
                      )}
                    </dl>
                  </div>
                ))}
              </div>
            </div>
          )}

          {/* Tax Filings */}
          <div className="bg-white shadow overflow-hidden sm:rounded-lg">
            <div className="px-4 py-5 sm:px-6 bg-gradient-to-r from-blue-50 to-indigo-50">
              <h3 className="text-base sm:text-lg leading-6 font-medium text-gray-900">
                Tax Filings ({filings ? filings.length : 0})
              </h3>
              <p className="mt-1 text-xs sm:text-sm text-gray-500">Complete filing history with all related data</p>
            </div>
            {filings && filings.length > 0 ? (
              <div className="border-t border-gray-200 divide-y divide-gray-200">
                {filings.map((filing) => {
                  const isExpanded = expandedFilings.has(filing.id)
                  const totalDocuments = filing.documents?.length || 0
                  const totalProperties = filing.properties?.length || 0
                  const totalPayments = filing.payments?.length || 0

                  return (
                    <div key={filing.id} className="hover:bg-gray-50 transition-colors">
                      {/* Collapsed View - Summary */}
                      <div
                        className="p-4 sm:p-6 cursor-pointer"
                        onClick={() => toggleFiling(filing.id)}
                      >
                        <div className="flex flex-col space-y-4 lg:flex-row lg:items-center lg:justify-between lg:space-y-0">
                          <div className="flex items-center space-x-3 sm:space-x-4">
                            <div className="flex-shrink-0">
                              <div className="h-10 w-10 sm:h-12 sm:w-12 rounded-lg bg-gradient-to-br from-blue-500 to-indigo-600 flex items-center justify-center text-white font-bold text-base sm:text-lg">
                                {filing.year}
                              </div>
                            </div>
                            <div className="min-w-0 flex-1">
                              <h4 className="text-base sm:text-lg font-semibold text-gray-900">Tax Year {filing.year}</h4>
                              <div className="flex flex-wrap items-center gap-x-2 sm:gap-x-4 gap-y-1 mt-1">
                                <span className="text-xs sm:text-sm text-gray-600 whitespace-nowrap">
                                  Income: <span className="font-medium text-gray-900">${filing.income?.toLocaleString() || 0}</span>
                                </span>
                                <span className="hidden sm:inline text-sm text-gray-500">•</span>
                                <span className="text-xs sm:text-sm text-gray-600 whitespace-nowrap">
                                  {filing.maritalStatus || 'Status Unknown'}
                                </span>
                                {totalDocuments > 0 && (
                                  <>
                                    <span className="hidden sm:inline text-sm text-gray-500">•</span>
                                    <span className="text-xs sm:text-sm text-gray-600 whitespace-nowrap">{totalDocuments} doc{totalDocuments !== 1 ? 's' : ''}</span>
                                  </>
                                )}
                              </div>
                            </div>
                          </div>
                          <div className="flex flex-wrap items-center gap-2 sm:gap-3">
                            <span className={`px-2 sm:px-3 py-1 text-xs font-semibold rounded-full whitespace-nowrap ${filing.status?.isCompleted
                                ? 'bg-green-100 text-green-800'
                                : 'bg-yellow-100 text-yellow-800'
                              }`}>
                              {filing.status?.status || 'IN_PROCESS'}
                            </span>
                            <button
                              onClick={(e) => {
                                e.stopPropagation()
                                setSelectedFilingForSignature(filing.id)
                                setSignatureModalOpen(true)
                              }}
                              className="px-2 sm:px-3 py-1 text-xs font-medium text-white bg-purple-600 rounded-md hover:bg-purple-700 flex items-center space-x-1 whitespace-nowrap"
                            >
                              <svg className="w-3 h-3 sm:w-4 sm:h-4" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                                <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M15.232 5.232l3.536 3.536m-2.036-5.036a2.5 2.5 0 113.536 3.536L6.5 21.036H3v-3.572L16.732 3.732z" />
                              </svg>
                              <span className="hidden sm:inline">Send Signature</span>
                              <span className="sm:hidden">Signature</span>
                            </button>
                            {!filing.status?.isCompleted && (
                              <button
                                onClick={(e) => {
                                  e.stopPropagation()
                                  handleMarkFilingCompleted(filing.id)
                                }}
                                disabled={completingFiling === filing.id}
                                className="px-2 sm:px-3 py-1 text-xs font-medium text-white bg-blue-600 rounded-md hover:bg-blue-700 disabled:bg-gray-400 disabled:cursor-not-allowed whitespace-nowrap"
                              >
                                {completingFiling === filing.id ? 'Marking...' : 'Mark Completed'}
                              </button>
                            )}
                            <svg
                              className={`h-5 w-5 sm:h-6 sm:w-6 text-gray-400 transition-transform ${isExpanded ? 'rotate-180' : ''}`}
                              fill="none"
                              viewBox="0 0 24 24"
                              stroke="currentColor"
                            >
                              <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M19 9l-7 7-7-7" />
                            </svg>
                          </div>
                        </div>
                      </div>

                      {/* Expanded View - Full Details */}
                      {isExpanded && (
                        <div className="px-6 pb-6 space-y-6 border-t border-gray-100">

                          {/* Quick Stats */}
                          <div className="grid grid-cols-2 md:grid-cols-4 gap-4 pt-4">
                            <div className="bg-gray-50 rounded-lg p-4 border border-gray-200">
                              <p className="text-xs font-bold text-gray-600 uppercase">Documents</p>
                              <p className="text-2xl font-bold text-gray-900 mt-1">{totalDocuments}</p>
                            </div>
                            <div className="bg-gray-50 rounded-lg p-4 border border-gray-200">
                              <p className="text-xs font-bold text-gray-600 uppercase">Payments</p>
                              <p className="text-2xl font-bold text-gray-900 mt-1">{totalPayments}</p>
                            </div>
                            <div className="bg-gray-50 rounded-lg p-4 border border-gray-200">
                              <p className="text-xs font-bold text-gray-600 uppercase">Properties</p>
                              <p className="text-2xl font-bold text-gray-900 mt-1">{totalProperties}</p>
                            </div>
                            <div className="bg-gray-50 rounded-lg p-4 border border-gray-200">
                              <p className="text-xs font-bold text-gray-600 uppercase">Progress</p>
                              <p className="text-2xl font-bold text-gray-900 mt-1">Step {filing.status?.latestStep || 0}</p>
                            </div>
                          </div>

                          {/* Filing Details */}
                          <div className="bg-gray-50 rounded-lg p-5">
                            <h5 className="text-sm font-semibold text-gray-900 mb-4 flex items-center">
                              <svg className="w-5 h-5 mr-2 text-gray-600" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                                <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M9 12h6m-6 4h6m2 5H7a2 2 0 01-2-2V5a2 2 0 012-2h5.586a1 1 0 01.707.293l5.414 5.414a1 1 0 01.293.707V19a2 2 0 01-2 2z" />
                              </svg>
                              Filing Information
                            </h5>
                            <div className="grid grid-cols-2 md:grid-cols-3 gap-4">
                              <div>
                                <p className="text-xs font-bold text-gray-600 uppercase">Total Income</p>
                                <p className="text-sm text-gray-900 mt-1">${filing.income?.toLocaleString() || 0}</p>
                              </div>
                              <div>
                                <p className="text-xs font-bold text-gray-600 uppercase">Marital Status</p>
                                <p className="text-sm text-gray-900 mt-1">{filing.maritalStatus || '-'}</p>
                              </div>
                              <div>
                                <p className="text-xs font-bold text-gray-600 uppercase">Marketplace Insurance</p>
                                <p className="text-sm text-gray-900 mt-1">{filing.marketplaceInsurance ? 'Yes' : 'No'}</p>
                              </div>
                              <div>
                                <p className="text-xs font-bold text-gray-600 uppercase">Filing Created</p>
                                <p className="text-sm text-gray-900 mt-1">{new Date(filing.createdAt).toLocaleDateString()}</p>
                              </div>
                              {filing.updatedAt && (
                                <div>
                                  <p className="text-xs font-bold text-gray-600 uppercase">Last Updated</p>
                                  <p className="text-sm text-gray-900 mt-1">{new Date(filing.updatedAt).toLocaleDateString()}</p>
                                </div>
                              )}
                              {filing.status && (
                                <div>
                                  <p className="text-xs font-bold text-gray-600 uppercase">Completed</p>
                                  <p className="text-sm text-gray-900 mt-1">{filing.status.isCompleted ? 'Yes' : 'In Progress'}</p>
                                </div>
                              )}
                            </div>
                          </div>

                          {/* Source of Income & Deductions */}
                          {((filing.sourceOfIncome && filing.sourceOfIncome.length > 0) || (filing.deductions && filing.deductions.length > 0)) && (
                            <div className="bg-white rounded-lg border border-gray-200 p-5">
                              <h5 className="text-sm font-semibold text-gray-900 mb-4 flex items-center">
                                <svg className="w-5 h-5 mr-2 text-gray-600" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                                  <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M12 8c-1.657 0-3 .895-3 2s1.343 2 3 2 3 .895 3 2-1.343 2-3 2m0-8c1.11 0 2.08.402 2.599 1M12 8V7m0 1v8m0 0v1m0-1c-1.11 0-2.08-.402-2.599-1M21 12a9 9 0 11-18 0 9 9 0 0118 0z" />
                                </svg>
                                Income & Deductions
                              </h5>
                              <div className="space-y-4">
                                {filing.sourceOfIncome && filing.sourceOfIncome.length > 0 && (
                                  <div>
                                    <p className="text-xs font-bold text-gray-600 uppercase mb-2">Sources of Income</p>
                                    <div className="flex flex-wrap gap-2">
                                      {filing.sourceOfIncome.map((source, idx) => (
                                        <span key={idx} className="px-3 py-1 bg-gray-100 text-gray-800 text-sm rounded-full">
                                          {source}
                                        </span>
                                      ))}
                                    </div>
                                  </div>
                                )}
                                {filing.deductions && filing.deductions.length > 0 && (
                                  <div>
                                    <p className="text-xs font-bold text-gray-600 uppercase mb-2">Deductions</p>
                                    <div className="flex flex-wrap gap-2">
                                      {filing.deductions.map((deduction, idx) => (
                                        <span key={idx} className="px-3 py-1 bg-gray-100 text-gray-800 text-sm rounded-full">
                                          {deduction}
                                        </span>
                                      ))}
                                    </div>
                                  </div>
                                )}
                              </div>
                            </div>
                          )}

                          {/* Documents */}
                          <div className="bg-white rounded-lg border border-gray-200">
                            <div className="px-5 py-4 border-b border-gray-200 bg-gray-50">
                              <h5 className="text-sm font-semibold text-gray-900 flex items-center">
                                <svg className="w-5 h-5 mr-2 text-gray-600" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                                  <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M7 21h10a2 2 0 002-2V9.414a1 1 0 00-.293-.707l-5.414-5.414A1 1 0 0012.586 3H7a2 2 0 00-2 2v14a2 2 0 002 2z" />
                                </svg>
                                Documents ({filing.documents?.length || 0})
                              </h5>
                            </div>

                            {/* Upload Form */}
                            <div className="p-5 bg-gray-50 border-b border-gray-200">
                              <h6 className="text-xs font-semibold text-gray-700 uppercase mb-3">Upload New Document</h6>
                              <div className="flex flex-col sm:flex-row gap-3">
                                <div className="flex-1">
                                  <label className="block text-xs font-medium text-gray-700 mb-1">Document Type</label>
                                  <select
                                    value={selectedDocType}
                                    onChange={(e) => setSelectedDocType(e.target.value)}
                                    className="w-full px-3 py-2 text-sm border border-gray-300 rounded-md focus:outline-none focus:ring-blue-500 focus:border-blue-500"
                                  >
                                    <option value="W2">W-2</option>
                                    <option value="1099-NEC">1099-NEC</option>
                                    <option value="1099-MISC">1099-MISC</option>
                                    <option value="1099-INT">1099-INT</option>
                                    <option value="1099-DIV">1099-DIV</option>
                                    <option value="1099-B">1099-B</option>
                                    <option value="1099-R">1099-R</option>
                                    <option value="1099-K">1099-K</option>
                                    <option value="1098-E">1098-E (Student Loan Interest)</option>
                                    <option value="1098-T">1098-T (Tuition)</option>
                                    <option value="1098-MTG">1098 (Mortgage Interest)</option>
                                    <option value="1099-S">1099-S (Real Estate)</option>
                                    <option value="1095-A">1095-A (Health Insurance)</option>
                                    <option value="SSA-1099">SSA-1099 (Social Security)</option>
                                    <option value="Schedule C">Schedule C (Business)</option>
                                    <option value="Schedule E">Schedule E (Rental)</option>
                                    <option value="OTHER">Other</option>
                                  </select>
                                </div>
                                <div className="flex-1">
                                  <label className="block text-xs font-medium text-gray-700 mb-1">File</label>
                                  <input
                                    type="file"
                                    onChange={(e) => setSelectedFile(e.target.files?.[0] || null)}
                                    accept=".pdf,.jpg,.jpeg,.png"
                                    className="w-full px-3 py-2 text-sm border border-gray-300 rounded-md focus:outline-none focus:ring-blue-500 focus:border-blue-500"
                                  />
                                </div>
                                <div className="flex items-end">
                                  <button
                                    onClick={() => handleUploadDocument(filing.id)}
                                    disabled={!selectedFile || uploadingDoc === filing.id}
                                    className="px-4 py-2 bg-blue-600 text-white text-sm font-medium rounded-md hover:bg-blue-700 disabled:bg-gray-300 disabled:cursor-not-allowed"
                                  >
                                    {uploadingDoc === filing.id ? 'Uploading...' : 'Upload'}
                                  </button>
                                </div>
                              </div>
                            </div>

                            {/* Documents Table */}
                            {filing.documents && filing.documents.length > 0 ? (
                              <div className="overflow-x-auto">
                                <table className="min-w-full divide-y divide-gray-200">
                                  <thead className="bg-gray-50">
                                    <tr>
                                      <th className="px-5 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">Document Name</th>
                                      <th className="px-5 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">Type</th>
                                      <th className="px-5 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">Uploaded</th>
                                      <th className="px-5 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">Actions</th>
                                    </tr>
                                  </thead>
                                  <tbody className="bg-white divide-y divide-gray-200">
                                    {filing.documents.map((doc) => (
                                      <tr key={doc.id} className="hover:bg-gray-50">
                                        <td className="px-5 py-4 text-sm font-medium text-gray-900">{doc.name}</td>
                                        <td className="px-5 py-4 text-sm text-gray-600">
                                          <span className="inline-flex items-center px-2.5 py-0.5 rounded text-xs font-medium bg-gray-100 text-gray-800">
                                            {doc.type}
                                          </span>
                                        </td>
                                        <td className="px-5 py-4 text-sm text-gray-600">{new Date(doc.createdAt).toLocaleDateString()}</td>
                                        <td className="px-5 py-4 text-sm">
                                          <div className="flex items-center gap-2">
                                            <button
                                              onClick={() => handleDownloadDocument(doc.id)}
                                              className="text-blue-600 hover:text-blue-900 font-medium"
                                            >
                                              Download
                                            </button>
                                            <span className="text-gray-300">|</span>
                                            <button
                                              onClick={() => handleDeleteDocument(filing.id, doc.id, doc.name)}
                                              className="text-red-600 hover:text-red-900 font-medium"
                                            >
                                              Delete
                                            </button>
                                          </div>
                                        </td>
                                      </tr>
                                    ))}
                                  </tbody>
                                </table>
                              </div>
                            ) : (
                              <div className="p-5 text-center text-sm text-gray-500">
                                No documents uploaded yet
                              </div>
                            )}
                          </div>

                          {/* Properties */}
                          {filing.properties && filing.properties.length > 0 && (
                            <div className="bg-white rounded-lg border border-gray-200">
                              <div className="px-5 py-4 border-b border-gray-200 bg-gray-50">
                                <h5 className="text-sm font-semibold text-gray-900 flex items-center">
                                  <svg className="w-5 h-5 mr-2 text-gray-600" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                                    <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M3 12l2-2m0 0l7-7 7 7M5 10v10a1 1 0 001 1h3m10-11l2 2m-2-2v10a1 1 0 01-1 1h-3m-6 0a1 1 0 001-1v-4a1 1 0 011-1h2a1 1 0 011 1v4a1 1 0 001 1m-6 0h6" />
                                  </svg>
                                  Properties ({filing.properties.length})
                                </h5>
                              </div>
                              <div className="p-5 space-y-4">
                                {filing.properties.map((property) => (
                                  <div key={property.id} className="bg-gray-50 rounded-lg p-5 border border-gray-200">
                                    <div className="grid grid-cols-1 md:grid-cols-2 gap-4 mb-4">
                                      <div>
                                        <p className="text-xs font-bold text-gray-600 uppercase mb-1">Property Address</p>
                                        <p className="text-sm text-gray-900">{property.address1}{property.address2 && `, ${property.address2}`}</p>
                                        <p className="text-sm text-gray-700">{property.city}, {property.state} {property.zipcode}</p>
                                      </div>
                                      <div className="space-y-2">
                                        <div className="flex justify-between">
                                          <span className="text-xs font-bold text-gray-600 uppercase">Purchase Price</span>
                                          <span className="text-sm text-gray-900">${property.purchasePrice.toLocaleString()}</span>
                                        </div>
                                        <div className="flex justify-between">
                                          <span className="text-xs font-bold text-gray-600 uppercase">Closing Cost</span>
                                          <span className="text-sm text-gray-900">${property.closingCost.toLocaleString()}</span>
                                        </div>
                                        {property.rents && (
                                          <div className="flex justify-between">
                                            <span className="text-xs font-bold text-gray-600 uppercase">Rents</span>
                                            <span className="text-sm text-gray-900">${property.rents.toLocaleString()}</span>
                                          </div>
                                        )}
                                        {property.royalties && (
                                          <div className="flex justify-between">
                                            <span className="text-xs font-bold text-gray-600 uppercase">Royalties</span>
                                            <span className="text-sm text-gray-900">${property.royalties.toLocaleString()}</span>
                                          </div>
                                        )}
                                      </div>
                                    </div>
                                    {property.expenses && property.expenses.length > 0 && (
                                      <div className="mt-4 pt-4 border-t border-gray-200">
                                        <p className="text-xs font-bold text-gray-600 uppercase mb-3">Expenses ({property.expenses.length})</p>
                                        <div className="grid grid-cols-1 md:grid-cols-2 gap-2">
                                          {property.expenses.map((expense) => (
                                            <div key={expense.id} className="flex justify-between items-center bg-white rounded px-3 py-2 border border-gray-200">
                                              <span className="text-sm text-gray-700">{expense.name}</span>
                                              <span className="text-sm font-semibold text-gray-900">${expense.amount.toLocaleString()}</span>
                                            </div>
                                          ))}
                                        </div>
                                      </div>
                                    )}
                                  </div>
                                ))}
                              </div>
                            </div>
                          )}

                          {/* IRA Contributions & Charitable Contributions */}
                          {((filing.iraContributions && filing.iraContributions.length > 0) || (filing.charities && filing.charities.length > 0)) && (
                            <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
                              {/* IRA Contributions */}
                              {filing.iraContributions && filing.iraContributions.length > 0 && (
                                <div className="bg-white rounded-lg border border-gray-200">
                                  <div className="px-5 py-4 border-b border-gray-200 bg-gray-50">
                                    <h5 className="text-sm font-semibold text-gray-900 flex items-center">
                                      <svg className="w-5 h-5 mr-2 text-gray-600" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                                        <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M12 8c-1.657 0-3 .895-3 2s1.343 2 3 2 3 .895 3 2-1.343 2-3 2m0-8c1.11 0 2.08.402 2.599 1M12 8V7m0 1v8m0 0v1m0-1c-1.11 0-2.08-.402-2.599-1M21 12a9 9 0 11-18 0 9 9 0 0118 0z" />
                                      </svg>
                                      IRA Contributions ({filing.iraContributions.length})
                                    </h5>
                                  </div>
                                  <div className="p-4 space-y-3">
                                    {filing.iraContributions.map((ira) => (
                                      <div key={ira.id} className="flex justify-between items-center p-3 bg-gray-50 rounded border border-gray-200">
                                        <span className="text-sm text-gray-900">{ira.accountType}</span>
                                        <span className="text-sm font-bold text-gray-900">${ira.amount.toLocaleString()}</span>
                                      </div>
                                    ))}
                                  </div>
                                </div>
                              )}

                              {/* Charities */}
                              {filing.charities && filing.charities.length > 0 && (
                                <div className="bg-white rounded-lg border border-gray-200">
                                  <div className="px-5 py-4 border-b border-gray-200 bg-gray-50">
                                    <h5 className="text-sm font-semibold text-gray-900 flex items-center">
                                      <svg className="w-5 h-5 mr-2 text-gray-600" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                                        <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M4.318 6.318a4.5 4.5 0 000 6.364L12 20.364l7.682-7.682a4.5 4.5 0 00-6.364-6.364L12 7.636l-1.318-1.318a4.5 4.5 0 00-6.364 0z" />
                                      </svg>
                                      Charitable Contributions ({filing.charities.length})
                                    </h5>
                                  </div>
                                  <div className="p-4 space-y-3">
                                    {filing.charities.map((charity) => (
                                      <div key={charity.id} className="flex justify-between items-center p-3 bg-gray-50 rounded border border-gray-200">
                                        <span className="text-sm text-gray-900">{charity.name}</span>
                                        <span className="text-sm font-bold text-gray-900">${charity.contribution.toLocaleString()}</span>
                                      </div>
                                    ))}
                                  </div>
                                </div>
                              )}
                            </div>
                          )}

                          {/* Childcares */}
                          {filing.childcares && filing.childcares.length > 0 && (
                            <div className="bg-white rounded-lg border border-gray-200">
                              <div className="px-5 py-4 border-b border-gray-200 bg-gray-50">
                                <h5 className="text-sm font-semibold text-gray-900 flex items-center">
                                  <svg className="w-5 h-5 mr-2 text-gray-600" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                                    <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M17 20h5v-2a3 3 0 00-5.356-1.857M17 20H7m10 0v-2c0-.656-.126-1.283-.356-1.857M7 20H2v-2a3 3 0 015.356-1.857M7 20v-2c0-.656.126-1.283.356-1.857m0 0a5.002 5.002 0 019.288 0M15 7a3 3 0 11-6 0 3 3 0 016 0zm6 3a2 2 0 11-4 0 2 2 0 014 0zM7 10a2 2 0 11-4 0 2 2 0 014 0z" />
                                  </svg>
                                  Childcare Expenses ({filing.childcares.length})
                                </h5>
                              </div>
                              <div className="p-5 space-y-4">
                                {filing.childcares.map((childcare) => (
                                  <div key={childcare.id} className="bg-gray-50 rounded-lg p-4 border border-gray-200">
                                    <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
                                      <div>
                                        <p className="text-xs font-bold text-gray-600 uppercase mb-1">Provider</p>
                                        <p className="text-sm text-gray-900">{childcare.name}</p>
                                        <p className="text-xs text-gray-600 mt-1"><span className="font-bold">Tax ID:</span> {childcare.taxId}</p>
                                      </div>
                                      <div>
                                        <p className="text-xs font-bold text-gray-600 uppercase mb-1">Amount</p>
                                        <p className="text-lg font-bold text-gray-900">${childcare.amount.toLocaleString()}</p>
                                      </div>
                                      <div className="col-span-full">
                                        <p className="text-xs font-bold text-gray-600 uppercase mb-1">Address</p>
                                        <p className="text-sm text-gray-900">{childcare.address1}{childcare.address2 && `, ${childcare.address2}`}</p>
                                        <p className="text-sm text-gray-900">{childcare.city}, {childcare.state} {childcare.zipcode}</p>
                                      </div>
                                    </div>
                                  </div>
                                ))}
                              </div>
                            </div>
                          )}

                          {/* Payments */}
                          {filing.payments && filing.payments.length > 0 && (
                            <div className="bg-white rounded-lg border border-gray-200">
                              <div className="px-5 py-4 border-b border-gray-200 bg-gray-50">
                                <h5 className="text-sm font-semibold text-gray-900 flex items-center">
                                  <svg className="w-5 h-5 mr-2 text-gray-600" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                                    <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M3 10h18M7 15h1m4 0h1m-7 4h12a3 3 0 003-3V8a3 3 0 00-3-3H6a3 3 0 00-3 3v8a3 3 0 003 3z" />
                                  </svg>
                                  Payments ({filing.payments.length})
                                </h5>
                              </div>
                              <div className="p-5 space-y-4">
                                {filing.payments.map((payment) => (
                                  <div key={payment.id} className="bg-gray-50 rounded-lg p-5 border border-gray-200">
                                    <div className="flex justify-between items-start mb-3">
                                      <div>
                                        <p className="text-2xl font-bold text-gray-900">${payment.amount.toLocaleString()}</p>
                                        {payment.originalAmount && payment.originalAmount !== payment.amount && (
                                          <p className="text-xs text-gray-600 mt-1"><span className="font-bold">Original:</span> ${payment.originalAmount.toLocaleString()}</p>
                                        )}
                                        {payment.discountAmount && (
                                          <p className="text-sm text-gray-700 mt-1"><span className="font-bold">Savings:</span> ${payment.discountAmount.toLocaleString()}</p>
                                        )}
                                      </div>
                                      <span className="px-3 py-1 text-xs font-bold rounded bg-gray-200 text-gray-900">
                                        {payment.status.toUpperCase()}
                                      </span>
                                    </div>
                                    {payment.discountCode && (
                                      <div className="mb-3">
                                        <span className="text-xs text-gray-700">
                                          <span className="font-bold">Code:</span> {payment.discountCode}
                                        </span>
                                      </div>
                                    )}
                                    {payment.items && payment.items.length > 0 && (
                                      <div className="mt-3 pt-3 border-t border-gray-200">
                                        <p className="text-xs font-bold text-gray-600 uppercase mb-2">Line Items</p>
                                        <div className="space-y-2">
                                          {payment.items.map((item) => (
                                            <div key={item.id} className="flex justify-between items-center bg-white rounded px-3 py-2 border border-gray-200">
                                              <span className="text-sm text-gray-700">{item.name} <span className="text-gray-500">×{item.quantity}</span></span>
                                              <span className="text-sm font-bold text-gray-900">${item.unitAmount.toLocaleString()}</span>
                                            </div>
                                          ))}
                                        </div>
                                      </div>
                                    )}
                                    <p className="text-xs text-gray-500 mt-3">{new Date(payment.createdAt).toLocaleString()}</p>
                                  </div>
                                ))}
                              </div>
                            </div>
                          )}

                          {/* Discounts */}
                          {filing.discounts && filing.discounts.length > 0 && (
                            <div className="bg-white rounded-lg border border-gray-200">
                              <div className="px-5 py-4 border-b border-gray-200 bg-gray-50">
                                <h5 className="text-sm font-semibold text-gray-900 flex items-center">
                                  <svg className="w-5 h-5 mr-2 text-gray-600" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                                    <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M12 8c-1.657 0-3 .895-3 2s1.343 2 3 2 3 .895 3 2-1.343 2-3 2m0-8c1.11 0 2.08.402 2.599 1M12 8V7m0 1v8m0 0v1m0-1c-1.11 0-2.08-.402-2.599-1M21 12a9 9 0 11-18 0 9 9 0 0118 0z" />
                                  </svg>
                                  Discounts Applied ({filing.discounts.length})
                                </h5>
                              </div>
                              <div className="p-5 space-y-3">
                                {filing.discounts.map((discount) => (
                                  <div key={discount.id} className="bg-gray-50 rounded-lg p-4 border border-gray-200">
                                    <div className="flex justify-between items-start">
                                      <div>
                                        <p className="text-sm font-bold text-gray-900">{discount.code || 'Discount Applied'}</p>
                                        <div className="mt-2 space-y-1">
                                          <p className="text-xs text-gray-700"><span className="font-bold">Original:</span> ${discount.originalAmount.toLocaleString()}</p>
                                          <p className="text-xs text-gray-700"><span className="font-bold">Final:</span> ${discount.finalAmount.toLocaleString()}</p>
                                        </div>
                                        <p className="text-xs text-gray-500 mt-2">{new Date(discount.appliedAt).toLocaleString()}</p>
                                      </div>
                                      <div className="text-right">
                                        <p className="text-xl font-bold text-gray-900">-${discount.discountAmount.toLocaleString()}</p>
                                        <p className="text-xs text-gray-600 font-bold">SAVED</p>
                                      </div>
                                    </div>
                                  </div>
                                ))}
                              </div>
                            </div>
                          )}
                        </div>
                      )}
                    </div>
                  )
                })}
              </div>
            ) : (
              <div className="border-t border-gray-200 px-4 py-5 sm:px-6 text-center">
                <p className="text-gray-500">No tax filings found</p>
              </div>
            )}
          </div>
        </div>
      </main>

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

      {/* Signature Request Modal */}
      {signatureModalOpen && selectedFilingForSignature && clientData && (
        <SignatureModal
          isOpen={signatureModalOpen}
          onClose={() => {
            setSignatureModalOpen(false)
            setSelectedFilingForSignature(null)
          }}
          onSubmit={handleSendSignatureRequest}
          filing={clientData.filings.find(f => f.id === selectedFilingForSignature)!}
          client={clientData.client}
          spouse={clientData.spouse}
          isSending={sendingSignature}
        />
      )}
    </div>
  )
}

// Signature Modal Component
function SignatureModal({
  isOpen,
  onClose,
  onSubmit,
  filing,
  client,
  spouse,
  isSending
}: {
  isOpen: boolean
  onClose: () => void
  onSubmit: (data: {
    pdfPath: string;
    taxPayerEmail: string;
    taxPayerName: string;
    taxPayerSsn: string;
    spouseName: string;
    spouseEmail: string;
    grossIncome: number;
    totalTax: number;
    taxWithHeld: number;
    refund: number;
    owed: number;
    spouseSignature: boolean;
  }) => void
  filing: { income: number | null; year: number; [key: string]: unknown }
  client: { email?: string | null; firstName?: string | null; lastName?: string | null; ssn?: string | null; [key: string]: unknown }
  spouse: { email?: string | null; firstName?: string | null; lastName?: string | null; [key: string]: unknown } | null
  isSending: boolean
}) {
  const [formData, setFormData] = useState({
    pdfPath: '',
    taxPayerEmail: client.email || '',
    taxPayerName: `${client.firstName || ''} ${client.lastName || ''}`.trim(),
    taxPayerSsn: client.ssn || '',
    spouseName: spouse ? `${spouse.firstName || ''} ${spouse.lastName || ''}`.trim() : '',
    spouseEmail: spouse?.email || '',
    grossIncome: filing.income || 0,
    totalTax: 0,
    taxWithHeld: 0,
    refund: 0,
    owed: 0,
    spouseSignature: !!spouse,
  })

  const handleSubmit = (e: React.FormEvent) => {
    e.preventDefault()
    onSubmit(formData)
  }

  if (!isOpen) return null

  return (
    <div className="fixed inset-0 z-50 overflow-y-auto">
      <div className="flex min-h-screen items-center justify-center p-4">
        <div className="fixed inset-0 bg-black bg-opacity-50 transition-opacity" onClick={onClose}></div>

        <div className="relative bg-white rounded-lg shadow-xl max-w-2xl w-full max-h-[90vh] overflow-y-auto">
          {/* Header */}
          <div className="sticky top-0 bg-white border-b border-gray-200 px-6 py-4 flex items-center justify-between">
            <div>
              <h3 className="text-lg font-semibold text-gray-900">Send for Signature</h3>
              <p className="text-sm text-gray-500">Tax Year {filing.year}</p>
            </div>
            <button
              onClick={onClose}
              className="text-gray-400 hover:text-gray-600"
            >
              <svg className="w-6 h-6" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M6 18L18 6M6 6l12 12" />
              </svg>
            </button>
          </div>

          {/* Form */}
          <form onSubmit={handleSubmit} className="p-6 space-y-6">
            {/* PDF Path */}
            <div>
              <label className="block text-sm font-medium text-gray-700 mb-2">
                PDF Document Path <span className="text-red-500">*</span>
              </label>
              <input
                type="text"
                required
                value={formData.pdfPath}
                onChange={(e) => setFormData({ ...formData, pdfPath: e.target.value })}
                placeholder="gs://bucket-name/path/to/form-8879.pdf"
                className="w-full px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-blue-500 focus:border-blue-500"
              />
              <p className="mt-1 text-xs text-gray-500">Full path to the Form 8879 PDF in cloud storage</p>
            </div>

            {/* Taxpayer Information */}
            <div className="border-t border-gray-200 pt-4">
              <h4 className="text-sm font-semibold text-gray-900 mb-4">Taxpayer Information</h4>
              <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
                <div>
                  <label className="block text-sm font-medium text-gray-700 mb-2">
                    Name <span className="text-red-500">*</span>
                  </label>
                  <input
                    type="text"
                    required
                    value={formData.taxPayerName}
                    onChange={(e) => setFormData({ ...formData, taxPayerName: e.target.value })}
                    className="w-full px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-blue-500 focus:border-blue-500"
                  />
                </div>
                <div>
                  <label className="block text-sm font-medium text-gray-700 mb-2">
                    Email <span className="text-red-500">*</span>
                  </label>
                  <input
                    type="email"
                    required
                    value={formData.taxPayerEmail}
                    onChange={(e) => setFormData({ ...formData, taxPayerEmail: e.target.value })}
                    className="w-full px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-blue-500 focus:border-blue-500"
                  />
                </div>
                <div>
                  <label className="block text-sm font-medium text-gray-700 mb-2">
                    SSN <span className="text-red-500">*</span>
                  </label>
                  <input
                    type="text"
                    required
                    value={formData.taxPayerSsn}
                    onChange={(e) => setFormData({ ...formData, taxPayerSsn: e.target.value })}
                    className="w-full px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-blue-500 focus:border-blue-500"
                  />
                </div>
              </div>
            </div>

            {/* Spouse Information */}
            {formData.spouseSignature && (
              <div className="border-t border-gray-200 pt-4">
                <h4 className="text-sm font-semibold text-gray-900 mb-4">Spouse Information</h4>
                <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
                  <div>
                    <label className="block text-sm font-medium text-gray-700 mb-2">
                      Name <span className="text-red-500">*</span>
                    </label>
                    <input
                      type="text"
                      required={formData.spouseSignature}
                      value={formData.spouseName}
                      onChange={(e) => setFormData({ ...formData, spouseName: e.target.value })}
                      className="w-full px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-blue-500 focus:border-blue-500"
                    />
                  </div>
                  <div>
                    <label className="block text-sm font-medium text-gray-700 mb-2">
                      Email <span className="text-red-500">*</span>
                    </label>
                    <input
                      type="email"
                      required={formData.spouseSignature}
                      value={formData.spouseEmail}
                      onChange={(e) => setFormData({ ...formData, spouseEmail: e.target.value })}
                      className="w-full px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-blue-500 focus:border-blue-500"
                    />
                  </div>
                </div>
              </div>
            )}

            {/* Tax Amounts */}
            <div className="border-t border-gray-200 pt-4">
              <h4 className="text-sm font-semibold text-gray-900 mb-4">Tax Information</h4>
              <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
                <div>
                  <label className="block text-sm font-medium text-gray-700 mb-2">Gross Income</label>
                  <input
                    type="number"
                    step="0.01"
                    value={formData.grossIncome}
                    onChange={(e) => setFormData({ ...formData, grossIncome: parseFloat(e.target.value) || 0 })}
                    className="w-full px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-blue-500 focus:border-blue-500"
                  />
                </div>
                <div>
                  <label className="block text-sm font-medium text-gray-700 mb-2">Total Tax</label>
                  <input
                    type="number"
                    step="0.01"
                    value={formData.totalTax}
                    onChange={(e) => setFormData({ ...formData, totalTax: parseFloat(e.target.value) || 0 })}
                    className="w-full px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-blue-500 focus:border-blue-500"
                  />
                </div>
                <div>
                  <label className="block text-sm font-medium text-gray-700 mb-2">Tax Withheld</label>
                  <input
                    type="number"
                    step="0.01"
                    value={formData.taxWithHeld}
                    onChange={(e) => setFormData({ ...formData, taxWithHeld: parseFloat(e.target.value) || 0 })}
                    className="w-full px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-blue-500 focus:border-blue-500"
                  />
                </div>
                <div>
                  <label className="block text-sm font-medium text-gray-700 mb-2">Refund</label>
                  <input
                    type="number"
                    step="0.01"
                    value={formData.refund}
                    onChange={(e) => setFormData({ ...formData, refund: parseFloat(e.target.value) || 0 })}
                    className="w-full px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-blue-500 focus:border-blue-500"
                  />
                </div>
                <div>
                  <label className="block text-sm font-medium text-gray-700 mb-2">Amount Owed</label>
                  <input
                    type="number"
                    step="0.01"
                    value={formData.owed}
                    onChange={(e) => setFormData({ ...formData, owed: parseFloat(e.target.value) || 0 })}
                    className="w-full px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-blue-500 focus:border-blue-500"
                  />
                </div>
              </div>
            </div>

            {/* Footer */}
            <div className="border-t border-gray-200 pt-4 flex justify-end space-x-3">
              <button
                type="button"
                onClick={onClose}
                disabled={isSending}
                className="px-4 py-2 text-sm font-medium text-gray-700 bg-white border border-gray-300 rounded-md hover:bg-gray-50 disabled:opacity-50"
              >
                Cancel
              </button>
              <button
                type="submit"
                disabled={isSending}
                className="px-4 py-2 text-sm font-medium text-white bg-purple-600 rounded-md hover:bg-purple-700 disabled:bg-gray-400 flex items-center space-x-2"
              >
                {isSending ? (
                  <>
                    <svg className="animate-spin h-4 w-4" fill="none" viewBox="0 0 24 24">
                      <circle className="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" strokeWidth="4"></circle>
                      <path className="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"></path>
                    </svg>
                    <span>Sending...</span>
                  </>
                ) : (
                  <>
                    <svg className="w-4 h-4" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                      <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M15.232 5.232l3.536 3.536m-2.036-5.036a2.5 2.5 0 113.536 3.536L6.5 21.036H3v-3.572L16.732 3.732z" />
                    </svg>
                    <span>Send for Signature</span>
                  </>
                )}
              </button>
            </div>
          </form>
        </div>
      </div>
    </div>
  )
}

export default function ClientDetailPage() {
  return (
    <ProtectedRoute>
      <ClientDetailContent />
    </ProtectedRoute>
  );
}
