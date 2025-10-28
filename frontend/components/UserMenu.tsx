'use client';

import React, { useState, useRef, useEffect } from 'react';
import Image from 'next/image';
import Link from 'next/link';
import { useAuth } from '@/contexts/AuthContext';

interface TenantAccess {
  tenantId: string;
  tenantName: string;
  role: string;
  isActive: boolean;
}

export const UserMenu: React.FC = () => {
    const { user, employee, signOut } = useAuth();
    const [isOpen, setIsOpen] = useState(false);
    const [tenants, setTenants] = useState<TenantAccess[]>([]);
    const menuRef = useRef<HTMLDivElement>(null);

    useEffect(() => {
        const handleClickOutside = (event: MouseEvent) => {
            if (menuRef.current && !menuRef.current.contains(event.target as Node)) {
                setIsOpen(false);
            }
        };

        document.addEventListener('mousedown', handleClickOutside);
        return () => document.removeEventListener('mousedown', handleClickOutside);
    }, []);

    // Fetch user's tenant access
    useEffect(() => {
        async function fetchTenants() {
            if (!user || !employee) return;

            try {
                const apiUrl = process.env.NEXT_PUBLIC_API_URL || 'http://localhost:8081';
                const idToken = await user.getIdToken();

                const response = await fetch(`${apiUrl}/api/v1/employees/me/tenants`, {
                    headers: {
                        'Authorization': `Bearer ${idToken}`,
                        'Content-Type': 'application/json',
                    },
                });

                if (response.ok) {
                    const data = await response.json();
                    const activeTenants = (data || []).filter((t: TenantAccess) => t.isActive);
                    setTenants(activeTenants);
                }
            } catch (err) {
                console.error('Failed to fetch tenants:', err);
            }
        }

        fetchTenants();
    }, [user, employee]);

    const handleSignOut = async () => {
        try {
            await signOut();
            setIsOpen(false);
        } catch (error) {
            console.error('Failed to sign out:', error);
        }
    };

    if (!user) return null;

    const defaultTenant = tenants[0];

    return (
        <div className="relative" ref={menuRef}>
            <button
                onClick={() => setIsOpen(!isOpen)}
                className="flex items-center gap-2 p-2 rounded-lg hover:bg-gray-100 transition-colors"
            >
                <Image
                    src={user.photoURL || '/default-avatar.svg'}
                    alt={user.displayName || 'User'}
                    width={32}
                    height={32}
                    className="w-8 h-8 rounded-full"
                />
                <span className="text-sm font-medium text-gray-700 hidden sm:block">
                    {user.displayName || user.email}
                </span>
                <svg
                    className={`w-4 h-4 text-gray-500 transition-transform ${isOpen ? 'rotate-180' : ''}`}
                    fill="none"
                    stroke="currentColor"
                    viewBox="0 0 24 24"
                >
                    <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M19 9l-7 7-7-7" />
                </svg>
            </button>

            {isOpen && (
                <div className="absolute right-0 mt-2 w-64 bg-white rounded-xl shadow-lg border border-gray-200 py-2 z-50">
                    {/* User Info */}
                    <div className="px-4 py-3 border-b border-gray-100">
                        <p className="text-sm font-medium text-gray-900">
                            {employee?.firstName && employee?.lastName
                                ? `${employee.firstName} ${employee.lastName}`
                                : user.displayName
                            }
                        </p>
                        <p className="text-xs text-gray-500 mt-0.5">{user.email}</p>
                        {employee?.role && (
                            <p className="text-xs text-blue-600 font-medium capitalize mt-1">
                                {employee.role}
                            </p>
                        )}
                    </div>

                    {/* Navigation Links */}
                    <div className="py-2">
                        <Link
                            href="/"
                            onClick={() => setIsOpen(false)}
                            className="flex items-center px-4 py-2 text-sm text-gray-700 hover:bg-gray-50 transition-colors"
                        >
                            <svg className="w-4 h-4 mr-3 text-gray-400" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                                <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M3 12l2-2m0 0l7-7 7 7M5 10v10a1 1 0 001 1h3m10-11l2 2m-2-2v10a1 1 0 01-1 1h-3m-6 0a1 1 0 001-1v-4a1 1 0 011-1h2a1 1 0 011 1v4a1 1 0 001 1m-6 0h6" />
                            </svg>
                            Home
                        </Link>

                        {defaultTenant && (
                            <>
                                <div className="border-t border-gray-100 my-2"></div>
                                <div className="px-4 py-2">
                                    <p className="text-xs font-medium text-gray-500 uppercase tracking-wide">Quick Access</p>
                                </div>

                                <Link
                                    href={`/${defaultTenant.tenantId}/clients`}
                                    onClick={() => setIsOpen(false)}
                                    className="flex items-center px-4 py-2 text-sm text-gray-700 hover:bg-blue-50 hover:text-blue-700 transition-colors"
                                >
                                    <svg className="w-4 h-4 mr-3 text-blue-500" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                                        <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M17 20h5v-2a3 3 0 00-5.356-1.857M17 20H7m10 0v-2c0-.656-.126-1.283-.356-1.857M7 20H2v-2a3 3 0 015.356-1.857M7 20v-2c0-.656.126-1.283.356-1.857m0 0a5.002 5.002 0 019.288 0M15 7a3 3 0 11-6 0 3 3 0 016 0zm6 3a2 2 0 11-4 0 2 2 0 014 0zM7 10a2 2 0 11-4 0 2 2 0 014 0z" />
                                    </svg>
                                    Clients
                                </Link>

                                <Link
                                    href={`/${defaultTenant.tenantId}/affiliates`}
                                    onClick={() => setIsOpen(false)}
                                    className="flex items-center px-4 py-2 text-sm text-gray-700 hover:bg-purple-50 hover:text-purple-700 transition-colors"
                                >
                                    <svg className="w-4 h-4 mr-3 text-purple-500" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                                        <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M12 4.354a4 4 0 110 5.292M15 21H3v-1a6 6 0 0112 0v1zm0 0h6v-1a6 6 0 00-9-5.197M13 7a4 4 0 11-8 0 4 4 0 018 0z" />
                                    </svg>
                                    Affiliates
                                </Link>

                                <Link
                                    href={`/${defaultTenant.tenantId}/commissions`}
                                    onClick={() => setIsOpen(false)}
                                    className="flex items-center px-4 py-2 text-sm text-gray-700 hover:bg-green-50 hover:text-green-700 transition-colors"
                                >
                                    <svg className="w-4 h-4 mr-3 text-green-500" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                                        <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M12 8c-1.657 0-3 .895-3 2s1.343 2 3 2 3 .895 3 2-1.343 2-3 2m0-8c1.11 0 2.08.402 2.599 1M12 8V7m0 1v8m0 0v1m0-1c-1.11 0-2.08-.402-2.599-1M21 12a9 9 0 11-18 0 9 9 0 0118 0z" />
                                    </svg>
                                    Commissions
                                </Link>

                                {employee?.role === 'admin' && (
                                    <Link
                                        href={`/${defaultTenant.tenantId}/employees`}
                                        onClick={() => setIsOpen(false)}
                                        className="flex items-center px-4 py-2 text-sm text-gray-700 hover:bg-orange-50 hover:text-orange-700 transition-colors"
                                    >
                                        <svg className="w-4 h-4 mr-3 text-orange-500" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                                            <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M21 13.255A23.931 23.931 0 0112 15c-3.183 0-6.22-.62-9-1.745M16 6V4a2 2 0 00-2-2h-4a2 2 0 00-2 2v2m4 6h.01M5 20h14a2 2 0 002-2V8a2 2 0 00-2-2H5a2 2 0 00-2 2v10a2 2 0 002 2z" />
                                        </svg>
                                        Employees
                                    </Link>
                                )}
                            </>
                        )}

                        {employee?.role === 'admin' && (
                            <>
                                <div className="border-t border-gray-100 my-2"></div>
                                <Link
                                    href="/admin/tenants"
                                    onClick={() => setIsOpen(false)}
                                    className="flex items-center px-4 py-2 text-sm text-gray-700 hover:bg-gray-50 transition-colors"
                                >
                                    <svg className="w-4 h-4 mr-3 text-gray-500" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                                        <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M10.325 4.317c.426-1.756 2.924-1.756 3.35 0a1.724 1.724 0 002.573 1.066c1.543-.94 3.31.826 2.37 2.37a1.724 1.724 0 001.065 2.572c1.756.426 1.756 2.924 0 3.35a1.724 1.724 0 00-1.066 2.573c.94 1.543-.826 3.31-2.37 2.37a1.724 1.724 0 00-2.572 1.065c-.426 1.756-2.924 1.756-3.35 0a1.724 1.724 0 00-2.573-1.066c-1.543.94-3.31-.826-2.37-2.37a1.724 1.724 0 00-1.065-2.572c-1.756-.426-1.756-2.924 0-3.35a1.724 1.724 0 001.066-2.573c-.94-1.543.826-3.31 2.37-2.37.996.608 2.296.07 2.572-1.065z" />
                                        <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M15 12a3 3 0 11-6 0 3 3 0 016 0z" />
                                    </svg>
                                    Account Management
                                </Link>
                            </>
                        )}
                    </div>

                    {/* Sign Out */}
                    <div className="border-t border-gray-100 pt-2">
                        <button
                            onClick={handleSignOut}
                            className="w-full flex items-center px-4 py-2 text-sm text-red-600 hover:bg-red-50 transition-colors"
                        >
                            <svg className="w-4 h-4 mr-3" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                                <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M17 16l4-4m0 0l-4-4m4 4H7m6 4v1a3 3 0 01-3 3H6a3 3 0 01-3-3V7a3 3 0 013-3h4a3 3 0 013 3v1" />
                            </svg>
                            Sign out
                        </button>
                    </div>
                </div>
            )}
        </div>
    );
};
