'use client';

import React, { createContext, useContext, useEffect, useState } from 'react';
import {
    User,
    signInWithPopup,
    signOut as firebaseSignOut,
    onAuthStateChanged
} from 'firebase/auth';
import { auth, googleProvider } from '@/lib/firebase';
import { employeeService, Employee, TenantAccess } from '@/services/employeeService';

interface AuthContextType {
    user: User | null;
    employee: Employee | null;
    tenantAccess: TenantAccess[] | null;
    loading: boolean;
    employeeLoading: boolean;
    signInWithGoogle: () => Promise<void>;
    signOut: () => Promise<void>;
    refreshEmployee: () => Promise<void>;
}

const AuthContext = createContext<AuthContextType>({
    user: null,
    employee: null,
    tenantAccess: null,
    loading: true,
    employeeLoading: false,
    signInWithGoogle: async () => { },
    signOut: async () => { },
    refreshEmployee: async () => { },
});

export const useAuth = () => {
    const context = useContext(AuthContext);
    if (context === undefined) {
        throw new Error('useAuth must be used within an AuthProvider');
    }
    return context;
};

export const AuthProvider: React.FC<{ children: React.ReactNode }> = ({ children }) => {
    const [user, setUser] = useState<User | null>(null);
    const [employee, setEmployee] = useState<Employee | null>(null);
    const [tenantAccess, setTenantAccess] = useState<TenantAccess[] | null>(null);
    const [loading, setLoading] = useState(true);
    const [employeeLoading, setEmployeeLoading] = useState(false);

    const createEmployeeIfNeeded = async (firebaseUser: User) => {
        setEmployeeLoading(true);
        try {
            const firstName = firebaseUser.displayName?.split(' ')[0] || undefined;
            const lastName = firebaseUser.displayName?.split(' ').slice(1).join(' ') || undefined;

            const employeeData = {
                firebaseUid: firebaseUser.uid,
                email: firebaseUser.email || '',
                firstName,
                lastName,
                role: 'admin' as const
            };

            // Get the Firebase ID token to include in the request
            const idToken = await firebaseUser.getIdToken();
            const result = await employeeService.createEmployee(employeeData, idToken);
            setEmployee(result.employee);
        } catch (error) {
            console.error('Error creating employee:', error);
            // If employee creation fails, we might still want to continue
            // You can choose to handle this differently based on requirements
        } finally {
            setEmployeeLoading(false);
        }
    };

    const refreshEmployee = async () => {
        if (!user) return;

        setEmployeeLoading(true);
        try {
            const idToken = await user.getIdToken();
            const employeeData = await employeeService.getCurrentEmployee(idToken);
            setEmployee(employeeData);

            // Also fetch tenant access
            try {
                const tenantData = await employeeService.getEmployeeTenants(idToken);
                setTenantAccess(tenantData);
            } catch (tenantError) {
                console.error('Error fetching tenant access:', tenantError);
                setTenantAccess([]);
            }
        } catch (error) {
            console.error('Error fetching employee:', error);
            setEmployee(null);
            setTenantAccess(null);
        } finally {
            setEmployeeLoading(false);
        }
    };

    useEffect(() => {
        // Only set up auth listener if Firebase is configured
        if (!auth) {
            console.warn('Firebase auth not configured, skipping auth listener setup');
            setLoading(false);
            return;
        }

        try {
            const unsubscribe = onAuthStateChanged(auth, async (firebaseUser) => {
                setUser(firebaseUser);

                if (firebaseUser) {
                    // Try to get existing employee first
                    try {
                        const idToken = await firebaseUser.getIdToken();
                        const employeeData = await employeeService.getCurrentEmployee(idToken);
                        setEmployee(employeeData);

                        // Also fetch tenant access
                        try {
                            const tenantData = await employeeService.getEmployeeTenants(idToken);
                            setTenantAccess(tenantData);
                        } catch (tenantError) {
                            console.error('Error fetching tenant access:', tenantError);
                            setTenantAccess([]);
                        }
                    } catch (error) {
                        // If employee doesn't exist, create one
                        await createEmployeeIfNeeded(firebaseUser);
                    }
                } else {
                    setEmployee(null);
                    setTenantAccess(null);
                }

                setLoading(false);
            });

            return () => unsubscribe();
        } catch (error) {
            console.error('Error setting up auth listener:', error);
            setLoading(false);
        }
    }, []);

    const signInWithGoogle = async () => {
        if (!auth || !googleProvider) {
            throw new Error('Firebase not configured. Please set up your environment variables.');
        }

        try {
            const result = await signInWithPopup(auth, googleProvider);
            // The onAuthStateChanged listener will handle employee creation/fetching
        } catch (error) {
            console.error('Error signing in with Google:', error);
            throw error;
        }
    };

    const signOut = async () => {
        if (!auth) {
            throw new Error('Firebase not configured.');
        }

        try {
            await firebaseSignOut(auth);
            setEmployee(null);
            setTenantAccess(null);
        } catch (error) {
            console.error('Error signing out:', error);
            throw error;
        }
    };

    const value = {
        user,
        employee,
        tenantAccess,
        loading,
        employeeLoading,
        signInWithGoogle,
        signOut,
        refreshEmployee,
    };

    return (
        <AuthContext.Provider value={value}>
            {children}
        </AuthContext.Provider>
    );
};