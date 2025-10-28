interface Employee {
    id: string;
    firebaseUid: string;
    email: string;
    firstName: string | null;
    lastName: string | null;
    role: string;
    isActive: boolean;
    createdAt: string;
    updatedAt: string;
}

interface TenantAccess {
    tenantId: string;
    tenantName: string;
    role: string;
    isActive: boolean;
}

interface CreateEmployeeRequest {
    firebaseUid: string;
    email: string;
    firstName?: string;
    lastName?: string;
    role?: string;
    tenantIds?: string[];
}interface CreateEmployeeResponse {
    success: boolean;
    message: string;
    employee: Employee;
}

class EmployeeService {
    private baseUrl: string;

    constructor() {
        this.baseUrl = process.env.NEXT_PUBLIC_API_URL || 'http://localhost:8081';
    }

    async createEmployee(data: CreateEmployeeRequest, idToken?: string): Promise<CreateEmployeeResponse> {
        const headers: Record<string, string> = {
            'Content-Type': 'application/json',
        };

        // Add Authorization header if idToken is provided
        if (idToken) {
            headers['Authorization'] = `Bearer ${idToken}`;
        }

        const response = await fetch(`${this.baseUrl}/api/v1/employees`, {
            method: 'POST',
            headers,
            body: JSON.stringify(data),
        });

        if (!response.ok) {
            const errorData = await response.json().catch(() => ({}));
            throw new Error(errorData.error || 'Failed to create employee');
        }

        return response.json();
    }

    async getCurrentEmployee(idToken: string): Promise<Employee> {
        const response = await fetch(`${this.baseUrl}/api/v1/employees/me`, {
            headers: {
                'Authorization': `Bearer ${idToken}`,
            },
        });

        if (!response.ok) {
            const errorData = await response.json().catch(() => ({}));
            throw new Error(errorData.error || 'Failed to get employee');
        }

        return response.json();
    }

    async updateEmployee(idToken: string, data: Partial<Pick<Employee, 'firstName' | 'lastName'>>): Promise<Employee> {
        const response = await fetch(`${this.baseUrl}/api/v1/employees/me`, {
            method: 'PUT',
            headers: {
                'Content-Type': 'application/json',
                'Authorization': `Bearer ${idToken}`,
            },
            body: JSON.stringify(data),
        });

        if (!response.ok) {
            const errorData = await response.json().catch(() => ({}));
            throw new Error(errorData.error || 'Failed to update employee');
        }

        return response.json();
    }

    async getEmployeeTenants(idToken: string): Promise<TenantAccess[]> {
        const response = await fetch(`${this.baseUrl}/api/v1/employees/me/tenants`, {
            headers: {
                'Authorization': `Bearer ${idToken}`,
            },
        });

        if (!response.ok) {
            const errorData = await response.json().catch(() => ({}));
            throw new Error(errorData.error || 'Failed to get employee tenants');
        }

        return response.json();
    }
}

export const employeeService = new EmployeeService();
export type { Employee, TenantAccess, CreateEmployeeRequest, CreateEmployeeResponse };
