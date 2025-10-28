'use client';

import React from 'react';

export const EnvironmentCheck: React.FC = () => {
    const firebaseConfigured = !!(
        process.env.NEXT_PUBLIC_FIREBASE_API_KEY &&
        process.env.NEXT_PUBLIC_FIREBASE_PROJECT_ID &&
        process.env.NEXT_PUBLIC_FIREBASE_AUTH_DOMAIN
    );

    if (process.env.NODE_ENV === 'development') {
        return (
            <div className="fixed bottom-4 right-4 bg-white border border-gray-300 rounded-lg p-3 shadow-lg text-xs max-w-xs">
                <h4 className="font-semibold text-gray-800 mb-2">Debug Info</h4>
                <div className="space-y-1">
                    <div className={`${firebaseConfigured ? 'text-green-600' : 'text-red-600'}`}>
                        Firebase: {firebaseConfigured ? '✓ Configured' : '✗ Not configured'}
                    </div>
                    <div className="text-gray-600">
                        API URL: {process.env.NEXT_PUBLIC_API_URL || 'http://localhost:8081'}
                    </div>
                </div>
            </div>
        );
    }

    return null;
};
