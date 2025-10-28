'use client';

import React from 'react';
import { GoogleSignInButton } from '@/components/GoogleSignInButton';

export const LoginPage: React.FC = () => {
    return (
        <div className="min-h-screen bg-gray-50 flex flex-col justify-center py-12 sm:px-6 lg:px-8">
            <div className="sm:mx-auto sm:w-full sm:max-w-md">
                <div className="text-center mb-8">
                    <img src="/logo.png" alt="WellTaxPro" className="h-32 mx-auto" />
                </div>
            </div>

            <div className="sm:mx-auto sm:w-full sm:max-w-md">
                <div className="bg-white py-8 px-4 shadow sm:rounded-lg sm:px-10">
                    <div className="space-y-6">
                        <div className="text-center">
                            <GoogleSignInButton className="w-full" />
                        </div>

                        <div className="mt-6">
                            <div className="relative">
                                <div className="absolute inset-0 flex items-center">
                                    <div className="w-full border-t border-gray-300" />
                                </div>
                                <div className="relative flex justify-center text-sm">
                                    <span className="px-2 bg-white text-gray-500">Secure authentication</span>
                                </div>
                            </div>
                        </div>

                        <div className="text-center text-xs text-gray-500">
                            By signing in, you agree to our terms of service and privacy policy.
                        </div>
                    </div>
                </div>
            </div>
        </div>
    );
};
