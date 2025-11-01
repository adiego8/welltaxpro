'use client'

import { useEffect, useState } from 'react'
import { useParams, usePathname, useRouter } from 'next/navigation'
import { useAuth } from '@/contexts/AuthContext'

interface TenantAccess {
  tenantId: string
  tenantName: string
  role: string
  isActive: boolean
}

export function TenantSwitcher() {
  const params = useParams()
  const pathname = usePathname()
  const router = useRouter()
  const { user } = useAuth()
  const currentTenantId = params.tenantId as string

  const [tenants, setTenants] = useState<TenantAccess[]>([])
  const [isOpen, setIsOpen] = useState(false)
  const [loading, setLoading] = useState(true)

  // Fetch employee's tenant access
  useEffect(() => {
    async function fetchTenants() {
      if (!user) {
        setLoading(false)
        return
      }

      try {
        const apiUrl = process.env.NEXT_PUBLIC_API_URL || 'http://localhost:8081'
        const idToken = await user.getIdToken()

        const response = await fetch(`${apiUrl}/api/v1/employees/me/tenants`, {
          headers: {
            'Authorization': `Bearer ${idToken}`,
            'Content-Type': 'application/json',
          },
        })

        if (!response.ok) {
          throw new Error('Failed to fetch tenants')
        }

        const data = await response.json()
        // Filter to only active tenants
        const activeTenants = (data || []).filter((t: TenantAccess) => t.isActive)
        setTenants(activeTenants)
      } catch (err) {
        console.error('Failed to fetch tenants:', err)
      } finally {
        setLoading(false)
      }
    }

    fetchTenants()
  }, [user])

  // Get current tenant info
  const currentTenant = tenants.find(t => t.tenantId === currentTenantId)

  // Handle tenant switch
  const handleTenantSwitch = (newTenantId: string) => {
    if (newTenantId === currentTenantId) {
      setIsOpen(false)
      return
    }

    // Replace current tenantId in pathname with new one
    const newPath = pathname.replace(`/${currentTenantId}`, `/${newTenantId}`)
    router.push(newPath)
    setIsOpen(false)
  }

  // Don't show if loading
  if (loading) {
    return null
  }

  // Show even with one tenant for testing (remove this condition later)
  // if (tenants.length <= 1) {
  //   return null
  // }

  return (
    <div className="relative">
      <button
        onClick={() => setIsOpen(!isOpen)}
        className="flex items-center gap-2 px-3 py-2 text-sm font-medium text-gray-700 bg-white border border-gray-300 rounded-lg hover:bg-gray-50 focus:outline-none focus:ring-2 focus:ring-blue-500"
      >
        <svg className="w-4 h-4 text-gray-500" fill="none" viewBox="0 0 24 24" stroke="currentColor">
          <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M19 21V5a2 2 0 00-2-2H7a2 2 0 00-2 2v16m14 0h2m-2 0h-5m-9 0H3m2 0h5M9 7h1m-1 4h1m4-4h1m-1 4h1m-5 10v-5a1 1 0 011-1h2a1 1 0 011 1v5m-4 0h4" />
        </svg>
        <span className="hidden sm:inline">{currentTenant?.tenantName || currentTenantId}</span>
        <svg className={`w-4 h-4 text-gray-500 transition-transform ${isOpen ? 'rotate-180' : ''}`} fill="none" viewBox="0 0 24 24" stroke="currentColor">
          <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M19 9l-7 7-7-7" />
        </svg>
      </button>

      {isOpen && (
        <>
          {/* Backdrop */}
          <div
            className="fixed inset-0 z-10"
            onClick={() => setIsOpen(false)}
          />

          {/* Dropdown */}
          <div className="absolute right-0 z-20 mt-2 w-64 bg-white rounded-lg shadow-lg border border-gray-200 py-1">
            <div className="px-4 py-2 text-xs font-semibold text-gray-500 uppercase tracking-wide border-b border-gray-200">
              Switch Account
            </div>
            <div className="max-h-64 overflow-y-auto">
              {tenants.map((tenant) => (
                <button
                  key={tenant.tenantId}
                  onClick={() => handleTenantSwitch(tenant.tenantId)}
                  className={`w-full text-left px-4 py-3 hover:bg-gray-50 transition-colors ${
                    tenant.tenantId === currentTenantId ? 'bg-blue-50' : ''
                  }`}
                >
                  <div className="flex items-center justify-between">
                    <div className="flex-1 min-w-0">
                      <div className={`text-sm font-medium truncate ${
                        tenant.tenantId === currentTenantId ? 'text-blue-700' : 'text-gray-900'
                      }`}>
                        {tenant.tenantName}
                      </div>
                      <div className="text-xs text-gray-500 capitalize mt-0.5">
                        {tenant.role}
                      </div>
                    </div>
                    {tenant.tenantId === currentTenantId && (
                      <svg className="w-5 h-5 text-blue-600 flex-shrink-0 ml-2" fill="currentColor" viewBox="0 0 20 20">
                        <path fillRule="evenodd" d="M16.707 5.293a1 1 0 010 1.414l-8 8a1 1 0 01-1.414 0l-4-4a1 1 0 011.414-1.414L8 12.586l7.293-7.293a1 1 0 011.414 0z" clipRule="evenodd" />
                      </svg>
                    )}
                  </div>
                </button>
              ))}
            </div>
          </div>
        </>
      )}
    </div>
  )
}
