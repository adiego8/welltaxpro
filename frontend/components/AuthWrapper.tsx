'use client';

import dynamic from 'next/dynamic';
import React from 'react';

// Dynamically import AuthProvider to ensure it only runs on client
const DynamicAuthProvider = dynamic(
    () => import('@/contexts/AuthContext').then(mod => ({ default: mod.AuthProvider })),
    {
        ssr: false,
        loading: () => (
            <div className="min-h-screen bg-gray-50 flex items-center justify-center">
                <div className="text-center">
                    <div className="w-8 h-8 border-2 border-gray-300 border-t-blue-600 rounded-full animate-spin mx-auto mb-4" />
                    <p className="text-gray-600">Loading...</p>
                </div>
            </div>
        ),
    }
);

interface AuthWrapperProps {
    children: React.ReactNode;
}

export const AuthWrapper: React.FC<AuthWrapperProps> = ({ children }) => {
    return (
        <DynamicAuthProvider>
            {children}
        </DynamicAuthProvider>
    );
};
