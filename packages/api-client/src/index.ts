const API_BASE = process.env.NEXT_PUBLIC_API_URL || 'http://localhost:8080';
const AUTH_BASE = process.env.NEXT_PUBLIC_AUTH_URL || 'http://localhost:8081';

interface RequestOptions {
  method: string;
  path: string;
  body?: any;
  token?: string;
  baseUrl?: string;
}

class APIError extends Error {
  status: number;
  code: string;

  constructor(message: string, status: number, code: string) {
    super(message);
    this.name = 'APIError';
    this.status = status;
    this.code = code;
  }
}

async function request<T>({ method, path, body, token, baseUrl }: RequestOptions): Promise<T> {
  const url = (baseUrl || API_BASE) + path;
  const headers: Record<string, string> = {
    'Content-Type': 'application/json',
  };

  if (token) {
    headers['Authorization'] = `Bearer ${token}`;
  }

  const res = await fetch(url, {
    method,
    headers,
    body: body ? JSON.stringify(body) : undefined,
  });

  const data = await res.json();

  if (!res.ok) {
    throw new APIError(data.error || 'Request failed', res.status, data.code || 'UNKNOWN');
  }

  return data as T;
}

// ===== Auth API =====

export const authApi = {
  register(email: string, name: string, password: string) {
    return request<{ user: any; access_token: string; refresh_token: string; expires_in: number }>({
      method: 'POST', path: '/auth/register', baseUrl: AUTH_BASE,
      body: { email, name, password },
    });
  },

  login(email: string, password: string) {
    return request<{ user: any; access_token: string; refresh_token: string; expires_in: number }>({
      method: 'POST', path: '/auth/login', baseUrl: AUTH_BASE,
      body: { email, password },
    });
  },

  refresh(refreshToken: string) {
    return request<{ access_token: string; expires_in: number }>({
      method: 'POST', path: '/auth/refresh', baseUrl: AUTH_BASE,
      body: { refresh_token: refreshToken },
    });
  },

  logout(token: string) {
    return request({ method: 'POST', path: '/auth/logout', baseUrl: AUTH_BASE, token });
  },

  getMe(token: string) {
    return request<any>({ method: 'GET', path: '/auth/me', baseUrl: AUTH_BASE, token });
  },
};

// ===== Projects API =====

export const projectsApi = {
  list(token: string) {
    return request<{ projects: any[]; count: number }>({
      method: 'GET', path: '/api/v1/projects', token,
    });
  },

  get(token: string, projectId: string) {
    return request<any>({ method: 'GET', path: `/api/v1/projects/${projectId}`, token });
  },

  create(token: string, data: {
    name: string; org_id?: string; repo_url?: string; runtime?: string;
    framework?: string; build_command?: string; start_command?: string;
  }) {
    return request<any>({ method: 'POST', path: '/api/v1/projects', token, body: data });
  },

  update(token: string, projectId: string, data: any) {
    return request<any>({ method: 'PUT', path: `/api/v1/projects/${projectId}`, token, body: data });
  },

  delete(token: string, projectId: string) {
    return request<any>({ method: 'DELETE', path: `/api/v1/projects/${projectId}`, token });
  },
};

// ===== Deployments API =====

export const deploymentsApi = {
  list(token: string, projectId: string) {
    return request<{ deployments: any[]; count: number }>({
      method: 'GET', path: `/api/v1/projects/${projectId}/deployments`, token,
    });
  },

  get(token: string, projectId: string, deployId: string) {
    return request<any>({
      method: 'GET', path: `/api/v1/projects/${projectId}/deployments/${deployId}`, token,
    });
  },

  trigger(token: string, projectId: string, ref?: string) {
    return request<any>({
      method: 'POST', path: `/api/v1/projects/${projectId}/deploy`, token,
      body: { ref: ref || 'main', source: 'dashboard' },
    });
  },

  cancel(token: string, projectId: string, deployId: string) {
    return request<any>({
      method: 'POST', path: `/api/v1/projects/${projectId}/deployments/${deployId}/cancel`, token,
    });
  },

  rollback(token: string, projectId: string, deployId: string) {
    return request<any>({
      method: 'POST', path: `/api/v1/projects/${projectId}/deployments/${deployId}/rollback`, token,
    });
  },

  getLogs(token: string, projectId: string, deployId: string) {
    return request<{ logs: string[] }>({
      method: 'GET', path: `/api/v1/projects/${projectId}/deployments/${deployId}/logs`, token,
    });
  },
};

// ===== Environment Variables API =====

export const envApi = {
  list(token: string, projectId: string) {
    return request<{ env_vars: any[] }>({
      method: 'GET', path: `/api/v1/projects/${projectId}/env`, token,
    });
  },

  set(token: string, projectId: string, key: string, value: string, target?: string[]) {
    return request<any>({
      method: 'POST', path: `/api/v1/projects/${projectId}/env`, token,
      body: { key, value, target },
    });
  },

  delete(token: string, projectId: string, key: string) {
    return request<any>({
      method: 'DELETE', path: `/api/v1/projects/${projectId}/env/${key}`, token,
    });
  },
};

// ===== Organizations API =====

export const orgsApi = {
  list(token: string) {
    return request<{ organizations: any[] }>({ method: 'GET', path: '/api/v1/orgs', token });
  },

  create(token: string, name: string, slug: string) {
    return request<any>({ method: 'POST', path: '/api/v1/orgs', token, body: { name, slug } });
  },

  get(token: string, orgId: string) {
    return request<any>({ method: 'GET', path: `/api/v1/orgs/${orgId}`, token });
  },

  listMembers(token: string, orgId: string) {
    return request<{ members: any[] }>({ method: 'GET', path: `/api/v1/orgs/${orgId}/members`, token });
  },
};

// ===== Domains API =====

export const domainsApi = {
  list(token: string, projectId: string) {
    return request<{ domains: any[] }>({
      method: 'GET', path: `/api/v1/projects/${projectId}/domains`, token,
    });
  },

  add(token: string, projectId: string, domain: string) {
    return request<any>({
      method: 'POST', path: `/api/v1/projects/${projectId}/domains`, token,
      body: { domain },
    });
  },

  remove(token: string, projectId: string, domainId: string) {
    return request<any>({
      method: 'DELETE', path: `/api/v1/projects/${projectId}/domains/${domainId}`, token,
    });
  },
};

export { APIError };
