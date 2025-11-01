'use client'

import { useState } from 'react'
import Link from 'next/link'
import { useRouter } from 'next/navigation'
import { ProtectedRoute } from '@/components/ProtectedRoute'
import { UserMenu } from '@/components/UserMenu'
import { useAuth } from '@/contexts/AuthContext'

function NewTenantContent() {
  const router = useRouter()
  const { user } = useAuth()

  const [loading, setLoading] = useState(false)
  const [error, setError] = useState<string | null>(null)

  const [formData, setFormData] = useState({
    tenantId: '',
    tenantName: '',
    dbHost: '',
    dbPort: '5432',
    dbUser: '',
    dbPassword: '',
    dbName: '',
    dbSslMode: 'require',
    schemaPrefix: '',
    adapterType: 'mywelltax',
    storageProvider: '',
    storageBucket: '',
    storageCredentialsSecret: '',
    storageCredentialsPath: '',
    docusignIntegrationKey: '',
    docusignClientId: '',
    docusignPrivateKeySecret: '',
    docusignApiUrl: 'https://demo.docusign.net/restapi',
    notes: '',
  })

  const handleChange = (e: React.ChangeEvent<HTMLInputElement | HTMLSelectElement | HTMLTextAreaElement>) => {
    setFormData({
      ...formData,
      [e.target.name]: e.target.value,
    })
  }

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault()
    setError(null)
    setLoading(true)

    try {
      const apiUrl = process.env.NEXT_PUBLIC_API_URL || 'http://localhost:8081'
      const idToken = await user?.getIdToken()

      const response = await fetch(`${apiUrl}/api/v1/admin/tenants`, {
        method: 'POST',
        headers: {
          'Authorization': `Bearer ${idToken}`,
          'Content-Type': 'application/json',
        },
        body: JSON.stringify({
          ...formData,
          dbPort: parseInt(formData.dbPort),
          notes: formData.notes || null,
        }),
      })

      if (!response.ok) {
        const errorText = await response.text()
        throw new Error(errorText || 'Failed to create tenant')
      }

      await response.json()
      router.push('/admin/tenants')
    } catch (err) {
      setError(err instanceof Error ? err.message : 'An error occurred')
    } finally {
      setLoading(false)
    }
  }

  return (
    <div className="min-h-screen bg-gray-50">
      <nav className="bg-white shadow-sm">
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
                href="/admin/tenants"
                className="text-xs sm:text-sm text-gray-600 hover:text-gray-900"
              >
                <span className="hidden sm:inline">Account Management</span>
                <span className="sm:hidden">Accounts</span>
              </Link>
              <span className="text-gray-300">|</span>
              <span className="text-xs sm:text-sm font-semibold text-blue-600">New Account</span>
              <UserMenu />
            </div>
          </div>
        </div>
      </nav>

      <div className="max-w-4xl mx-auto py-6 sm:px-6 lg:px-8">
        <div className="px-4 py-6 sm:px-0">
          <div className="mb-6">
            <h1 className="text-xl sm:text-2xl font-bold text-gray-900">Add New Account Connection</h1>
            <p className="mt-1 text-xs sm:text-sm text-gray-600">
              Configure database connection and integration settings for a new account.
            </p>
          </div>

          {error && (
            <div className="mb-6 rounded-md bg-red-50 p-4">
              <div className="text-sm text-red-800">{error}</div>
            </div>
          )}

          <form onSubmit={handleSubmit} className="space-y-8">
            {/* Basic Information */}
            <div className="bg-white shadow sm:rounded-lg">
              <div className="px-4 py-5 sm:p-6">
                <h3 className="text-lg font-medium leading-6 text-gray-900 mb-4">
                  Basic Information
                </h3>
                <div className="grid grid-cols-1 gap-6 sm:grid-cols-2">
                  <div>
                    <label htmlFor="tenantId" className="block text-sm font-medium text-gray-700">
                      Tenant ID <span className="text-red-500">*</span>
                    </label>
                    <input
                      type="text"
                      name="tenantId"
                      id="tenantId"
                      required
                      value={formData.tenantId}
                      onChange={handleChange}
                      className="mt-1 block w-full rounded-md border-gray-300 shadow-sm focus:border-blue-500 focus:ring-blue-500 sm:text-sm"
                      placeholder="mywelltax"
                    />
                    <p className="mt-1 text-xs text-gray-500">Unique identifier used in API routes</p>
                  </div>

                  <div>
                    <label htmlFor="tenantName" className="block text-sm font-medium text-gray-700">
                      Tenant Name <span className="text-red-500">*</span>
                    </label>
                    <input
                      type="text"
                      name="tenantName"
                      id="tenantName"
                      required
                      value={formData.tenantName}
                      onChange={handleChange}
                      className="mt-1 block w-full rounded-md border-gray-300 shadow-sm focus:border-blue-500 focus:ring-blue-500 sm:text-sm"
                      placeholder="MyWellTax"
                    />
                  </div>

                  <div className="sm:col-span-2">
                    <label htmlFor="adapterType" className="block text-sm font-medium text-gray-700">
                      Adapter Type <span className="text-red-500">*</span>
                    </label>
                    <select
                      name="adapterType"
                      id="adapterType"
                      required
                      value={formData.adapterType}
                      onChange={handleChange}
                      className="mt-1 block w-full rounded-md border-gray-300 shadow-sm focus:border-blue-500 focus:ring-blue-500 sm:text-sm"
                    >
                      <option value="mywelltax">MyWellTax</option>
                      <option value="drake">Drake</option>
                      <option value="lacerte">Lacerte</option>
                      <option value="proseries">ProSeries</option>
                      <option value="ultratax">UltraTax</option>
                    </select>
                  </div>
                </div>
              </div>
            </div>

            {/* Database Configuration */}
            <div className="bg-white shadow sm:rounded-lg">
              <div className="px-4 py-5 sm:p-6">
                <h3 className="text-lg font-medium leading-6 text-gray-900 mb-4">
                  Database Configuration
                </h3>
                <div className="grid grid-cols-1 gap-6 sm:grid-cols-2">
                  <div>
                    <label htmlFor="dbHost" className="block text-sm font-medium text-gray-700">
                      Database Host <span className="text-red-500">*</span>
                    </label>
                    <input
                      type="text"
                      name="dbHost"
                      id="dbHost"
                      required
                      value={formData.dbHost}
                      onChange={handleChange}
                      className="mt-1 block w-full rounded-md border-gray-300 shadow-sm focus:border-blue-500 focus:ring-blue-500 sm:text-sm"
                      placeholder="localhost"
                    />
                  </div>

                  <div>
                    <label htmlFor="dbPort" className="block text-sm font-medium text-gray-700">
                      Database Port <span className="text-red-500">*</span>
                    </label>
                    <input
                      type="number"
                      name="dbPort"
                      id="dbPort"
                      required
                      value={formData.dbPort}
                      onChange={handleChange}
                      className="mt-1 block w-full rounded-md border-gray-300 shadow-sm focus:border-blue-500 focus:ring-blue-500 sm:text-sm"
                    />
                  </div>

                  <div>
                    <label htmlFor="dbUser" className="block text-sm font-medium text-gray-700">
                      Database User <span className="text-red-500">*</span>
                    </label>
                    <input
                      type="text"
                      name="dbUser"
                      id="dbUser"
                      required
                      value={formData.dbUser}
                      onChange={handleChange}
                      className="mt-1 block w-full rounded-md border-gray-300 shadow-sm focus:border-blue-500 focus:ring-blue-500 sm:text-sm"
                      placeholder="postgres"
                    />
                  </div>

                  <div>
                    <label htmlFor="dbPassword" className="block text-sm font-medium text-gray-700">
                      Database Password <span className="text-red-500">*</span>
                    </label>
                    <input
                      type="password"
                      name="dbPassword"
                      id="dbPassword"
                      required
                      value={formData.dbPassword}
                      onChange={handleChange}
                      className="mt-1 block w-full rounded-md border-gray-300 shadow-sm focus:border-blue-500 focus:ring-blue-500 sm:text-sm"
                    />
                    <p className="mt-1 text-xs text-gray-500">Will be encrypted before storage</p>
                  </div>

                  <div>
                    <label htmlFor="dbName" className="block text-sm font-medium text-gray-700">
                      Database Name <span className="text-red-500">*</span>
                    </label>
                    <input
                      type="text"
                      name="dbName"
                      id="dbName"
                      required
                      value={formData.dbName}
                      onChange={handleChange}
                      className="mt-1 block w-full rounded-md border-gray-300 shadow-sm focus:border-blue-500 focus:ring-blue-500 sm:text-sm"
                      placeholder="mywelltax"
                    />
                  </div>

                  <div>
                    <label htmlFor="dbSslMode" className="block text-sm font-medium text-gray-700">
                      SSL Mode <span className="text-red-500">*</span>
                    </label>
                    <select
                      name="dbSslMode"
                      id="dbSslMode"
                      required
                      value={formData.dbSslMode}
                      onChange={handleChange}
                      className="mt-1 block w-full rounded-md border-gray-300 shadow-sm focus:border-blue-500 focus:ring-blue-500 sm:text-sm"
                    >
                      <option value="disable">Disable</option>
                      <option value="require">Require</option>
                      <option value="verify-ca">Verify CA</option>
                      <option value="verify-full">Verify Full</option>
                    </select>
                  </div>

                  <div className="sm:col-span-2">
                    <label htmlFor="schemaPrefix" className="block text-sm font-medium text-gray-700">
                      Schema Prefix <span className="text-red-500">*</span>
                    </label>
                    <input
                      type="text"
                      name="schemaPrefix"
                      id="schemaPrefix"
                      required
                      value={formData.schemaPrefix}
                      onChange={handleChange}
                      className="mt-1 block w-full rounded-md border-gray-300 shadow-sm focus:border-blue-500 focus:ring-blue-500 sm:text-sm"
                      placeholder="taxes"
                    />
                    <p className="mt-1 text-xs text-gray-500">Schema or table prefix used in tenant database</p>
                  </div>
                </div>
              </div>
            </div>

            {/* Storage Configuration */}
            <div className="bg-white shadow sm:rounded-lg">
              <div className="px-4 py-5 sm:p-6">
                <h3 className="text-lg font-medium leading-6 text-gray-900 mb-4">
                  Storage Configuration (Optional)
                </h3>
                <div className="grid grid-cols-1 gap-6 sm:grid-cols-2">
                  <div>
                    <label htmlFor="storageProvider" className="block text-sm font-medium text-gray-700">
                      Storage Provider
                    </label>
                    <select
                      name="storageProvider"
                      id="storageProvider"
                      value={formData.storageProvider}
                      onChange={handleChange}
                      className="mt-1 block w-full rounded-md border-gray-300 shadow-sm focus:border-blue-500 focus:ring-blue-500 sm:text-sm"
                    >
                      <option value="">None</option>
                      <option value="gcs">Google Cloud Storage</option>
                      <option value="s3">Amazon S3</option>
                      <option value="azure">Azure Blob Storage</option>
                    </select>
                  </div>

                  <div>
                    <label htmlFor="storageBucket" className="block text-sm font-medium text-gray-700">
                      Storage Bucket
                    </label>
                    <input
                      type="text"
                      name="storageBucket"
                      id="storageBucket"
                      value={formData.storageBucket}
                      onChange={handleChange}
                      className="mt-1 block w-full rounded-md border-gray-300 shadow-sm focus:border-blue-500 focus:ring-blue-500 sm:text-sm"
                      placeholder="my-bucket"
                    />
                  </div>

                  <div>
                    <label htmlFor="storageCredentialsSecret" className="block text-sm font-medium text-gray-700">
                      Credentials Secret Path
                    </label>
                    <input
                      type="text"
                      name="storageCredentialsSecret"
                      id="storageCredentialsSecret"
                      value={formData.storageCredentialsSecret}
                      onChange={handleChange}
                      className="mt-1 block w-full rounded-md border-gray-300 shadow-sm focus:border-blue-500 focus:ring-blue-500 sm:text-sm"
                      placeholder="projects/PROJECT/secrets/NAME/versions/VERSION"
                    />
                    <p className="mt-1 text-xs text-gray-500">GCP Secret Manager path</p>
                  </div>

                  <div>
                    <label htmlFor="storageCredentialsPath" className="block text-sm font-medium text-gray-700">
                      Credentials File Path
                    </label>
                    <input
                      type="text"
                      name="storageCredentialsPath"
                      id="storageCredentialsPath"
                      value={formData.storageCredentialsPath}
                      onChange={handleChange}
                      className="mt-1 block w-full rounded-md border-gray-300 shadow-sm focus:border-blue-500 focus:ring-blue-500 sm:text-sm"
                      placeholder="/tmp/service-account.json"
                    />
                    <p className="mt-1 text-xs text-gray-500">Local file path (for development)</p>
                  </div>
                </div>
              </div>
            </div>

            {/* DocuSign Configuration */}
            <div className="bg-white shadow sm:rounded-lg">
              <div className="px-4 py-5 sm:p-6">
                <h3 className="text-lg font-medium leading-6 text-gray-900 mb-4">
                  DocuSign Configuration (Optional)
                </h3>
                <div className="grid grid-cols-1 gap-6 sm:grid-cols-2">
                  <div>
                    <label htmlFor="docusignIntegrationKey" className="block text-sm font-medium text-gray-700">
                      Integration Key
                    </label>
                    <input
                      type="text"
                      name="docusignIntegrationKey"
                      id="docusignIntegrationKey"
                      value={formData.docusignIntegrationKey}
                      onChange={handleChange}
                      className="mt-1 block w-full rounded-md border-gray-300 shadow-sm focus:border-blue-500 focus:ring-blue-500 sm:text-sm"
                    />
                  </div>

                  <div>
                    <label htmlFor="docusignClientId" className="block text-sm font-medium text-gray-700">
                      Client ID
                    </label>
                    <input
                      type="text"
                      name="docusignClientId"
                      id="docusignClientId"
                      value={formData.docusignClientId}
                      onChange={handleChange}
                      className="mt-1 block w-full rounded-md border-gray-300 shadow-sm focus:border-blue-500 focus:ring-blue-500 sm:text-sm"
                    />
                  </div>

                  <div>
                    <label htmlFor="docusignPrivateKeySecret" className="block text-sm font-medium text-gray-700">
                      Private Key Secret Path
                    </label>
                    <input
                      type="text"
                      name="docusignPrivateKeySecret"
                      id="docusignPrivateKeySecret"
                      value={formData.docusignPrivateKeySecret}
                      onChange={handleChange}
                      className="mt-1 block w-full rounded-md border-gray-300 shadow-sm focus:border-blue-500 focus:ring-blue-500 sm:text-sm"
                      placeholder="projects/PROJECT/secrets/NAME/versions/VERSION"
                    />
                  </div>

                  <div>
                    <label htmlFor="docusignApiUrl" className="block text-sm font-medium text-gray-700">
                      API URL
                    </label>
                    <select
                      name="docusignApiUrl"
                      id="docusignApiUrl"
                      value={formData.docusignApiUrl}
                      onChange={handleChange}
                      className="mt-1 block w-full rounded-md border-gray-300 shadow-sm focus:border-blue-500 focus:ring-blue-500 sm:text-sm"
                    >
                      <option value="https://demo.docusign.net/restapi">Demo</option>
                      <option value="https://na3.docusign.net/restapi">Production (NA3)</option>
                    </select>
                  </div>
                </div>
              </div>
            </div>

            {/* Notes */}
            <div className="bg-white shadow sm:rounded-lg">
              <div className="px-4 py-5 sm:p-6">
                <h3 className="text-lg font-medium leading-6 text-gray-900 mb-4">
                  Notes
                </h3>
                <div>
                  <label htmlFor="notes" className="block text-sm font-medium text-gray-700">
                    Additional Notes
                  </label>
                  <textarea
                    name="notes"
                    id="notes"
                    rows={3}
                    value={formData.notes}
                    onChange={handleChange}
                    className="mt-1 block w-full rounded-md border-gray-300 shadow-sm focus:border-blue-500 focus:ring-blue-500 sm:text-sm"
                    placeholder="Any additional information about this tenant..."
                  />
                </div>
              </div>
            </div>

            {/* Actions */}
            <div className="flex justify-end gap-3">
              <Link
                href="/admin/tenants"
                className="inline-flex items-center px-4 py-2 border border-gray-300 rounded-md shadow-sm text-sm font-medium text-gray-700 bg-white hover:bg-gray-50 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-blue-500"
              >
                Cancel
              </Link>
              <button
                type="submit"
                disabled={loading}
                className="inline-flex items-center px-4 py-2 border border-transparent rounded-md shadow-sm text-sm font-medium text-white bg-blue-600 hover:bg-blue-700 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-blue-500 disabled:opacity-50 disabled:cursor-not-allowed"
              >
                {loading ? (
                  <>
                    <svg className="animate-spin -ml-1 mr-2 h-4 w-4 text-white" fill="none" viewBox="0 0 24 24">
                      <circle className="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" strokeWidth="4"></circle>
                      <path className="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"></path>
                    </svg>
                    Creating...
                  </>
                ) : (
                  'Create Account'
                )}
              </button>
            </div>
          </form>
        </div>
      </div>
    </div>
  )
}

export default function NewTenantPage() {
  return (
    <ProtectedRoute requireAdmin>
      <NewTenantContent />
    </ProtectedRoute>
  )
}
