'use client';

import { useEffect, useState } from 'react';
import { useParams, useRouter } from 'next/navigation';
import { useAuth } from '@/contexts/AuthContext';

interface Filing {
  id: string;
  year: number;
  status: {
    status: string;
    isCompleted: boolean;
    updatedAt: string;
  };
  filingType?: string;
  maritalStatus?: string;
  income?: number;
  sourceOfIncome?: string[];
  deductions?: string[];
  createdAt: string;
  updatedAt: string;
  documents?: Document[];
  properties?: any[];
  payments?: any[];
  iraContributions?: any[];
  charities?: any[];
  childcares?: any[];
}

interface Document {
  id: string;
  name: string;
  type: string;
  createdAt: string;
  filePath: string;
}

interface Spouse {
  id: string;
  firstName: string;
  middleName?: string;
  lastName: string;
  email?: string;
  phone?: string;
  dob: string;
  ssn: string;
  isDeath: boolean;
  deathDate?: string;
  createdAt: string;
}

interface Dependent {
  id: string;
  firstName: string;
  middleName?: string;
  lastName: string;
  dob: string;
  ssn: string;
  relationship: string;
  timeWithApplicant: string;
  exclusiveClaim: boolean;
  documents: string[];
  createdAt: string;
}

interface Client {
  id: string;
  authUserName: string;
  firstName: string;
  middleName?: string;
  lastName: string;
  email: string;
  phone?: string;
  address1?: string;
  address2?: string;
  city?: string;
  state?: string;
  zipcode?: string;
  dob?: string;
  ssn?: string;
  role?: string;
  createdAt: string;
}

interface ClientData {
  client: Client;
  spouse?: Spouse;
  dependents?: Dependent[];
  filings: Filing[];
}

export default function TenantUserPortalPage() {
  const params = useParams();
  const router = useRouter();
  const { user, loading: authLoading, signOut } = useAuth();
  const tenantId = params.tenantId as string;

  const [clientData, setClientData] = useState<ClientData | null>(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [expandedFilings, setExpandedFilings] = useState<Set<string>>(new Set());

  useEffect(() => {
    if (!authLoading && !user) {
      router.push(`/signin/${tenantId}`);
      return;
    }

    async function fetchClientData() {
      if (!user) return;

      try {
        const apiUrl = process.env.NEXT_PUBLIC_API_URL || 'http://localhost:8081';
        const idToken = await user.getIdToken();

        const response = await fetch(`${apiUrl}/api/v1/${tenantId}/user/profile`, {
          headers: {
            'Authorization': `Bearer ${idToken}`,
            'Content-Type': 'application/json',
          },
        });

        if (response.status === 404) {
          setError('You are not registered for portal access. Please contact your tax professional.');
          setLoading(false);
          return;
        }

        if (!response.ok) {
          throw new Error('Failed to fetch profile data');
        }

        const data = await response.json();
        setClientData(data);
      } catch (err) {
        console.error('Failed to fetch client data:', err);
        setError('Failed to load your profile. Please try again later.');
      } finally {
        setLoading(false);
      }
    }

    if (user) {
      fetchClientData();
    }
  }, [user, authLoading, tenantId, router]);

  const handleSignOut = async () => {
    await signOut();
    router.push(`/signin/${tenantId}`);
  };

  const handleDocumentDownload = async (documentId: string) => {
    if (!user) return;

    try {
      const apiUrl = process.env.NEXT_PUBLIC_API_URL || 'http://localhost:8081';
      const idToken = await user.getIdToken();

      const response = await fetch(`${apiUrl}/api/v1/${tenantId}/user/documents/${documentId}/download`, {
        headers: {
          'Authorization': `Bearer ${idToken}`,
        },
      });

      if (!response.ok) {
        throw new Error('Failed to download document');
      }

      const contentDisposition = response.headers.get('Content-Disposition');
      let filename = 'document';
      if (contentDisposition) {
        const match = contentDisposition.match(/filename="(.+)"/);
        if (match) filename = match[1];
      }

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
      console.error('Failed to download document:', err);
      alert('Failed to download document. Please try again.');
    }
  };

  const toggleFiling = (filingId: string) => {
    setExpandedFilings(prev => {
      const newSet = new Set(prev);
      if (newSet.has(filingId)) {
        newSet.delete(filingId);
      } else {
        newSet.add(filingId);
      }
      return newSet;
    });
  };

  if (authLoading || loading) {
    return (
      <div className="min-h-screen bg-gray-50 flex items-center justify-center">
        <div className="text-center">
          <div className="w-8 h-8 border-2 border-gray-300 border-t-blue-600 rounded-full animate-spin mx-auto mb-4" />
          <p className="text-gray-600">Loading your information...</p>
        </div>
      </div>
    );
  }

  if (error) {
    return (
      <div className="min-h-screen bg-gray-50 flex items-center justify-center p-4">
        <div className="max-w-md w-full bg-white rounded-lg shadow-lg p-8 text-center">
          <div className="w-16 h-16 bg-red-100 rounded-full flex items-center justify-center mx-auto mb-4">
            <svg className="w-8 h-8 text-red-600" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M12 9v2m0 4h.01m-6.938 4h13.856c1.54 0 2.502-1.667 1.732-3L13.732 4c-.77-1.333-2.694-1.333-3.464 0L3.34 16c-.77 1.333.192 3 1.732 3z" />
            </svg>
          </div>
          <h2 className="text-xl font-bold text-gray-900 mb-2">Access Denied</h2>
          <p className="text-gray-600 mb-6">{error}</p>
          <button
            onClick={handleSignOut}
            className="w-full px-4 py-2 bg-blue-600 text-white rounded-lg hover:bg-blue-700 transition-colors"
          >
            Sign Out
          </button>
        </div>
      </div>
    );
  }

  if (!clientData) {
    return null;
  }

  const { client, spouse, dependents, filings } = clientData;

  return (
    <div className="min-h-screen bg-gray-50">
      {/* Header */}
      <header className="bg-white shadow-sm">
        <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-4 flex justify-between items-center">
          <h1 className="text-2xl font-bold text-gray-900">My Tax Portal</h1>
          <button
            onClick={handleSignOut}
            className="px-4 py-2 text-sm font-medium text-gray-700 hover:text-gray-900 hover:bg-gray-100 rounded-md transition-colors"
          >
            Sign Out
          </button>
        </div>
      </header>

      <main className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-8">
        {/* Client Information */}
        <div className="bg-white shadow overflow-hidden sm:rounded-lg mb-6">
          <div className="px-4 py-5 sm:px-6">
            <h3 className="text-lg leading-6 font-medium text-gray-900">Personal Information</h3>
            <p className="mt-1 text-sm text-gray-500">Your personal details on file</p>
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
            </dl>
          </div>
        </div>

        {/* Spouse */}
        {spouse && (
          <div className="bg-white shadow overflow-hidden sm:rounded-lg mb-6">
            <div className="px-4 py-5 sm:px-6">
              <h3 className="text-lg leading-6 font-medium text-gray-900">Spouse Information</h3>
              <p className="mt-1 text-sm text-gray-500">Spouse details</p>
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
                  <dd className="mt-1 text-sm text-gray-900 sm:mt-0 sm:col-span-2">
                    {new Date(spouse.dob).toLocaleDateString()}
                  </dd>
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
                        <dd className="mt-1 text-sm text-gray-900 sm:mt-0 sm:col-span-2">
                          {new Date(spouse.deathDate).toLocaleDateString()}
                        </dd>
                      </div>
                    )}
                  </>
                )}
              </dl>
            </div>
          </div>
        )}

        {/* Dependents */}
        {dependents && dependents.length > 0 && (
          <div className="bg-white shadow overflow-hidden sm:rounded-lg mb-6">
            <div className="px-4 py-5 sm:px-6">
              <h3 className="text-lg leading-6 font-medium text-gray-900">Dependents ({dependents.length})</h3>
              <p className="mt-1 text-sm text-gray-500">Dependent information</p>
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
                      <dt className="text-sm font-bold text-gray-600">Date of Birth</dt>
                      <dd className="mt-1 text-sm text-gray-900 sm:mt-0 sm:col-span-2">
                        {new Date(dep.dob).toLocaleDateString()}
                      </dd>
                    </div>
                    <div className="bg-white px-4 py-5 sm:grid sm:grid-cols-3 sm:gap-4 sm:px-6">
                      <dt className="text-sm font-bold text-gray-600">SSN</dt>
                      <dd className="mt-1 text-sm text-gray-900 sm:mt-0 sm:col-span-2">{dep.ssn}</dd>
                    </div>
                    <div className="bg-gray-50 px-4 py-5 sm:grid sm:grid-cols-3 sm:gap-4 sm:px-6">
                      <dt className="text-sm font-bold text-gray-600">Relationship</dt>
                      <dd className="mt-1 text-sm text-gray-900 sm:mt-0 sm:col-span-2">{dep.relationship}</dd>
                    </div>
                    <div className="bg-white px-4 py-5 sm:grid sm:grid-cols-3 sm:gap-4 sm:px-6">
                      <dt className="text-sm font-bold text-gray-600">Time with Applicant</dt>
                      <dd className="mt-1 text-sm text-gray-900 sm:mt-0 sm:col-span-2">{dep.timeWithApplicant}</dd>
                    </div>
                    <div className="bg-gray-50 px-4 py-5 sm:grid sm:grid-cols-3 sm:gap-4 sm:px-6">
                      <dt className="text-sm font-bold text-gray-600">Exclusive Claim</dt>
                      <dd className="mt-1 text-sm text-gray-900 sm:mt-0 sm:col-span-2">
                        {dep.exclusiveClaim ? 'Yes' : 'No'}
                      </dd>
                    </div>
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
            <h3 className="text-lg leading-6 font-medium text-gray-900">
              Tax Filings ({filings ? filings.length : 0})
            </h3>
            <p className="mt-1 text-sm text-gray-500">Your complete filing history</p>
          </div>
          {filings && filings.length > 0 ? (
            <div className="border-t border-gray-200 divide-y divide-gray-200">
              {filings.map((filing) => {
                const isExpanded = expandedFilings.has(filing.id);
                const totalDocuments = filing.documents?.length || 0;

                return (
                  <div key={filing.id} className="hover:bg-gray-50 transition-colors">
                    {/* Collapsed View */}
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
                            <div className="flex items-center gap-4 mt-1">
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
                                  <span className="text-sm text-gray-600">{totalDocuments} document{totalDocuments !== 1 ? 's' : ''}</span>
                                </>
                              )}
                            </div>
                          </div>
                        </div>
                        <div className="flex items-center gap-3">
                          <span className={`px-3 py-1 text-xs font-semibold rounded-full ${
                            filing.status?.isCompleted
                              ? 'bg-green-100 text-green-800'
                              : 'bg-yellow-100 text-yellow-800'
                          }`}>
                            {filing.status?.status || 'IN_PROCESS'}
                          </span>
                          <svg
                            className={`w-5 h-5 text-gray-400 transform transition-transform ${isExpanded ? 'rotate-180' : ''}`}
                            fill="none"
                            viewBox="0 0 24 24"
                            stroke="currentColor"
                          >
                            <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M19 9l-7 7-7-7" />
                          </svg>
                        </div>
                      </div>
                    </div>

                    {/* Expanded View */}
                    {isExpanded && (
                      <div className="px-6 pb-6 pt-2 bg-gray-50 border-t border-gray-200">
                        {/* Filing Details */}
                        <div className="grid grid-cols-1 md:grid-cols-2 gap-4 mb-4">
                          {filing.sourceOfIncome && filing.sourceOfIncome.length > 0 && (
                            <div>
                              <h5 className="text-sm font-semibold text-gray-700 mb-2">Sources of Income</h5>
                              <ul className="list-disc list-inside text-sm text-gray-600 space-y-1">
                                {filing.sourceOfIncome.map((source, idx) => (
                                  <li key={idx}>{source}</li>
                                ))}
                              </ul>
                            </div>
                          )}
                          {filing.deductions && filing.deductions.length > 0 && (
                            <div>
                              <h5 className="text-sm font-semibold text-gray-700 mb-2">Deductions</h5>
                              <ul className="list-disc list-inside text-sm text-gray-600 space-y-1">
                                {filing.deductions.map((deduction, idx) => (
                                  <li key={idx}>{deduction}</li>
                                ))}
                              </ul>
                            </div>
                          )}
                        </div>

                        {/* Documents */}
                        {filing.documents && filing.documents.length > 0 && (
                          <div className="mt-4">
                            <h5 className="text-sm font-semibold text-gray-700 mb-3">Documents</h5>
                            <div className="space-y-2">
                              {filing.documents.map((doc) => (
                                <div
                                  key={doc.id}
                                  className="flex items-center justify-between p-3 bg-white border border-gray-200 rounded-lg"
                                >
                                  <div className="flex items-center gap-3">
                                    <svg className="w-8 h-8 text-blue-500" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                                      <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M7 21h10a2 2 0 002-2V9.414a1 1 0 00-.293-.707l-5.414-5.414A1 1 0 0012.586 3H7a2 2 0 00-2 2v14a2 2 0 002 2z" />
                                    </svg>
                                    <div>
                                      <p className="font-medium text-gray-900">{doc.name}</p>
                                      <p className="text-sm text-gray-500">
                                        {doc.type} • {new Date(doc.createdAt).toLocaleDateString()}
                                      </p>
                                    </div>
                                  </div>
                                  <button
                                    onClick={() => handleDocumentDownload(doc.id)}
                                    className="px-4 py-2 text-sm font-medium text-blue-600 hover:text-blue-700 hover:bg-blue-50 rounded-md transition-colors"
                                  >
                                    Download
                                  </button>
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
            <div className="p-8 text-center text-gray-500">
              <p>No filings found. Your filings will appear here once available.</p>
            </div>
          )}
        </div>
      </main>
    </div>
  );
}
