'use client'

import { useEffect, useState } from 'react'
import { useRouter, useSearchParams } from 'next/navigation'

export default function PortalVerifyPage({ params }: { params: { tenantId: string } }) {
  const router = useRouter()
  const searchParams = useSearchParams()
  const [lastFourSSN, setLastFourSSN] = useState('')
  const [loading, setLoading] = useState(false)
  const [error, setError] = useState('')
  const [validatingToken, setValidatingToken] = useState(true)

  const tenantId = params.tenantId
  const magicToken = searchParams.get('token')

  useEffect(() => {
    // Validate that we have a token
    if (!magicToken) {
      setError('Invalid or missing access link. Please request a new link.')
      setValidatingToken(false)
      return
    }

    // Validate token format (JWT)
    const jwtPattern = /^[A-Za-z0-9-_]+\.[A-Za-z0-9-_]+\.[A-Za-z0-9-_]+$/
    if (!jwtPattern.test(magicToken)) {
      setError('Invalid access link format. Please request a new link.')
      setValidatingToken(false)
      return
    }

    setValidatingToken(false)
  }, [magicToken])

  const handleVerify = async (e: React.FormEvent) => {
    e.preventDefault()
    setError('')

    // Validate SSN format
    if (!/^\d{4}$/.test(lastFourSSN)) {
      setError('Please enter exactly 4 digits')
      return
    }

    setLoading(true)

    try {
      const apiUrl = process.env.NEXT_PUBLIC_API_URL || 'http://localhost:8080'
      const response = await fetch(`${apiUrl}/api/v1/portal/exchange`, {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
        },
        body: JSON.stringify({
          magicToken,
          lastFourSSN,
        }),
      })

      if (!response.ok) {
        const errorText = await response.text()

        if (response.status === 401) {
          if (errorText.includes('already been used')) {
            setError('This link has already been used. Please request a new access link.')
          } else if (errorText.includes('Identity verification failed')) {
            setError('Identity verification failed. Please check your SSN and try again.')
          } else {
            setError('Your access link has expired. Please request a new link.')
          }
        } else {
          setError('Verification failed. Please try again or request a new link.')
        }
        setLoading(false)
        return
      }

      const data = await response.json()

      // Store session token
      sessionStorage.setItem('portalToken', data.sessionToken)
      sessionStorage.setItem('tokenExpiry', String(Date.now() + data.expiresIn * 1000))

      // Redirect to dashboard
      router.push(`/${tenantId}/portal/dashboard`)
    } catch (err) {
      console.error('Verification error:', err)
      setError('An error occurred. Please try again.')
      setLoading(false)
    }
  }

  if (validatingToken) {
    return (
      <div className="min-h-screen bg-gradient-to-br from-blue-50 to-indigo-100 flex items-center justify-center p-4">
        <div className="bg-white rounded-2xl shadow-xl p-8 max-w-md w-full">
          <div className="text-center">
            <div className="animate-spin rounded-full h-12 w-12 border-b-2 border-blue-600 mx-auto"></div>
            <p className="mt-4 text-gray-600">Validating access link...</p>
          </div>
        </div>
      </div>
    )
  }

  if (error && !magicToken) {
    return (
      <div className="min-h-screen bg-gradient-to-br from-blue-50 to-indigo-100 flex items-center justify-center p-4">
        <div className="bg-white rounded-2xl shadow-xl p-8 max-w-md w-full">
          <div className="text-center">
            <div className="mx-auto flex items-center justify-center h-12 w-12 rounded-full bg-red-100 mb-4">
              <svg className="h-6 w-6 text-red-600" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M6 18L18 6M6 6l12 12" />
              </svg>
            </div>
            <h2 className="text-xl font-semibold text-gray-900 mb-2">Invalid Access Link</h2>
            <p className="text-gray-600 mb-6">{error}</p>
            <p className="text-sm text-gray-500">Please contact your tax preparer for a new access link.</p>
          </div>
        </div>
      </div>
    )
  }

  return (
    <div className="min-h-screen bg-gradient-to-br from-blue-50 to-indigo-100 flex items-center justify-center p-4">
      <div className="bg-white rounded-2xl shadow-xl p-8 max-w-md w-full">
        {/* Header */}
        <div className="text-center mb-8">
          <div className="mx-auto flex items-center justify-center h-12 w-12 rounded-full bg-blue-100 mb-4">
            <svg className="h-6 w-6 text-blue-600" fill="none" viewBox="0 0 24 24" stroke="currentColor">
              <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M9 12l2 2 4-4m5.618-4.016A11.955 11.955 0 0112 2.944a11.955 11.955 0 01-8.618 3.04A12.02 12.02 0 003 9c0 5.591 3.824 10.29 9 11.622 5.176-1.332 9-6.03 9-11.622 0-1.042-.133-2.052-.382-3.016z" />
              </svg>
          </div>
          <h1 className="text-2xl font-bold text-gray-900 mb-2">
            Verify Your Identity
          </h1>
          <p className="text-gray-600 text-sm">
            For your security, please verify your identity by entering the last 4 digits of your Social Security Number
          </p>
        </div>

        {/* Error Message */}
        {error && (
          <div className="mb-6 bg-red-50 border border-red-200 rounded-lg p-4">
            <div className="flex">
              <div className="flex-shrink-0">
                <svg className="h-5 w-5 text-red-400" viewBox="0 0 20 20" fill="currentColor">
                  <path fillRule="evenodd" d="M10 18a8 8 0 100-16 8 8 0 000 16zM8.707 7.293a1 1 0 00-1.414 1.414L8.586 10l-1.293 1.293a1 1 0 101.414 1.414L10 11.414l1.293 1.293a1 1 0 001.414-1.414L11.414 10l1.293-1.293a1 1 0 00-1.414-1.414L10 8.586 8.707 7.293z" clipRule="evenodd" />
                </svg>
              </div>
              <div className="ml-3">
                <p className="text-sm text-red-800">{error}</p>
              </div>
            </div>
          </div>
        )}

        {/* Form */}
        <form onSubmit={handleVerify}>
          <div className="mb-6">
            <label htmlFor="ssn" className="block text-sm font-medium text-gray-700 mb-2">
              Last 4 digits of SSN
            </label>
            <input
              type="text"
              id="ssn"
              maxLength={4}
              pattern="\d{4}"
              value={lastFourSSN}
              onChange={(e) => {
                const value = e.target.value.replace(/\D/g, '')
                setLastFourSSN(value)
              }}
              className="w-full px-4 py-3 border border-gray-300 rounded-lg focus:ring-2 focus:ring-blue-500 focus:border-transparent text-center text-2xl tracking-widest font-mono"
              placeholder="••••"
              required
              autoFocus
              disabled={loading}
            />
            <p className="mt-2 text-xs text-gray-500">
              This information is encrypted and used only for verification purposes
            </p>
          </div>

          <button
            type="submit"
            disabled={loading || lastFourSSN.length !== 4}
            className="w-full bg-blue-600 text-white py-3 px-4 rounded-lg font-medium hover:bg-blue-700 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-blue-500 disabled:opacity-50 disabled:cursor-not-allowed transition-colors"
          >
            {loading ? (
              <span className="flex items-center justify-center">
                <svg className="animate-spin -ml-1 mr-3 h-5 w-5 text-white" fill="none" viewBox="0 0 24 24">
                  <circle className="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" strokeWidth="4"></circle>
                  <path className="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"></path>
                </svg>
                Verifying...
              </span>
            ) : (
              'Verify & Continue'
            )}
          </button>
        </form>

        {/* Security Note */}
        <div className="mt-6 pt-6 border-t border-gray-200">
          <div className="flex items-start">
            <svg className="h-5 w-5 text-gray-400 mt-0.5 mr-2" fill="none" viewBox="0 0 24 24" stroke="currentColor">
              <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M12 15v2m-6 4h12a2 2 0 002-2v-6a2 2 0 00-2-2H6a2 2 0 00-2 2v6a2 2 0 002 2zm10-10V7a4 4 0 00-8 0v4h8z" />
            </svg>
            <p className="text-xs text-gray-500">
              Your access link is valid for 24 hours and can only be used once. Your personal information is protected with bank-level encryption.
            </p>
          </div>
        </div>
      </div>
    </div>
  )
}
