'use client';

import { useEffect, useState } from 'react';
import { useRouter, useParams } from 'next/navigation';

interface ClientData {
  client: {
    id: string;
    firstName: string | null;
    middleName: string | null;
    lastName: string | null;
    email: string;
    phone: string | null;
    dob: string | null;
    ssn: string | null;
    address1: string | null;
    address2: string | null;
    city: string | null;
    state: string | null;
    zipcode: number | null;
    role: string;
    createdAt: string;
  };
  spouse: {
    id: string;
    userId: string;
    firstName: string;
    middleName: string | null;
    lastName: string;
    email: string | null;
    phone: string | null;
    dob: string;
    ssn: string;
    isDeath: boolean;
    deathDate: string | null;
    createdAt: string;
  } | null;
  dependents: Array<{
    id: string;
    userId: string;
    firstName: string;
    middleName: string | null;
    lastName: string;
    dob: string;
    ssn: string;
    relationship: string;
    timeWithApplicant: string;
    exclusiveClaim: boolean;
    createdAt: string;
    updatedAt: string | null;
  }>;
  filings: Array<{
    id: string;
    year: number;
    userId: string;
    maritalStatus: string | null;
    spouseId: string | null;
    sourceOfIncome: string[];
    deductions: string[];
    income: number | null;
    marketplaceInsurance: boolean | null;
    createdAt: string;
    updatedAt: string | null;
    status?: {
      id: string;
      filingId: string;
      latestStep: number;
      isCompleted: boolean;
      status: string;
    };
    documents?: Array<{
      id: string;
      name: string;
      type: string;
      createdAt: string;
    }>;
    properties?: Array<{
      id: string;
      address1: string;
      address2: string | null;
      city: string;
      state: string;
      zipcode: string;
      purchasePrice: number;
      closingCost: number;
      purchaseDate: string;
      rents: number | null;
      royalties: number | null;
      expenses?: Array<{
        id: string;
        name: string;
        amount: number;
      }>;
    }>;
    iraContributions?: Array<{
      id: string;
      accountType: string;
      amount: number;
    }>;
    charities?: Array<{
      id: string;
      name: string;
      contribution: number;
    }>;
    childcares?: Array<{
      id: string;
      name: string;
      amount: number;
      taxId: string;
      address1: string;
      address2: string | null;
      city: string;
      state: string;
      zipcode: string;
    }>;
    payments?: Array<{
      id: string;
      amount: number;
      originalAmount: number | null;
      discountAmount: number | null;
      discountCode: string | null;
      status: string;
      createdAt: string;
      items?: Array<{
        id: string;
        name: string;
        quantity: number;
        unitAmount: number;
      }>;
    }>;
    discounts?: Array<{
      id: string;
      code: string | null;
      originalAmount: number;
      discountAmount: number;
      finalAmount: number;
      appliedAt: string;
    }>;
  }>;
}

export default function PortalDashboard() {
  const router = useRouter();
  const params = useParams();
  const tenantId = params.tenantId as string;

  const [clientData, setClientData] = useState<ClientData | null>(null);
  const [expandedFilings, setExpandedFilings] = useState<Set<string>>(new Set());
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [clientEmail, setClientEmail] = useState<string>('');

  useEffect(() => {
    // Check if user has valid session
    const token = sessionStorage.getItem('portalToken');
    const email = sessionStorage.getItem('portalEmail');

    if (!token) {
      router.push(`/${tenantId}/portal`);
      return;
    }

    if (email) {
      setClientEmail(email);
    }

    fetchClientData(token);
  }, [tenantId, router]);

  const fetchClientData = async (token: string) => {
    try {
      const apiUrl = process.env.NEXT_PUBLIC_API_URL || 'http://localhost:8080';
      const response = await fetch(
        `${apiUrl}/api/v1/${tenantId}/portal/client`,
        {
          headers: {
            'Authorization': `Bearer ${token}`
          }
        }
      );

      if (response.status === 401) {
        // Token expired, redirect to portal entry
        sessionStorage.clear();
        router.push(`/${tenantId}/portal`);
        return;
      }

      if (!response.ok) {
        throw new Error('Failed to fetch client data');
      }

      const data = await response.json();
      setClientData(data);
      setLoading(false);
    } catch (err) {
      console.error('Error fetching client data:', err);
      setError('Failed to load your information. Please try again.');
      setLoading(false);
    }
  };

  const toggleFiling = (filingId: string) => {
    setExpandedFilings(prev => {
      const next = new Set(prev);
      if (next.has(filingId)) {
        next.delete(filingId);
      } else {
        next.add(filingId);
      }
      return next;
    });
  };

  const handleDownloadDocument = async (documentId: string) => {
    const token = sessionStorage.getItem('portalToken');
    if (!token) {
      setError('Session expired. Please login again.');
      return;
    }

    try {
      const apiUrl = process.env.NEXT_PUBLIC_API_URL || 'http://localhost:8080';
      const response = await fetch(
        `${apiUrl}/api/v1/${tenantId}/portal/documents/${documentId}/download`,
        {
          headers: {
            'Authorization': `Bearer ${token}`
          }
        }
      );

      if (!response.ok) {
        throw new Error('Failed to download document');
      }

      // Get filename from Content-Disposition header or use default
      const contentDisposition = response.headers.get('Content-Disposition');
      let filename = 'document.pdf';
      if (contentDisposition) {
        const match = contentDisposition.match(/filename="?(.+)"?/);
        if (match) {
          filename = match[1];
        }
      }

      // Download the file
      const blob = await response.blob();
      const url = window.URL.createObjectURL(blob);
      const a = document.createElement('a');
      a.href = url;
      a.download = filename;
      document.body.appendChild(a);
      a.click();
      window.URL.revokeObjectURL(url);
      document.body.removeChild(a);
    } catch (err) {
      console.error('Error downloading document:', err);
      setError('Failed to download document. Please try again.');
    }
  };

  const handleLogout = () => {
    sessionStorage.clear();
    router.push(`/${tenantId}/portal`);
  };

  const getStatusColor = (status: string) => {
    switch (status.toUpperCase()) {
      case 'COMPLETED':
        return 'bg-green-100 text-green-800';
      case 'IN_PROGRESS':
        return 'bg-yellow-100 text-yellow-800';
      case 'PENDING':
        return 'bg-gray-100 text-gray-800';
      default:
        return 'bg-blue-100 text-blue-800';
    }
  };

  if (loading) {
    return (
      <div className="min-h-screen bg-gray-50 flex items-center justify-center">
        <div className="animate-spin rounded-full h-12 w-12 border-b-2 border-blue-600"></div>
      </div>
    );
  }

  return (
    <div className="min-h-screen bg-gray-50">
      {/* Header */}
      <div className="bg-white shadow">
        <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-4">
          <div className="flex justify-between items-center">
            <div className="flex items-center space-x-4">
              <img src="/logo.png" alt="Logo" className="h-12" />
              <div>
                <h1 className="text-2xl font-bold text-gray-900">Client Portal</h1>
                <p className="text-sm text-gray-500">{clientEmail}</p>
              </div>
            </div>
            <button
              onClick={handleLogout}
              className="px-4 py-2 text-sm font-medium text-gray-700 bg-white border border-gray-300 rounded-md hover:bg-gray-50"
            >
              Logout
            </button>
          </div>
        </div>
      </div>

      <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-8">
        {error ? (
          <div className="rounded-md bg-red-50 p-4 mb-6">
            <p className="text-sm text-red-800">{error}</p>
          </div>
        ) : null}

        {!clientData ? (
          <div className="bg-white shadow rounded-lg p-8 text-center">
            <p className="text-gray-500">Loading your information...</p>
          </div>
        ) : (
          <>
            {/* Client Information */}
            <div className="bg-white shadow overflow-hidden sm:rounded-lg mb-6">
              <div className="px-4 py-5 sm:px-6">
                <h3 className="text-lg leading-6 font-medium text-gray-900">Your Information</h3>
                <p className="mt-1 max-w-2xl text-sm text-gray-500">Personal details</p>
              </div>
              <div className="border-t border-gray-200">
                <dl>
                  <div className="bg-gray-50 px-4 py-5 sm:grid sm:grid-cols-3 sm:gap-4 sm:px-6">
                    <dt className="text-sm font-bold text-gray-600">Full Name</dt>
                    <dd className="mt-1 text-sm text-gray-900 sm:mt-0 sm:col-span-2">
                      {clientData.client.firstName} {clientData.client.middleName && `${clientData.client.middleName} `}{clientData.client.lastName}
                    </dd>
                  </div>
                  <div className="bg-white px-4 py-5 sm:grid sm:grid-cols-3 sm:gap-4 sm:px-6">
                    <dt className="text-sm font-bold text-gray-600">Email</dt>
                    <dd className="mt-1 text-sm text-gray-900 sm:mt-0 sm:col-span-2">{clientData.client.email}</dd>
                  </div>
                  <div className="bg-gray-50 px-4 py-5 sm:grid sm:grid-cols-3 sm:gap-4 sm:px-6">
                    <dt className="text-sm font-bold text-gray-600">Phone</dt>
                    <dd className="mt-1 text-sm text-gray-900 sm:mt-0 sm:col-span-2">{clientData.client.phone || '-'}</dd>
                  </div>
                  <div className="bg-white px-4 py-5 sm:grid sm:grid-cols-3 sm:gap-4 sm:px-6">
                    <dt className="text-sm font-bold text-gray-600">Date of Birth</dt>
                    <dd className="mt-1 text-sm text-gray-900 sm:mt-0 sm:col-span-2">
                      {clientData.client.dob ? new Date(clientData.client.dob).toLocaleDateString() : '-'}
                    </dd>
                  </div>
                  <div className="bg-gray-50 px-4 py-5 sm:grid sm:grid-cols-3 sm:gap-4 sm:px-6">
                    <dt className="text-sm font-bold text-gray-600">SSN</dt>
                    <dd className="mt-1 text-sm text-gray-900 sm:mt-0 sm:col-span-2">{clientData.client.ssn || '-'}</dd>
                  </div>
                  <div className="bg-white px-4 py-5 sm:grid sm:grid-cols-3 sm:gap-4 sm:px-6">
                    <dt className="text-sm font-bold text-gray-600">Address</dt>
                    <dd className="mt-1 text-sm text-gray-900 sm:mt-0 sm:col-span-2">
                      {clientData.client.address1 || '-'}<br />
                      {clientData.client.address2 && <>{clientData.client.address2}<br /></>}
                      {clientData.client.city && clientData.client.state && `${clientData.client.city}, ${clientData.client.state} ${clientData.client.zipcode || ''}`}
                    </dd>
                  </div>
                </dl>
              </div>
            </div>

            {/* Spouse Information */}
            {clientData.spouse && (
              <div className="bg-white shadow overflow-hidden sm:rounded-lg mb-6">
                <div className="px-4 py-5 sm:px-6">
                  <h3 className="text-lg leading-6 font-medium text-gray-900">Spouse Information</h3>
                  <p className="mt-1 max-w-2xl text-sm text-gray-500">Spouse details</p>
                </div>
                <div className="border-t border-gray-200">
                  <dl>
                    <div className="bg-gray-50 px-4 py-5 sm:grid sm:grid-cols-3 sm:gap-4 sm:px-6">
                      <dt className="text-sm font-bold text-gray-600">Full Name</dt>
                      <dd className="mt-1 text-sm text-gray-900 sm:mt-0 sm:col-span-2">
                        {clientData.spouse.firstName} {clientData.spouse.middleName && `${clientData.spouse.middleName} `}{clientData.spouse.lastName}
                      </dd>
                    </div>
                    <div className="bg-white px-4 py-5 sm:grid sm:grid-cols-3 sm:gap-4 sm:px-6">
                      <dt className="text-sm font-bold text-gray-600">Email</dt>
                      <dd className="mt-1 text-sm text-gray-900 sm:mt-0 sm:col-span-2">{clientData.spouse.email || '-'}</dd>
                    </div>
                    <div className="bg-gray-50 px-4 py-5 sm:grid sm:grid-cols-3 sm:gap-4 sm:px-6">
                      <dt className="text-sm font-bold text-gray-600">Phone</dt>
                      <dd className="mt-1 text-sm text-gray-900 sm:mt-0 sm:col-span-2">{clientData.spouse.phone || '-'}</dd>
                    </div>
                    <div className="bg-white px-4 py-5 sm:grid sm:grid-cols-3 sm:gap-4 sm:px-6">
                      <dt className="text-sm font-bold text-gray-600">Date of Birth</dt>
                      <dd className="mt-1 text-sm text-gray-900 sm:mt-0 sm:col-span-2">{new Date(clientData.spouse.dob).toLocaleDateString()}</dd>
                    </div>
                    <div className="bg-gray-50 px-4 py-5 sm:grid sm:grid-cols-3 sm:gap-4 sm:px-6">
                      <dt className="text-sm font-bold text-gray-600">SSN</dt>
                      <dd className="mt-1 text-sm text-gray-900 sm:mt-0 sm:col-span-2">{clientData.spouse.ssn}</dd>
                    </div>
                  </dl>
                </div>
              </div>
            )}

            {/* Dependents */}
            {clientData.dependents && clientData.dependents.length > 0 && (
              <div className="bg-white shadow overflow-hidden sm:rounded-lg mb-6">
                <div className="px-4 py-5 sm:px-6">
                  <h3 className="text-lg leading-6 font-medium text-gray-900">Dependents ({clientData.dependents.length})</h3>
                  <p className="mt-1 max-w-2xl text-sm text-gray-500">Dependent information</p>
                </div>
                <div className="border-t border-gray-200 divide-y divide-gray-200">
                  {clientData.dependents.map((dep, idx) => (
                    <div key={dep.id}>
                      <div className="px-4 py-3 bg-gray-50">
                        <h4 className="text-sm font-semibold text-gray-900">Dependent {idx + 1}</h4>
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
                      </dl>
                    </div>
                  ))}
                </div>
              </div>
            )}

            {/* Tax Filings */}
            <div className="bg-white shadow overflow-hidden sm:rounded-lg">
              <div className="px-4 py-5 sm:px-6 bg-gradient-to-r from-blue-50 to-indigo-50">
                <h3 className="text-lg leading-6 font-medium text-gray-900">
                  Tax Filings ({clientData.filings ? clientData.filings.length : 0})
                </h3>
                <p className="mt-1 text-sm text-gray-500">Complete filing history</p>
              </div>
              {clientData.filings && clientData.filings.length > 0 ? (
                <div className="border-t border-gray-200 divide-y divide-gray-200">
                  {clientData.filings.map((filing) => {
                    const isExpanded = expandedFilings.has(filing.id);
                    const totalDocuments = filing.documents?.length || 0;
                    const totalProperties = filing.properties?.length || 0;
                    const totalPayments = filing.payments?.length || 0;

                    return (
                      <div key={filing.id} className="hover:bg-gray-50 transition-colors">
                        {/* Collapsed View - Summary */}
                        <div
                          className="p-6 cursor-pointer"
                          onClick={() => toggleFiling(filing.id)}
                        >
                          <div className="flex items-center justify-between">
                            <div className="flex items-center space-x-4">
                              <div className="flex-shrink-0">
                                <div className="h-12 w-12 rounded-lg bg-gradient-to-br from-blue-500 to-indigo-600 flex items-center justify-center text-white font-bold text-lg">
                                  {filing.year}
                                </div>
                              </div>
                              <div>
                                <h4 className="text-lg font-semibold text-gray-900">Tax Year {filing.year}</h4>
                                <div className="flex items-center space-x-4 mt-1">
                                  <span className="text-sm text-gray-600">
                                    Income: <span className="font-medium text-gray-900">${filing.income?.toLocaleString() || 0}</span>
                                  </span>
                                  <span className="text-sm text-gray-500">•</span>
                                  <span className="text-sm text-gray-600">
                                    {filing.maritalStatus || 'Status Unknown'}
                                  </span>
                                  {totalDocuments > 0 && (
                                    <>
                                      <span className="text-sm text-gray-500">•</span>
                                      <span className="text-sm text-gray-600">{totalDocuments} doc{totalDocuments !== 1 ? 's' : ''}</span>
                                    </>
                                  )}
                                </div>
                              </div>
                            </div>
                            <div className="flex items-center space-x-3">
                              <span className={`px-3 py-1 text-xs font-semibold rounded-full ${
                                filing.status?.isCompleted
                                  ? 'bg-green-100 text-green-800'
                                  : 'bg-yellow-100 text-yellow-800'
                              }`}>
                                {filing.status?.status || 'IN_PROGRESS'}
                              </span>
                              <svg
                                className={`h-6 w-6 text-gray-400 transition-transform ${isExpanded ? 'rotate-180' : ''}`}
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

                            {/* Documents - With Download */}
                            {filing.documents && filing.documents.length > 0 && (
                              <div className="bg-white rounded-lg border border-gray-200">
                                <div className="px-5 py-4 border-b border-gray-200 bg-gray-50">
                                  <h5 className="text-sm font-semibold text-gray-900 flex items-center">
                                    <svg className="w-5 h-5 mr-2 text-gray-600" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                                      <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M7 21h10a2 2 0 002-2V9.414a1 1 0 00-.293-.707l-5.414-5.414A1 1 0 0012.586 3H7a2 2 0 00-2 2v14a2 2 0 002 2z" />
                                    </svg>
                                    Documents ({filing.documents.length})
                                  </h5>
                                </div>
                                <div className="p-5 space-y-2">
                                  {filing.documents.map((doc) => (
                                    <div key={doc.id} className="flex items-center justify-between p-3 bg-gray-50 rounded-lg hover:bg-gray-100 transition">
                                      <div className="flex items-center space-x-3 flex-1">
                                        <svg className="h-8 w-8 text-blue-500" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                                          <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M7 21h10a2 2 0 002-2V9.414a1 1 0 00-.293-.707l-5.414-5.414A1 1 0 0012.586 3H7a2 2 0 00-2 2v14a2 2 0 002 2z" />
                                        </svg>
                                        <div className="flex-1">
                                          <h4 className="text-sm font-medium text-gray-900">{doc.name}</h4>
                                          <p className="text-xs text-gray-500">{doc.type} • {new Date(doc.createdAt).toLocaleDateString()}</p>
                                        </div>
                                      </div>
                                      <button
                                        onClick={() => handleDownloadDocument(doc.id)}
                                        className="ml-4 px-3 py-1.5 bg-blue-600 text-white text-xs font-medium rounded hover:bg-blue-700 flex items-center space-x-1"
                                      >
                                        <svg className="w-4 h-4" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                                          <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M4 16v1a3 3 0 003 3h10a3 3 0 003-3v-1m-4-4l-4 4m0 0l-4-4m4 4V4" />
                                        </svg>
                                        <span>Download</span>
                                      </button>
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
                                      <p className="text-xs text-gray-500 mt-3">{new Date(payment.createdAt).toLocaleString()}</p>
                                    </div>
                                  ))}
                                </div>
                              </div>
                            )}
                          </div>
                        )}
                      </div>
                    );
                  })}
                </div>
              ) : (
                <div className="border-t border-gray-200 px-4 py-5 sm:px-6 text-center">
                  <p className="text-gray-500">No tax filings found</p>
                </div>
              )}
            </div>
          </>
        )}
      </div>
    </div>
  );
}
