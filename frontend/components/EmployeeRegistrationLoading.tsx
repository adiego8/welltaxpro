'use client';

import React from 'react';

export const EmployeeRegistrationLoading: React.FC = () => {
    return (
        <div className="min-h-screen bg-gray-50 flex items-center justify-center">
            <div className="max-w-md mx-auto text-center p-6 bg-white rounded-lg shadow-lg">
                <div className="w-12 h-12 border-4 border-gray-300 border-t-blue-600 rounded-full animate-spin mx-auto mb-4" />
                <h2 className="text-xl font-semibold text-gray-900 mb-2">Setting up your account</h2>
                <p className="text-gray-600">
                    We&apos;re registering your account in our system. This will just take a moment...
                </p>
            </div>
        </div>
    );
};
